package main

import (
	"time"
	"./cachet"
)

func main() {
	monitors := []*cachet.Monitor{
		/*&cachet.Monitor{
			Url: "https://nodegear.io/ping",
			MetricId: 1,
			Threshold: 80.0,
		},*/
		&cachet.Monitor{
			Url: "http://localhost:1337",
			MetricId: 1,
			Threshold: 80.0,
			ExpectedStatusCode: 200,
		},
	}

	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		for _, monitor := range monitors {
			go monitor.Run()
		}
	}
}
