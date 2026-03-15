package model

type Category struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"` // OPEX / CAPEX / PENDAPATAN
	MultiplierEnabled bool   `json:"multiplier_enabled"`
	Status            string `json:"status"` // ACTIVE / INACTIVE
}
