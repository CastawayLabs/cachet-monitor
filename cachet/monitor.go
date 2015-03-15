package cachet

import (
	"fmt"
	"time"
	"net/http"
)

const timeout = time.Duration(time.Second)

type Monitor struct {
	Url string `json:"url"`
	MetricId int `json:"metric_id"`
}

func (monitor *Monitor) Run() {
	reqStart := getMs()
	err := monitor.doRequest()
	lag := getMs() - reqStart

	failed := false
	if err != nil {
		failed = true
	}

	if failed == true {
		fmt.Println("Req failed")
	}

	SendMetric(1, lag)
}

func (monitor *Monitor) doRequest() error {
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(monitor.Url) // http://127.0.0.1:1337
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
