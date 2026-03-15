package service

import (
	"context"
	"fmt"
	"strconv"

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
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		status := fmt.Sprintf("%v", row[3])
		if status != "ACTIVE" {
			continue
		}

		targetModal, _ := strconv.ParseInt(fmt.Sprintf("%v", row[4]), 10, 64)

		sites = append(sites, model.Site{
			ID:          fmt.Sprintf("%v", row[0]),
			Name:        fmt.Sprintf("%v", row[1]),
			Location:    fmt.Sprintf("%v", row[2]),
			Status:      status,
			TargetModal: targetModal,
		})
	}
	return sites, nil
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
				report.OperationalCost += (labor + transport)
			}
		case "OPERASIONAL":
			report.OperationalCost += amountRaw
		}
	}

	report.NetProfit = report.GrossIncome - report.OperationalCost
	if targetModal > 0 {
		report.ROITracking = (float64(report.NetProfit) / float64(targetModal)) * 100
	}

	return report, nil
}
