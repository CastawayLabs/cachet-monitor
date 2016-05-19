Cachet Monitor plugin
=====================

This is a monitoring plugin for CachetHQ.

![screenshot](https://castawaylabs.github.io/cachet-monitor/screenshot.png)

Features
--------

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag every second * config.Interval
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if in Partial Outage
- [x] Can be run on multiple servers and geo regions

Configuration
-------------

```
{
  "api_url": "https://demo.cachethq.io/api/v1",
  "api_token": "<API TOKEN>",
  "interval": 60,
  "monitors": [
    {
      "name": "Name of your monitor",
      "url": "Ping URL",
      "metric_id": <metric id from cachet>,
      "component_id": <component id from cachet>,
      "threshold": 80,
      "expected_status_code": 200,
      "strict_tls": true
    }
  ],
  "insecure_api": false
}
```

*Notes:*

- `metric_id` is optional
- `insecure_api` if true it will ignore HTTPS certificate errors (eg if self-signed)
- `strict_tls` if false (true is default) it will ignore HTTPS certificate errors (eg if monitor uses self-signed certificate)
- `component_id` is optional
- `threshold` is a percentage
- `expected_status_code` is a http response code
- GET request will be performed on the `url`

Installation
------------

1. Download binary from release page
2. Create your configuration ([example](https://raw.githubusercontent.com/CastawayLabs/cachet-monitor/master/example.config.json))
3. `cachet-monitor -c /etc/cachet-monitor.config.json`

tip: run in background using `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

```
Usage of cachet-monitor:
  -c="/etc/cachet-monitor.config.json": Config path
  -log="": Log path
  -name="": System Name
```

Environment variables
---------------------

| Name         | Example Value               | Description                 |
| ------------ | --------------------------- | --------------------------- |
| CACHET_API   | http://demo.cachethq.io/api | URL endpoint for cachet api |
| CACHET_TOKEN | randomvalue                 | API Authentication token    |
| CACHET_DEV   | 1                           | Strips logging              |

Vision and goals
----------------

We made this tool because we felt the need to have our own monitoring software (leveraging on Cachet).
The idea is a stateless program which collects data and pushes it to a central cachet instance.

This gives us power to have an army of geographically distributed loggers and reveal issues in both latency & downtime on client websites.
