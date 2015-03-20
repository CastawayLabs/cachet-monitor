package cachet

import (
	"fmt"
	"strconv"
	"encoding/json"
)

// Send lag metric point
func SendMetric(metricID int, delay int64) {
	if metricID <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := makeRequest("POST", "/metrics/" + strconv.Itoa(metricID) + "/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Could not log data point!\n%v\n", err)
		return
	}
}