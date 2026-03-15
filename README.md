# 🌴 SAWIT-X: Advanced Plantation Ledger

Advanced Plantation Ledger & Multi-Site Management System, integrated with Slack and Google Sheets.

## Project Structure

```
sawit-x/
├── cmd/
│   └── main.go              # Entry point for Cloud Function
├── internal/
│   ├── handler/            # Slack event & interaction handlers
│   ├── service/            # Core business logic
│   ├── client/             # Google Sheets API client
│   ├── middleware/         # Security & logging middlewares
│   └── model/              # Data structures
├── .env.example            # Environment variables template
├── go.mod                  # Go module definition
└── PRD.md                  # Project Requirements Document
```

## Setup & Development

1. **Prerequisites**: Go 1.26+, Slack App, Google Cloud Service Account.
2. **Environment**: Copy `.env.example` to `.env` and fill in the values.
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
