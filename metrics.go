package cachet

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// SendMetric sends lag metric point
func (monitor *CachetMonitor) SendMetric(metricID int, delay int64) error {
	if metricID <= 0 {
		return nil
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := monitor.makeRequest("POST", "/metrics/"+strconv.Itoa(metricID)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Could not log data point!\n%v\n", err)
	}

	return nil
}
