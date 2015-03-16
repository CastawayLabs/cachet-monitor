package cachet

import (
	"time"
)

type Component struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Status int `json:"status"`
	Link *string `json:"link"`
	Order *int `json:"order"`
	Group_id *int `json:"group_id"`
	Created_at *time.Time `json:"created_at"`
	Updated_at *time.Time `json:"updated_at"`
	Deleted_at *time.Time `json:"deleted_at"`
}