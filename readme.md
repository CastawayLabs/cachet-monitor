![screenshot](https://castawaylabs.github.io/cachet-monitor/screenshot.png)

## Features

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag to cachet graphs
- [x] HTTP Checks (body/status code)
- [x] DNS Checks
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if already in Partial Outage (works with distributed monitors)
- [x] Can be run on multiple servers and geo regions

## Example Configuration

**Note:** configuration can be in json or yaml format. [`example.config.json`](https://github.com/CastawayLabs/cachet-monitor/blob/master/example.config.json), [`example.config.yaml`](https://github.com/CastawayLabs/cachet-monitor/blob/master/example.config.yml) files.

```yaml
api:
  # cachet url
  url: https://demo.cachethq.io/api/v1
  # cachet api token
  token: 9yMHsdioQosnyVK4iCVR
  insecure: false
# https://golang.org/src/time/format.go#L57
date_format: 02/01/2006 15:04:05 MST
monitors:
  # http monitor example
  - name: google
    # test url
    target: https://google.com
    # strict certificate checking for https
    strict: true
    # HTTP method
    method: POST
    
    # set to update component (either component_id or metric_id are required)
    component_id: 1
    # set to post lag to cachet metric (graph)
    metric_id: 4

    # custom templates (see readme for details)
    template:
      investigating:
        subject: "{{ .Monitor.Name }} - {{ .SystemName }}"
        message: "{{ .Monitor.Name }} check **failed** (server time: {{ .now }})\n\n{{ .FailReason }}"
      fixed:
        subject: "I HAVE BEEN FIXED"
    
    # seconds between checks
    interval: 1
    # seconds for timeout
    timeout: 1
    # If % of downtime is over this threshold, open an incident
    threshold: 80

    # custom HTTP headers
    headers:
      Authorization: Basic <hash>
    # expected status code (either status code or body must be supplied)
    expected_status_code: 200
    # regex to match body
    expected_body: "P.*NG"
  # dns monitor example
  - name: dns
    # fqdn
    target: matej.me.
    # question type (A/AAAA/CNAME/...)
    question: mx
    type: dns
    # set component_id/metric_id
    component_id: 2
    # poll every 1s
    interval: 1
    timeout: 1
    # custom DNS server (defaults to system)
    dns: 8.8.4.4:53
    answers:
      # exact/regex check
      - regex: [1-9] alt[1-9].aspmx.l.google.com.
      - exact: 10 aspmx2.googlemail.com.
      - exact: 1 aspmx.l.google.com.
      - exact: 10 aspmx3.googlemail.com.
```

## Installation

1. Download binary from [release page](https://github.com/CastawayLabs/cachet-monitor/releases)
2. Create a configuration
3. `cachet-monitor -c /etc/cachet-monitor.yaml`

pro tip: run in background using `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

```
Usage:
  cachet-monitor (-c PATH | --config PATH) [--log=LOGPATH] [--name=NAME] [--immediate]
  cachet-monitor -h | --help | --version

Arguments:
  PATH     path to config.json
  LOGPATH  path to log output (defaults to STDOUT)
  NAME     name of this logger

Examples:
  cachet-monitor -c /root/cachet-monitor.json
  cachet-monitor -c /root/cachet-monitor.json --log=/var/log/cachet-monitor.log --name="development machine"

Options:
  -c PATH.json --config PATH     Path to configuration file
  -h --help                      Show this screen.
  --version                      Show version
  --immediate                    Tick immediately (by default waits for first defined interval)
  
Environment varaibles:
  CACHET_API      override API url from configuration
  CACHET_TOKEN    override API token from configuration
  CACHET_DEV      set to enable dev logging
```

## Templates

This package makes use of [`text/template`](https://godoc.org/text/template). [Default HTTP template](https://github.com/CastawayLabs/cachet-monitor/blob/master/http.go#L14)

The following variables are available:

| Root objects  |
| ------------- | -----------------
| `.SystemName` | system name
| `.API`        | `api` object from configuration
| `.Monitor`    | `monitor` object from configuration
| `.now`        | formatted date string

| Monitor variables  |
| ------------------ |
| `.Name`            |
| `.Target`          |
| `.Type`            |
| `.Strict`          |
| `.MetricID`        |
| ...                |

All monitor variables are available from `monitor.go`

## Vision and goals

We made this tool because we felt the need to have our own monitoring software (leveraging on Cachet).
The idea is a stateless program which collects data and pushes it to a central cachet instance.

This gives us power to have an army of geographically distributed loggers and reveal issues in both latency & downtime on client websites.

## Package usage

When using `cachet-monitor` as a package in another program, you should follow what `cli/main.go` does. It is important to call `Validate` on `CachetMonitor` and all the monitors inside.

[API Documentation](https://godoc.org/github.com/CastawayLabs/cachet-monitor)

# Contributions welcome

We'll happily accept contributions for the following (non exhaustive list).

- Implement ICMP check
- Implement TCP check
- Any bug fixes / code improvements
- Test cases

## License

MIT License

Copyright (c) 2016 Castaway Labs LLC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.