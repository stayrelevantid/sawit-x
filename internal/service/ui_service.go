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
func (s *UIService) BuildSiteSelectionModal(sites []model.Site, channelID string) slack.ModalViewRequest {
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
		PrivateMetadata: fmt.Sprintf(`{"channel_id":"%s"}`, channelID),
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

	pencatatanBtn := slack.NewButtonBlockElement("mode_pencatatan", "PENCATATAN", txt("✍️ Pencatatan Baru"))
	pencatatanBtn.Style = slack.StylePrimary

	return slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           txt("🌴 Menu Kebun"),
		Close:           txt("Tutup"),
		CallbackID:      "mode_selection_modal",
		PrivateMetadata: string(stateJSON),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewHeaderBlock(txt(fmt.Sprintf("🌴 Kebun: %s", state.SiteName))),
				slack.NewContextBlock("", md("Silakan pilih transaksi baru atau laporan yang ingin Anda akses.")),
				slack.NewDividerBlock(),

				// Group 1: Transaksi Baru
				slack.NewSectionBlock(md("*✍️ Transaksi Baru*\n_Catat panen, pengeluaran kebun, kasbon, atau investasi._"), nil, nil),
				slack.NewActionBlock(
					"group_transaksi_block",
					pencatatanBtn,
				),
				slack.NewDividerBlock(),

				// Group 2: Laporan & Keuangan
				slack.NewSectionBlock(md("*📊 Rekap & Keuangan*\n_Ringkasan performa laba/rugi kebun dan rekap saldo kasbon pegawai._"), nil, nil),
				slack.NewActionBlock(
					"group_keuangan_block",
					slack.NewButtonBlockElement("view_report", "REKAP", txt("📊 Lihat Rekap Kebun")),
					slack.NewButtonBlockElement("view_rekap_hutang_pegawai", "CREW_DEBT_LIST", txt("📋 Rekap Hutang Pegawai")),
				),
				slack.NewDividerBlock(),

				// Group 3: Riwayat Hasil Panen
				slack.NewSectionBlock(md("*🌾 Riwayat Hasil Panen*\n_Daftar histori timbangan panen, harga/kg, dan pendapatan bersih._"), nil, nil),
				slack.NewActionBlock(
					"group_panen_block",
					slack.NewButtonBlockElement("view_list_panen_1_tahun_ini", "PANEN_YEAR_THIS", txt("📅 Panen 1 Tahun Ini")),
					slack.NewButtonBlockElement("view_list_panen_1_tahun_lalu", "PANEN_YEAR_LAST", txt("📅 Panen 1 Tahun Lalu")),
				),
				slack.NewDividerBlock(),

				// Group 4: Log Perawatan Kebun
				slack.NewSectionBlock(md("*🌿 Perawatan Kebun*\n_Histori pembelian pupuk dan log penyemprotan gulma/hama._"), nil, nil),
				slack.NewActionBlock(
					"group_perawatan_block",
					slack.NewButtonBlockElement("view_list_pupuk", "PUPUK_LIST", txt("🧪 List Pembelian Pupuk")),
					slack.NewButtonBlockElement("view_list_semprot", "SEMPROT_LIST", txt("🌧️ List Penyemprotan")),
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
		detail = fmt.Sprintf("*Kebun:* %s\n*Pemanen:* %s\n*Berat:* %d Kg\n*Harga:* Rp%s\n*Perhitungan:*\n> Gross: %d Kg x Rp%s = Rp%s\n> Biaya: Rp%s (Upah) + Rp%s (Bensin/Timbang)\n*Net Profit:* Rp%s\n*Catatan:* %s",
			entry.SiteName, entry.CrewName, entry.Weight, formatRupiah(entry.UnitPrice),
			entry.Weight, formatRupiah(entry.UnitPrice), formatRupiah(entry.AmountRaw),
			formatRupiah(entry.LaborCost), formatRupiah(entry.TransportCost), formatRupiah(entry.AmountFinal), entry.Notes)
	case model.ModuleOperasional:
		detail = fmt.Sprintf("*Kebun:* %s\n*Kategori:* %s\n*PJ:* %s\n*Nominal:* Rp%s\n*Keterangan:* %s",
			entry.SiteName, entry.CategoryName, entry.CrewName, formatRupiah(entry.AmountRaw), entry.Notes)
	case model.ModulePiutang:
		action := entry.CategoryID // PINJAM or BAYAR
		prevBalance := entry.AmountFinal - entry.AmountRaw
		if action == "BAYAR" {
			prevBalance = entry.AmountFinal + entry.AmountRaw
		}
		detail = fmt.Sprintf("*Kebun:* %s\n*Pegawai:* %s\n*Aksi:* %s\n*Nominal:* Rp%s\n*Perhitungan Saldo:*\n> Saldo Awal: Rp%s\n> %s: Rp%s\n*Saldo Akhir:* Rp%s\n*Catatan:* %s",
			entry.SiteName, entry.CrewName, action, formatRupiah(entry.AmountRaw),
			formatRupiah(prevBalance), action, formatRupiah(entry.AmountRaw), formatRupiah(entry.AmountFinal), entry.Notes)
	case model.ModuleInvestasi:
		detail = fmt.Sprintf("*Kebun:* %s\n*Kategori:* %s\n*Nominal:* Rp%s\n*Keterangan:* %s",
			entry.SiteName, entry.CategoryName, formatRupiah(entry.AmountRaw), entry.Notes)
	default:
		detail = fmt.Sprintf("*Kebun:* %s\n*Nominal:* Rp%s", entry.SiteName, formatRupiah(entry.AmountRaw))
	}

	// Add event date to the detail
	detail = fmt.Sprintf("*Tanggal:* %s\n%s", entry.EventDate.Format("02 Jan 2006"), detail)

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
				slack.NewHeaderBlock(txt(fmt.Sprintf("📊 Rekap: %s", siteName))),
				slack.NewContextBlock("", md(fmt.Sprintf("_Data per tanggal: %s_", time.Now().Format("02 Jan 2006, 15:04")))),
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
					slack.NewContextBlock("", md(fmt.Sprintf("_Data per tanggal: %s_", time.Now().Format("02 Jan 2006, 15:04")))),
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

// BuildListPanenModal builds a modal to show harvest data for a specific year.
func (s *UIService) BuildListPanenModal(siteName string, targetYear int, panenList []model.LogEntry) slack.ModalViewRequest {
	blocks := []slack.Block{
		slack.NewHeaderBlock(txt(fmt.Sprintf("📅 List Panen %d", targetYear))),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(panenList)))),
		slack.NewDividerBlock(),
	}

	limit := 30
	if len(panenList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Tidak ada data panen di tahun ini."), nil, nil))
	} else {
		count := 0
		var totalWeight int64
		var totalNet int64
		for _, item := range panenList {
			totalWeight += item.Weight
			totalNet += item.AmountFinal
		}

		for i := len(panenList) - 1; i >= 0; i-- {
			if count >= limit {
				blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
				break
			}
			p := panenList[i]
			priceStr := "-"
			if p.UnitPrice > 0 {
				priceStr = formatRupiah(p.UnitPrice)
			}
			detail := fmt.Sprintf("🌾 *%s* | _%s_\n⚖️ *Berat:* %s Kg  •  🏷️ *Harga:* Rp%s /Kg\n💵 *Net:* Rp%s",
				p.EventDate.Format("02 Jan 2006"),
				p.CrewName,
				formatRupiah(p.Weight),
				priceStr,
				formatRupiah(p.AmountFinal),
			)
			if p.Notes != "" {
				detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
			}
			blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
			blocks = append(blocks, slack.NewDividerBlock())
			count++
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("📊 *Total Panen %d:* %s Kg | *Total Net:* Rp%s", targetYear, formatRupiah(totalWeight), formatRupiah(totalNet)))))
	}

	return slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: txt("🌾 Riwayat Panen"),
		Close: txt("Tutup"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}

// BuildListPanenMessage builds a message-based report for harvest data.
func (s *UIService) BuildListPanenMessage(siteName string, targetYear int, panenList []model.LogEntry) slack.Message {
	blocks := []slack.Block{
		slack.NewSectionBlock(md(fmt.Sprintf("📅 *LIST PANEN %d*", targetYear)), nil, nil),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(panenList)))),
		slack.NewDividerBlock(),
	}

	limit := 30
	if len(panenList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Tidak ada data panen di tahun ini."), nil, nil))
	} else {
		count := 0
		var totalWeight int64
		var totalNet int64
		for _, item := range panenList {
			totalWeight += item.Weight
			totalNet += item.AmountFinal
		}

		for i := len(panenList) - 1; i >= 0; i-- {
			if count >= limit {
				blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
				break
			}
			p := panenList[i]
			priceStr := "-"
			if p.UnitPrice > 0 {
				priceStr = formatRupiah(p.UnitPrice)
			}
			detail := fmt.Sprintf("🌾 *%s* | _%s_\n⚖️ *Berat:* %s Kg  •  🏷️ *Harga:* Rp%s /Kg\n💵 *Net:* Rp%s",
				p.EventDate.Format("02 Jan 2006"),
				p.CrewName,
				formatRupiah(p.Weight),
				priceStr,
				formatRupiah(p.AmountFinal),
			)
			if p.Notes != "" {
				detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
			}
			blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
			blocks = append(blocks, slack.NewDividerBlock())
			count++
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("📊 *Total Panen %d:* %s Kg | *Total Net:* Rp%s", targetYear, formatRupiah(totalWeight), formatRupiah(totalNet)))))
	}

	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		},
	}
}

// BuildListPupukModal builds a modal displaying fertilizer purchase logs.
func (s *UIService) BuildListPupukModal(siteName string, pupukList []model.PupukLogEntry) slack.ModalViewRequest {
	blocks := []slack.Block{
		slack.NewHeaderBlock(txt("🧪 List Pembelian Pupuk")),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(pupukList)))),
		slack.NewDividerBlock(),
	}

	if len(pupukList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada riwayat pembelian pupuk tercatat."), nil, nil))
	} else {
		limit := 30
		count := 0
		var totalNominal int64
		for i := len(pupukList) - 1; i >= 0; i-- {
			p := pupukList[i]
			totalNominal += p.Amount
			if count < limit {
				detail := fmt.Sprintf("🧪 *%s* | *PJ:* %s\n*Nominal:* Rp%s", p.EventDate.Format("02 Jan 2006"), p.CrewName, formatRupiah(p.Amount))
				if p.Notes != "" {
					detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
				}
				blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
				blocks = append(blocks, slack.NewDividerBlock())
				count++
			}
		}
		if len(pupukList) > limit {
			blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Akumulasi Pembelian: Rp%s_", formatRupiah(totalNominal)))))
	}

	return slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: txt("🧪 Rekap Pupuk"),
		Close: txt("Tutup"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}

// BuildListPupukMessage builds a message response for fertilizer purchase logs.
func (s *UIService) BuildListPupukMessage(siteName string, pupukList []model.PupukLogEntry) slack.Message {
	blocks := []slack.Block{
		slack.NewSectionBlock(md("🧪 *LIST PEMBELIAN PUPUK*"), nil, nil),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(pupukList)))),
		slack.NewDividerBlock(),
	}

	if len(pupukList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada riwayat pembelian pupuk tercatat."), nil, nil))
	} else {
		limit := 30
		count := 0
		var totalNominal int64
		for i := len(pupukList) - 1; i >= 0; i-- {
			p := pupukList[i]
			totalNominal += p.Amount
			if count < limit {
				detail := fmt.Sprintf("🧪 *%s* | *PJ:* %s\n*Nominal:* Rp%s", p.EventDate.Format("02 Jan 2006"), p.CrewName, formatRupiah(p.Amount))
				if p.Notes != "" {
					detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
				}
				blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
				blocks = append(blocks, slack.NewDividerBlock())
				count++
			}
		}
		if len(pupukList) > limit {
			blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Akumulasi Pembelian: Rp%s_", formatRupiah(totalNominal)))))
	}

	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		},
	}
}

// BuildCrewDebtModal builds a modal displaying crew debt summary per person.
func (s *UIService) BuildCrewDebtModal(siteName string, summaries []model.CrewDebtSummary) slack.ModalViewRequest {
	blocks := []slack.Block{
		slack.NewHeaderBlock(txt("📋 Rekap Hutang Pegawai")),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d pegawai_", siteName, len(summaries)))),
		slack.NewDividerBlock(),
	}

	if len(summaries) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada data pegawai / hutang tercatat."), nil, nil))
	} else {
		var totalOutstanding int64
		for _, c := range summaries {
			totalOutstanding += c.OutstandingDebt
			status := "🟢 Lunas / Tidak ada hutang"
			if c.OutstandingDebt > 0 {
				status = fmt.Sprintf("🔴 *Sisa Hutang:* Rp%s", formatRupiah(c.OutstandingDebt))
			} else if c.OutstandingDebt < 0 {
				status = fmt.Sprintf("🔵 *Lebih Bayar:* Rp%s", formatRupiah(-c.OutstandingDebt))
			}

			detail := fmt.Sprintf("👤 *%s* (_%s_)\n• Pinjam: Rp%s | Bayar: Rp%s\n%s",
				c.CrewName, c.Role, formatRupiah(c.TotalPinjam), formatRupiah(c.TotalBayar), status)

			blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
			blocks = append(blocks, slack.NewDividerBlock())
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Sisa Utang Beredar: Rp%s_", formatRupiah(totalOutstanding)))))
	}

	return slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: txt("📋 Hutang Pegawai"),
		Close: txt("Tutup"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}

// BuildCrewDebtMessage builds a message response for crew debt summary per person.
func (s *UIService) BuildCrewDebtMessage(siteName string, summaries []model.CrewDebtSummary) slack.Message {
	blocks := []slack.Block{
		slack.NewSectionBlock(md("📋 *REKAP HUTANG PEGAWAI*"), nil, nil),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d pegawai_", siteName, len(summaries)))),
		slack.NewDividerBlock(),
	}

	if len(summaries) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada data pegawai / hutang tercatat."), nil, nil))
	} else {
		var totalOutstanding int64
		for _, c := range summaries {
			totalOutstanding += c.OutstandingDebt
			status := "🟢 Lunas"
			if c.OutstandingDebt > 0 {
				status = fmt.Sprintf("🔴 *Utang:* Rp%s", formatRupiah(c.OutstandingDebt))
			} else if c.OutstandingDebt < 0 {
				status = fmt.Sprintf("🔵 *Lebih Bayar:* Rp%s", formatRupiah(-c.OutstandingDebt))
			}

			detail := fmt.Sprintf("👤 *%s* (_%s_)\n• Pinjam: Rp%s | Bayar: Rp%s | %s",
				c.CrewName, c.Role, formatRupiah(c.TotalPinjam), formatRupiah(c.TotalBayar), status)

			blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
			blocks = append(blocks, slack.NewDividerBlock())
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Sisa Utang Beredar: Rp%s_", formatRupiah(totalOutstanding)))))
	}

	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		},
	}
}

// BuildListSemprotModal builds a modal displaying spraying operational logs.
func (s *UIService) BuildListSemprotModal(siteName string, semprotList []model.SemprotLogEntry) slack.ModalViewRequest {
	blocks := []slack.Block{
		slack.NewHeaderBlock(txt("🌧️ List Penyemprotan")),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(semprotList)))),
		slack.NewDividerBlock(),
	}

	if len(semprotList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada riwayat penyemprotan tercatat."), nil, nil))
	} else {
		limit := 30
		count := 0
		var totalNominal int64
		for i := len(semprotList) - 1; i >= 0; i-- {
			p := semprotList[i]
			totalNominal += p.Amount
			if count < limit {
				detail := fmt.Sprintf("🌧️ *%s* | *PJ:* %s\n*Nominal:* Rp%s", p.EventDate.Format("02 Jan 2006"), p.CrewName, formatRupiah(p.Amount))
				if p.Notes != "" {
					detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
				}
				blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
				blocks = append(blocks, slack.NewDividerBlock())
				count++
			}
		}
		if len(semprotList) > limit {
			blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Akumulasi Biaya Semprot: Rp%s_", formatRupiah(totalNominal)))))
	}

	return slack.ModalViewRequest{
		Type:  slack.VTModal,
		Title: txt("🌧️ Rekap Semprot"),
		Close: txt("Tutup"),
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}
}

// BuildListSemprotMessage builds a message response for spraying operational logs.
func (s *UIService) BuildListSemprotMessage(siteName string, semprotList []model.SemprotLogEntry) slack.Message {
	blocks := []slack.Block{
		slack.NewSectionBlock(md("🌧️ *LIST PENYEMPROTAN*"), nil, nil),
		slack.NewContextBlock("", md(fmt.Sprintf("_Kebun: %s | Total: %d transaksi_", siteName, len(semprotList)))),
		slack.NewDividerBlock(),
	}

	if len(semprotList) == 0 {
		blocks = append(blocks, slack.NewSectionBlock(md("😔 Belum ada riwayat penyemprotan tercatat."), nil, nil))
	} else {
		limit := 30
		count := 0
		var totalNominal int64
		for i := len(semprotList) - 1; i >= 0; i-- {
			p := semprotList[i]
			totalNominal += p.Amount
			if count < limit {
				detail := fmt.Sprintf("🌧️ *%s* | *PJ:* %s\n*Nominal:* Rp%s", p.EventDate.Format("02 Jan 2006"), p.CrewName, formatRupiah(p.Amount))
				if p.Notes != "" {
					detail += fmt.Sprintf("\n📝 _Catatan: %s_", p.Notes)
				}
				blocks = append(blocks, slack.NewSectionBlock(md(detail), nil, nil))
				blocks = append(blocks, slack.NewDividerBlock())
				count++
			}
		}
		if len(semprotList) > limit {
			blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Data dibatasi %d transaksi terbaru_", limit))))
		}
		blocks = append(blocks, slack.NewContextBlock("", md(fmt.Sprintf("_Total Akumulasi Biaya Semprot: Rp%s_", formatRupiah(totalNominal)))))
	}

	return slack.Message{
		Msg: slack.Msg{
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		},
	}
}


