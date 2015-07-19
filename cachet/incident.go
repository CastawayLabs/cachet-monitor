package cachet

import (
	"encoding/json"
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

// IncidentData is a response when creating/updating an incident
type IncidentData struct {
	Incident Incident `json:"data"`
}

// IncidentList - from API /incidents
type IncidentList struct {
	Incidents []Incident `json:"data"`
}

// GetIncidents - Get list of incidents
func GetIncidents() []Incident {
	_, body, err := makeRequest("GET", "/incidents", nil)
	if err != nil {
		Logger.Printf("Cannot get incidents: %v\n", err)
		return []Incident{}
	}

	var data IncidentList
	err = json.Unmarshal(body, &data)
	if err != nil {
		Logger.Printf("Cannot parse incidents: %v\n", err)
		panic(err)
	}

	return data.Incidents
}

// Send - Create or Update incident
func (incident *Incident) Send() {
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

	resp, body, err := makeRequest(requestType, requestURL, jsonBytes)
	if err != nil {
		Logger.Printf("Cannot create/update incident: %v\n", err)
		return
	}

	Logger.Println(strconv.Itoa(resp.StatusCode) + " " + string(body))

	var data IncidentData
	err = json.Unmarshal(body, &data)
	if err != nil {
		Logger.Println("Cannot parse incident body.", string(body))
		panic(err)
	} else {
		incident.ID = data.Incident.ID
		incident.Component = data.Incident.Component
	}

	if resp.StatusCode != 200 {
		Logger.Println("Could not create/update incident!")
	}
}

// GetSimilarIncidentID gets the same incident.
// Updates incident.ID
func (incident *Incident) GetSimilarIncidentID() {
	incidents := GetIncidents()

	for _, inc := range incidents {
		if incident.Name == inc.Name && incident.Message == inc.Message && incident.Status == inc.Status {
			incident.ID = inc.ID
			Logger.Printf("Updated incident id to %v\n", inc.ID)
			break
		}
	}
}

func (incident *Incident) fetchComponent() error {
	_, body, err := makeRequest("GET", "/components/"+string(*incident.ComponentID), nil)
	if err != nil {
		return err
	}

	var data ComponentData
	err = json.Unmarshal(body, &data)
	if err != nil {
		Logger.Println("Cannot parse component body. %v", string(body))
		panic(err)
	}

	incident.Component = &data.Component

	return nil
}

func (incident *Incident) UpdateComponent() {
	if incident.ComponentID == nil || len(*incident.ComponentID) == 0 {
		return
	}

	if incident.Component == nil {
		// fetch component
		if err := incident.fetchComponent(); err != nil {
			Logger.Printf("Cannot fetch component for incident. %v\n", err)
			return
		}
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

	resp, _, err := makeRequest("PUT", "/components/"+string(incident.Component.ID), jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		Logger.Printf("Could not update component: (resp code %d) %v", resp.StatusCode, err)
		return
	}
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
