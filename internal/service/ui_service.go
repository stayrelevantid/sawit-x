package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/indragiri/sawit-x/internal/model"
	"github.com/slack-go/slack"
)

type UIService struct{}

func NewUIService() *UIService {
	return &UIService{}
}

// helper to create plain text objects
func txt(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
}

// helper to create markdown text objects
func md(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
}

// BuildSiteSelectionModal builds the first modal to select a plantation site.
func (s *UIService) BuildSiteSelectionModal(sites []model.Site) slack.ModalViewRequest {
	var siteOptions []*slack.OptionBlockObject
	for _, site := range sites {
		siteOptions = append(siteOptions, slack.NewOptionBlockObject(
			site.ID,
			txt(site.Name),
			txt(site.Location),
		))
	}

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("🌴 SAWIT-X"),
		Close:           txt("Batal"),
		Submit:          txt("Lanjut"),
		CallbackID:      "site_selection_modal",
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt("Pilih Lokasi Kebun")),
				slack.NewInputBlock(
					"site_selection_block",
					txt("Lokasi"),
					nil,
					slack.NewOptionsSelectBlockElement(
						slack.OptTypeStatic,
						txt("Pilih kebun..."),
						"site_id",
						siteOptions...,
					),
				),
			},
		},
	}
}

// BuildModeSelectionModal builds the choice between Recording and Reporting.
func (s *UIService) BuildModeSelectionModal(state model.TransactionState) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("🌴 Pilih Mode"),
		Close:           txt("Batal"),
		CallbackID:      "mode_selection_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Kebun: %s", state.SiteName))),
				slack.NewSectionBlock(md("Apa yang ingin Anda lakukan?"), nil, nil),
				slack.NewActionBlock(
					"mode_selection_block",
					slack.NewButtonBlockElement("mode_pencatatan", "PENCATATAN", txt("✍️ Pencatatan Baru")),
					slack.NewButtonBlockElement("view_report", "REKAP", txt("📊 Lihat Rekap")),
				),
			},
		},
	}
}

// BuildModuleSelectionModal builds the second modal for choosing a module.
func (s *UIService) BuildModuleSelectionModal(state model.TransactionState) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	panenOption := slack.NewOptionBlockObject(model.ModulePanen, txt("🌾 Panen"), txt("Catat hasil panen dan biaya logistik"))
	opsOption := slack.NewOptionBlockObject(model.ModuleOperasional, txt("💰 Operasional"), txt("Catat pengeluaran kebun"))
	piutangOption := slack.NewOptionBlockObject(model.ModulePiutang, txt("📋 Piutang"), txt("Kelola pinjaman pegawai"))
	investasiOption := slack.NewOptionBlockObject(model.ModuleInvestasi, txt("🚀 Investasi"), txt("Catat modal balik / pembelian lahan"))

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("🌴 SAWIT-X"),
		Close:           txt("Batal"),
		Submit:          txt("Pilih"),
		CallbackID:      "module_selection_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Kebun: %s", state.SiteName))),
				slack.NewSectionBlock(md("Pilih jenis pencatatan:"), nil, nil),
				slack.NewInputBlock(
					"module_block",
					txt("Modul"),
					nil,
					slack.NewOptionsSelectBlockElement(
						slack.OptTypeStatic,
						txt("Pilih modul..."),
						"module_type",
						panenOption, opsOption, piutangOption, investasiOption,
					),
				),
			},
		},
	}
}

// BuildInvestasiModal builds the Investasi module modal.
func (s *UIService) BuildInvestasiModal(state model.TransactionState, currentTarget int64) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	today := time.Now().Format("2006-01-02")
	datePicker := slack.NewDatePickerBlockElement("event_date")
	datePicker.InitialDate = today

	amountInput := slack.NewPlainTextInputBlockElement(txt("Contoh: 200000000"), "amount_raw")
	if currentTarget > 0 {
		amountInput.InitialValue = strconv.FormatInt(currentTarget, 10)
	}

	title := "🚀 Set Modal Awal"
	if currentTarget > 0 {
		title = "🚀 Update Investasi"
	}

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt(title),
		Close:           txt("Kembali"),
		Submit:          txt("Simpan"),
		CallbackID:      "investasi_entry_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Investasi — %s", state.SiteName))),
				slack.NewInputBlock("date_block", txt("Tanggal"), nil, datePicker),
				slack.NewInputBlock(
					"amount_block",
					txt("Nominal Modal (Rp)"),
					nil,
					amountInput,
				),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"notes_block",
						txt("Keterangan"),
						nil,
						slack.NewPlainTextInputBlockElement(txt(`Misal: "Beli Lahan 2 Ha"`), "notes"),
					)
					b.Optional = true
					return b
				}(),
			},
		},
	}
}

// BuildPanenModal builds the Panen module modal.
// Fields: Tanggal, Multi-select Pemanen, Berat (Kg), Harga/Kg, Upah Panen, Bensin/Timbang.
func (s *UIService) BuildPanenModal(state model.TransactionState, crew []model.Crew) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	var crewOptions []*slack.OptionBlockObject
	for _, c := range crew {
		crewOptions = append(crewOptions, slack.NewOptionBlockObject(
			c.ID, txt(c.Name), txt(c.Role),
		))
	}

	today := time.Now().Format("2006-01-02")
	datePicker := slack.NewDatePickerBlockElement("event_date")
	datePicker.InitialDate = today

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("🌾 Modul Panen"),
		Close:           txt("Kembali"),
		Submit:          txt("Simpan"),
		CallbackID:      "panen_entry_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Panen — %s", state.SiteName))),
				slack.NewInputBlock("date_block", txt("Tanggal"), nil, datePicker),
				slack.NewInputBlock(
					"crew_block",
					txt("Pemanen"),
					nil,
					slack.NewOptionsMultiSelectBlockElement(
						slack.MultiOptTypeStatic,
						txt("Pilih pemanen..."),
						"crew_id",
						crewOptions...,
					),
				),
				slack.NewInputBlock(
					"weight_block",
					txt("Berat Total (Kg)"),
					nil,
					slack.NewPlainTextInputBlockElement(txt("Contoh: 1250"), "weight"),
				),
				slack.NewInputBlock(
					"unit_price_block",
					txt("Harga per Kg (Rp)"),
					nil,
					slack.NewPlainTextInputBlockElement(txt("Contoh: 2400"), "unit_price"),
				),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"labor_block",
						txt("Upah Panen (Rp)"),
						nil,
						slack.NewPlainTextInputBlockElement(txt("Contoh: 150000"), "labor_cost"),
					)
					b.Optional = true
					return b
				}(),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"transport_block",
						txt("Bensin/Timbang (Rp)"),
						nil,
						slack.NewPlainTextInputBlockElement(txt("Contoh: 50000"), "transport_cost"),
					)
					b.Optional = true
					return b
				}(),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"notes_block",
						txt("Catatan"),
						nil,
						slack.NewPlainTextInputBlockElement(txt("Keterangan tambahan..."), "notes"),
					)
					b.Optional = true
					return b
				}(),
			},
		},
	}
}

// BuildOperasionalModal builds the Operasional module modal.
// Fields: Kategori Biaya, Penanggung Jawab (single), Nominal, Keterangan.
func (s *UIService) BuildOperasionalModal(state model.TransactionState, categories []model.Category, crew []model.Crew) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	var catOptions []*slack.OptionBlockObject
	for _, cat := range categories {
		catOptions = append(catOptions, slack.NewOptionBlockObject(
			cat.ID, txt(cat.Name), txt(cat.Type),
		))
	}

	var crewOptions []*slack.OptionBlockObject
	for _, c := range crew {
		crewOptions = append(crewOptions, slack.NewOptionBlockObject(
			c.ID, txt(c.Name), txt(c.Role),
		))
	}

	today := time.Now().Format("2006-01-02")
	datePicker := slack.NewDatePickerBlockElement("event_date")
	datePicker.InitialDate = today

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("💰 Operasional"),
		Close:           txt("Kembali"),
		Submit:          txt("Simpan"),
		CallbackID:      "operasional_entry_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Operasional — %s", state.SiteName))),
				slack.NewInputBlock("date_block", txt("Tanggal"), nil, datePicker),
				slack.NewInputBlock(
					"category_block",
					txt("Kategori Biaya"),
					nil,
					slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, txt("Pilih kategori..."), "category_id", catOptions...),
				),
				slack.NewInputBlock(
					"crew_block",
					txt("Penanggung Jawab"),
					nil,
					slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, txt("Pilih pegawai..."), "crew_id", crewOptions...),
				),
				slack.NewInputBlock(
					"amount_block",
					txt("Nominal (Rp)"),
					nil,
					slack.NewPlainTextInputBlockElement(txt("Contoh: 200000"), "amount_raw"),
				),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"notes_block",
						txt("Keterangan"),
						nil,
						slack.NewPlainTextInputBlockElement(txt(`Misal: "Beli NPK 12-12-17"`), "notes"),
					)
					b.Optional = true
					return b
				}(),
			},
		},
	}
}

// BuildPiutangCrewSelectModal builds the first Piutang step: choose a crew member.
func (s *UIService) BuildPiutangCrewSelectModal(state model.TransactionState, crew []model.Crew) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	var crewOptions []*slack.OptionBlockObject
	for _, c := range crew {
		crewOptions = append(crewOptions, slack.NewOptionBlockObject(
			c.ID, txt(c.Name), txt(c.Role),
		))
	}

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("📋 Piutang"),
		Close:           txt("Batal"),
		Submit:          txt("Cek Saldo"),
		CallbackID:      "piutang_crew_select_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt("Pilih Pegawai")),
				slack.NewInputBlock(
					"crew_block",
					txt("Nama Pegawai"),
					nil,
					slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, txt("Pilih pegawai..."), "crew_id", crewOptions...),
				),
			},
		},
	}
}

// BuildPiutangActionModal builds the second Piutang step: show balance + Pinjam/Bayar choice + nominal.
func (s *UIService) BuildPiutangActionModal(state model.TransactionState, crewName string, balance int64) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	balanceText := fmt.Sprintf("*%s* — Saldo Piutang: *Rp%s*", crewName, formatRupiah(balance))
	if balance == 0 {
		balanceText = fmt.Sprintf("*%s* — Tidak ada piutang tercatat.", crewName)
	}

	pinjamOption := slack.NewOptionBlockObject("PINJAM", txt("💸 Pinjam"), txt("Tambah pinjaman baru"))
	bayarOption := slack.NewOptionBlockObject("BAYAR", txt("✅ Bayar / Potong"), txt("Kurangi saldo piutang"))

	today := time.Now().Format("2006-01-02")
	datePicker := slack.NewDatePickerBlockElement("event_date")
	datePicker.InitialDate = today

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("📋 Piutang"),
		Close:           txt("Batal"),
		Submit:          txt("Simpan"),
		CallbackID:      "piutang_action_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt("Detail Piutang")),
				slack.NewSectionBlock(md(balanceText), nil, nil),
				slack.NewDividerBlock(),
				slack.NewInputBlock("date_block", txt("Tanggal"), nil, datePicker),
				slack.NewInputBlock(
					"action_block",
					txt("Aksi"),
					nil,
					slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, txt("Pilih aksi..."), "piutang_action", pinjamOption, bayarOption),
				),
				slack.NewInputBlock(
					"amount_block",
					txt("Nominal (Rp)"),
					nil,
					slack.NewPlainTextInputBlockElement(txt("Contoh: 500000"), "amount_raw"),
				),
				func() *slack.InputBlock {
					b := slack.NewInputBlock(
						"notes_block",
						txt("Keterangan"),
						nil,
						slack.NewPlainTextInputBlockElement(txt("Keterangan opsional..."), "notes"),
					)
					b.Optional = true
					return b
				}(),
			},
		},
	}
}

// BuildSuccessResponse returns a message block for successful logging.
func (s *UIService) BuildSuccessResponse(entry model.LogEntry) slack.Message {
	var detail string
	switch entry.ModuleType {
	case model.ModulePanen:
		detail = fmt.Sprintf("*Kebun:* %s\n*Pemanen:* %s\n*Berat:* %d Kg\n*Harga:* Rp%s\n*Perhitungan:*\n> Gross: %d Kg x Rp%s = Rp%s\n> Biaya: Rp%s (Upah) + Rp%s (Bensin/Timbang)\n*Net Profit:* Rp%s",
			entry.SiteName, entry.CrewName, entry.Weight, formatRupiah(entry.UnitPrice),
			entry.Weight, formatRupiah(entry.UnitPrice), formatRupiah(entry.AmountRaw),
			formatRupiah(entry.LaborCost), formatRupiah(entry.TransportCost), formatRupiah(entry.AmountFinal))
	case model.ModuleOperasional:
		detail = fmt.Sprintf("*Kebun:* %s\n*Kategori:* %s\n*PJ:* %s\n*Nominal:* Rp%s\n*Keterangan:* %s",
			entry.SiteName, entry.CategoryName, entry.CrewName, formatRupiah(entry.AmountRaw), entry.Notes)
	case model.ModulePiutang:
		action := entry.CategoryID // PINJAM or BAYAR
		prevBalance := entry.AmountFinal - entry.AmountRaw
		if action == "BAYAR" {
			prevBalance = entry.AmountFinal + entry.AmountRaw
		}
		detail = fmt.Sprintf("*Kebun:* %s\n*Pegawai:* %s\n*Aksi:* %s\n*Nominal:* Rp%s\n*Perhitungan Saldo:*\n> Saldo Awal: Rp%s\n> %s: Rp%s\n*Saldo Akhir:* Rp%s",
			entry.SiteName, entry.CrewName, action, formatRupiah(entry.AmountRaw),
			formatRupiah(prevBalance), action, formatRupiah(entry.AmountRaw), formatRupiah(entry.AmountFinal))
	case model.ModuleInvestasi:
		detail = fmt.Sprintf("*Kebun:* %s\n*Kategori:* %s\n*Nominal:* Rp%s\n*Keterangan:* %s",
			entry.SiteName, entry.CategoryName, formatRupiah(entry.AmountRaw), entry.Notes)
	default:
		detail = fmt.Sprintf("*Kebun:* %s\n*Nominal:* Rp%s", entry.SiteName, formatRupiah(entry.AmountRaw))
	}

	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(
						md(fmt.Sprintf("✅ *Data Berhasil Dicatat!*\n\n%s", detail)),
						nil, nil,
					),
				},
			},
		},
	}
}

// BuildReportModal builds a dashboard-style modal for site performance.
func (s *UIService) BuildReportModal(siteName string, report model.SiteReport) slack.ModalViewRequest {
	return slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: txt("📊 Rekap Performa"),
		Close: txt("Tutup"),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("Kebun: %s", siteName))),
				slack.NewSectionBlock(md("_Agregasi seluruh transaksi yang tercatat di sistem._"), nil, nil),
				slack.NewDividerBlock(),

				// Section 1: Panen
				slack.NewSectionBlock(md("🌾 *PANEN TBS*"), nil, nil),
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					md(fmt.Sprintf("*Total Berat:*\n%d Kg", report.TotalWeight)),
					md(fmt.Sprintf("*Gross Income:*\nRp%s", formatRupiah(report.GrossIncome))),
					md(fmt.Sprintf("*Total Upah:*\nRp%s", formatRupiah(report.TotalUpah))),
					md(fmt.Sprintf("*Total Transport:*\nRp%s", formatRupiah(report.TotalTransport))),
				}, nil),
				slack.NewDividerBlock(),

				// Section 2: Operasional
				slack.NewSectionBlock(md("💰 *OPERASIONAL*"), nil, nil),
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					md(fmt.Sprintf("*Biaya Ops Mandiri:*\nRp%s", formatRupiah(report.TotalOperasional))),
					md(fmt.Sprintf("*Total Pengeluaran:*\nRp%s", formatRupiah(report.OperationalCost))),
				}, nil),
				slack.NewDividerBlock(),

				// Section 3: Piutang (Utang)
				slack.NewSectionBlock(md("📋 *UTANG / PIUTANG*"), nil, nil),
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					md(fmt.Sprintf("*Total Pinjam:*\nRp%s", formatRupiah(report.TotalPinjam))),
					md(fmt.Sprintf("*Total Bayar:*\nRp%s", formatRupiah(report.TotalBayar))),
					md(fmt.Sprintf("*Utang Beredar:*\nRp%s", formatRupiah(report.OutstandingDebt))),
				}, nil),
				slack.NewDividerBlock(),

				// Section 4: Finansial & ROI
				slack.NewSectionBlock(md("📈 *FINANSIAL & ROI*"), nil, nil),
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					md(fmt.Sprintf("*Profit Akumulasi:*\nRp%s", formatRupiah(report.NetProfit))),
					md(fmt.Sprintf("*Sisa Modal:*\nRp%s", formatRupiah(report.RemainingCapital))),
					md(fmt.Sprintf("*ROI Tracking:*\n%.2f%%", report.ROITracking)),
					md(fmt.Sprintf("*BEP Projection:*\n%s", report.BEPProjection)),
				}, nil),
				slack.NewDividerBlock(),

				// Section: Detail Perhitungan
				slack.NewSectionBlock(md("📝 *DETAIL HITUNG*"), nil, nil),
				slack.NewSectionBlock(md(fmt.Sprintf(
					"• *Gross*: Rp%s\n"+
						"• *Biaya*: Rp%s (Panen + Ops)\n"+
						"• *Net*: Rp%s - Rp%s = *Rp%s*\n"+
						"• *ROI*: (Rp%s / Rp%s) × 100 = *%.2f%%*\n"+
						"• *Sisa*: Rp%s - Rp%s = *Rp%s*",
					formatRupiah(report.GrossIncome),
					formatRupiah(report.OperationalCost),
					formatRupiah(report.GrossIncome), formatRupiah(report.OperationalCost), formatRupiah(report.NetProfit),
					formatRupiah(report.NetProfit), formatRupiah(report.TargetModal), report.ROITracking,
					formatRupiah(report.TargetModal), formatRupiah(report.NetProfit), formatRupiah(report.RemainingCapital),
				)), nil, nil),

				slack.NewDividerBlock(),
				slack.NewContextBlock("", md(fmt.Sprintf("_Target Investasi: Rp%s_", formatRupiah(report.TargetModal)))),
			},
		},
	}
}

// BuildReportMessage builds a message-based report for site performance.
func (s *UIService) BuildReportMessage(siteName string, report model.SiteReport) slack.Message {
	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(md(fmt.Sprintf("📊 *REKAP PERFORMA: %s*", siteName)), nil, nil),
					slack.NewDividerBlock(),
					slack.NewSectionBlock(md(fmt.Sprintf(
						"🌾 *Panen:*\n• Berat: %d Kg\n• Gross: Rp%s\n• Upah+Trans: Rp%s\n\n"+
							"💰 *Operasional:*\n• Biaya Ops: Rp%s\n• Total Biaya: Rp%s\n\n"+
							"📋 *Piutang:*\n• Outst. Utang: Rp%s\n\n"+
							"📈 *Finansial:*\n• Profit: Rp%s\n• Sisa Modal: Rp%s\n• ROI: %.2f%%\n• *BEP: %s*",
						report.TotalWeight, formatRupiah(report.GrossIncome), formatRupiah(report.TotalUpah+report.TotalTransport),
						formatRupiah(report.TotalOperasional), formatRupiah(report.OperationalCost),
						formatRupiah(report.OutstandingDebt),
						formatRupiah(report.NetProfit), formatRupiah(report.RemainingCapital), report.ROITracking,
						report.BEPProjection,
					)), nil, nil),
					slack.NewDividerBlock(),
					slack.NewSectionBlock(md(fmt.Sprintf(
						"📝 *Detail Hitung:*\n"+
							"• Gross: Rp%s\n"+
							"• Biaya: Rp%s\n"+
							"• Net: Rp%s - Rp%s = *Rp%s*",
						formatRupiah(report.GrossIncome), formatRupiah(report.OperationalCost),
						formatRupiah(report.GrossIncome), formatRupiah(report.OperationalCost), formatRupiah(report.NetProfit),
					)), nil, nil),
					slack.NewContextBlock("", md(fmt.Sprintf("_Target Investasi: Rp%s_", formatRupiah(report.TargetModal)))),
				},
			},
		},
	}
}

// formatRupiah formats an int64 into a human-readable Rupiah string with dots.
func formatRupiah(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	result := ""
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}
