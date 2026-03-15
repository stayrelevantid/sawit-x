package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsReader is the interface for reading from and writing to Google Sheets.
// Services depend on this interface, not the concrete SheetsClient, enabling testability.
type SheetsReader interface {
	ReadSpreadsheet(readRange string) ([][]interface{}, error)
	AppendRow(sheetName string, row []interface{}) error
	UpdateCell(cellRange string, value interface{}) error
	UpdateRange(rangeName string, values [][]interface{}) error
}

// SheetsClient is the production implementation of SheetsReader using Google Sheets API v4.
type SheetsClient struct {
	Service       *sheets.Service
	SpreadsheetID string
}

// NewSheetsClient creates a new Google Sheets service.
// It supports credentials via:
// 1. GOOGLE_CREDENTIALS_JSON env var (base64-encoded JSON) - preferred for Cloud Functions
// 2. Application Default Credentials (ADC) / GOOGLE_APPLICATION_CREDENTIALS file
func NewSheetsClient(ctx context.Context) (*SheetsClient, error) {
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		return nil, fmt.Errorf("SPREADSHEET_ID is not set")
	}

	var opts []option.ClientOption

	// Prefer base64-encoded JSON credentials from env var (works without file upload)
	if credsB64 := os.Getenv("GOOGLE_CREDENTIALS_JSON"); credsB64 != "" {
		credsJSON, err := base64.StdEncoding.DecodeString(credsB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode GOOGLE_CREDENTIALS_JSON: %v", err)
		}
		opts = append(opts, option.WithCredentialsJSON(credsJSON))
		log.Printf("[SHEETS] Using credentials from GOOGLE_CREDENTIALS_JSON env var")
	} else {
		log.Printf("[SHEETS] Using Application Default Credentials / GOOGLE_APPLICATION_CREDENTIALS")
	}

	opts = append(opts, option.WithScopes(sheets.SpreadsheetsScope))
	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return &SheetsClient{
		Service:       srv,
		SpreadsheetID: spreadsheetID,
	}, nil
}

// ReadSpreadsheet reads a range from the spreadsheet.
func (c *SheetsClient) ReadSpreadsheet(readRange string) ([][]interface{}, error) {
	resp, err := c.Service.Spreadsheets.Values.Get(c.SpreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	return resp.Values, nil
}

// AppendRow appends a row to a sheet.
func (c *SheetsClient) AppendRow(sheetName string, row []interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	_, err := c.Service.Spreadsheets.Values.Append(c.SpreadsheetID, sheetName, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("unable to append row: %v", err)
	}

	return nil
}

// UpdateCell updates a single cell in the spreadsheet.
func (c *SheetsClient) UpdateCell(cellRange string, value interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{value}},
	}

	_, err := c.Service.Spreadsheets.Values.Update(c.SpreadsheetID, cellRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("unable to update cell: %v", err)
	}

	return nil
}

// UpdateRange updates a range of cells in the spreadsheet.
func (c *SheetsClient) UpdateRange(rangeName string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.Service.Spreadsheets.Values.Update(c.SpreadsheetID, rangeName, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("unable to update range: %v", err)
	}

	return nil
}
