package cachet

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

const timeout = time.Duration(time.Second)

// Monitor data model
type Monitor struct {
	Name               string        `json:"name"`
	URL                string        `json:"url"`
	MetricID           int           `json:"metric_id"`
	Threshold          float32       `json:"threshold"`
	ComponentID        *int          `json:"component_id"`
	ExpectedStatusCode int           `json:"expected_status_code"`
	StrictTLS          *bool         `json:"strict_tls"`
	Interval           time.Duration `json:"interval"`

	History        []bool    `json:"-"`
	LastFailReason *string   `json:"-"`
	Incident       *Incident `json:"-"`
	config         *CachetMonitor

	// Closed when mon.Stop() is called
	stopC chan bool
}

func (cfg *CachetMonitor) Run() {
	cfg.Logger.Printf("System: %s\nInterval: %d second(s)\nAPI: %s\n\n", cfg.SystemName, cfg.Interval, cfg.APIUrl)
	cfg.Logger.Printf("Starting %d monitors:\n", len(cfg.Monitors))
	for _, mon := range cfg.Monitors {
		cfg.Logger.Printf(" %s: GET %s & Expect HTTP %d\n", mon.Name, mon.URL, mon.ExpectedStatusCode)
		if mon.MetricID > 0 {
			cfg.Logger.Printf(" - Logs lag to metric id: %d\n", mon.MetricID)
		}
		if mon.ComponentID != nil && *mon.ComponentID > 0 {
			cfg.Logger.Printf(" - Updates component id: %d\n", *mon.ComponentID)
		}
	}

	cfg.Logger.Println()
	wg := &sync.WaitGroup{}

	for _, mon := range cfg.Monitors {
		wg.Add(1)
		mon.config = cfg
		mon.stopC = make(chan bool)

		go func(mon *Monitor) {
			if mon.Interval < 1 {
				mon.Interval = time.Duration(cfg.Interval)
			}

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
	for _, mon := range cfg.Monitors {
		mon.Stop()
	}

	wg.Wait()
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
		monitor.config.SendMetric(monitor.MetricID, lag)
	}
}

func (monitor *Monitor) Stop() {
	if monitor.Stopped() {
		return
	}

	close(monitor.stopC)
}

func (monitor *Monitor) StopC() <-chan bool {
	return monitor.stopC
}

func (monitor *Monitor) Stopped() bool {
	select {
	case <-monitor.stopC:
		return true
	default:
		return false
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
	monitor.config.Logger.Printf("%s %.2f%% Down at %v. Threshold: %.2f%%\n", monitor.URL, t, time.Now().UnixNano()/int64(time.Second), monitor.Threshold)

	if len(monitor.History) != 10 {
		// not enough data
		return
	}

	if t > monitor.Threshold && monitor.Incident == nil {
		// is down, create an incident
		monitor.config.Logger.Println("Creating incident...")

		component_id := json.Number(strconv.Itoa(*monitor.ComponentID))
		monitor.Incident = &Incident{
			Name:        monitor.Name + " - " + monitor.config.SystemName,
			Message:     monitor.Name + " check failed",
			ComponentID: &component_id,
		}

		if monitor.LastFailReason != nil {
			monitor.Incident.Message += "\n\n - " + *monitor.LastFailReason
		}

		// set investigating status
		monitor.Incident.SetInvestigating()

		// create/update incident
		monitor.config.SendIncident(monitor.Incident)
		monitor.config.UpdateComponent(monitor.Incident)
	} else if t < monitor.Threshold && monitor.Incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		monitor.config.Logger.Println("Updating incident to resolved...")

		component_id := json.Number(strconv.Itoa(*monitor.ComponentID))
		monitor.Incident = &Incident{
			Name:        monitor.Incident.Name,
			Message:     monitor.Name + " check succeeded",
			ComponentID: &component_id,
		}

		monitor.Incident.SetFixed()
		monitor.config.SendIncident(monitor.Incident)
		monitor.config.UpdateComponent(monitor.Incident)

		monitor.Incident = nil
	}
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
