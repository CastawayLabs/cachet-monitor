package cachet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Component Cachet model
type Component struct {
	ID            json.Number `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Status        json.Number `json:"status_id"`
	HumanStatus   string      `json:"-"`
	IncidentCount int         `json:"-"`
	CreatedAt     *string     `json:"created_at"`
	UpdatedAt     *string     `json:"updated_at"`
}

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
