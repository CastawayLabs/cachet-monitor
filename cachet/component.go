package cachet

import (
	"time"
)

// Component Cachet model
type Component struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Status int `json:"status"`
	Link *string `json:"link"`
	Order *int `json:"order"`
	GroupID *int `json:"group_id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}