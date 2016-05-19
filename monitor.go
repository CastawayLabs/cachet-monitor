package cachet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HttpTimeout = time.Duration(time.Second)
const DefaultInterval = 60
const DefaultTimeFormat = "15:04:05 Jan 2 MST"

// Monitor data model
type Monitor struct {
	Name          string        `json:"name"`
	URL           string        `json:"url"`
	Method        string        `json:"method"`
	StrictTLS     bool          `json:"strict_tls"`
	CheckInterval time.Duration `json:"interval"`

	MetricID    int `json:"metric_id"`
	ComponentID int `json:"component_id"`

	// Threshold = percentage
	Threshold          float32 `json:"threshold"`
	ExpectedStatusCode int     `json:"expected_status_code"`
	// compiled to Regexp
	ExpectedBody string `json:"expected_body"`
	bodyRegexp   *regexp.Regexp

	history        []bool
	lastFailReason string
	incident       *Incident
	config         *CachetMonitor

	// Closed when mon.Stop() is called
	stopC chan bool
}

func (mon *Monitor) Start(cfg *CachetMonitor, wg *sync.WaitGroup) {
	wg.Add(1)
	mon.config = cfg
	mon.stopC = make(chan bool)

	mon.config.Logger.Printf(" Starting %s: %d seconds check interval\n - %v %s", mon.Name, mon.CheckInterval, mon.Method, mon.URL)

	// print features
	if mon.ExpectedStatusCode > 0 {
		mon.config.Logger.Printf(" - Expect HTTP %d", mon.ExpectedStatusCode)
	}
	if len(mon.ExpectedBody) > 0 {
		mon.config.Logger.Printf(" - Expect Body to match \"%v\"", mon.ExpectedBody)
	}
	if mon.MetricID > 0 {
		mon.config.Logger.Printf(" - Log lag to metric id %d\n", mon.MetricID)
	}
	if mon.ComponentID > 0 {
		mon.config.Logger.Printf(" - Update component id %d\n\n", mon.ComponentID)
	}

	mon.Tick()

	ticker := time.NewTicker(mon.CheckInterval * time.Second)
	for {
		select {
		case <-ticker.C:
			mon.Tick()
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (monitor *Monitor) Stop() {
	if monitor.Stopped() {
		return
	}

	close(monitor.stopC)
}

func (monitor *Monitor) Stopped() bool {
	select {
	case <-monitor.stopC:
		return true
	default:
		return false
	}
}

func (monitor *Monitor) Tick() {
	reqStart := getMs()
	isUp := monitor.doRequest()
	lag := getMs() - reqStart

	if len(monitor.history) == 9 {
		monitor.config.Logger.Printf("%v is now saturated\n", monitor.Name)
	}
	if len(monitor.history) >= 10 {
		monitor.history = monitor.history[len(monitor.history)-9:]
	}
	monitor.history = append(monitor.history, isUp)
	monitor.AnalyseData()

	if isUp == true && monitor.MetricID > 0 {
		monitor.SendMetric(lag)
	}
}

func (monitor *Monitor) doRequest() bool {
	client := &http.Client{
		Timeout: HttpTimeout,
	}
	if monitor.StrictTLS == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(monitor.URL)
	if err != nil {
		monitor.lastFailReason = err.Error()

		return false
	}

	defer resp.Body.Close()

	if monitor.ExpectedStatusCode > 0 && resp.StatusCode != monitor.ExpectedStatusCode {
		monitor.lastFailReason = "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(monitor.ExpectedStatusCode)

		return false
	}

	if monitor.bodyRegexp != nil {
		// check body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			monitor.lastFailReason = err.Error()

			return false
		}

		match := monitor.bodyRegexp.Match(responseBody)
		if !match {
			monitor.lastFailReason = "Unexpected body: " + string(responseBody) + ". Expected to match " + monitor.ExpectedBody
		}

		return match
	}

	return true
}

// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (monitor *Monitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range monitor.history {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(monitor.history))) * 100
	monitor.config.Logger.Printf("%s %.2f%%/%.2f%% down at %v\n", monitor.Name, t, monitor.Threshold, time.Now().UnixNano()/int64(time.Second))

	if len(monitor.history) != 10 {
		// not saturated
		return
	}

	if t > monitor.Threshold && monitor.incident == nil {
		monitor.incident = &Incident{
			Name:        monitor.Name + " - " + monitor.config.SystemName,
			ComponentID: monitor.ComponentID,
			Message:     monitor.Name + " check **failed** - " + time.Now().Format(DefaultTimeFormat),
			Notify:      true,
		}

		if len(monitor.lastFailReason) > 0 {
			monitor.incident.Message += "\n\n `" + monitor.lastFailReason + "`"
		}

		// is down, create an incident
		monitor.config.Logger.Printf("%v creating incident. Monitor is down: %v", monitor.Name, monitor.lastFailReason)
		// set investigating status
		monitor.incident.SetInvestigating()
		// create/update incident
		if err := monitor.incident.Send(monitor.config); err != nil {
			monitor.config.Logger.Printf("Error sending incident: %v\n", err)
		}
	} else if t < monitor.Threshold && monitor.incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		monitor.config.Logger.Printf("%v resolved downtime incident", monitor.Name)

		// resolve incident
		monitor.incident.Message = "\n**Resolved** - " + time.Now().Format(DefaultTimeFormat) + "\n\n - - - \n\n" + monitor.incident.Message
		monitor.incident.SetFixed()
		monitor.incident.Send(monitor.config)

		monitor.lastFailReason = ""
		monitor.incident = nil
	}
}

func (monitor *Monitor) ValidateConfiguration() error {
	if len(monitor.ExpectedBody) > 0 {
		exp, err := regexp.Compile(monitor.ExpectedBody)
		if err != nil {
			return err
		}

		monitor.bodyRegexp = exp
	}

	if len(monitor.ExpectedBody) == 0 && monitor.ExpectedStatusCode == 0 {
		return errors.New("Nothing to check, both 'expected_body' and 'expected_status_code' fields empty")
	}

	if monitor.CheckInterval < 1 {
		monitor.CheckInterval = DefaultInterval
	}

	monitor.Method = strings.ToUpper(monitor.Method)
	switch monitor.Method {
	case "GET", "POST", "DELETE", "OPTIONS", "HEAD":
		break
	case "":
		monitor.Method = "GET"
	default:
		return fmt.Errorf("Unsupported check method: %v", monitor.Method)
	}

	if monitor.ComponentID == 0 && monitor.MetricID == 0 {
		return errors.New("component_id & metric_id are unset")
	}

	if monitor.Threshold <= 0 {
		monitor.Threshold = 100
	}

	return nil
}
