package cachet

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const timeout = time.Duration(time.Second)

// Monitor data model
type Monitor struct {
	Name               string  `json:"name"`
	URL                string  `json:"url"`
	MetricID           int     `json:"metric_id"`
	Threshold          float32 `json:"threshold"`
	ComponentID        *int    `json:"component_id"`
	ExpectedStatusCode int     `json:"expected_status_code"`
	StrictTLS          *bool   `json:"strict_tls"`

	History        []bool    `json:"-"`
	LastFailReason *string   `json:"-"`
	Incident       *Incident `json:"-"`
}

// Run loop
func (monitor *Monitor) Run() {
	reqStart := getMs()
	isUp := monitor.doRequest()
	lag := getMs() - reqStart

	if len(monitor.History) >= 10 {
		monitor.History = monitor.History[len(monitor.History)-9:]
	}
	monitor.History = append(monitor.History, isUp)
	monitor.AnalyseData()

	if isUp == true && monitor.MetricID > 0 {
		SendMetric(monitor.MetricID, lag)
	}
}

func (monitor *Monitor) doRequest() bool {
	client := &http.Client{
		Timeout: timeout,
	}
	if monitor.StrictTLS != nil && *monitor.StrictTLS == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(monitor.URL)
	if err != nil {
		errString := err.Error()
		monitor.LastFailReason = &errString
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != monitor.ExpectedStatusCode {
		failReason := "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(monitor.ExpectedStatusCode)
		monitor.LastFailReason = &failReason
		return false
	}

	return true
}

// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (monitor *Monitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range monitor.History {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(monitor.History))) * 100
	Logger.Printf("%s %.2f%% Down at %v. Threshold: %.2f%%\n", monitor.URL, t, time.Now().UnixNano()/int64(time.Second), monitor.Threshold)

	if len(monitor.History) != 10 {
		// not enough data
		return
	}

	if t > monitor.Threshold && monitor.Incident == nil {
		// is down, create an incident
		Logger.Println("Creating incident...")

		component_id := json.Number(strconv.Itoa(*monitor.ComponentID))
		monitor.Incident = &Incident{
			Name:        monitor.Name + " - " + Config.SystemName,
			Message:     monitor.Name + " check failed",
			ComponentID: &component_id,
		}

		if monitor.LastFailReason != nil {
			monitor.Incident.Message += "\n\n - " + *monitor.LastFailReason
		}

		// set investigating status
		monitor.Incident.SetInvestigating()

		// create/update incident
		monitor.Incident.Send()
		monitor.Incident.UpdateComponent()
	} else if t < monitor.Threshold && monitor.Incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		Logger.Println("Updating incident to resolved...")

		component_id := json.Number(strconv.Itoa(*monitor.ComponentID))
		monitor.Incident = &Incident{
			Name: monitor.Incident.Name,
			Message: monitor.Name + " check succeeded",
			ComponentID: &component_id,
		}

		monitor.Incident.SetFixed()
		monitor.Incident.Send()
		monitor.Incident.UpdateComponent()

		monitor.Incident = nil
	}
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
