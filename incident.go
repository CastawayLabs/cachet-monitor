package cachet

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Incident Cachet data model
type Incident struct {
	ID          json.Number  `json:"id"`
	Name        string       `json:"name"`
	Message     string       `json:"message"`
	Status      json.Number  `json:"status"` // 4?
	HumanStatus string       `json:"human_status"`
	Component   *Component   `json:"-"`
	ComponentID *json.Number `json:"component_id"`
	CreatedAt   *string      `json:"created_at"`
	UpdatedAt   *string      `json:"updated_at"`
}

// GetIncidents - Get list of incidents
func (monitor *CachetMonitor) GetIncidents() ([]Incident, error) {
	_, body, err := monitor.makeRequest("GET", "/incidents", nil)
	if err != nil {
		return []Incident{}, fmt.Errorf("Cannot get incidents: %v\n", err)
	}

	var data struct {
		Incidents []Incident `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []Incident{}, fmt.Errorf("Cannot parse incidents: %v\n", err)
	}

	return data.Incidents, nil
}

// Send - Create or Update incident
func (monitor *CachetMonitor) SendIncident(incident *Incident) error {
	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"name":         incident.Name,
		"message":      incident.Message,
		"status":       incident.Status,
		"component_id": incident.ComponentID,
		"notify":       true,
	})

	requestType := "POST"
	requestURL := "/incidents"
	if len(incident.ID) > 0 {
		requestType = "PUT"
		requestURL += "/" + string(incident.ID)
	}

	resp, body, err := monitor.makeRequest(requestType, requestURL, jsonBytes)
	if err != nil {
		return err
	}

	var data struct {
		Incident Incident `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return errors.New("Cannot parse incident body." + string(body))
	} else {
		incident.ID = data.Incident.ID
		incident.Component = data.Incident.Component
	}

	if resp.StatusCode != 200 {
		return errors.New("Could not create/update incident!")
	}

	return nil
}

func (monitor *CachetMonitor) fetchComponent(componentID string) (*Component, error) {
	_, body, err := monitor.makeRequest("GET", "/components/"+componentID, nil)
	if err != nil {
		return nil, err
	}

	var data struct {
		Component Component `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errors.New("Cannot parse component body. " + string(body))
	}

	return &data.Component, nil
}

func (monitor *CachetMonitor) UpdateComponent(incident *Incident) error {
	if incident.ComponentID == nil || len(*incident.ComponentID) == 0 {
		return nil
	}

	if incident.Component == nil {
		// fetch component
		component, err := monitor.fetchComponent(string(*incident.ComponentID))
		if err != nil {
			return fmt.Errorf("Cannot fetch component for incident. %v\n", err)
		}

		incident.Component = component
	}

	status, _ := strconv.Atoi(string(incident.Status))
	switch status {
	case 1, 2, 3:
		if incident.Component.Status == "3" {
			incident.Component.Status = "4"
		} else {
			incident.Component.Status = "3"
		}
	case 4:
		incident.Component.Status = "1"
	}

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"status": incident.Component.Status,
	})

	resp, _, err := monitor.makeRequest("PUT", "/components/"+string(incident.Component.ID), jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Could not update component: (resp code %d) %v", resp.StatusCode, err)
	}

	return nil
}

// SetInvestigating sets status to Investigating
func (incident *Incident) SetInvestigating() {
	incident.Status = "1"
	incident.HumanStatus = "Investigating"
}

// SetIdentified sets status to Identified
func (incident *Incident) SetIdentified() {
	incident.Status = "2"
	incident.HumanStatus = "Identified"
}

// SetWatching sets status to Watching
func (incident *Incident) SetWatching() {
	incident.Status = "3"
	incident.HumanStatus = "Watching"
}

// SetFixed sets status to Fixed
func (incident *Incident) SetFixed() {
	incident.Status = "4"
	incident.HumanStatus = "Fixed"
}
