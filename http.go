package cachet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (monitor *CachetMonitor) makeRequest(requestType string, url string, reqBody []byte) (*http.Response, []byte, error) {
	req, err := http.NewRequest(requestType, monitor.APIUrl+url, bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", monitor.APIToken)

	client := &http.Client{}
	if monitor.InsecureAPI == true {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, []byte{}, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return res, body, nil
}

// SendMetric sends lag metric point
func (monitor *Monitor) SendMetric(delay int64) error {
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

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
