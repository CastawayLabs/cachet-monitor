package cachet

import (
	"time"
)

type Incident struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Message string `json:"message"`
	Status int `json:"status"`// 4?
	Human_status string `json:"human_status"`
	Component *Component `json:"component"`
	Component_id *int `json:"component_id"`
	Created_at *time.Time `json:"created_at"`
	Updated_at *time.Time `json:"updated_at"`
}