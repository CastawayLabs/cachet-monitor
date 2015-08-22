Cachet Monitor plugin
=====================

This is a monitoring plugin for CachetHQ.

![screenshot](https://castawaylabs.github.io/cachet-monitor/screenshot.png)

Features
--------

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag every second
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if in Partial Outage
- [x] Can be run on multiple servers and geo regions

Docker Quickstart
-----------------

1. Create a configuration json
2. 
```
docker run -d \
  --name cachet-monitor \
  -h cachet-monitor \
  -v `pwd`/config.json:/etc/cachet-monitor.config.json \
  castawaylabs/cachet-monitor
```

Configuration
-------------

```
{
  "api_url": "https://demo.cachethq.io/api/v1",
  "api_token": "9yMHsdioQosnyVK4iCVR",
  "monitors": [
    {
      "name": "nodegear frontend",
      "url": "https://nodegear.io/ping",
      "metric_id": 0,
      "component_id": 0,
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

How to run
----------

Example:

1. Set up [Go](https://golang.org)
2. `go install github.com/castawaylabs/cachet-monitor`
3. `cachet-monitor -c https://raw.githubusercontent.com/CastawayLabs/cachet-monitor/master/example.config.json`

Production:

1. Download the example config and save to `/etc/cachet-monitor.config.json`
2. Run in background: `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

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
