package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const spreadsheetID = "1RBWoR_ZOfPZPKgL4RCnYcvI0xe2FFY8B_qPrxONTvVM"

func main() {
	ctx := context.Background()
	
	// Try to use se.json if exists, else ADC
	credFile := "/Users/muhammad.indragiri/Kerja/sawit-x/se.json"
	var opts []option.ClientOption
	if _, err := os.Stat(credFile); err == nil {
		opts = append(opts, option.WithCredentialsFile(credFile))
	}
	opts = append(opts, option.WithScopes(sheets.SpreadsheetsScope))

	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		log.Fatalf("Failed to create Sheets service: %v", err)
	}

	// 1. Clean existing data (Clear ranges)
	fmt.Println("🧹 Cleaning existing data for PRODUCTION baseline...")
	rangesToClear := []string{
		"Sites!A1:Z100",
		"Categories!A1:Z100",
		"Crew!A1:Z100",
		"X_LOG!A1:Z5000",
		"X_REKAP!A1:Z100",
	}
	rbce := &sheets.BatchClearValuesRequest{Ranges: rangesToClear}
	_, err = srv.Spreadsheets.Values.BatchClear(spreadsheetID, rbce).Do()
	if err != nil {
		log.Fatalf("Failed to clear sheets: %v", err)
	}

	// 2. Seed Production Baseline: Sites
	fmt.Println("🌱 Seeding Production Sites...")
	sitesData := [][]interface{}{
		{"id", "name", "location", "status", "target_modal", "created_at"},
		{"SITE_001", "Kebun Utama", "Lokasi Utama", "ACTIVE", 0, "2026-03-15"},
	}
	writeSheet(srv, "Sites!A1", sitesData)

	// 3. Seed Production Baseline: Categories
	fmt.Println("🌱 Seeding Production Categories...")
	categoriesData := [][]interface{}{
		{"id", "name", "type", "multiplier_enabled", "status"},
		{"CAT_PUPUK", "Pupuk", "OPEX", "TRUE", "ACTIVE"},
		{"CAT_BENSIN", "Bensin", "OPEX", "TRUE", "ACTIVE"},
		{"CAT_PRUNING", "Pruning", "OPEX", "FALSE", "ACTIVE"},
		{"CAT_SEMPROT", "Semprot", "OPEX", "FALSE", "ACTIVE"},
		{"CAT_GAJI", "Gaji / Upah Bulanan", "OPEX", "TRUE", "ACTIVE"},
		{"CAT_OPERASIONAL", "Biaya Ops Lainnya", "OPEX", "TRUE", "ACTIVE"},
		{"PANEN", "Panen TBS", "PANEN", "FALSE", "ACTIVE"},
		{"PINJAM", "Pinjam", "PIUTANG", "FALSE", "ACTIVE"},
		{"BAYAR", "Bayar / Potong", "PIUTANG", "FALSE", "ACTIVE"},
	}
	writeSheet(srv, "Categories!A1", categoriesData)

	// 4. Seed Production Baseline: Crew
	fmt.Println("🌱 Seeding Production Crew...")
	crewData := [][]interface{}{
		{"id", "name", "role", "site_id", "status"},
		{"CREW_001", "Mandor Utama", "Mandor", "SITE_001", "ACTIVE"},
	}
	writeSheet(srv, "Crew!A1", crewData)

	// 5. Seed X_LOG Headers (20 Columns)
	fmt.Println("🌱 Seeding X_LOG Headers...")
	logHeaders := [][]interface{}{
		{
			"log_id", "timestamp", "event_date", "module_type", "site_id",
			"site_name", "category_id", "category_name", "crew_id", "crew_name",
			"amount_raw", "amount_final", "weight", "unit_price", "labor_cost",
			"transport_cost", "notes", "slack_user_id", "slack_username", "channel_id",
		},
	}
	writeSheet(srv, "X_LOG!A1", logHeaders)

	// 6. Seed X_REKAP Headers (14 Columns)
	fmt.Println("🌱 Seeding X_REKAP Headers...")
	rekapHeaders := [][]interface{}{
		{
			"site_id", "site_name", "total_weight_kg", "gross_income_rp", "opex_rp",
			"net_profit_rp", "investasi_total_rp", "sisa_modal_rp", "roi_percent", "bep_projection",
			"total_pinjam", "total_bayar", "outstanding_debt", "last_updated",
		},
	}
	writeSheet(srv, "X_REKAP!A1", rekapHeaders)

	fmt.Println("✅ PRODUCTION baseline reset and seeding completed successfully!")
}

func writeSheet(srv *sheets.Service, rangeName string, values [][]interface{}) {
	vr := &sheets.ValueRange{Values: values}
	_, err := srv.Spreadsheets.Values.Update(spreadsheetID, rangeName, vr).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Failed to update range %s: %v", rangeName, err)
	}
}
