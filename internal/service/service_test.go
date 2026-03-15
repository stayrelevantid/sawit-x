package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/indragiri/sawit-x/internal/model"
	"github.com/indragiri/sawit-x/internal/service"
)

// mockSheetsClient is a fake SheetsClient for testing without real GCP calls.
type mockSheetsClient struct {
	readData [][]interface{}
	readErr  error
	appendedRows [][]interface{}
	appendErr    error
}

func (m *mockSheetsClient) ReadSpreadsheet(readRange string) ([][]interface{}, error) {
	return m.readData, m.readErr
}

func (m *mockSheetsClient) AppendRow(sheetName string, row []interface{}) error {
	m.appendedRows = append(m.appendedRows, row)
	return m.appendErr
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
