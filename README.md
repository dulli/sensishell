<!-- markdownlint-configure-file {
  "MD013": false,
  "MD033": false,
  "MD041": false
} -->

<div align="center">

:shell:

# sensishell

[![GitHub license](https://badgen.net/github/license/dulli/sensishell)](https://github.com/dulli/sensishell/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/dulli/sensishell.svg)](https://github.com/dulli/sensishell/releases/)
[![GoReportCard](https://goreportcard.com/badge/github.com/dulli/sensishell)](https://goreportcard.com/report/github.com/dulli/sensishell)

_sensishell_ is a simple way for _Linux_ systems to turn shell commands into a sensor that is compared to defined thresholds, e.g. to periodically check if the system is idle and suspend it - written in pure _Go_ and without any third-party dependencies.

---

[Getting started](#getting-started) •
[Configuration](#configuration) •
[Development](#development)

---

</div>

## Getting Started

Copy the compiled executable as to your target computer, create a config file (which is simply read from `STDIN`, see the example in the `config` folder or the section [below](#configuration)) and run it e.g. using the command line:

```bash
cat config/is-idle.conf | ./sensishell
```

The above example will output the sensor values using structured logging - if you want to easily use the values to actually do something, e.g. you can provide a command that should be executed after `n` cycles where the sensor was active (i.e. the threshold was exceeded):

```bash
cat config/is-idle.conf | ./sensishell -n 5 -c "echo system is idle"
```

Of course you can also specify the interval between the cycles or whether _sensishell_ should exit after a cycle limit was reached:

```console
$ bin/linux/amd64/sensishell -h
Usage of bin/linux/amd64/sensishell:
  -c string
        command to run after cycle limit is reached
  -n int
        number of maximum consecutive active cycles (default 3)
  -r    restart the cycles after cycle limit is reached (default true)
  -s int
        number of seconds to sleep between each cycle (default 5)
```

## Configuration

All sensor configuration is provided to _sensishell_ as input to `STDIN` in the following format where each line is an individual sensor and each field is separated by a single space character:

```
<sensor-id> <sensor-command> <sensor-comparator> <sensor-threshold>
```

Lines starting with a `#` are treated as comments. This results in the configuration format technically being a valid CSV-dialect that uses the space character as the delimiter.

An example configuration that defines different sensors to see if a system is idle would therefore look like this:

```python
# active users?
active-users "who | wc -l" == 0
# active samba sessions?
active-samba "smbstatus -j | jq '.sessions | length'" <= 0
# five minute load average?
active-load5 "cat /proc/loadavg | cut -d ' ' -f 3" <= 2.5
# active connections on port :22?
active-ssh "lsof -Pi :22 -sTCP:ESTABLISHED -t | wc -l" == 0
# active processes for snapraid or restic?
active-snapraid "pgrep -lc snapraid" == 0
active-restic "pgrep -lc restic" == 0
# minimum uptime in seconds?
minimum-uptime "cat /proc/uptime | cut -d ' ' -f 1" >= 300
```
Command outputs are parsed to floating point numbers to compute the sensor values and sensors are treated as active when the expression comparing the value to the thresholds with the defined comparator evaluates to true.

Valid comparators are:
```python
"==": Equal,
"!=": NotEqual,
"<":  Less,
">":  Greater,
"<=": LessOrEqual,
">=": GreaterOrEqual
```

<!-- if you are looking at the source of this readme, please note that while the above code blocks are syntax highlighted as Python, it does in fact need to be valid CSV as stated above; Python just fits the required highlight pretty good by coincidence  -->

## Development

The following sections will help you to get started if you want to help with the development of _sensishell_ or just want to modify or compile it for yourself. No dependencies are required, but the easiest way to get started is still to just spin up a [dev container](https://containers.dev/) from the included config.

<div align="center">

---

[App Structure](#structure) •
[Compiling](#compiling)

---

</div>

## Structure

This project is structured into multiple folders with different purposes:

### Runtime Environment

As the configuration is read from `STDIN`, _sensishell_ makes zero assumptions about its runtime environment.

### Development Environment

The complete source code is contained in `sensishell.go`.

- `.github`, `.devcontainer`, `.vscode`: Contain configurations for the development infrastructure
- `/bin`: Target folder for the compiled binaries, if the supplied `VS Code` build tasks are used, the resulting binaries will be ordered into subfolders of the format `<os>/<architecture>`

(These are only required if you want to contribute to the development of _sensishell_)

### Miscellaneous

- `/config`: Exemplary application configuration files
<!-- - `/docs`: Assets that contain or support the project's documentation -->

## Compiling

All _sensishell_ commands (executables) can be build using the default go toolchain:

```bash
go build ./...
```

### Releases

Additionally, a [`GoReleaser`](https://goreleaser.com/) configuration is provided to automate the building of production ready releases for all target platforms using:

```bash
goreleaser release --clean
```

<div align="center">
<small>

:shell: /ˈsen.sə.ʃel/ :shell:

</small>
</div>
