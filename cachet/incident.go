package cachet

import (
	"fmt"
	"strconv"
	"encoding/json"
)

type Incident struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Message string `json:"message"`
	Status int `json:"status"`// 4?
	Human_status string `json:"human_status"`
	Component *Component `json:"component"`
	Component_id *int `json:"component_id"`
	Created_at int `json:"created_at"`
	Updated_at int `json:"updated_at"`
}

type IncidentData struct {
	Incident Incident `json:"data"`
}

type IncidentList struct {
	Incidents []Incident `json:"data"`
}

func GetIncidents() []Incident {
	_, body, err := makeRequest("GET", "/incidents", nil)
	if err != nil {
		panic(err)
	}

	var data IncidentList
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Cannot parse incidents.")
		panic(err)
	}

	return data.Incidents
}

func (incident *Incident) Send() {
	jsonBytes, err := json.Marshal(incident)
	if err != nil {
		panic(err)
	}

	requestType := "POST"
	requestUrl := "/incidents"
	if incident.Id > 0 {
		requestType = "PUT"
		requestUrl = "/incidents/" + strconv.Itoa(incident.Id)
	}

	resp, body, err := makeRequest(requestType, requestUrl, jsonBytes)
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
		incident.Id = data.Incident.Id
	}

	fmt.Println("ID:"+strconv.Itoa(incident.Id))

	if resp.StatusCode != 200 {
		fmt.Println("Could not create/update incident!")
	}
}

func (incident *Incident) GetSimilarIncidentId() {
	incidents := GetIncidents()

	for _, inc := range incidents {
		if incident.Name == inc.Name && incident.Message == inc.Message && incident.Status == inc.Status && incident.Human_status == inc.Human_status {
			incident.Id = inc.Id
			fmt.Printf("Updated incident id to %v\n", inc.Id)
			break
		}
	}
}

func (incident *Incident) SetInvestigating() {
	incident.Status = 1
	incident.Human_status = "Investigating"
}

func (incident *Incident) SetIdentified() {
	incident.Status = 2
	incident.Human_status = "Identified"
}

func (incident *Incident) SetWatching() {
	incident.Status = 3
	incident.Human_status = "Watching"
}

func (incident *Incident) SetFixed() {
	incident.Status = 4
	incident.Human_status = "Fixed"
}