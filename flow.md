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
- Tombol: **[Lanjut]** | **[Batal]** | **[📦 Lihat Arsip]** *(Aksi: `btn_view_archived_sites`)*

**System Logic:**
1. Jika memilih kebun aktif & klik **[Lanjut]**: Ambil `site_id` dan `site_name` dari pilihan user, simpan ke `PrivateMetadata`, update view ke **Modal 2 (Mode Selection)**.
2. Jika klik **[📦 Lihat Arsip]**: Ambil daftar kebun berstatus `ARCHIVED`, update view ke **Modal Pilih Kebun Terarsip (`archived_site_selection_modal`)**.

---

### Step 3 — Modal 2: Pilih Mode (`mode_selection_modal`)

**UI Slack:**
- Header: *"🌴 Kebun: [Nama Kebun]"*
- Grouping Menu Terstruktur:
  - **✍️ Transaksi Baru:**
    - **[✍️ Pencatatan Baru]** *(Tombol utama hijau / Primary)*: Masuk ke menu modul (Panen/Ops/Piutang/Investasi).
  - **📊 Rekap & Keuangan:**
    - **[📊 Lihat Rekap Kebun]**: Masuk ke dashboard rekap performa keuangan kebun.
    - **[📋 Rekap Hutang Pegawai]**: Melihat rincian saldo piutang pegawai beserta tanggal kasbon/bayar terakhir dan shortcut riwayat lengkap.
    - **[📜 List Lengkap Kasbon]**: Melihat rincian seluruh mutasi pinjaman & pembayaran kasbon (tanggal, pegawai, tipe, nominal, sisa saldo, dan catatan).
  - **🌾 Riwayat Hasil Panen:**
    - **[📅 Panen 1 Tahun Ini]**: Melihat daftar riwayat panen tahun berjalan lengkap dengan berat, harga per Kg, dan pendapatan net.
    - **[📅 Panen 1 Tahun Lalu]**: Melihat daftar riwayat panen tahun sebelumnya lengkap dengan berat, harga per Kg, dan pendapatan net.
  - **🌿 Perawatan Kebun:**
    - **[🧪 List Pembelian Pupuk]**: Melihat daftar riwayat pembelian pupuk dan total akumulasi.
    - **[🌧️ List Penyemprotan]**: Melihat daftar riwayat pengeluaran operasional penyemprotan.
  - **⚙️ Pengaturan Kebun:**
    - **[📦 Arsipkan Kebun]** *(Tombol merah / Danger, aksi: `btn_archive_site`)*: Mengarsipkan kebun jika sudah dijual/nonaktif sehingga tidak ada mutasi baru.

---

### Step 4 — Modal 3: Pilih Modul (`module_selection_modal`)

*(Muncul jika user memilih "Pencatatan Baru" di Modal 2)*

**UI Slack:**
- Dropdown: Pilih jenis pencatatan

| `PIUTANG` | 📋 Piutang | Kelola pinjaman pegawai |
| `INVESTASI` | 🚀 Investasi | Catat modal balik / pembelian lahan |

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
| Kategori Biaya | Dropdown dinamis | Dari `X_MASTER[Categories]` (hanya tipe `OPEX`) |
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

### Step 5d — Modul Investasi (`investasi_entry_modal`)

**UI Slack (Modal Form):**

| Field | Tipe Input | Keterangan |
|---|---|---|
| Judul | Read-only | *"Set Modal Awal"* (jika 0) atau *"Update Investasi"* |
| Nominal Modal | Number input | Disarankan nominal total investasi lahan/aset |

**System Logic:** 
1. Update kolom `target_modal` di sheet `Sites`.
2. Catat audit log di `X_LOG` dengan `module_type = INVESTASI`.

---

### Step 6 — Konfirmasi & DM

Setelah submit tiap modul:
1. Respon Slack: **Clear modal** (modal tertutup otomatis).
2. Background: Bot kirim **DM konfirmasi** ke user.

**Contoh DM Panen:**
```
✅ Data Berhasil Dicatat!

Tanggal: 18 Mar 2026
Kebun: Kebun Induk
Pemanen: Jono, Slamet
Berat: 1250 Kg
Gross: Rp3.000.000
Net:   Rp2.750.000
```

3. Response Slack: **Clear modal** (modal tertutup otomatis).
4. Bot mengirimkan **Pesan Konfirmasi (Success DM)** ke user berisi ringkasan data yang baru dimasukkan + perhitungan otomatisnya.

---

### Step 7 — Rekap Performa (`view_report` action)

*(Muncul jika user memilih "Lihat Rekap" di Modal 2)*

**Output 1: Modal Dashboard**
- Header: *"Kebun: [Nama Kebun]"*
- **Seksi Panen**: Total Berat (Kg), Gross Income, Total Upah, Total Transport.
- **Seksi Operasional**: Biaya Ops Mandiri, Total Pengeluaran.
- **Seksi Utang**: Total Pinjam, Total Bayar, Utang Beredar.
- **Seksi Finansial**: Profit Akumulasi, Sisa Modal, ROI %, Proyeksi BEP.
- **Detail Perhitungan**: Breakdown rumus matematika (Gross - Biaya = Net).

**Output 2: X_REKAP Sync**
Sesuai request user, setiap kali report dibuka atau transaksi dicatat, bot melakukan sinkronisasi otomatis ke tab `X_REKAP` di Spreadsheet untuk dashboard permanen.
Bot mengirimkan pesan (DM) ke user dengan ringkasan yang sama sebagai arsip/history di chat.

---

### Step 8 — Flow Arsipkan Kebun (`btn_archive_site`)

*(Muncul jika user memilih "📦 Arsipkan Kebun" di Modal 2)*

**UI Slack (Modal Konfirmasi):**
- Callback ID: `archive_site_confirm_modal`
- Title: *"Konfirmasi Arsip Kebun"*
- Pesan peringatan: Menginfokan bahwa kebun yang diarsipkan tidak bisa menerima transaksi/pencatatan baru, namun seluruh data historis tetap bisa dilihat secara read-only.
- Tombol: **[Ya, Arsipkan]** (Danger) | **[Batal]**

**System Logic:**
1. User menekan submit `archive_site_confirm_modal`.
2. Backend memanggil `ArchiveSite(ctx, siteID)`.
3. Mengupdate baris kebun di sheet `Sites`:
   - Kolom D (`status`): diubah menjadi `ARCHIVED`
   - Kolom G (`archived_at`): diisi timestamp saat ini (format `YYYY-MM-DD HH:MM:SS`)
4. Mengirimkan pesan konfirmasi DM ke user dan notifikasi pengumuman ke Slack channel.

---

### Step 9 — Flow Akses Kebun Terarsip (Read-Only)

*(Dipicu dengan menekan "📦 Lihat Arsip" pada Modal 1)*

**UI Slack:**
1. **Modal Daftar Kebun Terarsip (`archived_site_selection_modal`)**:
   - Menampilkan dropdown daftar kebun yang berstatus `ARCHIVED`.
   - Jika tidak ada kebun terarsip, menampilkan empty-state modal pemberitahuan.
2. **Modal Mode Read-Only (`mode_selection_modal`)**:
   - Setelah memilih kebun terarsip, terbuka menu dashboard kebun.
   - **Pencegahan Transaksi Baru**: Tombol `[✍️ Pencatatan Baru]` dan `[📦 Arsipkan Kebun]` **disembunyikan**.
   - User tetap memiliki akses penuh ke fitur reporting & historis: `[📊 Lihat Rekap Kebun]`, `[📋 Rekap Hutang Pegawai]`, `[📜 List Lengkap Kasbon]`, `[📅 Panen 1 Tahun Ini/Lalu]`, dan `[🌿 Perawatan Kebun]`.

---

## 3. Skema Callback Modal

```
/sawit-x
    │
    ▼
site_selection_modal ──(btn_view_archived_sites)──▶ archived_site_selection_modal
    │                                                        │
    │ (Pilih kebun aktif)                                    │ (Pilih kebun arsip)
    ▼                                                        ▼
mode_selection_modal (Aktif)                        mode_selection_modal (Read-Only)
    │                                                        │
    ├── [✍️ Pencatatan Baru]                                  └── Fitur Rekap & Riwayat Saja
    │        ▼                                                   (Pencatatan & Arsip disembunyikan)
    │   module_selection_modal
    │        ├── PANEN       ──▶ panen_entry_modal
    │        │                        └── WriteLog (X_LOG)
    │        ├── OPERASIONAL ──▶ operasional_entry_modal
    │        │                        └── WriteLog (X_LOG)
    │        ├── PIUTANG     ──▶ piutang_crew_select_modal
    │        │                        │  GetCrewBalance (X_LOG)
    │        │                        ▼
    │        │                   piutang_action_modal
    │        │                        └── WriteLog (X_LOG)
    │        └── INVESTASI   ──▶ investasi_entry_modal
    │                                 └── WriteLog (X_LOG)
    │
    ├── [📊 Rekap & Riwayat] ──▶ view_report, panen, kasbon, pupuk, etc.
    │
    └── [📦 Arsipkan Kebun] (btn_archive_site)
             ▼
        archive_site_confirm_modal
             ▼
        ArchiveSite (Sites: status=ARCHIVED, archived_at=now)
```

---

## 4. Keamanan

| Mekanisme | Detail |
|---|---|
| **Slack HMAC Verification** | Setiap request diverifikasi via `X-Slack-Signature`. Request > 5 menit direject. |
| **Environment Variables** | Credentials disimpan di environment variable Cloud Functions (tidak hardcoded). |
| **Input Validation** | Nominal divalidasi tipe dan sign — input negatif atau non-integer ditolak inline. |