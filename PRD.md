# 🌴 PRD: Project SAWIT-X
### *Advanced Plantation Ledger & Multi-Site Management System*

---

| Field | Value |
|---|---|
| **Code Name** | SAWIT-X |
| **Status** | 🟢 Live / Production |
| **Maintainer** | - |
| **Platform** | GCP Cloud Functions (Go 1.26) |
| **Interface** | Slack Interactive App (Block Kit) |
| **Database** | Google Sheets (Headless API) |
| **Last Updated** | 2026-03-15 |

---

## 1. Latar Belakang & Visi Produk

### 1.1 Problem Statement

Pengelolaan keuangan perkebunan kelapa sawit skala kecil-menengah umumnya masih dilakukan secara manual (buku catatan fisik atau spreadsheet yang diisi secara ad-hoc). Masalah yang muncul:

- **Data tersebar** — catatan harian per mandor tidak terintegrasi ke dalam laporan terpadu.
- **Delay pencatatan** — transaksi sering baru dicatat hari/minggu berikutnya, rentan lupa dan salah nominal.
- **Kurang akuntabel** — tidak ada jejak audit (siapa mencatat, kapan, apa yang diubah).
- **Multi-site tidak skalabel** — saat kebun bertambah, sistem manual semakin tidak bisa diandalkan.

### 1.2 Visi

SAWIT-X dirancang sebagai **mesin pencatat finansial perkebunan** yang menghilangkan kerumitan manual. Fokus pada:

- **Skalabilitas** – mendukung multi-kebun tanpa perubahan kode.
- **Akurasi** – pencatatan nominal asli (Rupiah) dengan validasi otomatis.
- **Fleksibilitas** – kategori, lokasi, dan pegawai dikontrol penuh dari database tanpa deployment ulang.
- **Kemudahan adopsi** – antarmuka berbasis Slack yang sudah digunakan tim sehari-hari.

---

## 2. Fitur Inti (The X-Capabilities)

### 2.1 🛰️ Dynamic Site Engine

Mendukung ekspansi kebun tanpa batas kode. Sistem mendeteksi `Site_ID` secara dinamis dari sheet `X_MASTER`. Menambah kebun baru cukup dengan menambah baris di spreadsheet — tidak perlu redeploy atau ubah konfigurasi.

**Perilaku:**
- Saat slash command `/sawit-x` dipanggil, sistem membaca daftar kebun aktif dari `X_MASTER`.
- Kebun dengan status `INACTIVE` tidak akan muncul di dropdown.
- Order kebun di dropdown mengikuti urutan baris di spreadsheet.

### 2.2 🎭 Contextual UI (Slack Block Kit)

Antarmuka bertingkat yang sinkron real-time dengan database:

| Komponen | Fungsi |
|---|---|
| **Dropdown Site** | Daftar kebun aktif dari `X_MASTER[Sites]` |
| **Dropdown Kategori** | Kategori biaya dari `X_MASTER[Categories]` |
| **Dropdown Crew** | Nama pegawai dari `X_MASTER[Crew]` |
| **Date Picker** | Input tanggal dengan support *backdated logs* |
| **Modal Input** | Input nominal Rupiah (asli, tanpa auto-multiplier) |
| **Confirmation View** | Ringkasan transaksi sebelum submit final |

**Alur Interaksi Utama (New Flow):**
```
/sawit-x
  → Modal 1: Pilih Kebun
  → Modal 2: Pilih Mode (Pencatatan / Rekap)
  → Modal 3 (Pencatatan): Pilih Modul (Panen / Operasional / Piutang)
    → Modal 4a (Panen): Tanggal, Pemanen, Berat, Harga/Kg, Upah, Bensin
    → Modal 4b (Operasional): Kategori, Penanggung Jawab, Nominal, Keterangan
    → Modal 4c (Piutang): Pilih Pegawai → Tampilkan Saldo → Pinjam/Bayar + Nominal
  → Modal 3 (Rekap): Tampilkan Dashboard Performa
  → Aksi: Tulis ke X_LOG (untuk Pencatatan)
  → Response: Pesan sukses di DM Slack
```

### 2.3 🧮 Finance Intelligence

| Fitur | Deskripsi |
|---|---|
| **Auto-Multiplier** | Mengonversi input cepat (misal: `200`) menjadi ribuan (`200.000`). Mode ini *opsional* dan dapat dinonaktifkan di `X_MASTER`. |
| **Balance Tracking** | Memantau progres "Balik Modal" per lokasi kebun secara real-time dari tab `X_REKAP`. |
| **Audit Trail** | Setiap entri mencatat `user_slack_id`, `timestamp`, dan `channel_id` secara otomatis. |
| **ROI & BEP Engine** | Menghitung tingkat pengembalian investasi dan estimasi waktu balik modal berdasarkan profit rata-rata. |
| **Auto-Sync REKAP** | Sinkronisasi otomatis data agregat dari `X_LOG` ke `X_REKAP` tanpa campur tangan manual. |

---

## 3. Arsitektur Sistem

### 3.1 Diagram Alur

```
Slack User
    │
    │  /sawit-x (slash command)
    ▼
GCP Cloud Function (Go 1.26) ──── Slack HMAC Verification
    │
    ├── GET  /api/master-data  ──▶  Google Sheets API (X_MASTER)
    │                               └─ Sites, Categories, Crew
    │
    ├── POST /api/log          ──▶  Google Sheets API (X_LOG)
    │                               └─ Append row transaksi baru
    │
    └── GET  /api/rekap        ──▶  Google Sheets API (X_REKAP)
                                    └─ Aggregate & balance per site
```

### 3.2 Komponen Utama

| Komponen | Teknologi | Keterangan |
|---|---|---|
| **API Handler** | Go 1.26, `net/http` | Entry point untuk semua request dari Slack |
| **Slack Verifier** | HMAC-SHA256 | Validasi setiap inbound request dari Slack |
| **Sheets Client** | Google Sheets API v4 | Read/write ke spreadsheet |
| **Secret Manager** | GCP Secret Manager | Menyimpan Slack Signing Secret & Service Account credentials |
| **Deployment** | GCP Cloud Functions (Gen2) | Stateless, auto-scaling, serverless |

---

## 4. Skema Database (Google Sheets)

### 4.1 Tab: `X_LOG` — Ledger Transaksi

> Append-only. Tidak ada row yang boleh diedit/dihapus manual.

| Kolom | Tipe | Keterangan |
|---|---|---|
| `log_id` | STRING | UUID v4 yang di-generate sisi server |
| `timestamp` | DATETIME | Waktu server menerima request (UTC+7) |
| `event_date` | DATE | Tanggal kejadian (bisa backdated, input user) |
| `module_type` | ENUM | `PANEN` / `OPERASIONAL` / `PIUTANG` |
| `site_id` | STRING | Foreign key ke `X_MASTER[Sites].id` |
| `site_name` | STRING | Denormalized untuk kemudahan baca |
| `category_id` | STRING | Foreign key ke `X_MASTER[Categories].id` (atau `PINJAM`/`BAYAR` untuk Piutang) |
| `category_name` | STRING | Denormalized |
| `crew_id` | STRING | Foreign key ke `X_MASTER[Crew].id` (comma-separated untuk Panen) |
| `crew_name` | STRING | Denormalized |
| `amount_raw` | INTEGER | Gross income (Panen) atau nominal input |
| `amount_final` | INTEGER | Net income setelah dikurangi biaya (Panen) atau sama dengan raw |
| `weight` | INTEGER | Berat panen (Kg), khusus modul Panen |
| `unit_price` | INTEGER | Harga per Kg, khusus modul Panen |
| `labor_cost` | INTEGER | Total upah panen, khusus modul Panen |
| `transport_cost` | INTEGER | Biaya transport/timbang, khusus modul Panen |
| `notes` | STRING | Catatan opsional dari user |
| `slack_user_id` | STRING | ID user Slack yang submit |
| `slack_username` | STRING | Username Slack |
| `channel_id` | STRING | Channel tempat command dipanggil |

> **Total: 20 Kolom.** Urutan sesuai dengan `X_LOG` v1.1.0.

### 4.2 Tab: `X_MASTER` — Konfigurasi & Master Data

**Sheet: `Sites`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | STRING | Contoh: `SITE_001`, `SITE_002` |
| `name` | STRING | Contoh: `Kebun Induk`, `Kebun Plasma` |
| `location` | STRING | Deskripsi lokasi geografis |
| `status` | ENUM | `ACTIVE` / `INACTIVE` |
| `target_modal` | INTEGER | Target balik modal (Rupiah) |
| `created_at` | DATE | Tanggal kebun ditambahkan ke sistem |

**Sheet: `Categories`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | STRING | Contoh: `CAT_PUPUK`, `CAT_PRUNING` |
| `name` | STRING | Contoh: `Pupuk`, `Pruning`, `Semprot` |
| `type` | ENUM | `OPEX` / `CAPEX` / `PENDAPATAN` |
| `multiplier_enabled` | BOOLEAN | `TRUE` jika auto-multiplier aktif untuk kategori ini |
| `status` | ENUM | `ACTIVE` / `INACTIVE` |

**Sheet: `Crew`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | STRING | Contoh: `CREW_001` |
| `name` | STRING | Contoh: `Jono`, `Slamet`, `Adi` |
| `role` | STRING | Contoh: `Mandor`, `Buruh Harian` |
| `site_id` | STRING | Site utama crew (opsional, bisa multi-site) |
| `status` | ENUM | `ACTIVE` / `INACTIVE` |

### 4.3 Tab: `X_REKAP` — Rekap & Dashboard

> Tab ini diisi secara otomatis oleh server (MasterDataService) setiap kali ada perubahan data atau request report.

| Kolom | Tipe | Keterangan |
|---|---|---|
| `site_id` | STRING | ID Kebun |
| `site_name` | STRING | Nama Kebun |
| `total_weight_kg` | INTEGER | Total Produksi |
| `gross_income_rp` | INTEGER | Pendapatan Kotor |
| `opex_rp` | INTEGER | Total Biaya Operasional |
| `net_profit_rp` | INTEGER | Profit Bersih |
| `investasi_total_rp` | INTEGER | Modal Awal |
| `sisa_modal_rp` | INTEGER | Modal belum balik |
| `roi_percent` | FLOAT | ROI (%) |
| `bep_projection` | STRING | Estimasi BEP |
| `total_pinjam` | INTEGER | Total Utang Pegawai |
| `total_bayar` | INTEGER | Total Bayar Cicilan |
| `outstanding_debt` | INTEGER | Sisa Utang Beredar |
| `last_updated` | DATETIME | Timestamp Sinkronisasi |

> **Total: 14 Kolom.** Dashboard utama owner di Spreadsheet.

---

## 5. Spesifikasi API Endpoint

### 5.1 `POST /slack/events` — Slack Slash Command Handler

**Request (dari Slack):**
```
X-Slack-Request-Timestamp: <unix_timestamp>
X-Slack-Signature: v0=<hmac_sha256>
Content-Type: application/x-www-form-urlencoded

command=/sawit-x&user_id=U123ABC&channel_id=C456DEF&...
```

**Response awal (dalam 3 detik):**
```json
{ "response_type": "ephemeral", "text": "Memuat data kebun..." }
```

**Aksi selanjutnya:** trigger `views.open` ke Slack API dengan modal Block Kit.

### 5.2 `POST /slack/interactions` — Block Kit Interaction Handler

Menangani semua interaction dari modal Slack:
- `block_actions` — saat user memilih dropdown atau klik button.
- `view_submission` — saat user submit modal (data final ditulis ke `X_LOG`).
- `view_closed` — saat modal ditutup tanpa submit (no-op).

### 5.3 `GET /health` — Health Check

```json
{ "status": "ok", "version": "1.0.0", "timestamp": "2026-03-14T06:21:27Z" }
```

---

## 6. Keamanan

| Mekanisme | Detail |
|---|---|
| **Slack HMAC Verification** | Setiap request diverifikasi menggunakan `X-Slack-Signature` & `X-Slack-Request-Timestamp`. Request berumur > 5 menit langsung direject. |
| **GCP Secret Manager** | `SLACK_SIGNING_SECRET` dan `GOOGLE_CREDENTIALS` disimpan di Secret Manager, tidak di environment variable atau kode. |
| **Service Account IAM** | Service Account hanya diberikan role `roles/secretmanager.secretAccessor` dan akses terbatas ke Google Sheets API. |
| **Workload Identity** | Cloud Function menggunakan Workload Identity untuk autentikasi ke GCP tanpa file kunci JSON. |
| **Input Validation** | Semua input dari Slack divalidasi tipe dan panjangnya sebelum ditulis ke spreadsheet. |

---

## 7. Fase Eksekusi (Roadmap)

### Fase 1: Database Architecture ✅
- [x] Desain skema spreadsheet (`X_LOG`, `X_MASTER`, `X_REKAP`).
- [x] Setup Google Sheets dengan tab dan header yang benar.
- [x] Konfigurasi Service Account & IAM di GCP.
- [x] Pemberian akses Sheets ke Service Account (Editor).

### Fase 2: Backend Core (Go) ✅
- [x] Setup project struktur Go (module, handler, service, client).
- [x] Implementasi `SlackVerifier` middleware (HMAC-SHA256).
- [x] Implementasi `SheetsClient` untuk baca `X_MASTER`.
- [x] Implementasi `Dynamic Discovery` — load sites, categories, crew dari Sheets.
- [x] Implementasi `LogWriter` — append row ke `X_LOG`.
- [x] Unit test untuk semua service.

### Fase 3: Slack UX Orchestration ✅
- [x] Buat Slack App di `api.slack.com`.
- [x] Konfigurasi Slash Command `/sawit-x` mengarah ke Cloud Function URL.
- [x] Enable Interactivity & Shortcuts, set Request URL.
- [x] Implementasi **New Flow** — 3 modul terpisah: Panen, Operasional, Piutang.
- [x] Modal 1: Pilih Kebun (site_selection_modal).
- [x] Modal 2: Pilih Modul (module_selection_modal).
- [x] Modal 3a Panen: Tanggal, Multi-select Pemanen, Berat, Harga/Kg, Upah, Bensin (panen_entry_modal).
- [x] Modal 3b Operasional: Kategori, PJ, Nominal, Keterangan (operasional_entry_modal).
- [x] Modal 3c-1 Piutang: Pilih Pegawai (piutang_crew_select_modal).
- [x] Modal 3c-2 Piutang: Tampilkan saldo berjalan + Pinjam/Bayar + Nominal (piutang_action_modal).
- [x] Implementasi `GetCrewBalance` — hitung saldo piutang real-time dari X_LOG.

### Fase 4: Deployment & Observability ✅
- [x] Deploy ke GCP Cloud Functions Gen2 (region: `asia-southeast2`).
- [x] Setup Secret Manager untuk `SLACK_SIGNING_SECRET`.
- [x] End-to-end testing dengan Slack workspace nyata.

### Fase 5: Content & Dokumentasi ✅
- [x] README.md teknis untuk repository.
- [x] flow.md untuk dokumentasi interaksi sistem.
- [x] PRD.md update ke state terbaru.

### Fase 6: Advance Analytics & Dashboard ✅
- [x] Implementasi Phase 9: Investasi Awal & Proyeksi BEP.
- [x] Implementasi Phase 10: X_REKAP Synchronization.
- [x] Implementasi Phase 11: Detailed Financial Reports.
- [x] Implementasi Phase 12: UI/UX & Data Integrity Polish.

---

## 8. Struktur Repository

```
sawit-x/
├── cmd/
│   └── main.go              # Entry point Cloud Function
├── internal/
│   ├── handler/
│   │   ├── slack_events.go      # Handler /slack/events
│   │   └── slack_interactions.go # Handler /slack/interactions
│   ├── service/
│   │   ├── master_data.go       # Logic baca X_MASTER
│   │   └── log_writer.go        # Logic tulis X_LOG
│   ├── client/
│   │   └── sheets.go            # Google Sheets API client
│   ├── middleware/
│   │   └── slack_verifier.go    # HMAC verification
│   └── model/
│       ├── site.go
│       ├── category.go
│       ├── crew.go
│       └── log_entry.go
├── PRD.md
├── go.mod
├── go.sum
└── README.md
```

---

## 9. Technical Specs & Constraints

| Aspek | Detail |
|---|---|
| **Runtime** | Go 1.26 |
| **Cloud Provider** | GCP (Cloud Functions Gen2, Secret Manager) |
| **Region** | `asia-southeast2` (Jakarta) |
| **Max Execution Time** | 60 detik (Cloud Function timeout) |
| **Slack Response SLA** | < 3 detik untuk response awal (Slack hard limit) |
| **Scalability** | Stateless, serverless, auto-scaling, multi-site ready |
| **Database** | Google Sheets (tidak ada RDBMS — sesuai skala & biaya) |
| **Cost Target** | Di bawah free tier GCP selama volume < 2jt invocasi/bulan |

---

## 10. Definition of Done

Sebuah fase dianggap **selesai** jika:

1. ✅ Semua checklist pada fase tersebut sudah ter-centang.
2. ✅ End-to-end flow dapat dipanggil dari Slack nyata tanpa error.
3. ✅ Data tercatat dengan benar di Google Sheets (`X_LOG`).
4. ✅ Tidak ada secret/credential yang hardcoded di kode.
5. ✅ Kode sudah di-push ke repository (dengan commit message yang deskriptif).