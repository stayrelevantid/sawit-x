package model

import "time"

// ModuleType constants for transaction classification.
const (
	ModulePanen       = "PANEN"
	ModuleOperasional = "OPERASIONAL"
	ModulePiutang     = "PIUTANG"
	ModuleInvestasi   = "INVESTASI"
)

type LogEntry struct {
	LogID         string    `json:"log_id"`
	Timestamp     time.Time `json:"timestamp"`
	EventDate     time.Time `json:"event_date"`
	ModuleType    string    `json:"module_type"`   // PANEN / OPERASIONAL / PIUTANG
	SiteID        string    `json:"site_id"`
	SiteName      string    `json:"site_name"`
	CategoryID    string    `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CrewID        string    `json:"crew_id"`       // Comma-separated IDs for multi-select (Panen)
	CrewName      string    `json:"crew_name"`     // Comma-separated names
	AmountRaw     int64     `json:"amount_raw"`    // Gross income (Panen) or expense amount
	AmountFinal   int64     `json:"amount_final"`  // Net income after costs (Panen) or same as raw
	Weight        int64     `json:"weight"`        // Kg (Panen)
	UnitPrice     int64     `json:"unit_price"`    // Price per Kg (Panen)
	LaborCost     int64     `json:"labor_cost"`    // Upah Panen
	TransportCost int64     `json:"transport_cost"` // Bensin/Timbang (Panen)
	Notes         string    `json:"notes"`
	SlackUserID   string    `json:"slack_user_id"`
	SlackUsername string    `json:"slack_username"`
	ChannelID     string    `json:"channel_id"`
}
