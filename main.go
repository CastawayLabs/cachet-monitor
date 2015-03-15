package main

import (
	"time"
	"./cachet"
)

func main() {
	monitors := []cachet.Monitor{
		cachet.Monitor{
			Url: "https://nodegear.io/ping",
			MetricId: 1,
		},
	}

	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		for _, monitor := range monitors {
			go monitor.Run()
		}
	}
}
