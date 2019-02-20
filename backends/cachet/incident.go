package cachetbackend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/castawaylabs/cachet-monitor/backends"
	"github.com/castawaylabs/cachet-monitor/monitors"
)

// "encoding/json"
// "fmt"
// "strconv"

// "github.com/sirupsen/logrus"

// Incident Cachet data model
type Incident struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Visible int    `json:"visible"`
	Notify  bool   `json:"notify"`

	ComponentID     int `json:"component_id"`
	ComponentStatus int `json:"component_status"`
}

// Send - Create or Update incident
func (incident *Incident) Send(backend backends.BackendInterface) error {
	requestURL := "/incidents"
	requestMethod := "POST"
	jsonBytes, _ := json.Marshal(incident)

	if incident.ID > 0 {
		// create an incident update
		requestMethod = "PUT"
		requestURL += "/" + strconv.Itoa(incident.ID)
	}

	resp, body, err := backend.NewRequest(requestMethod, requestURL, jsonBytes)
	if err != nil {
		return err
	}

	var data struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body.(CachetResponse).Data, &data); err != nil {
		return fmt.Errorf("Cannot parse incident body: %v, %v", err, string(body.(CachetResponse).Data))
	}

	incident.ID = data.ID
	if resp.StatusCode != 200 {
		return fmt.Errorf("Could not update/create incident!")
	}

	return nil
}

func (api *CachetBackend) getTemplateData(monitor *monitors.AbstractMonitor) map[string]interface{} {
	return map[string]interface{}{
		// "SystemName": monitor.config.SystemName,
		"Monitor": monitor,
		"now":     time.Now().Format(api.DateFormat),
		// "incident":   monitor.incident,
	}
}
