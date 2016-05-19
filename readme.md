![screenshot](https://castawaylabs.github.io/cachet-monitor/screenshot.png)

Features
--------

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag to cachet graphs
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if already in Partial Outage (works well with distributed monitoring)
- [x] Can be run on multiple servers and geo regions

Configuration
-------------

```
{
  // URL for the API. Note: Must end with /api/v1
  "api_url": "https://<cachet domain>/api/v1",
  // Your API token for Cachet
  "api_token": "<cachet api token>",
  // optional, false default, set if your certificate is self-signed/untrusted
  "insecure_api": false,
  "monitors": [{
    // required, friendly name for your monitor
    "name": "Name of your monitor",
    // required, url to probe
    "url": "Ping URL",
    // optional, http method (defaults GET)
    "method": "get",
    // self-signed ssl certificate
    "strict_tls": true,
    // seconds between checks
    "interval": 10,
    // post lag to cachet metric (graph)
    // note either metric ID or component ID are required
    "metric_id": <metric id>,
    // post incidents to this component
    "component_id": <component id>,
    // If % of downtime is over this threshold, open an incident
    "threshold": 80,
    // optional, expected status code (either status code or body must be supplied)
    "expected_status_code": 200,
    // optional, regular expression to match body content
    "expected_body": "P.*NG"
  }],
  // optional, system name to identify bot (uses hostname by default)
  "system_name": "",
  // optional, defaults to stdout
  "log_path": ""
}
```

Installation
------------

1. Download binary from [release page](https://github.com/CastawayLabs/cachet-monitor/releases)
2. Create your configuration ([example](https://raw.githubusercontent.com/CastawayLabs/cachet-monitor/master/example.config.json))
3. `cachet-monitor -c /etc/cachet-monitor.config.json`

pro tip: run in background using `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

```
Usage of cachet-monitor:
  -c="/etc/cachet-monitor.config.json": Config path
  -log="": Log path
  -name="": System Name
```

Environment variables
---------------------

| Name         | Example Value                  | Description                 |
| ------------ | ------------------------------ | --------------------------- |
| CACHET_API   | http://demo.cachethq.io/api/v1 | URL endpoint for cachet api |
| CACHET_TOKEN | APIToken123                    | API Authentication token    |
| CACHET_DEV   | 1                              | Strips logging              |

Vision and goals
----------------

We made this tool because we felt the need to have our own monitoring software (leveraging on Cachet).
The idea is a stateless program which collects data and pushes it to a central cachet instance.

This gives us power to have an army of geographically distributed loggers and reveal issues in both latency & downtime on client websites.
