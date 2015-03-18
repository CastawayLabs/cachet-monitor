Cachet Monitor plugin
=====================

This is a monitoring plugin for CachetHQ.

How to run:
-----------

Example:

1. Set up [Go](https://golang.org)
2. `go install github.com/castawaylabs/cachet-monitor`
3. `cachet-monitor -c https://raw.githubusercontent.com/CastawayLabs/cachet-monitor/master/example.config.json`

Production:

1. Download the example config and save to `/etc/cachet-monitor.config.json`
2. Run in background: `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

Environment variables:
----------------------

| Name         | Example Value               | Description                 |
| ------------ | --------------------------- | --------------------------- |
| CACHET_API   | http://demo.cachethq.io/api | URL endpoint for cachet api |
| CACHET_TOKEN | randomvalue                 | API Authentication token    |