package cachet

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

// Static config
var Config CachetConfig

// Central logger
var Logger *log.Logger

// CachetConfig is the monitoring tool configuration
type CachetConfig struct {
	APIUrl      string     `json:"api_url"`
	APIToken    string     `json:"api_token"`
	Interval    int64      `json:"interval"`
	Monitors    []*Monitor `json:"monitors"`
	SystemName  string     `json:"system_name"`
	LogPath     string     `json:"log_path"`
	InsecureAPI bool       `json:"insecure_api"`
}

func New() error {
	var configPath string
	var systemName string
	var logPath string
	flag.StringVar(&configPath, "c", "/etc/cachet-monitor.config.json", "Config path")
	flag.StringVar(&systemName, "name", "", "System Name")
	flag.StringVar(&logPath, "log", "", "Log path")
	flag.Parse()

	var data []byte

	// test if its a url
	url, err := url.ParseRequestURI(configPath)
	if err == nil && len(url.Scheme) > 0 {
		// download config
		response, err := http.Get(configPath)
		if err != nil {
			return errors.New("Cannot download network config: " + err.Error())
		}

		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)

		fmt.Println("Downloaded network configuration.")
	} else {
		data, err = ioutil.ReadFile(configPath)
		if err != nil {
			return errors.New("Config file '" + configPath + "' missing!")
		}
	}

	if err := json.Unmarshal(data, &Config); err != nil {
		return errors.New("Cannot parse config!")
	}

	if len(systemName) > 0 {
		Config.SystemName = systemName
	}
	if len(Config.SystemName) == 0 {
		// get hostname
		Config.SystemName = getHostname()
	}
	if Config.Interval <= 0 {
		Config.Interval = 60
	}

	if len(os.Getenv("CACHET_API")) > 0 {
		Config.APIUrl = os.Getenv("CACHET_API")
	}
	if len(os.Getenv("CACHET_TOKEN")) > 0 {
		Config.APIToken = os.Getenv("CACHET_TOKEN")
	}

	if len(Config.APIToken) == 0 || len(Config.APIUrl) == 0 {
		return errors.New("API URL or API Token not set. cachet-monitor won't be able to report incidents.\n\nPlease set:\n CACHET_API and CACHET_TOKEN environment variable to override settings.\n\nGet help at https://github.com/CastawayLabs/cachet-monitor\n")
	}

	if len(Config.Monitors) == 0 {
		return errors.New("No monitors defined!\nSee sample configuration: https://github.com/CastawayLabs/cachet-monitor/blob/master/example.config.json\n")
	}

	if len(logPath) > 0 {
		Config.LogPath = logPath
	}

	var logWriter io.Writer
	logWriter = os.Stdout
	if len(Config.LogPath) > 0 {
		logWriter, err = os.Create(Config.LogPath)
		if err != nil {
			return errors.New("Unable to open file '" + Config.LogPath + "' for logging\n")
		}
	}

	flags := log.Llongfile | log.Ldate | log.Ltime
	if len(os.Getenv("DEVELOPMENT")) > 0 {
		flags = 0
	}

	Logger = log.New(logWriter, "", flags)

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
