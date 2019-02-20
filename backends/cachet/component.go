package cachetbackend

// Incident Cachet data model
type Component struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Visible int    `json:"visible"`
	Notify  bool   `json:"notify"`

	ComponentID     int `json:"component_id"`
	ComponentStatus int `json:"component_status"`
}
