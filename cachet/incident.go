package cachet

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Incident Cachet data model
type Incident struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Message     string     `json:"message"`
	Status      int        `json:"status"` // 4?
	HumanStatus string     `json:"human_status"`
	Component   *Component `json:"component"`
	ComponentID *int       `json:"component_id"`
	CreatedAt   int        `json:"created_at"`
	UpdatedAt   int        `json:"updated_at"`
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
		panic(err)
	}

	var data IncidentList
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Cannot parse incidents.")
	}

	return data.Incidents
}

// Send - Create or Update incident
func (incident *Incident) Send() {
	jsonBytes, err := json.Marshal(incident)
	if err != nil {
		panic(err)
	}

	requestType := "POST"
	requestURL := "/incidents"
	if incident.ID > 0 {
		requestType = "PUT"
		requestURL = "/incidents/" + strconv.Itoa(incident.ID)
	}

	resp, body, err := makeRequest(requestType, requestURL, jsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Println(strconv.Itoa(resp.StatusCode) + " " + string(body))

	var data IncidentData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Cannot parse incident body.")
		panic(err)
	} else {
		incident.ID = data.Incident.ID
	}

	fmt.Println("ID:" + strconv.Itoa(incident.ID))

	if resp.StatusCode != 200 {
		fmt.Println("Could not create/update incident!")
	}
}

// GetSimilarIncidentId gets the same incident.
// Updates incident.ID
func (incident *Incident) GetSimilarIncidentID() {
	incidents := GetIncidents()

	for _, inc := range incidents {
		if incident.Name == inc.Name && incident.Message == inc.Message && incident.Status == inc.Status && incident.HumanStatus == inc.HumanStatus {
			incident.ID = inc.ID
			fmt.Printf("Updated incident id to %v\n", inc.ID)
			break
		}
	}
}

// SetInvestigating sets status to Investigating
func (incident *Incident) SetInvestigating() {
	incident.Status = 1
	incident.HumanStatus = "Investigating"
}

// SetIdentified sets status to Identified
func (incident *Incident) SetIdentified() {
	incident.Status = 2
	incident.HumanStatus = "Identified"
}

// SetWatching sets status to Watching
func (incident *Incident) SetWatching() {
	incident.Status = 3
	incident.HumanStatus = "Watching"
}

// SetFixed sets status to Fixed
func (incident *Incident) SetFixed() {
	incident.Status = 4
	incident.HumanStatus = "Fixed"
}
