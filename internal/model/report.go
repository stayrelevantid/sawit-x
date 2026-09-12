package model

import "time"

type SiteReport struct {
	TotalWeight      int64   `json:"total_weight"`      // Panen Kg
	GrossIncome      int64   `json:"gross_income"`      // Total Rp from Panen
	TotalUpah        int64   `json:"total_upah"`        // Total Rp Upah Panen
	TotalTransport   int64   `json:"total_transport"`   // Total Rp Bensin/Timbang
	TotalOperasional int64   `json:"total_operasional"` // Total Rp from Operasional Module
	OperationalCost  int64   `json:"operational_cost"`  // Sum of Upah + Transport + Operasional
	NetProfit        int64   `json:"net_profit"`        // Gross - Ops Cost
	TargetModal      int64   `json:"target_modal"`      // Target from Sites sheet + Investasi
	RemainingCapital int64   `json:"remaining_capital"` // Sisa modal yang belum balik (Target - NetProfit)
	ROITracking      float64 `json:"roi_tracking"`      // Percentage (Net/Target)

	// Metrik Produktivitas Sawit (Komprehensif)
	HarvestCount     int   `json:"harvest_count"`      // Berapa kali putaran panen
	AvgHarvestWeight int64 `json:"avg_harvest_weight"` // Rata-rata Kg per putaran panen
	AvgPricePerKg    int64 `json:"avg_price_per_kg"`   // Rata-rata harga jual TBS per Kg
	CostPerKg        int64 `json:"cost_per_kg"`        // Biaya operasional per Kg
	ProfitPerKg      int64 `json:"profit_per_kg"`      // Keuntungan bersih per Kg

	// Rincian Biaya Operasional Terurai
	TotalPupukCost    int64 `json:"total_pupuk_cost"`     // Belanja Pupuk
	TotalSemprotCost  int64 `json:"total_semprot_cost"`   // Belanja Racun / Obat Semprot
	TotalOtherOpsCost int64 `json:"total_other_ops_cost"` // Biaya Operasional Lainnya

	// Status Kesehatan & Narasi Bahasa Bayi
	HealthStatus     string `json:"health_status"`      // e.g. "🟢 UNTUNG & SEHAT"
	HealthEmoji      string `json:"health_emoji"`       // e.g. "🟢"
	SummaryNarration string `json:"summary_narration"`  // Narasi bahasa bayi ringkas

	// Piutang Summary
	TotalPinjam     int64  `json:"total_pinjam"`
	TotalBayar      int64  `json:"total_bayar"`
	OutstandingDebt int64  `json:"outstanding_debt"` // Total Pinjam - Total Bayar
	BEPProjection   string `json:"bep_projection"`   // e.g. "Estimasi 18 bulan lagi"
}

type PupukLogEntry struct {
	EventDate time.Time `json:"event_date"`
	CrewName  string    `json:"crew_name"`
	Amount    int64     `json:"amount"`
	Notes     string    `json:"notes"`
}

type CrewDebtSummary struct {
	CrewID          string     `json:"crew_id"`
	CrewName        string     `json:"crew_name"`
	Role            string     `json:"role"`
	TotalPinjam     int64      `json:"total_pinjam"`
	TotalBayar      int64      `json:"total_bayar"`
	OutstandingDebt int64      `json:"outstanding_debt"`
	LastPinjamDate  *time.Time `json:"last_pinjam_date,omitempty"`
	LastBayarDate   *time.Time `json:"last_bayar_date,omitempty"`
}

type SemprotLogEntry struct {
	EventDate time.Time `json:"event_date"`
	CrewName  string    `json:"crew_name"`
	Amount    int64     `json:"amount"`
	Notes     string    `json:"notes"`
}

type HutangLogEntry struct {
	EventDate  time.Time `json:"event_date"`
	CrewName   string    `json:"crew_name"`
	CategoryID string    `json:"category_id"` // "PINJAM" or "BAYAR"
	Amount     int64     `json:"amount"`
	Balance    int64     `json:"balance"`
	Notes      string    `json:"notes"`
}


