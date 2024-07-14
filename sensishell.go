package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	version = "dev"
	commit  = "NONE"
	date    = "UNKNOWN"
	builtBy = "UNKNOWN"
)

const (
	Equal Comparator = iota
	NotEqual
	Less
	Greater
	LessOrEqual
	GreaterOrEqual
)

type Comparator uint8

var comparatorMap = map[string]Comparator{
	"==": Equal,
	"!=": NotEqual,
	"<":  Less,
	">":  Greater,
	"<=": LessOrEqual,
	">=": GreaterOrEqual,
}

type Sensor struct {
	Id         string
	Command    string
	Limit      float64
	Comparator Comparator
}

func main() {
	metaVersion := version
	metaBuild := fmt.Sprintf("commit %s, built at %s by %s", commit, date, builtBy)

	slog.Debug("Build",
		slog.String("app-version", metaVersion),
		slog.String("app-build", metaBuild),
	)

	cycleLimit := flag.Int("n", 1, "number of maximum consecutive active cycles")
	cycleInterval := flag.Int("s", 5, "number of seconds to sleep between each cycle")
	finalCommand := flag.String("c", "", "command to run after cycle limit is reached")
	finalRestart := flag.Bool("r", true, "restart the cycles after cycle limit is reached")
	flag.Parse()

	sensorList := []Sensor{}
	r := csv.NewReader(os.Stdin)
	r.Comma = ' '
	r.Comment = '#'
	for {
		sensorFields, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			slog.Error("Config",
				slog.Any("config-fields", sensorFields),
				slog.Any("config-error", err),
			)
			log.Fatal("config error")
		}

		sensorComparator, ok := comparatorMap[sensorFields[2]]
		if !ok {
			log.Fatal("unknown comparator")
		}

		sensorLimit, err := strconv.ParseFloat(sensorFields[3], 64)
		if err != nil {
			log.Fatal(err)
		}

		sensorList = append(sensorList, Sensor{
			Id:         sensorFields[0],
			Command:    sensorFields[1],
			Limit:      sensorLimit,
			Comparator: Comparator(sensorComparator),
		})
	}

	if len(sensorList) == 0 {
		log.Fatal("no sensors configured")
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		cycle := 0
		for {
			allActive := true
			for _, sensor := range sensorList {
				cmdHandler := exec.Command("bash", "-c", sensor.Command)
				cmdStdout, cmdErr := cmdHandler.Output()
				cmdOutput := strings.TrimSpace(string(cmdStdout))
				if cmdOutput == "" {
					slog.Warn("Not Available",
						slog.String("sensor-id", sensor.Id),
						slog.String("sensor-cmd", sensor.Command),
						slog.Any("sensor-error", cmdErr),
					)
					continue
				}

				sensorValue := 0.0
				sensorActive := false
				sensorValue, cmdErr = strconv.ParseFloat(cmdOutput, 64)
				if cmdErr != nil {
					slog.Warn("Not Parsed",
						slog.String("sensor-id", sensor.Id),
						slog.String("sensor-output", cmdOutput),
						slog.Any("sensor-error", cmdErr),
					)
					continue
				}

				switch sensor.Comparator {
				case Equal:
					sensorActive = sensorValue == sensor.Limit
				case NotEqual:
					sensorActive = sensorValue != sensor.Limit
				case Less:
					sensorActive = sensorValue < sensor.Limit
				case Greater:
					sensorActive = sensorValue > sensor.Limit
				case LessOrEqual:
					sensorActive = sensorValue <= sensor.Limit
				case GreaterOrEqual:
					sensorActive = sensorValue >= sensor.Limit
				}

				slog.Info("Reading",
					slog.String("sensor-id", sensor.Id),
					slog.Float64("sensor-value", sensorValue),
					slog.Bool("sensor-active", sensorActive),
					slog.Any("sensor-error", cmdErr),
				)

				allActive = allActive && sensorActive
			}

			if allActive {
				cycle = cycle + 1
			} else {
				cycle = 0
			}
			slog.Info("Cycle",
				slog.Bool("active-state", allActive),
				slog.Int("active-cycle", cycle),
			)
			if cycle == *cycleLimit {
				slog.Info("Limit")
				if *finalCommand != "" {
					finalOutput, err := exec.Command("bash", "-c", *finalCommand).Output()
					slog.Info("Final",
						slog.String("final-command", string(finalOutput)),
						slog.Any("final-error", err),
					)
				}
				if *finalRestart {
					cycle = 0
					slog.Info("Restarting")
				} else {
					done <- syscall.SIGTERM
					break
				}
			}

			time.Sleep(time.Second * time.Duration(*cycleInterval))
		}
	}()
	sig := <-done
	slog.Info("Closing",
		slog.Any("exit-signal", sig),
	)
}
