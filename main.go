
package main

import (
	"fmt"
	// "time"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"time"
)

const timeout = time.Duration(time.Second)

func main() {
	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		reqStart := time.Now().UnixNano() / int64(time.Millisecond)
		doRequest()
		reqEnd := time.Now().UnixNano() / int64(time.Millisecond)
		go sendMetric(reqEnd - reqStart)
	}
}

func doRequest() error {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get("https://nodegear.io/ping") // http://127.0.0.1:1337
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func sendMetric(delay int64) {
	js := &map[string]interface{}{
		"value": delay,
	}

	jsonBytes, err := json.Marshal(&js)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://demo.cachethq.io/api/metrics/1/points", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", "5wQt9MnJXmhnQsDI8Hmv")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
