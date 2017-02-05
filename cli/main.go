package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"strings"

	"github.com/Sirupsen/logrus"
	cachet "github.com/castawaylabs/cachet-monitor"
	docopt "github.com/docopt/docopt-go"
	yaml "gopkg.in/yaml.v2"
)

const usage = `cachet-monitor

Usage:
  cachet-monitor (-c PATH | --config PATH) [--log=LOGPATH] [--name=NAME]
  cachet-monitor -h | --help | --version
  cachet-monitor print-config

Arguments:
  PATH     path to config.yml
  LOGPATH  path to log output (defaults to STDOUT)
  NAME     name of this logger

Examples:
  cachet-monitor -c /root/cachet-monitor.yml
  cachet-monitor -c /root/cachet-monitor.yml --log=/var/log/cachet-monitor.log --name="development machine"

Options:
  -c PATH.yml --config PATH     Path to configuration file
  -h --help                     Show this screen.
  --version                     Show version
  print-config                  Print example configuration
  
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
		cfg.APIUrl = os.Getenv("CACHET_API")
	}
	if len(os.Getenv("CACHET_TOKEN")) > 0 {
		cfg.APIToken = os.Getenv("CACHET_TOKEN")
	}
	if len(os.Getenv("CACHET_DEV")) > 0 {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := cfg.ValidateConfiguration(); err != nil {
		panic(err)
	}

	logrus.Infof("System: %s\nAPI: %s\nMonitors: %d\n\n", cfg.SystemName, cfg.APIUrl, len(cfg.Monitors))

	wg := &sync.WaitGroup{}
	for _, mon := range cfg.Monitors {
		l := logrus.WithFields(logrus.Fields{
			"name":     mon.Name,
			"interval": mon.CheckInterval,
			"method":   mon.Method,
			"url":      mon.URL,
			"timeout":  mon.HttpTimeout,
		})
		l.Info(" Starting monitor")

		// print features
		if mon.ExpectedStatusCode > 0 {
			l.Infof(" - Expect HTTP %d", mon.ExpectedStatusCode)
		}
		if len(mon.ExpectedBody) > 0 {
			l.Infof(" - Expect Body to match \"%v\"", mon.ExpectedBody)
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
		mon.Stop()
	}

	wg.Wait()
}

func getLogger(logPath *string) *os.File {
	if logPath == nil || len(*logPath) == 0 {
		return os.Stdout
	}

	if file, err := os.Create(logPath); err != nil {
		logrus.Errorf("Unable to open file '%v' for logging: \n%v", logPath, err)
		os.Exit(1)
	} else {
		return file
	}
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

	// test file path for yml
	if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		err = yaml.Unmarshal(data, &cfg)
	} else {
		err = json.Unmarshal(data, &cfg)
	}

	if err != nil {
		logrus.Warnf("Unable to parse configuration file")
	}

	return &cfg, err
}
