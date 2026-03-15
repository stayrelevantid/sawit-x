package model

type SiteReport struct {
	TotalWeight      int64   `json:"total_weight"`      // Panen Kg
	GrossIncome      int64   `json:"gross_income"`      // Total Rp from Panen
	TotalUpah        int64   `json:"total_upah"`        // Total Rp Upah Panen
	TotalTransport   int64   `json:"total_transport"`   // Total Rp Bensin/Timbang
	TotalOperasional int64   `json:"total_operasional"` // Total Rp from Operasional Module
	OperationalCost  int64   `json:"operational_cost"`  // Sum of Upah + Transport + Operasional
	NetProfit        int64   `json:"net_profit"`        // Gross - Ops Cost
	TargetModal      int64   `json:"target_modal"`      // Target from Sites sheet
	RemainingCapital int64   `json:"remaining_capital"` // Sisa modal yang belum balik (Target - NetProfit)
	ROITracking      float64 `json:"roi_tracking"`      // Percentage (Net/Target)

	// Piutang Summary
	TotalPinjam     int64 `json:"total_pinjam"`
	TotalBayar      int64 `json:"total_bayar"`
	OutstandingDebt int64 `json:"outstanding_debt"` // Total Pinjam - Total Bayar
	BEPProjection   string `json:"bep_projection"`   // e.g. "Estimasi 18 bulan lagi"
}
