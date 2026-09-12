package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/indragiri/sawit-x/internal/model"
	"github.com/indragiri/sawit-x/internal/service"
)

// mockSheetsClient is a fake SheetsClient for testing without real GCP calls.
type mockSheetsClient struct {
	readData     [][]interface{}
	readRangeMap map[string][][]interface{}
	readErr      error
	appendedRows [][]interface{}
	appendErr    error
}

func (m *mockSheetsClient) ReadSpreadsheet(readRange string) ([][]interface{}, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if m.readRangeMap != nil {
		if data, ok := m.readRangeMap[readRange]; ok {
			return data, nil
		}
	}
	return m.readData, nil
}

func (m *mockSheetsClient) AppendRow(sheetName string, row []interface{}) error {
	m.appendedRows = append(m.appendedRows, row)
	return m.appendErr
}

func (m *mockSheetsClient) UpdateCell(cellRange string, value interface{}) error {
	return nil
}

func (m *mockSheetsClient) UpdateRange(rangeName string, values [][]interface{}) error {
	return nil
}

// ---- MasterDataService Tests ----

func TestGetActiveSites_OnlyReturnsActive(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"SITE_001", "Kebun Induk", "Kalimantan", "ACTIVE", "50000000"},
			{"SITE_002", "Kebun Plasma", "Kalimantan", "INACTIVE", "30000000"},
			{"SITE_003", "Kebun Baru", "Kalimantan", "ACTIVE", "20000000"},
		},
	}

	svc := service.NewMasterDataService(mock)
	sites, err := svc.GetActiveSites(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sites) != 2 {
		t.Errorf("expected 2 active sites, got %d", len(sites))
	}
	if sites[0].ID != "SITE_001" || sites[1].ID != "SITE_003" {
		t.Errorf("unexpected site IDs: %v", sites)
	}
}

func TestGetActiveSites_SheetError(t *testing.T) {
	mock := &mockSheetsClient{
		readErr: errors.New("sheets API error"),
	}
	svc := service.NewMasterDataService(mock)
	_, err := svc.GetActiveSites(context.Background())
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestGetActiveSites_ShortRowSkipped(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"SITE_001", "Kebun Induk", "Kalimantan"}, // Only 3 cols — should be skipped
		},
	}
	svc := service.NewMasterDataService(mock)
	sites, err := svc.GetActiveSites(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 sites (short row skipped), got %d", len(sites))
	}
}

func TestGetActiveCategories_OnlyReturnsActive(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"CAT_PUPUK", "Pupuk", "OPEX", "TRUE", "ACTIVE"},
			{"CAT_OLD", "Tanaman Tua", "OPEX", "FALSE", "INACTIVE"},
		},
	}
	svc := service.NewMasterDataService(mock)
	cats, err := svc.GetActiveCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("expected 1 active category, got %d", len(cats))
	}
	if !cats[0].MultiplierEnabled {
		t.Error("expected MultiplierEnabled to be true")
	}
}

func TestGetActiveCrew_OnlyReturnsActive(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"CREW_001", "Jono", "Mandor", "SITE_001", "ACTIVE"},
			{"CREW_002", "Slamet", "Buruh Harian", "SITE_001", "INACTIVE"},
		},
	}
	svc := service.NewMasterDataService(mock)
	crew, err := svc.GetActiveCrew(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(crew) != 1 || crew[0].Name != "Jono" {
		t.Errorf("expected 1 active crew 'Jono', got %v", crew)
	}
}

// ---- LogService Tests ----

func TestWriteLog_AppendsCorrectColumns(t *testing.T) {
	mock := &mockSheetsClient{}
	svc := service.NewLogService(mock)

	entry := model.LogEntry{
		LogID:         "uuid-1234",
		ModuleType:    "OPERASIONAL",
		SiteID:        "SITE_001",
		SiteName:      "Kebun Induk",
		CategoryID:    "CAT_PUPUK",
		CategoryName:  "Pupuk",
		CrewID:        "CREW_001",
		CrewName:      "Jono",
		AmountRaw:     200,
		AmountFinal:   200000,
		Notes:         "Test catatan",
		SlackUserID:   "U123",
		SlackUsername: "jono.mandor",
		ChannelID:     "C456",
	}

	err := svc.WriteLog(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.appendedRows) != 1 {
		t.Fatalf("expected 1 appended row, got %d", len(mock.appendedRows))
	}
	row := mock.appendedRows[0]
	// New schema: 20 columns (added module_type at index 3)
	if len(row) != 20 {
		t.Errorf("expected 20 columns per new schema, got %d", len(row))
	}
	// Spot check
	if row[0] != "uuid-1234" {
		t.Errorf("expected log_id 'uuid-1234', got %v", row[0])
	}
	if row[3] != "OPERASIONAL" {
		t.Errorf("expected module_type 'OPERASIONAL', got %v", row[3])
	}
	// amount_raw is now at index 10 (was 9 before module_type addition)
	if row[10] != int64(200) {
		t.Errorf("expected amount_raw 200, got %v", row[10])
	}
	if row[11] != int64(200000) {
		t.Errorf("expected amount_final 200000, got %v", row[11])
	}
}

func TestWriteLog_SheetsError(t *testing.T) {
	mock := &mockSheetsClient{appendErr: errors.New("append failed")}
	svc := service.NewLogService(mock)
	err := svc.WriteLog(context.Background(), model.LogEntry{})
	if err == nil {
		t.Error("expected an error on append failure, got nil")
	}
}

// ---- GetCrewBalance Tests ----

func TestGetCrewBalance_ReturnsCorrectBalance(t *testing.T) {
	// The mock returns X_LOG rows with 12 columns.
	// col[3]=module_type, col[6]=category_id (PINJAM/BAYAR), col[8]=crew_id, col[10]=amount_raw
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			// Pinjam 500000 for CREW_001
			{"log-1", "2026-03-15", "2026-03-15", "PIUTANG", "SITE_001",
				"Kebun Induk", "PINJAM", "Pinjam", "CREW_001", "Jono", "500000", "500000"},
			// Bayar 200000 for CREW_001
			{"log-2", "2026-03-16", "2026-03-16", "PIUTANG", "SITE_001",
				"Kebun Induk", "BAYAR", "Bayar", "CREW_001", "Jono", "200000", "200000"},
			// Different crew — should be ignored
			{"log-3", "2026-03-16", "2026-03-16", "PIUTANG", "SITE_001",
				"Kebun Induk", "PINJAM", "Pinjam", "CREW_002", "Slamet", "100000", "100000"},
		},
	}

	svc := service.NewMasterDataService(mock)
	balance, err := svc.GetCrewBalance(context.Background(), "CREW_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 500000 - 200000 = 300000
	if balance != 300000 {
		t.Errorf("expected balance 300000, got %d", balance)
	}
}

func TestGetCrewBalance_SheetError(t *testing.T) {
	mock := &mockSheetsClient{readErr: errors.New("sheets error")}
	svc := service.NewMasterDataService(mock)
	_, err := svc.GetCrewBalance(context.Background(), "CREW_001")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetListSemprot_ReturnsCorrectEntries(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"log-1", "2026-03-15", "2026-03-15", "OPERASIONAL", "SITE_001",
				"Kebun Induk", "CAT_SEMPROT", "Semprot Herbisida", "CREW_001", "Jono", "150000", "150000", "0", "0", "0", "0", "Semprot blok A"},
			{"log-2", "2026-03-16", "2026-03-16", "OPERASIONAL", "SITE_001",
				"Kebun Induk", "CAT_PUPUK", "Pupuk NPK", "CREW_001", "Jono", "300000", "300000"},
			{"log-3", "2026-03-17", "2026-03-17", "OPERASIONAL", "SITE_002",
				"Kebun Plasma", "CAT_SEMPROT", "Semprot", "CREW_002", "Slamet", "200000", "200000"},
		},
	}

	svc := service.NewMasterDataService(mock)
	list, err := svc.GetListSemprot(context.Background(), "SITE_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 semprot entry for SITE_001, got %d", len(list))
	}
	if list[0].CrewName != "Jono" || list[0].Amount != 150000 || list[0].Notes != "Semprot blok A" {
		t.Errorf("unexpected semprot entry: %+v", list[0])
	}
}

func TestGetListSemprot_SheetError(t *testing.T) {
	mock := &mockSheetsClient{readErr: errors.New("sheets error")}
	svc := service.NewMasterDataService(mock)
	_, err := svc.GetListSemprot(context.Background(), "SITE_001")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetListPanen_ParsesUnitPriceAndFallback(t *testing.T) {
	thisYearDate := fmt.Sprintf("%d-05-10", time.Now().Year())
	lastYearDate := fmt.Sprintf("%d-05-10", time.Now().Year()-1)

	mock := &mockSheetsClient{
		readData: [][]interface{}{
			// Panen this year with explicit UnitPrice (2500)
			{"log-panen-1", "ts", thisYearDate, "PANEN", "SITE_001", "Kebun Induk", "CAT_PANEN", "Panen", "CREW_001", "Budi", "2500000", "2200000", "1000", "2500", "200000", "100000", "Blok A"},
			// Panen this year with UnitPrice 0, should fallback to amountRaw / weight (3000000 / 1200 = 2500)
			{"log-panen-2", "ts", thisYearDate, "PANEN", "SITE_001", "Kebun Induk", "CAT_PANEN", "Panen", "CREW_002", "Slamet", "3000000", "2700000", "1200", "0", "200000", "100000", "Blok B"},
			// Panen last year
			{"log-panen-3", "ts", lastYearDate, "PANEN", "SITE_001", "Kebun Induk", "CAT_PANEN", "Panen", "CREW_001", "Budi", "1000000", "900000", "500", "2000", "50000", "50000", "Blok C"},
			// Different site
			{"log-panen-4", "ts", thisYearDate, "PANEN", "SITE_002", "Kebun Plasma", "CAT_PANEN", "Panen", "CREW_001", "Budi", "5000000", "4500000", "2000", "2500", "300000", "200000", "Blok D"},
		},
	}

	svc := service.NewMasterDataService(mock)

	// Test this year (offset = 0)
	panenThisYear, err := svc.GetListPanen(context.Background(), "SITE_001", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(panenThisYear) != 2 {
		t.Fatalf("expected 2 panen entries for SITE_001 this year, got %d", len(panenThisYear))
	}
	if panenThisYear[0].UnitPrice != 2500 || panenThisYear[0].Weight != 1000 {
		t.Errorf("panen 1: expected UnitPrice 2500, Weight 1000, got %+v", panenThisYear[0])
	}
	if panenThisYear[1].UnitPrice != 2500 || panenThisYear[1].Weight != 1200 {
		t.Errorf("panen 2 fallback: expected UnitPrice 2500, got %+v", panenThisYear[1])
	}

	// Test last year (offset = -1)
	panenLastYear, err := svc.GetListPanen(context.Background(), "SITE_001", -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(panenLastYear) != 1 {
		t.Fatalf("expected 1 panen entry for last year, got %d", len(panenLastYear))
	}
	if panenLastYear[0].UnitPrice != 2000 {
		t.Errorf("panen last year: expected UnitPrice 2000, got %d", panenLastYear[0].UnitPrice)
	}
}

func TestUIService_BuildModeSelectionModal_HasGroupedBlocks(t *testing.T) {
	uis := service.NewUIService()
	state := model.TransactionState{
		SiteID:   "SITE_001",
		SiteName: "Kebun Induk",
	}

	modal := uis.BuildModeSelectionModal(state)
	if modal.CallbackID != "mode_selection_modal" {
		t.Errorf("expected callback id mode_selection_modal, got %s", modal.CallbackID)
	}

	// Ensure blocks are generated
	if len(modal.Blocks.BlockSet) < 10 {
		t.Errorf("expected at least 10 blocks for grouped menu, got %d", len(modal.Blocks.BlockSet))
	}
}

func TestUIService_BuildListPanenModal_ShowsUnitPrice(t *testing.T) {
	uis := service.NewUIService()
	panenList := []model.LogEntry{
		{
			EventDate:   time.Now(),
			CrewName:    "Budi, Slamet",
			Weight:      1500,
			UnitPrice:   2450,
			AmountFinal: 3200000,
			Notes:       "Timbangan luar",
		},
	}

	modal := uis.BuildListPanenModal("Kebun Induk", time.Now().Year(), panenList)
	if len(modal.Blocks.BlockSet) == 0 {
		t.Fatal("expected blocks in modal, got 0")
	}
	modalJSON, _ := json.Marshal(modal)
	if !strings.Contains(string(modalJSON), "Rp2.450") {
		t.Errorf("expected modal to contain formatted unit price Rp2.450, got %s", string(modalJSON))
	}

	msg := uis.BuildListPanenMessage("Kebun Induk", time.Now().Year(), panenList)
	if len(msg.Blocks.BlockSet) == 0 {
		t.Fatal("expected blocks in message, got 0")
	}
	msgJSON, _ := json.Marshal(msg)
	if !strings.Contains(string(msgJSON), "Rp2.450") {
		t.Errorf("expected message to contain formatted unit price Rp2.450, got %s", string(msgJSON))
	}
}

func TestGetCrewDebtSummaries_CalculatesBalancesAndDates(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Crew!A2:E": {
				{"CREW_001", "Jono", "Pemanen", "SITE_001", "ACTIVE"},
				{"CREW_002", "Slamet", "Supir", "SITE_001", "ACTIVE"},
			},
			"X_LOG!A2:L": {
				{"log-1", "ts", "2026-03-01", "PIUTANG", "SITE_001", "Kebun Induk", "PINJAM", "Pinjam", "CREW_001", "Jono", "500000", "500000"},
				{"log-2", "ts", "2026-03-10", "PIUTANG", "SITE_001", "Kebun Induk", "PINJAM", "Pinjam", "CREW_001", "Jono", "300000", "800000"},
				{"log-3", "ts", "2026-03-15", "PIUTANG", "SITE_001", "Kebun Induk", "BAYAR", "Bayar", "CREW_001", "Jono", "200000", "600000"},
			},
		},
	}

	svc := service.NewMasterDataService(mock)
	summaries, err := svc.GetCrewDebtSummaries(context.Background(), "SITE_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	jono := summaries[0]
	if jono.CrewName != "Jono" {
		t.Errorf("expected first crew to be Jono, got %s", jono.CrewName)
	}
	if jono.TotalPinjam != 800000 || jono.TotalBayar != 200000 || jono.OutstandingDebt != 600000 {
		t.Errorf("unexpected sums for Jono: pinjam=%d bayar=%d outst=%d", jono.TotalPinjam, jono.TotalBayar, jono.OutstandingDebt)
	}
	if jono.LastPinjamDate == nil || jono.LastPinjamDate.Format("2006-01-02") != "2026-03-10" {
		t.Errorf("expected Jono last pinjam date 2026-03-10, got %v", jono.LastPinjamDate)
	}
	if jono.LastBayarDate == nil || jono.LastBayarDate.Format("2006-01-02") != "2026-03-15" {
		t.Errorf("expected Jono last bayar date 2026-03-15, got %v", jono.LastBayarDate)
	}

	slamet := summaries[1]
	if slamet.OutstandingDebt != 0 || slamet.LastPinjamDate != nil || slamet.LastBayarDate != nil {
		t.Errorf("expected zero debt and nil dates for Slamet, got %+v", slamet)
	}
}

func TestGetListHutang_ReturnsCorrectEntries(t *testing.T) {
	mock := &mockSheetsClient{
		readData: [][]interface{}{
			{"log-1", "ts", "2026-03-05", "PIUTANG", "SITE_001", "Kebun Induk", "PINJAM", "Pinjam", "CREW_001", "Jono", "500000", "500000", "", "", "", "", "Kasbon darurat"},
			{"log-2", "ts", "2026-03-12", "PIUTANG", "SITE_001", "Kebun Induk", "BAYAR", "Bayar", "CREW_001", "Jono", "200000", "300000", "", "", "", "", "Potong panen"},
			{"log-3", "ts", "2026-03-15", "OPERASIONAL", "SITE_001", "Kebun Induk", "CAT_PUPUK", "Pupuk", "CREW_002", "Slamet", "100000", "100000"},
			{"log-4", "ts", "2026-03-20", "PIUTANG", "SITE_002", "Kebun Plasma", "PINJAM", "Pinjam", "CREW_003", "Anto", "250000", "250000"},
		},
	}

	svc := service.NewMasterDataService(mock)
	list, err := svc.GetListHutang(context.Background(), "SITE_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 piutang entries for SITE_001, got %d", len(list))
	}

	if list[0].CrewName != "Jono" || list[0].CategoryID != "PINJAM" || list[0].Amount != 500000 || list[0].Balance != 500000 || list[0].Notes != "Kasbon darurat" {
		t.Errorf("unexpected entry 0: %+v", list[0])
	}
	if list[1].CrewName != "Jono" || list[1].CategoryID != "BAYAR" || list[1].Amount != 200000 || list[1].Balance != 300000 || list[1].Notes != "Potong panen" {
		t.Errorf("unexpected entry 1: %+v", list[1])
	}
}

func TestUIService_BuildCrewDebtModal_ShowsDatesAndAction(t *testing.T) {
	uis := service.NewUIService()
	pinjamDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	bayarDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	summaries := []model.CrewDebtSummary{
		{
			CrewID:          "CREW_001",
			CrewName:        "Jono",
			Role:            "Pemanen",
			TotalPinjam:     800000,
			TotalBayar:      200000,
			OutstandingDebt: 600000,
			LastPinjamDate:  &pinjamDate,
			LastBayarDate:   &bayarDate,
		},
	}

	modal := uis.BuildCrewDebtModal("Kebun Induk", summaries)
	modalJSON, _ := json.Marshal(modal)
	modalStr := string(modalJSON)

	if !strings.Contains(modalStr, "10 Mar 2026") {
		t.Errorf("expected modal to show pinjam date 10 Mar 2026, got %s", modalStr)
	}
	if !strings.Contains(modalStr, "15 Mar 2026") {
		t.Errorf("expected modal to show bayar date 15 Mar 2026, got %s", modalStr)
	}
	if !strings.Contains(modalStr, "view_list_hutang_lengkap") {
		t.Errorf("expected modal to contain action button view_list_hutang_lengkap, got %s", modalStr)
	}

	msg := uis.BuildCrewDebtMessage("Kebun Induk", summaries)
	msgJSON, _ := json.Marshal(msg)
	msgStr := string(msgJSON)
	if !strings.Contains(msgStr, "10 Mar 2026") {
		t.Errorf("expected message to show pinjam date 10 Mar 2026, got %s", msgStr)
	}
}

func TestUIService_BuildListHutangModalAndMessage(t *testing.T) {
	uis := service.NewUIService()
	eventDate := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)

	entries := []model.HutangLogEntry{
		{
			EventDate:  eventDate,
			CrewName:   "Jono",
			CategoryID: "PINJAM",
			Amount:     500000,
			Balance:    500000,
			Notes:      "Kasbon berobat",
		},
		{
			EventDate:  eventDate.AddDate(0, 0, 2),
			CrewName:   "Jono",
			CategoryID: "BAYAR",
			Amount:     200000,
			Balance:    300000,
			Notes:      "Cicilan pertama",
		},
	}

	modal := uis.BuildListHutangModal("Kebun Induk", entries)
	modalJSON, _ := json.Marshal(modal)
	modalStr := string(modalJSON)

	if !strings.Contains(modalStr, "PINJAM") || !strings.Contains(modalStr, "BAYAR") {
		t.Errorf("expected modal to contain PINJAM and BAYAR, got %s", modalStr)
	}
	if !strings.Contains(modalStr, "Kasbon berobat") || !strings.Contains(modalStr, "Cicilan pertama") {
		t.Errorf("expected modal to contain notes, got %s", modalStr)
	}
	if !strings.Contains(modalStr, "Sisa Utang: Rp300.000") {
		t.Errorf("expected modal to display total sisa utang Rp300.000, got %s", modalStr)
	}

	msg := uis.BuildListHutangMessage("Kebun Induk", entries)
	msgJSON, _ := json.Marshal(msg)
	msgStr := string(msgJSON)
	if !strings.Contains(msgStr, "LIST LENGKAP KASBON") || !strings.Contains(msgStr, "PEMBAYARAN") {
		t.Errorf("expected message header, got %s", msgStr)
	}
	if !strings.Contains(msgStr, "Sisa Utang: Rp300.000") {
		t.Errorf("expected message to show sisa utang Rp300.000, got %s", msgStr)
	}
}

func TestUIService_BuildModeSelectionModal_HasListHutangButton(t *testing.T) {
	uis := service.NewUIService()
	state := model.TransactionState{SiteID: "SITE_001", SiteName: "Kebun Induk"}
	modal := uis.BuildModeSelectionModal(state)
	modalJSON, _ := json.Marshal(modal)
	if !strings.Contains(string(modalJSON), "view_list_hutang_lengkap") {
		t.Errorf("expected mode selection modal to contain view_list_hutang_lengkap button, got %s", string(modalJSON))
	}
}

func TestGetSiteReport_ComprehensiveAndBabyLanguage(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_001", "Kebun Induk", "Kalimantan", "ACTIVE", "100000000"},
			},
			"X_LOG!A2:Q": {
				// Panen 1: 2.000 kg, Gross 5.000.000, Upah 1.000.000, Trans 500.000
				{"LOG_01", "2026-01-10T10:00:00Z", "2026-01-10", "PANEN", "SITE_001", "Kebun Induk", "CAT_PANEN", "Panen", "C_01", "Budi", "5000000", "3500000", "2000", "2500", "1000000", "500000", "Panen Putaran 1"},
				// Panen 2: 3.000 kg, Gross 7.500.000, Upah 1.500.000, Trans 750.000
				{"LOG_02", "2026-02-10T10:00:00Z", "2026-02-10", "PANEN", "SITE_001", "Kebun Induk", "CAT_PANEN", "Panen", "C_01", "Budi", "7500000", "5250000", "3000", "2500", "1500000", "750000", "Panen Putaran 2"},
				// Operasional: Pupuk
				{"LOG_03", "2026-01-15T10:00:00Z", "2026-01-15", "OPERASIONAL", "SITE_001", "Kebun Induk", "CAT_PUPUK", "Pupuk", "C_02", "Jono", "2000000", "-2000000", "", "", "", "", "Beli pupuk NPK"},
				// Operasional: Semprot
				{"LOG_04", "2026-01-20T10:00:00Z", "2026-01-20", "OPERASIONAL", "SITE_001", "Kebun Induk", "CAT_SEMPROT", "Semprot", "C_02", "Jono", "1000000", "-1000000", "", "", "", "", "Beli racun herbisida"},
				// Operasional: Lainnya
				{"LOG_05", "2026-02-01T10:00:00Z", "2026-02-01", "OPERASIONAL", "SITE_001", "Kebun Induk", "CAT_LAIN", "Perawatan", "C_02", "Jono", "500000", "-500000", "", "", "", "", "Perbaikan jalan kebun"},
				// Piutang: Pinjam & Bayar
				{"LOG_06", "2026-01-05T10:00:00Z", "2026-01-05", "PIUTANG", "SITE_001", "Kebun Induk", "PINJAM", "Kasbon", "C_01", "Budi", "1000000", "-1000000", "", "", "", "", "Kasbon awal"},
				{"LOG_07", "2026-02-05T10:00:00Z", "2026-02-05", "PIUTANG", "SITE_001", "Kebun Induk", "BAYAR", "Bayar Kasbon", "C_01", "Budi", "400000", "400000", "", "", "", "", "Cicil kasbon"},
				// Investasi: Tambah Modal
				{"LOG_08", "2026-01-01T10:00:00Z", "2026-01-01", "INVESTASI", "SITE_001", "Kebun Induk", "CAT_INV", "Investasi", "", "", "10000000", "-10000000", "", "", "", "", "Beli bibit sisipan"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Validasi Panen
	if report.HarvestCount != 2 {
		t.Errorf("expected HarvestCount 2, got %d", report.HarvestCount)
	}
	if report.TotalWeight != 5000 {
		t.Errorf("expected TotalWeight 5000, got %d", report.TotalWeight)
	}
	if report.AvgHarvestWeight != 2500 {
		t.Errorf("expected AvgHarvestWeight 2500, got %d", report.AvgHarvestWeight)
	}
	if report.GrossIncome != 12500000 {
		t.Errorf("expected GrossIncome 12500000, got %d", report.GrossIncome)
	}
	if report.AvgPricePerKg != 2500 {
		t.Errorf("expected AvgPricePerKg 2500, got %d", report.AvgPricePerKg)
	}

	// 2. Validasi Biaya Operasional Terurai
	if report.TotalUpah != 2500000 {
		t.Errorf("expected TotalUpah 2500000, got %d", report.TotalUpah)
	}
	if report.TotalTransport != 1250000 {
		t.Errorf("expected TotalTransport 1250000, got %d", report.TotalTransport)
	}
	if report.TotalPupukCost != 2000000 {
		t.Errorf("expected TotalPupukCost 2000000, got %d", report.TotalPupukCost)
	}
	if report.TotalSemprotCost != 1000000 {
		t.Errorf("expected TotalSemprotCost 1000000, got %d", report.TotalSemprotCost)
	}
	if report.TotalOtherOpsCost != 500000 {
		t.Errorf("expected TotalOtherOpsCost 500000, got %d", report.TotalOtherOpsCost)
	}
	expectedOpex := int64(2500000 + 1250000 + 2000000 + 1000000 + 500000)
	if report.OperationalCost != expectedOpex {
		t.Errorf("expected OperationalCost %d, got %d", expectedOpex, report.OperationalCost)
	}
	if report.CostPerKg != expectedOpex/5000 {
		t.Errorf("expected CostPerKg %d, got %d", expectedOpex/5000, report.CostPerKg)
	}

	// 3. Validasi Keuntungan Bersih & Modal
	expectedNet := 12500000 - expectedOpex // 12.500.000 - 7.250.000 = 5.250.000
	if report.NetProfit != expectedNet {
		t.Errorf("expected NetProfit %d, got %d", expectedNet, report.NetProfit)
	}
	if report.ProfitPerKg != expectedNet/5000 {
		t.Errorf("expected ProfitPerKg %d, got %d", expectedNet/5000, report.ProfitPerKg)
	}
	// Baris INVESTASI di X_LOG hanya audit trail: modal murni dari Sites col E.
	expectedTarget := int64(100000000) // 100.000.000
	if report.TargetModal != expectedTarget {
		t.Errorf("expected TargetModal %d, got %d", expectedTarget, report.TargetModal)
	}
	expectedRemaining := expectedTarget - expectedNet
	if report.RemainingCapital != expectedRemaining {
		t.Errorf("expected RemainingCapital %d, got %d", expectedRemaining, report.RemainingCapital)
	}

	// 4. Validasi Kasbon Pegawai
	if report.TotalPinjam != 1000000 || report.TotalBayar != 400000 || report.OutstandingDebt != 600000 {
		t.Errorf("kasbon mismatch: pinjam=%d bayar=%d outst=%d", report.TotalPinjam, report.TotalBayar, report.OutstandingDebt)
	}

	// 5. Validasi Status Kesehatan & Narasi Bahasa Bayi
	if !strings.Contains(report.HealthStatus, "SEHAT") {
		t.Errorf("expected health status to indicate SEHAT, got %s", report.HealthStatus)
	}
	if !strings.Contains(report.SummaryNarration, "5.000 Kg") || !strings.Contains(report.SummaryNarration, "5.250.000") {
		t.Errorf("expected summary narration to contain key details, got %s", report.SummaryNarration)
	}

	// 6. Validasi UI Modal & Message Generation
	uis := service.NewUIService()
	modal := uis.BuildReportModal("Kebun Induk", report)
	modalJSON, _ := json.Marshal(modal)
	modalStr := string(modalJSON)

	if !strings.Contains(modalStr, "1. HASIL PANEN SAWIT") || !strings.Contains(modalStr, "2. PENGELUARAN") {
		t.Errorf("expected modal to contain structured numbered sections, got %s", modalStr)
	}
	if !strings.Contains(modalStr, "5.250.000") || !strings.Contains(modalStr, "Untung Bersih (Kantong)") {
		t.Errorf("expected modal to contain Untung Bersih, got %s", modalStr)
	}

	msg := uis.BuildReportMessage("Kebun Induk", report)
	msgJSON, _ := json.Marshal(msg)
	msgStr := string(msgJSON)
	if !strings.Contains(msgStr, "REKAP PERFORMA KEBUN") || !strings.Contains(msgStr, "Kesimpulan Ringkas") {
		t.Errorf("expected message to contain executive summary, got %s", msgStr)
	}
}

func TestGetSiteReport_BEPTercapai(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_002", "Kebun Berjaya", "Sumatera", "ACTIVE", "10000000"},
			},
			"X_LOG!A2:Q": {
				{"LOG_01", "2026-01-01T10:00:00Z", "2026-01-01", "PANEN", "SITE_002", "Kebun Berjaya", "CAT_PANEN", "Panen", "", "", "25000000", "20000000", "10000", "2500", "3000000", "2000000", "Panen Raya"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.RemainingCapital != 0 {
		t.Errorf("expected RemainingCapital 0 when BEP reached, got %d", report.RemainingCapital)
	}
	if !strings.Contains(report.HealthStatus, "BALIK MODAL PENUH") {
		t.Errorf("expected health status to indicate BALIK MODAL PENUH, got %s", report.HealthStatus)
	}
	if !strings.Contains(report.SummaryNarration, "kembali 100%") {
		t.Errorf("expected summary narration to mention kembali 100%%, got %s", report.SummaryNarration)
	}
}

func TestGetSiteReport_Defisit(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_003", "Kebun Baru Beli", "Riau", "ACTIVE", "50000000"},
			},
			"X_LOG!A2:Q": {
				// Panen kecil
				{"LOG_01", "2026-01-10T10:00:00Z", "2026-01-10", "PANEN", "SITE_003", "Kebun Baru Beli", "CAT_PANEN", "Panen", "", "", "1000000", "500000", "500", "2000", "300000", "200000", "Panen Awal"},
				// Operasional besar
				{"LOG_02", "2026-01-15T10:00:00Z", "2026-01-15", "OPERASIONAL", "SITE_003", "Kebun Baru Beli", "CAT_PUPUK", "Pupuk", "", "", "5000000", "-5000000", "", "", "", "", "Pupuk Awal"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.NetProfit >= 0 {
		t.Errorf("expected negative NetProfit (deficit), got %d", report.NetProfit)
	}
	if !strings.Contains(report.HealthStatus, "DEFISIT") {
		t.Errorf("expected health status to indicate DEFISIT, got %s", report.HealthStatus)
	}
	if !strings.Contains(report.SummaryNarration, "defisit") {
		t.Errorf("expected summary narration to mention defisit, got %s", report.SummaryNarration)
	}
}

func TestGetSiteReport_BEPTrailingWindow(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_004", "Kebun Terkini", "Kalimantan", "ACTIVE", "100000000"},
			},
			"X_LOG!A2:Q": {
				// Panen lama (di luar window 90 hari): profit 5.000.000
				{"LOG_01", "2026-01-01T10:00:00Z", "2026-01-01", "PANEN", "SITE_004", "Kebun Terkini", "CAT_PANEN", "Panen", "", "", "6000000", "5000000", "1000", "6000", "800000", "200000", "Panen Januari"},
				// Panen baru (dalam window): profit 5.000.000
				{"LOG_02", "2026-08-20T10:00:00Z", "2026-08-20", "PANEN", "SITE_004", "Kebun Terkini", "CAT_PANEN", "Panen", "", "", "6000000", "5000000", "1000", "6000", "800000", "200000", "Panen Agustus"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.NetProfit != 10000000 {
		t.Errorf("expected NetProfit 10000000, got %d", report.NetProfit)
	}
	// Window 90 hari hanya memuat panen Agustus (profit 5jt / 90 hari),
	// bukan seluruh rentang Jan-Agu (10jt / 231 hari).
	if !strings.Contains(report.BEPProjection, "4.4 tahun") {
		t.Errorf("expected BEP projection based on trailing window (~4.4 tahun), got %s", report.BEPProjection)
	}
}

func TestGetSiteReport_BEPWindowFallback(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_005", "Kebun Fallback", "Sumatera", "ACTIVE", "6000000"},
			},
			"X_LOG!A2:Q": {
				// Satu-satunya panen di luar window 90 hari → fallback per periode penuh
				{"LOG_01", "2026-01-01T10:00:00Z", "2026-01-01", "PANEN", "SITE_005", "Kebun Fallback", "CAT_PANEN", "Panen", "", "", "4000000", "3000000", "1000", "3000", "700000", "300000", "Panen Awal"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.NetProfit != 3000000 {
		t.Errorf("expected NetProfit 3000000, got %d", report.NetProfit)
	}
	// Remaining 3.000.000 / (3.000.000 per 30 hari) -> ~1,0 bulan
	if !strings.Contains(report.BEPProjection, "1.0 bulan") {
		t.Errorf("expected BEP fallback projection ~1.0 bulan, got %s", report.BEPProjection)
	}
}

func TestGetSiteReport_BEPCapNinetyNineYears(t *testing.T) {
	mock := &mockSheetsClient{
		readRangeMap: map[string][][]interface{}{
			"Sites!A2:E": {
				{"SITE_006", "Kebun Sangat Besar", "Kalimantan", "ACTIVE", "100000000000"},
			},
			"X_LOG!A2:Q": {
				{"LOG_01", "2026-08-20T10:00:00Z", "2026-08-20", "PANEN", "SITE_006", "Kebun Sangat Besar", "CAT_PANEN", "Panen", "", "", "1000000", "0", "1000", "1000", "0", "0", "Panen Kecil"},
			},
		},
	}

	mds := service.NewMasterDataService(mock)
	report, err := mds.GetSiteReport(context.Background(), "SITE_006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(report.BEPProjection, "> 99 tahun") {
		t.Errorf("expected BEP projection capped at > 99 tahun, got %s", report.BEPProjection)
	}
}
