package cachetbackend

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/castawaylabs/cachet-monitor/monitors"
	"github.com/sirupsen/logrus"
)

const DefaultTimeFormat = "15:04:05 Jan 2 MST"

type CachetBackend struct {
	URL        string `json:"url" yaml:"url"`
	Token      string `json:"token" yaml:"token"`
	Insecure   bool   `json:"insecure" yaml:"insecure"`
	DateFormat string `json:"date_format" yaml:"date_format"`
}

type CachetResponse struct {
	Data json.RawMessage `json:"data"`
}

func (api CachetBackend) ValidateMonitor(mon *monitors.AbstractMonitor) []string {
	errs := []string{}

	params := mon.Params

	componentID, componentIDOk := params["component_id"]
	metricID, metricIDOk := params["metric_id"]
	if !componentIDOk && !metricIDOk {
		errs = append(errs, "component_id and metric_id is unset")
	}

	if _, ok := componentID.(int); !ok && componentIDOk {
		errs = append(errs, "component_id not integer")
	}
	if _, ok := metricID.(int); !ok && metricIDOk {
		errs = append(errs, "metric_id not integer")
	}

	return errs
}

func (api CachetBackend) Validate() []string {
	errs := []string{}

	if len(api.URL) == 0 {
		errs = append(errs, "Cachet API URL invalid")
	}
	if len(api.Token) == 0 {
		errs = append(errs, "Cachet API Token invalid")
	}

	if len(api.DateFormat) == 0 {
		api.DateFormat = DefaultTimeFormat
	}

	return errs
}

// TODO: test
func (api CachetBackend) Ping() error {
	resp, _, err := api.NewRequest("GET", "/ping", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("API Responded with non-200 status code")
	}

	defer resp.Body.Close()

	return nil
}

// TODO: test
// NewRequest wraps http.NewRequest
func (api CachetBackend) NewRequest(requestType, url string, reqBody []byte) (*http.Response, interface{}, error) {
	req, err := http.NewRequest(requestType, api.URL+url, bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", api.Token)

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: api.Insecure}
	client := &http.Client{
		Transport: transport,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, CachetResponse{}, err
	}
	defer res.Body.Close()
	defer req.Body.Close()

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)

	return res, body, err
}

func (mon CachetBackend) Describe() []string {
	features := []string{"Cachet API"}

	return features
}

func (api CachetBackend) SendMetric(monitor monitors.MonitorInterface, lag int64) error {
	mon := monitor.GetMonitor()
	if _, ok := mon.Params["metric_id"]; !ok {
		return nil
	}

	metricID := mon.Params["metric_id"].(int)

	// report lag
	logrus.Debugf("Sending lag metric ID: %d RTT %vms", metricID, lag)

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"value":     lag,
		"timestamp": time.Now().Unix(),
	})

	resp, _, err := api.NewRequest("POST", "/metrics/"+strconv.Itoa(metricID)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		logrus.Warnf("Could not log metric! ID: %d, err: %v", metricID, err)
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	return nil
}

func (api CachetBackend) UpdateMonitor(mon monitors.MonitorInterface, status, previousStatus monitors.MonitorStatus, errs []error) error {
	monitor := mon.GetMonitor()
	l := logrus.WithFields(logrus.Fields{
		"monitor": monitor.Name,
		"time":    time.Now().Format(api.DateFormat),
	})

	errors := make([]string, len(errs))
	for i, err := range errs {
		errors[i] = err.Error()
	}

	fmt.Println("errs", errs)

	componentID := monitor.Params["component_id"].(int)
	incident, err := api.findIncident(componentID)
	if err != nil {
		l.Errorf("Couldn't find existing incidents: %v", err)
	}

	if incident == nil {
		// create a new one
		incident = &Incident{
			Name:        "",
			ComponentID: componentID,
			Message:     "",
			Notify:      true,
		}
	} else {
		// find component status
		component, err := api.getComponent(incident.ComponentID)
		if err != nil {
			panic(err)
		}

		incident.ComponentStatus = component.Status
	}

	tpls := monitor.Template
	tplData := api.getTemplateData(monitor)
	var tpl monitors.MessageTemplate

	if status == monitors.MonitorStatusDown {
		tpl = tpls.Investigating
		tplData["FailReason"] = strings.Join(errors, "\n - ")
		l.Warnf("updating component. Monitor is down: %v", tplData["FailReason"])
	} else {
		// was down, created an incident, its now ok, make it resolved.
		tpl = tpls.Fixed
		l.Warn("Resolving incident")
	}

	tplData["incident"] = incident
	subject, message := tpl.Exec(tplData)

	if incident.ID == 0 {
		incident.Name = subject
		incident.Message = message
	} else {
		incident.Message += "\n\n---\n\n" + subject + ":\n\n" + message
	}

	if status == monitors.MonitorStatusDown && (incident.ComponentStatus == 0 || incident.ComponentStatus > 2) {
		incident.Status = 1
		fmt.Println("incident status", incident.ComponentStatus)
		if incident.ComponentStatus >= 3 {
			// major outage
			incident.ComponentStatus = 4
		} else {
			incident.ComponentStatus = 3
		}
	} else if status == monitors.MonitorStatusUp {
		incident.Status = 4
		incident.ComponentStatus = 1
	}
	incident.Notify = true

	// create/update incident
	if err := incident.Send(api); err != nil {
		l.Errorf("Error sending incident: %v", err)
		return err
	}

	return nil
}

func (api CachetBackend) Tick(monitor monitors.MonitorInterface, status monitors.MonitorStatus, errs []error, lag int64) {
	mon := monitor.GetMonitor()
	if mon.GetLastStatus() == status || status == monitors.MonitorStatusNotSaturated {
		return
	}

	logrus.Infof("updating backend for monitor")
	lastStatus := mon.UpdateLastStatus(status)

	api.UpdateMonitor(monitor, status, lastStatus, errs)

	if _, ok := mon.Params["metric_id"]; ok && lag > 0 {
		api.SendMetric(monitor, lag)
	}
}

func (api CachetBackend) getComponent(componentID int) (*Component, error) {
	return nil, nil
}

func (api CachetBackend) findIncident(componentID int) (*Incident, error) {
	// fetch watching, identified & investigating
	statuses := []int{3, 2, 1}
	for _, status := range statuses {
		incidents, err := api.findIncidents(componentID, status)
		if err != nil {
			return nil, err
		}

		for _, incident := range incidents {
			incident.Status = status
			return incident, nil
		}
	}

	return nil, nil
}

func (api CachetBackend) findIncidents(componentID int, status int) ([]*Incident, error) {
	resp, body, err := api.NewRequest("GET", "/incidents?component_id="+strconv.Itoa(componentID)+"&status="+strconv.Itoa(status), nil)
	if err != nil {
		return nil, err
	}

	var data []*Incident
	if err := json.Unmarshal(body.(CachetResponse).Data, &data); err != nil {
		return nil, fmt.Errorf("Cannot find incidents: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Could not fetch incidents! %v", err)
	}

	return data, nil
}
