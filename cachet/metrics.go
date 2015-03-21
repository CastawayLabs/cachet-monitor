package cachet

import (
	"encoding/json"
	"strconv"
)

// SendMetric sends lag metric point
func SendMetric(metricID int, delay int64) {
	if metricID <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := makeRequest("POST", "/metrics/"+strconv.Itoa(metricID)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		Logger.Printf("Could not log data point!\n%v\n", err)
		return
	}
}
