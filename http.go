package cachet

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HTTPMonitor struct {
	*AbstractMonitor

	Method             string            `json:"method"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	Headers            map[string]string `json:"headers"`

	// compiled to Regexp
	ExpectedBody string `json:"expected_body"`
	bodyRegexp   *regexp.Regexp
}

func (monitor *HTTPMonitor) do() bool {
	client := &http.Client{
		Timeout: time.Duration(monitor.Timeout * time.Second),
	}
	if monitor.Strict == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	req, err := http.NewRequest(monitor.Method, monitor.Target, nil)
	for k, v := range monitor.Headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
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

func (monitor *HTTPMonitor) Validate() []string {
	errs := []string{}
	if len(monitor.ExpectedBody) > 0 {
		exp, err := regexp.Compile(monitor.ExpectedBody)
		if err != nil {
			errs = append(errs, "Regexp compilation failure: "+err.Error())
		} else {
			monitor.bodyRegexp = exp
		}
	}

	if len(monitor.ExpectedBody) == 0 && monitor.ExpectedStatusCode == 0 {
		errs = append(errs, "Both 'expected_body' and 'expected_status_code' fields empty")
	}

	if monitor.Interval < 1 {
		monitor.Interval = DefaultInterval
	}

	if monitor.Timeout < 1 {
		monitor.Timeout = DefaultTimeout
	}

	monitor.Method = strings.ToUpper(monitor.Method)
	switch monitor.Method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
		break
	case "":
		monitor.Method = "GET"
	default:
		errs = append(errs, "Unsupported check method: "+monitor.Method)
	}

	if monitor.ComponentID == 0 && monitor.MetricID == 0 {
		errs = append(errs, "component_id & metric_id are unset")
	}

	if monitor.Threshold <= 0 {
		monitor.Threshold = 100
	}

	return errs
}

func (mon *HTTPMonitor) GetMonitor() *AbstractMonitor {
	return mon.AbstractMonitor
}

// SendMetric sends lag metric point
/*func (monitor *Monitor) SendMetric(delay int64) error {
	if monitor.MetricID == 0 {
		return nil
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := monitor.config.makeRequest("POST", "/metrics/"+strconv.Itoa(monitor.MetricID)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Could not log data point!\n%v\n", err)
	}

	return nil
}
*/
