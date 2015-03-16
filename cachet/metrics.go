package cachet

import (
	"fmt"
	"bytes"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func SendMetric(metricId int, delay int64) {
	jsonBytes, err := json.Marshal(&map[string]interface{}{
		"value": delay,
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", apiUrl + "/metrics/" + strconv.Itoa(metricId) + "/points", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	_, _ = ioutil.ReadAll(resp.Body)
	// fmt.Println(strconv.Itoa(resp.StatusCode) + " " + string(body))

	if resp.StatusCode != 200 {
		fmt.Println("Could not log data point!")
	}
}
