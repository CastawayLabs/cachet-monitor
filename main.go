package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/castawaylabs/cachet-monitor/cachet"
)

func main() {
	config := cachet.Config
	log := cachet.Logger

	log.Printf("System: %s, API: %s\n", config.SystemName, config.APIUrl)
	log.Printf("Starting %d monitors:\n", len(config.Monitors))
	for _, mon := range config.Monitors {
		log.Printf(" %s: GET %s & Expect HTTP %d\n", mon.Name, mon.URL, mon.ExpectedStatusCode)
		if mon.MetricID > 0 {
			log.Printf(" - Logs lag to metric id: %d\n", mon.MetricID)
		}
	}

	log.Println()

	wg := &sync.WaitGroup{}
	for _, mon := range config.Monitors {
		wg.Add(1)
		go func(mon *cachet.Monitor) {
			ticker := time.NewTicker(mon.Interval * time.Second)
			for {
				select {
				case <-ticker.C:
					mon.Run()
				case <-mon.StopC():
					wg.Done()
					return
				}
			}
		}(mon)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	log.Println("Waiting monitors to end current operation")
	for _, mon := range config.Monitors {
		mon.Stop()
	}

	wg.Wait()
}
