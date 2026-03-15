package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const spreadsheetID = "1RBWoR_ZOfPZPKgL4RCnYcvI0xe2FFY8B_qPrxONTvVM"

func main() {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx,
		option.WithCredentialsFile("/Users/muhammad.indragiri/Kerja/sawit-x/se.json"),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	// Tambah baris baru ke Categories (utang & pelunasan)
	newRows := [][]interface{}{
		{"cat_4", "Utang Panen", "UTANG", "FALSE", "ACTIVE"},
		{"cat_5", "Pelunasan Utang", "PELUNASAN", "FALSE", "ACTIVE"},
	}
	vr := &sheets.ValueRange{Values: newRows}
	_, err = srv.Spreadsheets.Values.Append(spreadsheetID, "Categories!A:E", vr).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	fmt.Println("✅ Kategori Utang & Pelunasan berhasil ditambahkan!")
}
