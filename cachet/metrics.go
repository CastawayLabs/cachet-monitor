package cachet

import (
	"fmt"
	"strconv"
	"encoding/json"
)

func SendMetric(metricId int, delay int64) {
	if metricId <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := makeRequest("POST", "/metrics/" + strconv.Itoa(metricId) + "/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Could not log data point!\n%v\n", err)
		return
	}
}