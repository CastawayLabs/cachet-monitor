package main

import (
	"fmt"
	"github.com/castawaylabs/cachet-monitor/cachet"
	"time"
)

func main() {
	fmt.Printf("API: %s\n", cachet.Config.APIUrl)
	fmt.Printf("Starting %d monitors:\n", len(cachet.Config.Monitors))
	for _, monitor := range cachet.Config.Monitors {
		fmt.Printf(" %s: GET %s & Expect HTTP %d\n", monitor.Name, monitor.URL, monitor.ExpectedStatusCode)
		if monitor.MetricID > 0 {
			fmt.Printf(" - Logs lag to metric id: %d\n", monitor.MetricID)
		}
	}

	fmt.Println()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		for _, monitor := range cachet.Config.Monitors {
			go monitor.Run()
		}
	}
}
