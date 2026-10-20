package model

import "time"

type Site struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Location    string     `json:"location"`
	Status      string     `json:"status"` // ACTIVE / INACTIVE / ARCHIVED
	TargetModal int64      `json:"target_modal"`
	CreatedAt   time.Time  `json:"created_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
}

// IsArchived returns true if the site has been archived.
func (s Site) IsArchived() bool {
	return s.Status == "ARCHIVED"
}
