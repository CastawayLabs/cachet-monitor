package cachet

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/castawaylabs/cachet-monitor/backends"
	cachetbackend "github.com/castawaylabs/cachet-monitor/backends/cachet"
	"github.com/castawaylabs/cachet-monitor/monitors"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type CachetMonitor struct {
	RawMonitors []map[string]interface{} `json:"monitors" yaml:"monitors"`
	RawBackend  map[string]interface{}   `json:"backend" yaml:"backend"`

	SystemName string                      `json:"system_name" yaml:"system_name"`
	Backend    backends.BackendInterface   `json:"-" yaml:"-"`
	Monitors   []monitors.MonitorInterface `json:"-" yaml:"-"`
	Immediate  bool                        `json:"-" yaml:"-"`
}

func New(path string) (*CachetMonitor, error) {
	var cfg *CachetMonitor
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
		return nil, err
	}

	// get default type
	if backend, ok := cfg.RawBackend["type"].(string); !ok {
		err = errors.New("Cannot determine backend type")
	} else {
		switch backend {
		case "cachet":
			var backend cachetbackend.CachetBackend
			err = mapstructure.Decode(cfg.RawBackend, &backend)
			cfg.Backend = &backend
			// backend.config = cfg
		default:
			err = errors.New("Invalid backend type: %v" + backend)
		}
	}

	if errs := cfg.Backend.Validate(); len(errs) > 0 {
		logrus.Errorf("Backend validation errors: %v", errs)
		os.Exit(1)
	}

	if err != nil {
		logrus.Errorf("Unable to unmarshal backend: %v", err)
		return nil, err
	}

	cfg.Monitors = make([]monitors.MonitorInterface, len(cfg.RawMonitors))
	for index, rawMonitor := range cfg.RawMonitors {
		var t monitors.MonitorInterface

		// get default type
		monType := GetMonitorType("")
		if t, ok := rawMonitor["type"].(string); ok {
			monType = GetMonitorType(t)
		}

		switch monType {
		case "http":
			var mon monitors.HTTPMonitor
			err = mapstructure.Decode(rawMonitor, &mon)
			t = &mon
		case "dns":
			var mon monitors.DNSMonitor
			err = mapstructure.Decode(rawMonitor, &mon)
			t = &mon
		default:
			logrus.Errorf("Invalid monitor type (index: %d) %v", index, monType)
			continue
		}

		if err != nil {
			logrus.Errorf("Unable to unmarshal monitor to type (index: %d): %v", index, err)
			continue
		}

		mon := t.GetMonitor()
		mon.Type = monType
		cfg.Monitors[index] = t
	}

	return cfg, err
}

// Validate configuration
func (cfg *CachetMonitor) Validate() bool {
	valid := true

	if len(cfg.SystemName) == 0 {
		// get hostname
		cfg.SystemName = getHostname()
	}

	if len(cfg.Monitors) == 0 {
		logrus.Warnf("No monitors defined!\nSee help for example configuration")
		valid = false
	}

	for index, monitor := range cfg.Monitors {
		if errs := monitor.Validate(cfg.Backend.ValidateMonitor); len(errs) > 0 {
			logrus.Warnf("Monitor validation errors (index %d): %v", index, "\n - "+strings.Join(errs, "\n - "))
			valid = false
		}
	}

	return valid
}

// getHostname returns id of the current system
func getHostname() string {
	hostname, err := os.Hostname()
	if err == nil && len(hostname) > 0 {
		return hostname
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "unknown"
	}

	return addrs[0].String()
}

func GetMonitorType(t string) string {
	if len(t) == 0 {
		return "http"
	}

	return strings.ToLower(t)
}
