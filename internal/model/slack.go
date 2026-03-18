package model

import "github.com/slack-go/slack"

// SlackCommand represents an incoming Slash Command from Slack.
type SlackCommand struct {
	Token          string `schema:"token"`
	TeamID         string `schema:"team_id"`
	TeamDomain     string `schema:"team_domain"`
	EnterpriseID   string `schema:"enterprise_id"`
	EnterpriseName string `schema:"enterprise_name"`
	ChannelID      string `schema:"channel_id"`
	ChannelName    string `schema:"channel_name"`
	UserID         string `schema:"user_id"`
	UserName       string `schema:"user_name"`
	Command        string `schema:"command"`
	Text           string `schema:"text"`
	ResponseURL    string `schema:"response_url"`
	TriggerID      string `schema:"trigger_id"`
	APIAppID       string `schema:"api_app_id"`
}

// TransactionState stores the temporary data during modal progression.
type TransactionState struct {
	ModuleType   string `json:"module_type"`   // PANEN / OPERASIONAL / PIUTANG
	SiteID       string `json:"site_id"`
	SiteName     string `json:"site_name"`
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	CrewID       string `json:"crew_id"`
	CrewName     string `json:"crew_name"`
	EventDate    string `json:"event_date"`
	AmountRaw    string `json:"amount_raw"`
	Notes        string `json:"notes"`
	ChannelID    string `json:"channel_id"`
}

// InteractionPayload is a wrapper for Slack's interaction JSON.
type InteractionPayload struct {
	slack.InteractionCallback
}
