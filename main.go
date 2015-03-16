package main

import (
	"fmt"
	"time"
	"github.com/castawaylabs/cachet-monitor/cachet"
)

func main() {
	monitors := []*cachet.Monitor{
		/*&cachet.Monitor{
			Name: "nodegear frontend",
			Url: "https://nodegear.io/ping",
			MetricId: 1,
			Threshold: 80.0,
			ExpectedStatusCode: 200,
		},*/
		&cachet.Monitor{
			Name: "local test server",
			Url: "http://localhost:1337",
			Threshold: 80.0,
			ExpectedStatusCode: 200,
		},
	}

	fmt.Printf("Starting %d monitors:\n", len(monitors))
	for _, monitor := range monitors {
		fmt.Printf(" %s: GET %s & Expect HTTP %d\n", monitor.Name, monitor.Url, monitor.ExpectedStatusCode)
		if monitor.MetricId > 0 {
			fmt.Printf(" - Logs lag to metric id: %d\n", monitor.MetricId)
		}
	}

	fmt.Println()

	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		for _, monitor := range monitors {
			go monitor.Run()
		}
	}
}
