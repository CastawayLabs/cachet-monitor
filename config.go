package cachet

import (
	"errors"
	"log"
	"net"
	"os"
)

type CachetMonitor struct {
	Logger *log.Logger `json:"-"`

	APIUrl      string `json:"api_url"`
	APIToken    string `json:"api_token"`
	Interval    int64  `json:"interval"`
	SystemName  string `json:"system_name"`
	LogPath     string `json:"log_path"`
	InsecureAPI bool   `json:"insecure_api"`

	Monitors []*Monitor `json:"monitors"`
}

func (mon *CachetMonitor) ValidateConfiguration() error {
	if mon.Logger == nil {
		mon.Logger = log.New(os.Stdout, "", log.Llongfile|log.Ldate|log.Ltime)
	}

	if len(mon.SystemName) == 0 {
		// get hostname
		mon.SystemName = getHostname()
	}

	if mon.Interval <= 0 {
		mon.Interval = 60
	}

	if len(mon.APIToken) == 0 || len(mon.APIUrl) == 0 {
		return errors.New("API URL or API Token not set. cachet-monitor won't be able to report incidents.\n\nPlease set:\n CACHET_API and CACHET_TOKEN environment variable to override settings.\n\nGet help at https://github.com/castawaylabs/cachet-monitor\n")
	}

	if len(mon.Monitors) == 0 {
		return errors.New("No monitors defined!\nSee sample configuration: https://github.com/castawaylabs/cachet-monitor/blob/master/example.config.json\n")
	}

	return nil
}

// getHostname returns id of the current system
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		addrs, err := net.InterfaceAddrs()

		if err != nil {
			return "unknown"
		}

		for _, addr := range addrs {
			return addr.String()
		}
	}

	return hostname
}
