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
	Threshold float32 `json:"threshold"`
	ComponentId *int `json:"component_id"`
	ExpectedStatusCode int `json:"expected_status_code"`

	History []bool `json:"-"`
	Incident *Incident `json:"-"`
}

func (monitor *Monitor) Run() {
	reqStart := getMs()
	isUp := monitor.doRequest()
	lag := getMs() - reqStart

	if len(monitor.History) >= 10 {
		monitor.History = monitor.History[len(monitor.History)-9:]
	}
	monitor.History = append(monitor.History, isUp)
	monitor.AnalyseData()

	if isUp == true {
		SendMetric(monitor.MetricId, lag)
		return
	}
}

func (monitor *Monitor) doRequest() bool {
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(monitor.Url)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == monitor.ExpectedStatusCode
}

func (monitor *Monitor) AnalyseData() {
	// look at the past few incidents
	if len(monitor.History) != 10 {
		// not enough data
		return
	}

	numDown := 0
	for _, wasUp := range monitor.History {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(monitor.History))) * 100
	fmt.Printf("%s %.2f%% Down. Threshold: %.2f%%\n", monitor.Url, t, monitor.Threshold)
	if t > monitor.Threshold && monitor.Incident == nil {
		// is down, create an incident
		fmt.Println("Creating incident...")
		monitor.Incident = &Incident{}
	} else if t < monitor.Threshold && monitor.Incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		fmt.Println("Updating incident to resolved...")
		monitor.Incident = nil
	}
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
