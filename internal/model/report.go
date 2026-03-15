package model

type SiteReport struct {
	TotalWeight     int64   `json:"total_weight"`
	OperationalCost int64   `json:"operational_cost"`
	GrossIncome     int64   `json:"gross_income"`
	NetProfit       int64   `json:"net_profit"`
	TargetModal     int64   `json:"target_modal"`
	ROITracking     float64 `json:"roi_tracking"`
}
