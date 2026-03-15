// Package handler contains the HTTP handlers for Slack events and interactions.
// Handlers are registered in cmd/main.go via the SawitX root handler.
//
// Phase 3 will implement:
//   - SlackEventsHandler  — POST /slack/events
//   - SlackInteractionsHandler — POST /slack/interactions
package handler
