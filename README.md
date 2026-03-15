# 🌴 SAWIT-X: Advanced Plantation Ledger

Advanced Plantation Ledger & Multi-Site Management System, integrated with Slack and Google Sheets. Designed for scalability, transparency, and real-time financial tracking.

## 🚀 Key Features

*   **Multi-Site Ready**: Manage multiple plantation sites dynamically from a single dashboard.
*   **Modular Entry**: Dedicated modules for **Panen**, **Operasional**, **Piutang**, and **Investasi**.
*   **Smart Reports**: Detailed financial breakdown including Net Profit, ROI, and BEP Projections directly in Slack.
*   **Auto-Dashboard**: Tab `X_REKAP` in Google Sheets automatically syncs to provide a permanent management overview.
*   **Audit Trail**: Every transaction is logged with UUIDs and Slack user metadata in `X_LOG`.

## Project Structure

```
sawit-x/
├── cmd/
│   └── main.go              # Entry point for Cloud Function
├── internal/
│   ├── handler/            # Slack event & interaction handlers
│   ├── service/            # Core business logic (ROI, BEP, MasterData)
│   ├── client/             # Google Sheets API client
│   ├── middleware/         # Security & logging middlewares
│   └── model/              # Data structures
├── tools/
│   └── seed/               # Spreadsheet seeding & reset tool
├── flow.md                 # Detailed System Interaction Flow
├── PRD.md                  # Project Requirements & Technical Spec
└── README.md               # This file
```

## Setup & Development

1. **Prerequisites**: Go 1.26+, Slack App, Google Cloud Service Account.
2. **Environment**: Copy `.env.example` to `.env` and fill in the values (`SPREADSHEET_ID`, `SLACK_BOT_TOKEN`, etc.).
3. **Run Locally**:
   ```bash
   go run cmd/main.go
   ```

## Deployment

Deploys to GCP Cloud Functions Gen2.

```bash
gcloud functions deploy sawit-x \
  --gen2 \
  --runtime=go126 \
  --region=asia-southeast2 \
  --source=. \
  --entry-point=SawitX \
  --trigger-http
```
