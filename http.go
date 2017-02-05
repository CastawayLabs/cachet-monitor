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

// // Investigating template
// var HTTPTemplate = MessageTemplate{
// 	Subject: `{{ .Name }} - {{ .config.SystemName }}`,
// 	Message: `{{ .Name }} check **failed** - {{ .now }}

// {{ .lastFailReason }}`,
// }

// // Fixed template
// var HTTPTemplate = MessageTemplate{
// 	Subject: `{{ .Name }} - {{ .config.SystemName }}`,
// 	Message: `**Resolved** - {{ .now }}

// - - -

// {{ .incident.Message }}`,
// }

type HTTPMonitor struct {
	AbstractMonitor `mapstructure:",squash"`

	Method             string
	ExpectedStatusCode int `mapstructure:"expected_status_code"`
	Headers            map[string]string

	// compiled to Regexp
	ExpectedBody string `mapstructure:"expected_body"`
	bodyRegexp   *regexp.Regexp
}

func (monitor *HTTPMonitor) test() bool {
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

func (mon *HTTPMonitor) Validate() []string {
	errs := mon.AbstractMonitor.Validate()

	if len(mon.ExpectedBody) > 0 {
		exp, err := regexp.Compile(mon.ExpectedBody)
		if err != nil {
			errs = append(errs, "Regexp compilation failure: "+err.Error())
		} else {
			mon.bodyRegexp = exp
		}
	}

	if len(mon.ExpectedBody) == 0 && mon.ExpectedStatusCode == 0 {
		errs = append(errs, "Both 'expected_body' and 'expected_status_code' fields empty")
	}

	mon.Method = strings.ToUpper(mon.Method)
	switch mon.Method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
		break
	case "":
		mon.Method = "GET"
	default:
		errs = append(errs, "Unsupported HTTP method: "+mon.Method)
	}

	return errs
}

func (mon *HTTPMonitor) Describe() []string {
	features := mon.AbstractMonitor.Describe()
	features = append(features, "Method: "+mon.Method)

	return features
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
