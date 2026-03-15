# 🌴 SAWIT-X: System Flow
> *"Scaling the plantation, automating the ledger, staying relevant."*

---

## 1. Arsitektur Endpoint

| Endpoint | Method | Handler | Fungsi |
|---|---|---|---|
| `/health` | GET | `healthHandler` | Health check |
| `/slack/events` | POST | `SlackEventsHandler.HandleCommand` | Menerima slash command `/sawit-x` |
| `/slack/interactions` | POST | `SlackInteractionsHandler.HandleInteraction` | Menerima interaksi modal |

---

## 2. Flow Interaksi Lengkap

### Step 1 — Trigger: `/sawit-x`

**User Action:** Mengetik `/sawit-x` di Slack.

**System Logic:**
1. Slack kirim `POST /slack/events`.
2. `SlackVerifier` memverifikasi signature HMAC-SHA256 — request ditolak jika > 5 menit.
3. Server langsung respon **200 OK** (menghindari timeout 3 detik Slack).
4. Background goroutine: fetch daftar kebun aktif dari tab `Sites` di `X_MASTER`.
5. Buka **Modal 1: Pilih Lokasi Kebun**.

---

### Step 2 — Modal 1: Pilih Kebun (`site_selection_modal`)

**UI Slack:**
- Header: *"Pilih Lokasi Kebun"*
- Dropdown: Daftar kebun aktif dari `X_MASTER[Sites]` (hanya `Status == ACTIVE`)
- Tombol: **[Lanjut]** | **[Batal]**

**System Logic:**
1. Ambil `site_id` dan `site_name` dari pilihan user.
2. Simpan ke `PrivateMetadata` sebagai `TransactionState`.
3. Respon: **Update view** → tampilkan **Modal 2 (Mode Selection)**.

---

### Step 3 — Modal 2: Pilih Mode (`mode_selection_modal`)

**UI Slack:**
- Header: *"Kebun: [Nama Kebun]"*
- Action Buttons:
  - **[✍️ Pencatatan Baru]**: Masuk ke menu modul (Panen/Ops/Piutang).
  - **[📊 Lihat Rekap]**: Masuk ke dashboard performa.

---

### Step 4 — Modal 3: Pilih Modul (`module_selection_modal`)

*(Muncul jika user memilih "Pencatatan Baru" di Modal 2)*

**UI Slack:**
- Dropdown: Pilih jenis pencatatan

| Value | Label | Deskripsi |
|---|---|---|
| `PANEN` | 🌾 Panen | Catat hasil panen dan biaya logistik |
| `OPERASIONAL` | 💰 Operasional | Catat pengeluaran kebun |
| `PIUTANG` | 📋 Piutang | Kelola pinjaman pegawai |

**System Logic:** Buka modal form sesuai modul yang dipilih.

---

### Step 5a — Modul Panen (`panen_entry_modal`)

**UI Slack (Modal Form):**

| Field | Tipe Input | Keterangan |
|---|---|---|
| Tanggal | Date Picker | Default: hari ini |
| Pemanen | Multi-select dropdown | Pilih semua pegawai yang terlibat |
| Berat (Kg) | Number input | Berat hasil panen |
| Harga per Kg | Number input | Harga jual per kilogram |
| Upah Panen | Number input (opsional) | Biaya labor |
| Bensin/Timbang | Number input (opsional) | Biaya transport |
| Catatan | Text input (opsional) | Keterangan tambahan |

**Kalkulasi Backend:**
```
Gross_Income = Berat × Harga_per_Kg
Net_Income   = Gross_Income − (Upah_Panen + Bensin)
```

**Storage:** `X_LOG` — `module_type = PANEN`, `amount_raw = Gross_Income`, `amount_final = Net_Income`.

---

### Step 4b — Modul Operasional (`operasional_entry_modal`)

**UI Slack (Modal Form):**

| Field | Tipe Input | Keterangan |
|---|---|---|
| Tanggal | Date Picker | Default: hari ini |
| Kategori Biaya | Dropdown dinamis | Dari `X_MASTER[Categories]` |
| Penanggung Jawab | Dropdown pegawai | Dari `X_MASTER[Crew]` |
| Nominal | Number input | Nominal pengeluaran (Rupiah) |
| Keterangan | Text input (opsional) | Misal: *"Beli NPK 12-12-17"* |

**Storage:** `X_LOG` — `module_type = OPERASIONAL`, kolom kredit terisi.

---

### Step 5c — Modul Piutang (2 langkah)

#### Langkah 1: Pilih Pegawai (`piutang_crew_select_modal`)

**UI Slack:**
- Dropdown: Pilih nama pegawai dari `X_MASTER[Crew]`
- Tombol: **[Cek Saldo]**

**System Logic:** Fetch saldo berjalan pegawai dari `X_LOG` via `GetCrewBalance()`.

#### Langkah 2: Aksi Piutang (`piutang_action_modal`)

**UI Slack:**

| Field | Tipe Input | Keterangan |
|---|---|---|
| Info Saldo | Section (read-only) | Menampilkan saldo berjalan pegawai |
| Tanggal | Date Picker | Default: hari ini |
| Aksi | Dropdown | `[💸 Pinjam]` atau `[✅ Bayar / Potong]` |
| Nominal | Number input | Jumlah pinjaman atau pembayaran |
| Keterangan | Text input (opsional) | Catatan opsional |

**Storage:** `X_LOG` — `module_type = PIUTANG`, `category_id = PINJAM` atau `BAYAR`.

---

### Step 6 — Konfirmasi & DM

Setelah submit tiap modul:
1. Respon Slack: **Clear modal** (modal tertutup otomatis).
2. Background: Bot kirim **DM konfirmasi** ke user.

**Contoh DM Panen:**
```
✅ Data Berhasil Dicatat!

Kebun: Kebun Induk
Pemanen: Jono, Slamet
Berat: 1250 Kg
Gross: Rp3.000.000
Net:   Rp2.750.000
```

---

## 3. Skema Callback Modal

```
/sawit-x
    │
    ▼
site_selection_modal      ← Pilih Kebun
    │
    ▼
module_selection_modal    ← Pilih Modul
    │
    ├── PANEN      ──▶ panen_entry_modal
    │                        └── WriteLog (X_LOG)
    │
    ├── OPERASIONAL ──▶ operasional_entry_modal
    │                        └── WriteLog (X_LOG)
    │
    └── PIUTANG    ──▶ piutang_crew_select_modal
                             │  GetCrewBalance (X_LOG)
                             ▼
                        piutang_action_modal
                             └── WriteLog (X_LOG)
```

---

## 4. Keamanan

| Mekanisme | Detail |
|---|---|
| **Slack HMAC Verification** | Setiap request diverifikasi via `X-Slack-Signature`. Request > 5 menit direject. |
| **Environment Variables** | Credentials disimpan di environment variable Cloud Functions (tidak hardcoded). |
| **Input Validation** | Nominal divalidasi tipe dan sign — input negatif atau non-integer ditolak inline. |