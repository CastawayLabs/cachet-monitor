package cachet

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"fmt"
)

// Investigating template
var defaultHTTPInvestigatingTpl = MessageTemplate{
	Subject: `{{ .Monitor.Name }} - {{ .SystemName }}`,
	Message: `{{ .Monitor.Name }} check **failed** (server time: {{ .now }})

{{ .FailReason }}`,
}

// Fixed template
var defaultHTTPFixedTpl = MessageTemplate{
	Subject: `{{ .Monitor.Name }} - {{ .SystemName }}`,
	Message: `**Resolved** - {{ .now }}

- - -

{{ .incident.Message }}`,
}

type HTTPMonitor struct {
	AbstractMonitor `mapstructure:",squash"`

	Method             string
	ExpectedStatusCode int `mapstructure:"expected_status_code"`
	Headers            map[string]string

	// compiled to Regexp
	ExpectedBody string `mapstructure:"expected_body"`
	bodyRegexp   *regexp.Regexp

	// JSON data
	Data string `mapstructure:"data"`
}

// TODO: test
func (monitor *HTTPMonitor) test() bool {
	var dataBuffer *bytes.Buffer = nil
	var req *http.Request
	var err error
	if monitor.Data != "" {
		fmt.Println("Data: ", monitor.Data)
		dataBuffer = bytes.NewBuffer([]byte(monitor.Data))
		req, err = http.NewRequest(monitor.Method, monitor.Target, dataBuffer)
	} else {
	  req, err = http.NewRequest(monitor.Method, monitor.Target, nil)
	}
	fmt.Println("Target: ", monitor.Target)
	fmt.Println("Target: ", dataBuffer)
	for k, v := range monitor.Headers {
		fmt.Println(k, ": ", v)
		req.Header.Add(k, v)
	}

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: monitor.Strict == false}
	client := &http.Client{
		Timeout:   time.Duration(monitor.Timeout * time.Second),
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		monitor.lastFailReason = err.Error()
		return false
	}

	defer resp.Body.Close()

	if monitor.ExpectedStatusCode > 0 && resp.StatusCode != monitor.ExpectedStatusCode {
		monitor.lastFailReason = "Expected HTTP response status: " + strconv.Itoa(monitor.ExpectedStatusCode) + ", got: " + strconv.Itoa(resp.StatusCode)
		fmt.Println(monitor.lastFailReason)
		return false
	}

	if monitor.bodyRegexp != nil {
		// check response body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			monitor.lastFailReason = err.Error()
			fmt.Println(err.Error())
			return false
		}

		if !monitor.bodyRegexp.Match(responseBody) {
			monitor.lastFailReason = "Unexpected body: " + string(responseBody) + ".\nExpected to match: " + monitor.ExpectedBody
			fmt.Println(monitor.lastFailReason)
			return false
		}
	}

	return true
}

// TODO: test
func (mon *HTTPMonitor) Validate() []string {
	mon.Template.Investigating.SetDefault(defaultHTTPInvestigatingTpl)
	mon.Template.Fixed.SetDefault(defaultHTTPFixedTpl)

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
