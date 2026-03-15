// Package sawitx is the root package for the SAWIT-X Cloud Function.
package sawitx

import (
	"context"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/indragiri/sawit-x/internal/client"
	"github.com/indragiri/sawit-x/internal/handler"
	"github.com/indragiri/sawit-x/internal/middleware"
	"github.com/indragiri/sawit-x/internal/service"
)

var (
	eventsHandler      *handler.SlackEventsHandler
	interactionHandler *handler.SlackInteractionsHandler
	sheetsClient       client.SheetsReader
)

func init() {
	// Register the HTTP function with the Functions Framework.
	functions.HTTP("SawitX", SawitX)

	// Initialize dependencies once during function warm-up
	ctx := context.Background()
	var err error
	sheetsClient, err = client.NewSheetsClient(ctx)
	if err != nil {
		log.Printf("CRITICAL: Failed to initialize Sheets client: %v", err)
		// We don't exit here because Cloud Functions init shouldn't panic, 
		// but we'll check this in the handler
	}

	mdService := service.NewMasterDataService(sheetsClient)
	logService := service.NewLogService(sheetsClient)
	uiService := service.NewUIService()

	eventsHandler = handler.NewSlackEventsHandler(mdService, uiService)
	interactionHandler = handler.NewSlackInteractionsHandler(mdService, logService, uiService)
}

// SawitX is the root HTTP handler for the Cloud Function.
func SawitX(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] Raw Request URL: %s", r.URL.String())
	log.Printf("[DEBUG] Request Path: %s", r.URL.Path)
	log.Printf("[DEBUG] Headers: %v", r.Header)

	if sheetsClient == nil {
		log.Printf("[CRITICAL] Sheets client is nil during request handling")
		http.Error(w, "Service Initialization Error", http.StatusInternalServerError)
		return
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler)

	// Slack entry points with HMAC verification
	mux.HandleFunc("POST /slack/events", middleware.SlackVerifier(eventsHandler.HandleCommand))
	mux.HandleFunc("POST /slack/interactions", middleware.SlackVerifier(interactionHandler.HandleInteraction))

	mux.ServeHTTP(w, r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","version":"1.1.0"}`))
}
