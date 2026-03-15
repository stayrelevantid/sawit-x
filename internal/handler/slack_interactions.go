package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/indragiri/sawit-x/internal/model"
	"github.com/indragiri/sawit-x/internal/service"
	"github.com/slack-go/slack"
)

type SlackInteractionsHandler struct {
	masterDataService *service.MasterDataService
	logService        *service.LogService
	uiService         *service.UIService
	slackClient       *slack.Client
}

func NewSlackInteractionsHandler(mds *service.MasterDataService, ls *service.LogService, uis *service.UIService) *SlackInteractionsHandler {
	token := os.Getenv("SLACK_BOT_TOKEN")
	return &SlackInteractionsHandler{
		masterDataService: mds,
		logService:        ls,
		uiService:         uis,
		slackClient:       slack.New(token),
	}
}

// HandleInteraction is the main dispatcher for all Slack block-kit interactions.
func (h *SlackInteractionsHandler) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	var payload slack.InteractionCallback
	if err := json.Unmarshal([]byte(r.PostFormValue("payload")), &payload); err != nil {
		log.Printf("[INTERACTION] Error unmarshalling payload: %v", err)
		http.Error(w, "Error parsing payload", http.StatusBadRequest)
		return
	}

	switch payload.Type {
	case slack.InteractionTypeViewSubmission:
		h.handleViewSubmission(w, r, payload)
	case slack.InteractionTypeBlockActions:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// handleViewSubmission routes to the correct handler based on the modal's CallbackID.
func (h *SlackInteractionsHandler) handleViewSubmission(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	switch payload.View.CallbackID {
	case "site_selection_modal":
		h.handleSiteSelection(w, r, payload)
	case "module_selection_modal":
		h.handleModuleSelection(w, r, payload)
	case "panen_entry_modal":
		h.handlePanenEntry(w, r, payload)
	case "operasional_entry_modal":
		h.handleOperasionalEntry(w, r, payload)
	case "piutang_crew_select_modal":
		h.handlePiutangCrewSelect(w, r, payload)
	case "piutang_action_modal":
		h.handlePiutangAction(w, r, payload)
	default:
		log.Printf("[INTERACTION] Unknown callbackID: %s", payload.View.CallbackID)
		w.WriteHeader(http.StatusOK)
	}
}

// --- Step 1: Site Selection ---
// When user picks a site, update the view to the module selection modal.
func (h *SlackInteractionsHandler) handleSiteSelection(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()
	siteID := payload.View.State.Values["site_selection_block"]["site_id"].SelectedOption.Value
	siteName := payload.View.State.Values["site_selection_block"]["site_id"].SelectedOption.Text.Text

	// Validate site exists
	sites, _ := h.masterDataService.GetActiveSites(ctx)
	for _, s := range sites {
		if s.ID == siteID {
			siteName = s.Name
			break
		}
	}

	state := model.TransactionState{
		SiteID:   siteID,
		SiteName: siteName,
	}

	modal := h.uiService.BuildModuleSelectionModal(state)
	respondWithUpdateView(w, modal)
}

// --- Step 2: Module Selection ---
// When user picks a module, open the corresponding entry modal.
func (h *SlackInteractionsHandler) handleModuleSelection(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()

	var state model.TransactionState
	json.Unmarshal([]byte(payload.View.PrivateMetadata), &state)

	moduleType := payload.View.State.Values["module_block"]["module_type"].SelectedOption.Value
	state.ModuleType = moduleType

	switch moduleType {
	case model.ModulePanen:
		crew, _ := h.masterDataService.GetActiveCrew(ctx)
		modal := h.uiService.BuildPanenModal(state, crew)
		respondWithUpdateView(w, modal)

	case model.ModuleOperasional:
		categories, _ := h.masterDataService.GetActiveCategories(ctx)
		crew, _ := h.masterDataService.GetActiveCrew(ctx)
		modal := h.uiService.BuildOperasionalModal(state, categories, crew)
		respondWithUpdateView(w, modal)

	case model.ModulePiutang:
		crew, _ := h.masterDataService.GetActiveCrew(ctx)
		modal := h.uiService.BuildPiutangCrewSelectModal(state, crew)
		respondWithUpdateView(w, modal)

	default:
		log.Printf("[INTERACTION] Unknown module type: %s", moduleType)
		w.WriteHeader(http.StatusOK)
	}
}

// --- Step 3a: Panen Entry ---
func (h *SlackInteractionsHandler) handlePanenEntry(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()
	values := payload.View.State.Values

	var state model.TransactionState
	json.Unmarshal([]byte(payload.View.PrivateMetadata), &state)

	// Parse crew (multi-select)
	crewElements := values["crew_block"]["crew_id"].SelectedOptions
	var crewIDs, crewNames []string
	for _, opt := range crewElements {
		crewIDs = append(crewIDs, opt.Value)
		crewNames = append(crewNames, opt.Text.Text)
	}

	// Parse numeric fields
	eventDate := values["date_block"]["event_date"].SelectedDate
	weightStr := values["weight_block"]["weight"].Value
	priceStr := values["unit_price_block"]["unit_price"].Value
	laborStr := values["labor_block"]["labor_cost"].Value
	transportStr := values["transport_block"]["transport_cost"].Value
	notes := values["notes_block"]["notes"].Value

	parseNum := func(s string) (int64, bool) {
		n, err := strconv.ParseInt(strings.ReplaceAll(s, ".", ""), 10, 64)
		return n, err == nil && n >= 0
	}

	weight, okW := parseNum(weightStr)
	unitPrice, okP := parseNum(priceStr)

	if !okW || !okP {
		respondWithErrors(w, "weight_block", "Berat dan Harga harus berupa angka positif (contoh: 1250)")
		return
	}

	laborCost, _ := parseNum(laborStr)
	transportCost, _ := parseNum(transportStr)

	grossIncome := weight * unitPrice
	netIncome := grossIncome - laborCost - transportCost

	entry := model.LogEntry{
		LogID:         uuid.New().String(),
		Timestamp:     time.Now(),
		EventDate:     parseDate(eventDate),
		ModuleType:    model.ModulePanen,
		SiteID:        state.SiteID,
		SiteName:      state.SiteName,
		CategoryID:    "PANEN",
		CategoryName:  "Panen TBS",
		CrewID:        strings.Join(crewIDs, ", "),
		CrewName:      strings.Join(crewNames, ", "),
		AmountRaw:     grossIncome,
		AmountFinal:   netIncome,
		Weight:        weight,
		UnitPrice:     unitPrice,
		LaborCost:     laborCost,
		TransportCost: transportCost,
		Notes:         notes,
		SlackUserID:   payload.User.ID,
		SlackUsername: payload.User.Name,
	}

	if err := h.logService.WriteLog(ctx, entry); err != nil {
		log.Printf("[PANEN] Error writing log: %v", err)
	}

	respondClear(w)
	go h.sendSuccessDM(payload.User.ID, entry)
}

// --- Step 3b: Operasional Entry ---
func (h *SlackInteractionsHandler) handleOperasionalEntry(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()
	values := payload.View.State.Values

	var state model.TransactionState
	json.Unmarshal([]byte(payload.View.PrivateMetadata), &state)

	catOption := values["category_block"]["category_id"].SelectedOption
	catID := catOption.Value
	catName := catOption.Text.Text

	crewOption := values["crew_block"]["crew_id"].SelectedOption
	crewID := crewOption.Value
	crewName := crewOption.Text.Text

	eventDate := values["date_block"]["event_date"].SelectedDate
	amountStr := values["amount_block"]["amount_raw"].Value
	notes := values["notes_block"]["notes"].Value

	amount, err := strconv.ParseInt(strings.ReplaceAll(amountStr, ".", ""), 10, 64)
	if err != nil || amount < 0 {
		respondWithErrors(w, "amount_block", "Nominal harus berupa angka bulat positif (contoh: 200000)")
		return
	}

	entry := model.LogEntry{
		LogID:         uuid.New().String(),
		Timestamp:     time.Now(),
		EventDate:     parseDate(eventDate),
		ModuleType:    model.ModuleOperasional,
		SiteID:        state.SiteID,
		SiteName:      state.SiteName,
		CategoryID:    catID,
		CategoryName:  catName,
		CrewID:        crewID,
		CrewName:      crewName,
		AmountRaw:     amount,
		AmountFinal:   amount,
		Notes:         notes,
		SlackUserID:   payload.User.ID,
		SlackUsername: payload.User.Name,
	}

	if err := h.logService.WriteLog(ctx, entry); err != nil {
		log.Printf("[OPERASIONAL] Error writing log: %v", err)
	}

	respondClear(w)
	go h.sendSuccessDM(payload.User.ID, entry)
}

// --- Step 3c-1: Piutang — Crew Selection ---
// After selecting a crew member, fetch their balance and open the action modal.
func (h *SlackInteractionsHandler) handlePiutangCrewSelect(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()

	var state model.TransactionState
	json.Unmarshal([]byte(payload.View.PrivateMetadata), &state)

	crewOption := payload.View.State.Values["crew_block"]["crew_id"].SelectedOption
	crewID := crewOption.Value
	crewName := crewOption.Text.Text

	state.CrewID = crewID
	state.CrewName = crewName

	balance, err := h.masterDataService.GetCrewBalance(ctx, crewID)
	if err != nil {
		log.Printf("[PIUTANG] Error fetching balance for %s: %v", crewID, err)
		balance = 0
	}

	modal := h.uiService.BuildPiutangActionModal(state, crewName, balance)
	respondWithUpdateView(w, modal)
}

// --- Step 3c-2: Piutang — Action (Pinjam / Bayar) ---
func (h *SlackInteractionsHandler) handlePiutangAction(w http.ResponseWriter, r *http.Request, payload slack.InteractionCallback) {
	ctx := r.Context()
	values := payload.View.State.Values

	var state model.TransactionState
	json.Unmarshal([]byte(payload.View.PrivateMetadata), &state)

	actionOption := values["action_block"]["piutang_action"].SelectedOption
	actionID := actionOption.Value   // PINJAM or BAYAR
	actionName := actionOption.Text.Text

	eventDate := values["date_block"]["event_date"].SelectedDate
	amountStr := values["amount_block"]["amount_raw"].Value
	notes := values["notes_block"]["notes"].Value

	amount, err := strconv.ParseInt(strings.ReplaceAll(amountStr, ".", ""), 10, 64)
	if err != nil || amount <= 0 {
		respondWithErrors(w, "amount_block", "Nominal harus berupa angka positif (contoh: 500000)")
		return
	}

	entry := model.LogEntry{
		LogID:         uuid.New().String(),
		Timestamp:     time.Now(),
		EventDate:     parseDate(eventDate),
		ModuleType:    model.ModulePiutang,
		SiteID:        state.SiteID,
		SiteName:      state.SiteName,
		CategoryID:    actionID,
		CategoryName:  actionName,
		CrewID:        state.CrewID,
		CrewName:      state.CrewName,
		AmountRaw:     amount,
		AmountFinal:   amount,
		Notes:         notes,
		SlackUserID:   payload.User.ID,
		SlackUsername: payload.User.Name,
	}

	if err := h.logService.WriteLog(ctx, entry); err != nil {
		log.Printf("[PIUTANG] Error writing log: %v", err)
	}

	respondClear(w)
	go h.sendSuccessDM(payload.User.ID, entry)
}

// --- Helpers ---

func (h *SlackInteractionsHandler) sendSuccessDM(userID string, entry model.LogEntry) {
	msg := h.uiService.BuildSuccessResponse(entry)
	_, _, err := h.slackClient.PostMessage(userID, slack.MsgOptionBlocks(msg.Blocks.BlockSet...))
	if err != nil {
		log.Printf("[INTERACTION] Error sending success DM: %v", err)
	}
}

func respondWithUpdateView(w http.ResponseWriter, modal slack.ModalViewRequest) {
	resp := struct {
		ResponseAction string                 `json:"response_action"`
		View           slack.ModalViewRequest `json:"view"`
	}{
		ResponseAction: "update",
		View:           modal,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func respondClear(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"response_action":"clear"}`))
}

func respondWithErrors(w http.ResponseWriter, blockID, message string) {
	resp := struct {
		ResponseAction string            `json:"response_action"`
		Errors         map[string]string `json:"errors"`
	}{
		ResponseAction: "errors",
		Errors:         map[string]string{blockID: message},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseDate(d string) time.Time {
	t, _ := time.Parse("2006-01-02", d)
	return t
}

// unused — kept for backward compatibility reference
func (h *SlackInteractionsHandler) sendErrorMessage(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"response_type": "ephemeral", "text": "❌ ` + msg + `"}`))
}

// background handler for events (kept for context-awareness)
func (h *SlackInteractionsHandler) runBackground(ctx context.Context, f func(context.Context)) {
	go f(context.Background())
}
