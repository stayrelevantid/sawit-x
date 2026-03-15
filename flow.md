# SAWIT-X: System Flow
> Scaling the plantation, automating the ledger, staying relevant.

## Arsitektur

```
Slack → Cloud Functions Gen2 (Go) → Google Sheets API
                ↓
        SlackVerifier (HMAC-SHA256)
                ↓
        ┌───────────────┐
        │  /slack/events │──→ HandleCommand (Slash Command)
        │  /slack/inter  │──→ HandleInteraction (Modal Submit)
        │  /health       │──→ Health Check
        └───────────────┘
```

## Endpoint

| Endpoint | Method | Fungsi |
|----------|--------|--------|
| `/health` | GET | Health check (`{"status":"ok"}`) |
| `/slack/events` | POST | Menerima slash command `/sawit-x` |
| `/slack/interactions` | POST | Menerima interaksi modal (submit) |

## Flow Lengkap

### Step 1: Trigger — `/sawit-x`

**User** mengetik `/sawit-x` di Slack.

**System:**
1. Slack kirim POST ke `/slack/events`
2. `SlackVerifier` memverifikasi signature HMAC-SHA256
3. Server langsung respon **200 OK** (agar tidak timeout 3 detik)
4. Background goroutine:
   - Fetch data kebun dari tab **`Sites`** di Google Sheets
   - Filter hanya yang `Status == ACTIVE`
   - Buka **Modal 1: Pilih Lokasi Kebun**

### Step 2: Modal 1 — Pilih Kebun

**UI:**
- Header: "Pilih Lokasi Kebun"
- Dropdown: Daftar kebun aktif (nama + lokasi)
- Tombol: [Lanjut] [Batal]

**User** memilih kebun → klik [Lanjut]

**System:**
1. Slack kirim POST ke `/slack/interactions` (callback: `site_selection_modal`)
2. Handler ambil `site_id` yang dipilih
3. Fetch **Categories** dan **Crew** dari Sheets
4. Simpan `{site_id, site_name}` ke `PrivateMetadata`
5. Respon: **Update view** → tampilkan **Modal 2**

### Step 3: Modal 2 — Detail Transaksi

**UI:**
| Field | Tipe | Keterangan |
|-------|------|------------|
| Kategori | Dropdown | Panen TBS, Pupuk, Herbisida, Utang, dll |
| Personil | Dropdown | Daftar crew aktif (nama + role) |
| Tanggal | Date Picker | Default: hari ini |
| Nominal (Rupiah) | Text Input | Angka bulat positif |
| Catatan | Text Input | Opsional |

- Tombol: [Kirim] [Kembali]

**User** mengisi semua field → klik [Kirim]

### Step 4: Submit & Tulis ke Sheets

**System** (callback: `transaction_entry_modal`):
1. Parse semua input dari modal
2. Validasi: nominal harus angka positif (jika tidak → error inline di modal)
3. Buat `LogEntry` dengan UUID unik
4. **Tulis ke tab `X_LOG`** di Google Sheets (15 kolom)
5. Respon: **Clear modal** (modal ditutup)
6. Background: Kirim **DM konfirmasi** ke user

**DM Konfirmasi:**
```
✅ Data Berhasil Dicatat!

Site: Kebun Alpha
Kategori: Panen TBS
Crew: Budi Santoso
Nominal: Rp200000
```

## Skema Google Sheets

### Tab `Sites`
| A: ID | B: Name | C: Location | D: Status | E: Target_Modal |
|-------|---------|-------------|-----------|-----------------|
| site_1 | Kebun Alpha | Kalimantan Timur | ACTIVE | 1000 |

### Tab `Categories`
| A: ID | B: Name | C: Type | D: MultiplierEnabled | E: Status |
|-------|---------|---------|---------------------|-----------|
| cat_1 | Panen TBS | PANEN | TRUE | ACTIVE |
| cat_4 | Utang Panen | UTANG | FALSE | ACTIVE |

### Tab `Crew`
| A: ID | B: Name | C: Role | D: SiteID | E: Status |
|-------|---------|---------|-----------|-----------|
| crew_1 | Budi Santoso | Pemanen | site_1 | ACTIVE |

### Tab `X_LOG` (15 kolom)
| Kolom | Field |
|-------|-------|
| A | LogID (UUID) |
| B | Timestamp (RFC3339) |
| C | EventDate (YYYY-MM-DD) |
| D | SiteID |
| E | SiteName |
| F | CategoryID |
| G | CategoryName |
| H | CrewID |
| I | CrewName |
| J | AmountRaw |
| K | AmountFinal |
| L | Notes |
| M | SlackUserID |
| N | SlackUsername |
| O | ChannelID |

## Security

- Semua request Slack diverifikasi via **HMAC-SHA256** (`X-Slack-Signature`)
- Timestamp expire: 5 menit
- Credentials Google Sheets via env var `GOOGLE_CREDENTIALS_JSON` (base64)
- Environment variables via Cloud Functions config (bukan hardcoded)

## Teknologi

| Komponen | Stack |
|----------|-------|
| Runtime | Go 1.26 |
| Hosting | GCP Cloud Functions Gen2 (asia-southeast2) |
| Database | Google Sheets API v4 |
| Bot Framework | slack-go/slack |
| HTTP | Functions Framework Go |