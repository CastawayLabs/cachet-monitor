package cachet

import (
	"net"
	"os"

	"github.com/Sirupsen/logrus"
)

type CachetMonitor struct {
	Name     string     `json:"system_name"`
	API      CachetAPI  `json:"api"`
	Monitors []*Monitor `json:"monitors"`
}

func (cfg *CachetMonitor) Validate() bool {
	valid := true

	if len(cfg.Name) == 0 {
		// get hostname
		cfg.Name = getHostname()
	}

	if len(cfg.API.Token) == 0 || len(cfg.API.Url) == 0 {
		logrus.Warnf("API URL or API Token missing.\nGet help at https://github.com/castawaylabs/cachet-monitor")
		valid = false
	}

	if len(cfg.Monitors) == 0 {
		logrus.Warnf("No monitors defined!\nSee help for example configuration")
		valid = false
	}

	for _, monitor := range cfg.Monitors {
		if err := monitor.Validate(); !valid {
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
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		return addr.String()
	}
}
