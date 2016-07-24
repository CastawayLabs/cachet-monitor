package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"

	cachet "github.com/castawaylabs/cachet-monitor"
)

var configPath string
var systemName string
var logPath string

func main() {
	flag.StringVar(&configPath, "c", "/etc/cachet-monitor.config.json", "Config path")
	flag.StringVar(&systemName, "name", "", "System Name")
	flag.StringVar(&logPath, "log", "", "Log path")
	flag.Parse()

	cfg, err := getConfiguration(configPath)
	if err != nil {
		panic(err)
	}

	if len(systemName) > 0 {
		cfg.SystemName = systemName
	}
	if len(logPath) > 0 {
		cfg.LogPath = logPath
	}

	if len(os.Getenv("CACHET_API")) > 0 {
		cfg.APIUrl = os.Getenv("CACHET_API")
	}
	if len(os.Getenv("CACHET_TOKEN")) > 0 {
		cfg.APIToken = os.Getenv("CACHET_TOKEN")
	}

	if err := cfg.ValidateConfiguration(); err != nil {
		panic(err)
	}

	cfg.Logger.Printf("System: %s\nAPI: %s\nMonitors: %d\n\n", cfg.SystemName, cfg.APIUrl, len(cfg.Monitors))

	wg := &sync.WaitGroup{}
	for _, mon := range cfg.Monitors {
		cfg.Logger.Printf(" Starting %s: %d seconds check interval - %v %s (%d second/s timeout)", mon.Name, mon.CheckInterval, mon.Method, mon.URL, mon.HttpTimeout)

		// print features
		if len(mon.HttpHeaders) > 0 {
            for _, h := range mon.HttpHeaders {
    			cfg.Logger.Printf(" - Add HTTP-Header '%s' '%s'", h.Name, h.Value)
            }
		}
		if mon.ExpectedStatusCode > 0 {
			cfg.Logger.Printf(" - Expect HTTP %d", mon.ExpectedStatusCode)
		}
		if len(mon.ExpectedBody) > 0 {
			cfg.Logger.Printf(" - Expect Body to match \"%v\"", mon.ExpectedBody)
		}
		if mon.MetricID > 0 {
			cfg.Logger.Printf(" - Log lag to metric id %d\n", mon.MetricID)
		}
		if mon.ComponentID > 0 {
			cfg.Logger.Printf(" - Update component id %d\n\n", mon.ComponentID)
		}

		go mon.Start(cfg, wg)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	cfg.Logger.Println("Abort: Waiting monitors to finish")
	for _, mon := range cfg.Monitors {
		mon.Stop()
	}

	wg.Wait()
}

func getLogger(logPath string) *log.Logger {
	var logWriter = os.Stdout
	var err error

	if len(logPath) > 0 {
		logWriter, err = os.Create(logPath)
		if err != nil {
			fmt.Printf("Unable to open file '%v' for logging\n", logPath)
			os.Exit(1)
		}
	}

	flags := log.Llongfile | log.Ldate | log.Ltime
	if len(os.Getenv("CACHET_DEV")) > 0 {
		flags = 0
	}

	return log.New(logWriter, "", flags)
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
			return nil, errors.New("Cannot download network config: " + err.Error())
		}

		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)

		fmt.Println("Downloaded network configuration.")
	} else {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.New("Config file '" + path + "' missing!")
		}
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println(err)
		return nil, errors.New("Cannot parse config!")
	}

	cfg.Logger = getLogger(cfg.LogPath)

	return &cfg, nil
}
