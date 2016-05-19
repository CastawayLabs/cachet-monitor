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
	SystemName  string `json:"system_name"`
	LogPath     string `json:"log_path"`
	InsecureAPI bool   `json:"insecure_api"`

	Monitors []*Monitor `json:"monitors"`
}

func (cfg *CachetMonitor) ValidateConfiguration() error {
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stdout, "", log.Llongfile|log.Ldate|log.Ltime)
	}

	if len(cfg.SystemName) == 0 {
		// get hostname
		cfg.SystemName = getHostname()
	}

	if len(cfg.APIToken) == 0 || len(cfg.APIUrl) == 0 {
		return errors.New("API URL or API Token not set. cachet-monitor won't be able to report incidents.\n\nPlease set:\n CACHET_API and CACHET_TOKEN environment variable to override settings.\n\nGet help at https://github.com/castawaylabs/cachet-monitor\n")
	}

	if len(cfg.Monitors) == 0 {
		return errors.New("No monitors defined!\nSee sample configuration: https://github.com/castawaylabs/cachet-monitor/blob/master/example.config.json\n")
	}

	for _, monitor := range cfg.Monitors {
		if err := monitor.ValidateConfiguration(); err != nil {
			return err
		}
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
