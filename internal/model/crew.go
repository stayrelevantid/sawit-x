package model

type Crew struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	SiteID string `json:"site_id"`
	Status string `json:"status"` // ACTIVE / INACTIVE
}
