# Threaded TTOP Reporter

A simple CLI application that outputs the result of `LINES=100 top -H -n 120 -p 1 -d 2 -bw` as an HTML report.


## How to use

Install is just provided by Go for the time being.

```bash
go install github.com/rsvihladremio/threaded-top-reporter@latest

# Or build from source using the Makefile:
```bash
make build     # compile the binary to bin/ttoprep
make lint      # run linters
make security  # run security checks
make fmt       # format code
make test      # run tests
make all       # run full build, lint, security, fmt, and test
```
```

Minimal usage is simple: provide an input file to generate a report with the default title `Threaded Top Report` and output to `ttop.html`.

```bash
ttoprep ttop.txt
report 'Threaded Top Report' written to ttop.html
```

Custom Report Output

```bash
ttoprep ttop.txt -o out.html
report 'Threaded Top Report' written to out.html
```

Custom Report Title

```bash
ttoprep ttop.txt -n 'My Report'
report 'My Report' written to ttop.html
```

Extra report metadata

```bash
ttoprep ttop.txt -m '{"id":"40de949f-3741-476a-abcb-3214a14ac15e"}'
report 'Threaded Top Report' written to ttop.html
```

## How it works

1. The arguments from the CLI are read such as the name (-n) of the report and the output location (optional but is -o)
2. The input file is read and parsed fully by the CLI
3. HTML is produced and written to disk at the output location (default is ttop.html).


### Metadata

Metadata is provided under the title as "details" atm there are no plans to allow more complex structures than key=value pairs.

### Charts

Charts are provided by [echarts](https://echarts.apache.org/), with the initial goal is to have the following:

* see the per thread CPU performance in a graph (wip)
* memory usage (swap, total, free, cache, avail) over time in a graph (wip) 
* total CPU usage over time in a graph (iowait, sys, steal, nice, user) (wip)
* total threads and their state over time in a graph (total, sleeping, running, stopped, zombie)
* load avg graphed over time.
