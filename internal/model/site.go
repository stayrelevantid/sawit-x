package model

import "time"

type Site struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Status      string    `json:"status"` // ACTIVE / INACTIVE
	TargetModal int64     `json:"target_modal"`
	CreatedAt   time.Time `json:"created_at"`
}
