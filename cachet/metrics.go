package cachet

import (
	"fmt"
	"bytes"
	"strconv"
	"net/http"
	"encoding/json"
)

func SendMetric(metricId int, delay int64) {
	if metricId <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	client := &http.Client{}
	req, _ := http.NewRequest("POST", Config.API_Url + "/metrics/" + strconv.Itoa(metricId) + "/points", bytes.NewBuffer(jsonBytes))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", Config.API_Token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Could not log data point!\n%v\n", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Could not log data point!")
	}
}