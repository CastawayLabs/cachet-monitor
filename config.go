package cachet

import (
	"fmt"
	"net"
	"os"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
)

type CachetMonitor struct {
	SystemName  string            `json:"system_name"`
	API         CachetAPI         `json:"api"`
	RawMonitors []json.RawMessage `json:"monitors"`

	Monitors []MonitorInterface `json:"-"`
}

// Validate configuration
func (cfg *CachetMonitor) Validate() bool {
	valid := true

	if len(cfg.SystemName) == 0 {
		// get hostname
		cfg.SystemName = getHostname()
	}

	fmt.Println(cfg.API)
	if len(cfg.API.Token) == 0 || len(cfg.API.URL) == 0 {
		logrus.Warnf("API URL or API Token missing.\nGet help at https://github.com/castawaylabs/cachet-monitor")
		valid = false
	}

	if len(cfg.Monitors) == 0 {
		logrus.Warnf("No monitors defined!\nSee help for example configuration")
		valid = false
	}

	for _, monitor := range cfg.Monitors {
		if errs := monitor.Validate(); len(errs) > 0 {
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

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
