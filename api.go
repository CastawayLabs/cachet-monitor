package cachet

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
)

type CachetAPI struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	Insecure bool   `json:"insecure"`
}

func (api CachetAPI) Ping() error {
	resp, _, err := api.NewRequest("GET", "/ping", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("API Responded with non-200 status code")
	}

	return nil
}

func (api CachetAPI) NewRequest(requestType, url string, reqBody []byte) (*http.Response, []byte, error) {
	req, err := http.NewRequest(requestType, api.URL+url, bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", api.Token)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if api.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, []byte{}, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return res, body, nil
}
