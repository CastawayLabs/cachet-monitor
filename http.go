package cachet

import (
	"crypto/tls"
	"crypto/md5"
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

	// data
	Data string `mapstructure:"data"`
	ExpectedMd5Sum string `mapstructure:"expected_md5sum"`
	ExpectedLength int `mapstructure:"expected_length"`
}

// TODO: test
func (monitor *HTTPMonitor) test() bool {
	var req *http.Request
	var err error
	if monitor.Data != "" {
		fmt.Println("Data: ", monitor.Data)
		dataBuffer := strings.NewReader(monitor.Data)
		req, err = http.NewRequest(monitor.Method, monitor.Target, dataBuffer)
		fmt.Println("Target: ", dataBuffer)
	} else {
	  req, err = http.NewRequest(monitor.Method, monitor.Target, nil)
	}
	fmt.Println("Target: ", monitor.Target)
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

	// check response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	responseLength := len(string(responseBody))
	fmt.Println("Response: ", string(responseBody))
	fmt.Println("Response len: ", responseLength)
	if err != nil {
		monitor.lastFailReason = err.Error()
		fmt.Println(err.Error())
		return false
	}

	if monitor.ExpectedLength > 0 && responseLength != monitor.ExpectedLength {
		monitor.lastFailReason = "Expected response body length: " + strconv.Itoa(monitor.ExpectedLength) + ", got: " + strconv.Itoa(responseLength)
		return false
	}

	if monitor.ExpectedMd5Sum != "" {
		sum := fmt.Sprintf("%x", (md5.Sum(responseBody)))
		fmt.Println("Calculated sum", sum)
		fmt.Println("Expected sum", monitor.ExpectedMd5Sum)
		if strings.Compare(sum, monitor.ExpectedMd5Sum) != 0 {
			monitor.lastFailReason = "Expected respsone body MD5 checksum: " + monitor.ExpectedMd5Sum + ", got: " + sum
			return false
		}
	}

	if monitor.bodyRegexp != nil {
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
