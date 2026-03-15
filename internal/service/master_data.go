package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/indragiri/sawit-x/internal/client"
	"github.com/indragiri/sawit-x/internal/model"
)

type MasterDataService struct {
	sheetsClient client.SheetsReader
}

func NewMasterDataService(sheetsClient client.SheetsReader) *MasterDataService {
	return &MasterDataService{
		sheetsClient: sheetsClient,
	}
}

func (s *MasterDataService) GetActiveSites(ctx context.Context) ([]model.Site, error) {
	// X_MASTER tab: Sites — columns: id, name, location, status, target_modal
	rows, err := s.sheetsClient.ReadSpreadsheet("Sites!A2:E")
	if err != nil {
		return nil, err
	}

	var sites []model.Site
	for i, row := range rows {
		if len(row) < 2 { // Min id and name
			continue
		}
		
		id := fmt.Sprintf("%v", row[0])
		name := fmt.Sprintf("%v", row[1])
		
		status := "INACTIVE"
		if len(row) >= 4 {
			status = fmt.Sprintf("%v", row[3])
		}
		
		if status != "ACTIVE" {
			continue
		}

		var targetModal int64
		if len(row) >= 5 {
			val := fmt.Sprintf("%v", row[4])
			val = strings.ReplaceAll(val, ".", "")
			val = strings.ReplaceAll(val, ",", "")
			targetModal, _ = strconv.ParseInt(val, 10, 64)
		}

		log.Printf("[MASTER] Loaded active site: %s (%s), Target: %d", name, id, targetModal)

		sites = append(sites, model.Site{
			ID:          id,
			Name:        name,
			Location:    func() string { if len(row) >= 3 { return fmt.Sprintf("%v", row[2]) }; return "" }(),
			Status:      status,
			TargetModal: targetModal,
		})
		_ = i // future use
	}
	return sites, nil
}

func (s *MasterDataService) GetSiteByID(ctx context.Context, siteID string) (model.Site, error) {
	sites, err := s.GetActiveSites(ctx)
	if err != nil {
		return model.Site{}, err
	}
	for _, site := range sites {
		if site.ID == siteID {
			return site, nil
		}
	}
	return model.Site{}, fmt.Errorf("site %s not found", siteID)
}

// UpdateSiteTarget updates the TargetModal for a specific site in the Sites sheet.
func (s *MasterDataService) UpdateSiteTarget(ctx context.Context, siteID string, target int64) error {
	log.Printf("[MASTER] Updating target for site %s to %d", siteID, target)
	rows, err := s.sheetsClient.ReadSpreadsheet("Sites!A2:A")
	if err != nil {
		log.Printf("[MASTER] Error reading Sites sheet: %v", err)
		return err
	}

	for i, row := range rows {
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == siteID {
			// Row index in sheet is i+2 (because of A2 start)
			cellRange := fmt.Sprintf("Sites!E%d", i+2)
			log.Printf("[MASTER] Found site %s at row %d. Updating range %s", siteID, i+2, cellRange)
			err := s.sheetsClient.UpdateCell(cellRange, target)
			if err != nil {
				log.Printf("[MASTER] Error updating cell %s: %v", cellRange, err)
			}
			return err
		}
	}

	log.Printf("[MASTER] Site %s not found in A2:A range", siteID)
	return fmt.Errorf("site %s not found", siteID)
}

func (s *MasterDataService) GetActiveCategories(ctx context.Context) ([]model.Category, error) {
	return s.GetCategoriesByType(ctx, "")
}

func (s *MasterDataService) GetCategoriesByType(ctx context.Context, catType string) ([]model.Category, error) {
	// X_MASTER tab: Categories — columns: id, name, type, multiplier_enabled, status
	rows, err := s.sheetsClient.ReadSpreadsheet("Categories!A2:E")
	if err != nil {
		return nil, err
	}

	var categories []model.Category
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		status := fmt.Sprintf("%v", row[4])
		if status != "ACTIVE" {
			continue
		}

		rowType := fmt.Sprintf("%v", row[2])
		if catType != "" && rowType != catType {
			continue
		}

		multiplierEnabled := fmt.Sprintf("%v", row[3]) == "TRUE"

		categories = append(categories, model.Category{
			ID:                fmt.Sprintf("%v", row[0]),
			Name:              fmt.Sprintf("%v", row[1]),
			Type:              rowType,
			MultiplierEnabled: multiplierEnabled,
			Status:            status,
		})
	}
	return categories, nil
}

func (s *MasterDataService) GetActiveCrew(ctx context.Context) ([]model.Crew, error) {
	// X_MASTER tab: Crew — columns: id, name, role, site_id, status
	rows, err := s.sheetsClient.ReadSpreadsheet("Crew!A2:E")
	if err != nil {
		return nil, err
	}

	var crew []model.Crew
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		status := fmt.Sprintf("%v", row[4])
		if status != "ACTIVE" {
			continue
		}

		crew = append(crew, model.Crew{
			ID:     fmt.Sprintf("%v", row[0]),
			Name:   fmt.Sprintf("%v", row[1]),
			Role:   fmt.Sprintf("%v", row[2]),
			SiteID: fmt.Sprintf("%v", row[3]),
			Status: status,
		})
	}
	return crew, nil
}

// GetCrewBalance reads X_LOG and calculates the running debt balance for a crew member.
// Positive balance means the crew member has an outstanding debt (Pinjam).
// Negative balance means overpaid (should not happen normally).
// Columns: [0]=log_id, [1]=timestamp, [2]=event_date, [3]=module_type, [4]=site_id,
//
//	[5]=site_name, [6]=category_id, [7]=category_name, [8]=crew_id, [9]=crew_name,
//	[10]=amount_raw, [11]=amount_final, ...
func (s *MasterDataService) GetCrewBalance(ctx context.Context, crewID string) (int64, error) {
	rows, err := s.sheetsClient.ReadSpreadsheet("X_LOG!A2:L")
	if err != nil {
		return 0, err
	}

	var balance int64
	for _, row := range rows {
		if len(row) < 12 {
			continue
		}
		moduleType := fmt.Sprintf("%v", row[3])
		rowCrewID := fmt.Sprintf("%v", row[8])
		if moduleType != "PIUTANG" || rowCrewID != crewID {
			continue
		}
		// category_id holds "PINJAM" or "BAYAR"
		catID := fmt.Sprintf("%v", row[6])
		amount, _ := strconv.ParseInt(fmt.Sprintf("%v", row[10]), 10, 64)
		if catID == "PINJAM" {
			balance += amount
		} else if catID == "BAYAR" {
			balance -= amount
		}
	}
	return balance, nil
}
// GetSiteReport aggregates transaction data for a specific site from X_LOG.
func (s *MasterDataService) GetSiteReport(ctx context.Context, siteID string) (model.SiteReport, error) {
	// 1. Get Target Modal from Sites master
	sites, _ := s.GetActiveSites(ctx)
	var targetModal int64
	for _, site := range sites {
		if site.ID == siteID {
			targetModal = site.TargetModal
			break
		}
	}

	// 2. Fetch all logs for this site
	// Column order: [0]log_id, [1]timestamp, [2]event_date, [3]module_type, [4]site_id, [5]site_name,
	// [10]amount_raw (Gross/Expense), [11]amount_final (Net), [12]weight, [13]unit_price, [14]labor_cost, [15]transport_cost
	rows, err := s.sheetsClient.ReadSpreadsheet("X_LOG!A2:P")
	if err != nil {
		return model.SiteReport{}, err
	}

	var report model.SiteReport
	report.TargetModal = targetModal

	for _, row := range rows {
		if len(row) < 12 {
			continue
		}
		rowSiteID := fmt.Sprintf("%v", row[4])
		if rowSiteID != siteID {
			continue
		}

		moduleType := fmt.Sprintf("%v", row[3])
		amountRaw, _ := strconv.ParseInt(fmt.Sprintf("%v", row[10]), 10, 64)

		switch moduleType {
		case "PANEN":
			if len(row) >= 16 {
				weight, _ := strconv.ParseInt(fmt.Sprintf("%v", row[12]), 10, 64)
				labor, _ := strconv.ParseInt(fmt.Sprintf("%v", row[14]), 10, 64)
				transport, _ := strconv.ParseInt(fmt.Sprintf("%v", row[15]), 10, 64)

				report.TotalWeight += weight
				report.GrossIncome += amountRaw
				report.TotalUpah += labor
				report.TotalTransport += transport
				report.OperationalCost += (labor + transport)
			}
		case "OPERASIONAL":
			report.TotalOperasional += amountRaw
			report.OperationalCost += amountRaw
		case "PIUTANG":
			catID := fmt.Sprintf("%v", row[6])
			if catID == "PINJAM" {
				report.TotalPinjam += amountRaw
			} else if catID == "BAYAR" {
				report.TotalBayar += amountRaw
			}
		case "INVESTASI":
			report.TargetModal += amountRaw
		}
	}

	report.OutstandingDebt = report.TotalPinjam - report.TotalBayar

	report.NetProfit = report.GrossIncome - report.OperationalCost
	report.RemainingCapital = targetModal - report.NetProfit
	if report.RemainingCapital < 0 {
		report.RemainingCapital = 0
	}

	if targetModal > 0 {
		report.ROITracking = (float64(report.NetProfit) / float64(targetModal)) * 100
	}

	// 5. Calculate BEP Projection
	var firstDate, lastDate time.Time
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		rowSiteID := fmt.Sprintf("%v", row[4])
		if rowSiteID != siteID {
			continue
		}
		eventDateRaw := fmt.Sprintf("%v", row[2])
		eventDate, _ := time.Parse("2006-01-02", eventDateRaw)
		if !eventDate.IsZero() {
			if firstDate.IsZero() || eventDate.Before(firstDate) {
				firstDate = eventDate
			}
			if eventDate.After(lastDate) {
				lastDate = eventDate
			}
		}
	}

	if !firstDate.IsZero() && !lastDate.IsZero() {
		// Use months span
		days := lastDate.Sub(firstDate).Hours() / 24
		if days < 30 {
			days = 30 // Minimum 1 month for average
		}
		avgDailyProfit := float64(report.NetProfit) / days
		avgMonthlyProfit := avgDailyProfit * 30.44

		if report.RemainingCapital > 0 && avgMonthlyProfit > 0 {
			monthsRemaining := float64(report.RemainingCapital) / avgMonthlyProfit
			if monthsRemaining > 12 {
				report.BEPProjection = fmt.Sprintf("Estimasi %.1f tahun lagi", monthsRemaining/12)
			} else {
				report.BEPProjection = fmt.Sprintf("Estimasi %.1f bulan lagi", monthsRemaining)
			}
		} else if report.RemainingCapital <= 0 && targetModal > 0 {
			report.BEPProjection = "SUDAH BALIK MODAL (BEP) ✅"
		} else {
			report.BEPProjection = "Data belum mencukupi untuk estimasi"
		}
	} else {
		report.BEPProjection = "Belum ada data transaksi"
	}

	return report, nil
}

// SyncSiteReportToSheet writes the SiteReport to the X_REKAP tab.
func (s *MasterDataService) SyncSiteReportToSheet(ctx context.Context, siteID string, siteName string, report model.SiteReport) error {
	log.Printf("[REKAP] Syncing report for site %s to X_REKAP", siteID)

	// Ensure siteName is present
	if siteName == "" {
		site, err := s.GetSiteByID(ctx, siteID)
		if err == nil {
			siteName = site.Name
		} else {
			siteName = siteID // Fallback to ID if name not found
		}
	}

	// Range: X_REKAP!A2:A -> site_id is in col A
	rows, err := s.sheetsClient.ReadSpreadsheet("X_REKAP!A2:A")
	if err != nil {
		// If sheet doesn't exist or error reading, we'll try to find row index or append
		log.Printf("[REKAP] Warning: Could not read X_REKAP sheet: %v", err)
	}

	rowIndex := -1
	for i, row := range rows {
		if len(row) > 0 && fmt.Sprintf("%v", row[0]) == siteID {
			rowIndex = i + 2
			break
		}
	}

	values := []interface{}{
		siteID,
		siteName,
		report.TotalWeight,
		report.GrossIncome,
		report.OperationalCost,
		report.NetProfit,
		report.TargetModal,
		report.RemainingCapital,
		fmt.Sprintf("%.2f%%", report.ROITracking),
		report.BEPProjection,
		report.TotalPinjam,
		report.TotalBayar,
		report.OutstandingDebt,
		time.Now().Format("2006-01-02 15:04:05"),
	}

	if rowIndex != -1 {
		// Update existing row
		rangeName := fmt.Sprintf("X_REKAP!A%d", rowIndex)
		return s.sheetsClient.UpdateRange(rangeName, [][]interface{}{values})
	}

	// Append new row if not found
	return s.sheetsClient.AppendRow("X_REKAP", values)
}
