package service

import (
	"encoding/json"
	"fmt"
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
		Type:       slack.VTModal,
		Title:      txt("🌴 SAWIT-X"),
		Close:      txt("Batal"),
		Submit:     txt("Lanjut"),
		CallbackID: "site_selection_modal",
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

// BuildModuleSelectionModal builds the second modal for choosing a module (Panen/Operasional/Piutang).
func (s *UIService) BuildModuleSelectionModal(state model.TransactionState) slack.ModalViewRequest {
	stateJSON, _ := json.Marshal(state)

	panenOption := slack.NewOptionBlockObject("PANEN", txt("🌾 Panen"), txt("Catat hasil panen dan biaya logistik"))
	opsOption := slack.NewOptionBlockObject("OPERASIONAL", txt("💰 Operasional"), txt("Catat pengeluaran kebun"))
	piutangOption := slack.NewOptionBlockObject("PIUTANG", txt("📋 Piutang"), txt("Kelola pinjaman pegawai"))

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
						panenOption, opsOption, piutangOption,
					),
				),
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
		detail = fmt.Sprintf("*Kebun:* %s\n*Pemanen:* %s\n*Berat:* %d Kg\n*Gross:* Rp%s\n*Net:* Rp%s",
			entry.SiteName, entry.CrewName, entry.Weight,
			formatRupiah(entry.AmountRaw), formatRupiah(entry.AmountFinal))
	case model.ModuleOperasional:
		detail = fmt.Sprintf("*Kebun:* %s\n*Kategori:* %s\n*PJ:* %s\n*Nominal:* Rp%s",
			entry.SiteName, entry.CategoryName, entry.CrewName, formatRupiah(entry.AmountRaw))
	case model.ModulePiutang:
		action := entry.CategoryID // PINJAM or BAYAR
		detail = fmt.Sprintf("*Kebun:* %s\n*Pegawai:* %s\n*Aksi:* %s\n*Nominal:* Rp%s",
			entry.SiteName, entry.CrewName, action, formatRupiah(entry.AmountRaw))
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
