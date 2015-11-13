package main

import (
	"github.com/castawaylabs/cachet-monitor/cachet"
	"time"
)

func main() {
	cachet.LoadCachetConfigClassic()
	log := cachet.Logger

	log.Printf("System: %s, API: %s\n", cachet.Config.SystemName, cachet.Config.APIUrl)
	log.Printf("Starting %d monitors:\n", len(cachet.Config.Monitors))
	for _, mon := range cachet.Config.Monitors {
		log.Printf(" %s: GET %s & Expect HTTP %d\n", mon.Name, mon.URL, mon.ExpectedStatusCode)
		if mon.MetricID > 0 {
			log.Printf(" - Logs lag to metric id: %d\n", mon.MetricID)
		}
	}

	log.Println()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		for _, mon := range cachet.Config.Monitors {
			go mon.Run()
		}
	}
}
