package cachet

import (
	"fmt"
	"bytes"
	"io/ioutil"
	"strconv"
	"net/http"
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

func (incident *Incident) Send() {
	jsonBytes, err := json.Marshal(incident)
	if err != nil {
		panic(err)
	}

	var req *http.Request
	if incident.Id == 0 {
		req, err = http.NewRequest("POST", apiUrl + "/incidents", bytes.NewBuffer(jsonBytes))
	} else {
		req, err = http.NewRequest("PUT", apiUrl + "/incidents/" + strconv.Itoa(incident.Id), bytes.NewBuffer(jsonBytes))
	}

	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cachet-Token", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
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