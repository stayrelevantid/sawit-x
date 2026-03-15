package service

import (
	"context"
	"time"

	"github.com/indragiri/sawit-x/internal/client"
	"github.com/indragiri/sawit-x/internal/model"
)

type LogService struct {
	sheetsClient client.SheetsReader
}

func NewLogService(sheetsClient client.SheetsReader) *LogService {
	return &LogService{
		sheetsClient: sheetsClient,
	}
}

func (s *LogService) WriteLog(ctx context.Context, entry model.LogEntry) error {
	// Prepare row data as per expanded schema (20 columns).
	// Column order: log_id, timestamp, event_date, module_type, site_id, site_name,
	// category_id, category_name, crew_id, crew_name, amount_raw, amount_final,
	// weight, unit_price, labor_cost, transport_cost, notes,
	// slack_user_id, slack_username, channel_id
	row := []interface{}{
		entry.LogID,
		entry.Timestamp.Format(time.RFC3339),
		entry.EventDate.Format("2006-01-02"),
		entry.ModuleType,
		entry.SiteID,
		entry.SiteName,
		entry.CategoryID,
		entry.CategoryName,
		entry.CrewID,
		entry.CrewName,
		entry.AmountRaw,
		entry.AmountFinal,
		entry.Weight,
		entry.UnitPrice,
		entry.LaborCost,
		entry.TransportCost,
		entry.Notes,
		entry.SlackUserID,
		entry.SlackUsername,
		entry.ChannelID,
	}

	return s.sheetsClient.AppendRow("X_LOG", row)
}
