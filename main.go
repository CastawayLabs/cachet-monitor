package main

import (
	"fmt"
	"github.com/castawaylabs/cachet-monitor/cachet"
	"time"
)

func main() {
	config := cachet.Config

	fmt.Printf("System: %s, API: %s\n", config.SystemName, config.APIUrl)
	fmt.Printf("Starting %d monitors:\n", len(config.Monitors))
	for _, mon := range config.Monitors {
		fmt.Printf(" %s: GET %s & Expect HTTP %d\n", mon.Name, mon.URL, mon.ExpectedStatusCode)
		if mon.MetricID > 0 {
			fmt.Printf(" - Logs lag to metric id: %d\n", mon.MetricID)
		}
	}

	fmt.Println()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		for _, mon := range config.Monitors {
			go mon.Run()
		}
	}
}
