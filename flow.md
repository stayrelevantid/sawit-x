# SAWIT-X: The Ultimate System Flow

> *"Scaling the plantation, automating the ledger, staying relevant."*

---

## 1. Trigger & Discovery (The Startup)

Setiap interaksi dimulai dengan memastikan bot memiliki konteks data terbaru tanpa hard-coded di dalam aplikasi.

**User Action:** Mengetik `/sawit-x` di Slack.

**System Logic:**
1. GCF melakukan `FetchMasterData` dari tab `X_MASTER`.
2. Mengambil **List Dinamis**: Kebun, Kategori Biaya, dan Nama Pegawai.

**UI Slack:** Menampilkan menu pilihan → *"Pilih Lokasi Kebun"*.

---

## 2. Modul Panen (Multi-Worker & Logistics)

Digunakan untuk mencatat hasil produksi dengan detail biaya logistik yang transparan.

**UI Slack (Modal Form):**

| Field | Tipe Input | Keterangan |
|-------|-----------|------------|
| Tanggal | Date Picker | Default: Hari ini |
| Pemanen | Multi-select dropdown | Pilih semua pegawai yang terlibat |
| Berat (Kg) | Number input | Berat hasil panen |
| Harga per Kg | Number input | Harga jual per kilogram |
| Upah Panen | Number input | Biaya labor |
| Bensin/Timbang | Number input | Biaya transport |

**Backend Logic:**

```
Gross_Income = Berat × Harga
Net_Income   = Gross_Income - (Upah + Bensin)
```

**Storage:** Menulis ke `X_LOG` → Mencatat total net dan rincian logistik.

---

## 3. Modul Operasional (Expense with Accountability)

Digunakan untuk mencatat setiap pengeluaran kebun dengan penanggung jawab yang jelas.

**UI Slack (Modal Form):**

| Field | Tipe Input | Keterangan |
|-------|-----------|------------|
| Kategori Biaya | Dropdown dinamis | Pupuk, Bensin, Pruning, dll |
| Penanggung Jawab | Dropdown pegawai | Siapa yang belanja/pegang uang |
| Nominal | Number input | Mendukung normalisasi ribuan |
| Keterangan | Text input | Misal: "Beli NPK 12-12-17" |

**Storage:** Menulis ke `X_LOG` → Kolom Kredit terisi, melacak pengeluaran per personil.

---

## 4. Modul Piutang (Employee Debt Management)

Digunakan untuk mengelola pinjaman atau pembayaran utang pegawai.

**UI Slack:**

| Field | Tipe Input | Keterangan |
|-------|-----------|------------|
| Person | Dropdown pegawai | Pilih nama pegawai |
| Action | Tombol | `[Pinjam]` atau `[Bayar/Potong]` |

**System Logic:**
- Bot mengambil **saldo berjalan** pegawai tersebut dari Sheets.
- Menampilkan saldo sebagai info di Slack **sebelum** user menginput angka baru.

**Storage:** Menulis ke `X_LOG` dengan kategori `"Utang"`.

---

## 5. Reporting (Dashboard Visualization)

**Action:** Klik tombol `[Lihat Rekap]`.

**Output:** Bot melakukan agregasi data dan menampilkan ringkasan performa kebun:

| Metrik | Deskripsi |
|--------|-----------|
| Total Produksi | Jumlah hasil panen (Kg) |
| Operational Cost | Total biaya & upah |
| Net Profit | Laba bersih |
| ROI Tracking | Sisa target balik modal |

---

## 6. Skema Database Final (Google Sheets)

| Tab | Fungsi |
|-----|--------|
| `X_MASTER` | Config — data kebun, kategori, crew |
| `X_LOG` | Database utama — semua transaksi |
| `X_REKAP` | Rekap otomatis — ringkasan performa |