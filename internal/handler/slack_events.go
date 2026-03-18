package handler

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/schema"
	"github.com/indragiri/sawit-x/internal/model"
	"github.com/indragiri/sawit-x/internal/service"
	"github.com/slack-go/slack"
)

type SlackEventsHandler struct {
	masterDataService *service.MasterDataService
	uiService         *service.UIService
	slackClient       *slack.Client
}

func NewSlackEventsHandler(mds *service.MasterDataService, uis *service.UIService) *SlackEventsHandler {
	token := os.Getenv("SLACK_BOT_TOKEN")
	return &SlackEventsHandler{
		masterDataService: mds,
		uiService:         uis,
		slackClient:       slack.New(token),
	}
}

func (h *SlackEventsHandler) HandleCommand(w http.ResponseWriter, r *http.Request) {
	if err := r.PostFormValue("command"); err == "" {
		// Ensure form is parsed
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
		}
	}

	var cmd model.SlackCommand
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&cmd, r.PostForm); err != nil {
		log.Printf("Error decoding slash command: %v", err)
		http.Error(w, "Error decoding command", http.StatusInternalServerError)
		return
	}

	log.Printf("[COMMAND] Received %s from %s (TriggerID: %s)", cmd.Command, cmd.UserName, cmd.TriggerID)

	// Respond 200 OK immediately to satisfy Slack's 3s timeout
	w.WriteHeader(http.StatusOK)

	// Process the rest in a background goroutine
	go func(triggerID string) {
		log.Printf("[COMMAND] Processing background task for TriggerID: %s", triggerID)
		
		// Use a fresh context for background work
		ctx := context.Background()

		// Fetch active sites for the dropdown
		sites, err := h.masterDataService.GetActiveSites(ctx)
		if err != nil {
			log.Printf("[COMMAND] Error fetching sites: %v", err)
			return
		}

		if len(sites) == 0 {
			log.Printf("[COMMAND] No active sites found")
			return
		}

		// Build the site selection modal
		modal := h.uiService.BuildSiteSelectionModal(sites, cmd.ChannelID)

		_, err = h.slackClient.OpenView(triggerID, modal)
		if err != nil {
			log.Printf("[COMMAND] Error opening Slack view: %v", err)
		} else {
			log.Printf("[COMMAND] Successfully opened modal for TriggerID: %s", triggerID)
		}
	}(cmd.TriggerID)
}

func (h *SlackEventsHandler) sendErrorMessage(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Slack likes 200 OK even for errors if we want to show a message
	w.Write([]byte(`{"response_type": "ephemeral", "text": "❌ ` + msg + `"}`))
}
