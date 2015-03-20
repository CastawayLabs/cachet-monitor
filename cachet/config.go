package cachet

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// Static config
var Config CachetConfig

// CachetConfig is the monitoring tool configuration
type CachetConfig struct {
	APIUrl   string     `json:"api_url"`
	APIToken string     `json:"api_token"`
	Monitors []*Monitor `json:"monitors"`
}

func init() {
	var configPath string
	flag.StringVar(&configPath, "c", "/etc/cachet-monitor.config.json", "Config path")
	flag.Parse()

	var data []byte

	// test if its a url
	url, err := url.ParseRequestURI(configPath)
	if err == nil && len(url.Scheme) > 0 {
		// download config
		response, err := http.Get(configPath)
		if err != nil {
			fmt.Printf("Cannot download network config: %v\n", err)
			os.Exit(1)
		}

		defer response.Body.Close()

		data, _ = ioutil.ReadAll(response.Body)

		fmt.Println("Downloaded network configuration.")
	} else {
		data, err = ioutil.ReadFile(configPath)
		if err != nil {
			fmt.Println("Config file '" + configPath + "' missing!")
			os.Exit(1)
		}
	}

	err = json.Unmarshal(data, &Config)

	if err != nil {
		fmt.Println("Cannot parse config!")
		os.Exit(1)
	}

	if len(os.Getenv("CACHET_API")) > 0 {
		Config.APIUrl = os.Getenv("CACHET_API")
	}
	if len(os.Getenv("CACHET_TOKEN")) > 0 {
		Config.APIToken = os.Getenv("CACHET_TOKEN")
	}

	if len(Config.APIToken) == 0 || len(Config.APIUrl) == 0 {
		fmt.Printf("API URL or API Token not set. cachet-monitor won't be able to report incidents.\n\nPlease set:\n CACHET_API and CACHET_TOKEN environment variable to override settings.\n\nGet help at https://github.com/CastawayLabs/cachet-monitor\n")
		os.Exit(1)
	}

	if len(Config.Monitors) == 0 {
		fmt.Printf("No monitors defined!\nSee sample configuration: https://github.com/CastawayLabs/cachet-monitor/blob/master/example.config.json\n")
		os.Exit(1)
	}
}
