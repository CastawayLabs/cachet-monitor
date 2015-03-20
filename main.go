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
		fmt.Printf(" %s: GET %s & Expect HTTP %d\n", monitor.Name, monitor.Url, monitor.ExpectedStatusCode)
		if monitor.MetricId > 0 {
			fmt.Printf(" - Logs lag to metric id: %d\n", monitor.MetricId)
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