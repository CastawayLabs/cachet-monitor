package cachet

// Component Cachet model
type Component struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Status        int    `json:"status_id"`
	HumanStatus   string `json:"-"`
	IncidentCount int    `json:"-"`
	CreatedAt     int    `json:"created_at"`
	UpdatedAt     int    `json:"updated_at"`
}

// ComponentData json response model
type ComponentData struct {
	Component Component `json:"data"`
}
