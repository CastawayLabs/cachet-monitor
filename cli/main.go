package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	cachet "github.com/castawaylabs/cachet-monitor"
	docopt "github.com/docopt/docopt-go"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

const usage = `cachet-monitor

Usage:
  cachet-monitor (-c PATH | --config PATH) [--log=LOGPATH] [--name=NAME]
  cachet-monitor -h | --help | --version
  cachet-monitor print-config

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
  print-config                   Print example configuration
  
Environment varaibles:
  CACHET_API      override API url from configuration
  CACHET_TOKEN    override API token from configuration
  CACHET_DEV      set to enable dev logging`

func main() {
	arguments, _ := docopt.Parse(usage, nil, true, "cachet-monitor", false)

	cfg, err := getConfiguration(arguments["--config"].(string))
	if err != nil {
		logrus.Panicf("Unable to start (reading config): %v", err)
	}

	if name := arguments["--name"]; name != nil {
		cfg.SystemName = name.(string)
	}
	logrus.SetOutput(getLogger(arguments["--log"]))

	if len(os.Getenv("CACHET_API")) > 0 {
		cfg.API.URL = os.Getenv("CACHET_API")
	}
	if len(os.Getenv("CACHET_TOKEN")) > 0 {
		cfg.API.Token = os.Getenv("CACHET_TOKEN")
	}
	if len(os.Getenv("CACHET_DEV")) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if valid := cfg.Validate(); !valid {
		logrus.Errorf("Invalid configuration")
		os.Exit(1)
	}

	logrus.Infof("System: %s\nAPI: %s\nMonitors: %d\n\n", cfg.SystemName, cfg.API.URL, len(cfg.Monitors))
	logrus.Infof("Pinging cachet")
	if err := cfg.API.Ping(); err != nil {
		logrus.Warnf("Cannot ping cachet!\n%v", err)
	}

	wg := &sync.WaitGroup{}
	for _, monitorInterface := range cfg.Monitors {
		mon := monitorInterface.GetMonitor()
		l := logrus.WithFields(logrus.Fields{
			"name":     mon.Name,
			"interval": mon.Interval,
			"target":   mon.Target,
			"timeout":  mon.Timeout,
		})
		l.Info(" Starting monitor")

		// print features
		if mon.Type == "http" {
			httpMonitor := monitorInterface.(*cachet.HTTPMonitor)

			for k, v := range httpMonitor.Headers {
				logrus.Infof(" - HTTP-Header '%s' '%s'", k, v)
			}
			if httpMonitor.ExpectedStatusCode > 0 {
				l.Infof(" - Expect HTTP %d", httpMonitor.ExpectedStatusCode)
			}
			if len(httpMonitor.ExpectedBody) > 0 {
				l.Infof(" - Expect Body to match \"%v\"", httpMonitor.ExpectedBody)
			}
		}
		if mon.MetricID > 0 {
			l.Infof(" - Log lag to metric id %d\n", mon.MetricID)
		}
		if mon.ComponentID > 0 {
			l.Infof(" - Update component id %d\n\n", mon.ComponentID)
		}

		go mon.Start(cfg, wg)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	logrus.Warnf("Abort: Waiting monitors to finish")
	for _, mon := range cfg.Monitors {
		mon.(*cachet.AbstractMonitor).Stop()
	}

	wg.Wait()
}

func getLogger(logPath interface{}) *os.File {
	if logPath == nil || len(logPath.(string)) == 0 {
		return os.Stdout
	}

	file, err := os.Create(logPath.(string))
	if err != nil {
		logrus.Errorf("Unable to open file '%v' for logging: \n%v", logPath, err)
		os.Exit(1)
	}

	return file
}

func getConfiguration(path string) (*cachet.CachetMonitor, error) {
	var cfg cachet.CachetMonitor
	var data []byte

	// test if its a url
	url, err := url.ParseRequestURI(path)
	if err == nil && len(url.Scheme) > 0 {
		// download config
		response, err := http.Get(path)
		if err != nil {
			logrus.Warn("Unable to download network configuration")
			return nil, err
		}

		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)

		logrus.Info("Downloaded network configuration.")
	} else {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.New("Unable to open file: '" + path + "'")
		}
	}

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		err = yaml.Unmarshal(data, &cfg)
	} else {
		err = json.Unmarshal(data, &cfg)
	}

	if err != nil {
		logrus.Warnf("Unable to parse configuration file")
	}

	cfg.Monitors = make([]cachet.MonitorInterface, len(cfg.RawMonitors))
	for index, rawMonitor := range cfg.RawMonitors {
		var abstract cachet.AbstractMonitor
		var t cachet.MonitorInterface
		var err error

		monType := ""
		if t, ok := rawMonitor["type"].(string); ok {
			monType = t
		}

		switch monType {
		case "http", "":
			var s cachet.HTTPMonitor
			err = mapstructure.Decode(rawMonitor, &s)
			t = &s
		case "dns":
			// t = cachet.DNSMonitor
		case "icmp":
			// t = cachet.ICMPMonitor
		case "tcp":
			// t = cachet.TCPMonitor
		default:
			logrus.Errorf("Invalid monitor type (index: %d) %v", index, abstract.Type)
			continue
		}

		if err != nil {
			logrus.Errorf("Unable to unmarshal monitor to type (index: %d): %v", index, err)
			continue
		}

		cfg.Monitors[index] = t
	}

	return &cfg, err
}
