# 📋 TODO Task & Spesifikasi: Aplikasi Backoffice & Sistem Bagi Hasil (Pardis Barbershop)

Dokumen ini merangkum rencana kerja, arsitektur, aturan bisnis, formula perhitungan, serta checklist tugas implementasi untuk modul/aplikasi **Backoffice POS Phoenix** dengan sistem multi-cabang, bagi hasil fleksibel, dan slip gaji Excel.

---

## 📌 Ringkasan Eksekutif & Kebutuhan Utama

1. **Aplikasi / UI Terpisah**:
   - Sistem Backoffice dibuat terpisah dari aplikasi kasir utama (Main POS) agar fokus untuk keperluan manajerial, analitik, konfigurasi finansial, dan audit tanpa mengganggu operasional kasir di outlet.
2. **Multi-Branch (2 Cabang Pardis Barbershop)**:
   - **Cabang Klaseman**
   - **Cabang Ledok**
3. **Akses Khusus Owner (`Ipang`)**:
   - Hanya **Ipang** (Owner / Superadministrator) yang memiliki hak akses untuk mengubah formula dan persentase bagi hasil per-cabang dan melihat data konsolidasi kedua cabang.
4. **Dashboard & Analitik Historis**:
   - Chart/Dashboard tren pendapatan bulanan hingga **2 tahun terakhir ke belakang (24 bulan trailing window)** dari tanggal saat ini.
   - Filter data per-cabang maupun agregat gabungan kedua cabang.
5. **Ekspor Laporan Finansial (Excel)**:
   - Download report Excel pendapatan bulanan berjalan hingga 2 tahun ke belakang.
6. **Slip Gaji Karyawan (Excel Payroll)**:
   - Fitur khusus untuk men-generate slip gaji individual format Excel bagi setiap karyawan/barberman dari cabang Klaseman dan Ledok.
7. **Aturan Finansial & Bagi Hasil**:
   - **Jasa Layanan**: Dibagi hasilkan sesuai persentase yang dikonfigurasi per-cabang.
   - **Sisa Saldo Bulanan**: Jika total persentase bagi hasil kurang dari 100% (atau terdapat sisa setelah alokasi owner & karyawan), selisihnya otomatis dicatat sebagai **Saldo Kas Tidak Terpakai / Kas Cadangan Cabang**.
   - **Produk Retail**: **Tidak masuk** ke dalam pool bagi hasil omzet. Seluruh omzet produk terpisah.
   - **Komisi Penjualan Produk**: Karyawan/barberman yang berhasil menjual produk mendapatkan komisi langsung (variatif antara **Rp 5.000 – Rp 10.000** per item).
   - **Diskon & Bundling**: Mendukung pencatatan diskon nominal/persen serta kalkulasi paket bundling (misal: Haircut + Produk Pomade).

---

## 🏢 Multi-Branch & Matriks Otorisasi

| Entitas / Aktor | Hak Akses & Wewenang | Cabang |
|---|---|---|
| **Ipang (Owner / Superadmin)** | Akses penuh 2 cabang, kustomisasi persentase bagi hasil, kelola konfigurasi komisi produk, dashboard 2 tahun, download laporan excel & payroll slip | Klaseman & Ledok |
| **Karyawan / Barberman** | Menerima komisi bagi hasil jasa bulanan + komisi penjualan produk, menerima file slip gaji | Ditugaskan di salah satu cabang (Klaseman / Ledok) |
| **Kasir / Operator Cabang** | Menginput transaksi di Main POS, memilih barberman bertugas, menerapkan diskon/bundling | Khusus cabang lokal operasional |

---

## 📐 Formula & Logika Bisnis

### 1. Pemisahan Pool Pendapatan (Revenue Segregation)
$$\text{Total Transaksi Kotor} = \text{Omzet Jasa (Services)} + \text{Omzet Produk (Retail)} - \text{Diskon}$$

- **Pool Bagi Hasil (Jasa)**: Hanya dihitung dari net pendapatan jasa layanan rambut/grooming.
- **Pool Retail (Produk)**: Dikecualikan dari bagi hasil omzet.

### 2. Alokasi Bagi Hasil Jasa Bulanan
Untuk setiap cabang $B$ pada bulan $M$:
$$\text{Pendapatan Karyawan}_i = \text{Net Omzet Jasa}_B \times \%_i$$
$$\text{Porsi Owner} = \text{Net Omzet Jasa}_B \times \%_{\text{owner}}$$
$$\%_{\text{Sisa Saldo}} = 100\% - \left( \%_{\text{owner}} + \sum \%_{\text{karyawan}} \right)$$
$$\text{Saldo Tidak Terpakai (Cadangan Cabang)} = \text{Net Omzet Jasa}_B \times \%_{\text{Sisa Saldo}}$$

### 3. Komisi Produk Retail Karyawan
$$\text{Total Komisi Produk}_i = \sum (\text{Qty Terjual}_p \times \text{Rate Komisi}_p)$$
*Rate komisi produk bervariasi Rp 5.000 - Rp 10.000 sesuai jenis produk yang dijual.*

### 4. Total Gaji Bersih Karyawan (Take Home Pay)
$$\text{Total Gaji Bersih}_i = \text{Bagi Hasil Jasa}_i + \text{Total Komisi Produk}_i + \text{Bonus/Penyesuaian} - \text{Potongan}$$

---

## 📊 Simulasi Perhitungan Riil

### Cabang Klaseman (Bulan Januari)
- **Omzet Jasa**: Rp 10.000.000
- **Konfigurasi Bagi Hasil**:
  - Karyawan 1 (35%): **Rp 3.500.000**
  - Karyawan 2 (35%): **Rp 3.500.000**
  - Porsi Owner Ipang (20%): **Rp 2.000.000**
  - **Sisa Saldo Tidak Terpakai (10%)**: **Rp 1.000.000** *(disimpan sebagai kas cadangan cabang Klaseman)*
- **Penjualan Produk (Contoh)**:
  - Karyawan 1 menjual 10 pcs Pomade (komisi Rp 5.000/pcs): **Rp 50.000**
  - Karyawan 2 menjual 5 pcs Tonic (komisi Rp 10.000/pcs): **Rp 50.000**
- **Take Home Pay**:
  - Karyawan 1: Rp 3.500.000 + Rp 50.000 = **Rp 3.550.000**
  - Karyawan 2: Rp 3.500.000 + Rp 50.000 = **Rp 3.550.000**

---

### Cabang Ledok (Bulan Januari)
- **Omzet Jasa**: Rp 12.000.000
- **Konfigurasi Bagi Hasil**:
  - Karyawan 1 (40%): **Rp 4.800.000**
  - Karyawan 2 (30%): **Rp 3.600.000**
  - Porsi Owner Ipang (20%): **Rp 2.400.000**
  - **Sisa Saldo Tidak Terpakai (10%)**: **Rp 1.200.000** *(disimpan sebagai kas cadangan cabang Ledok)*
- **Penjualan Produk (Contoh)**:
  - Karyawan 1 menjual 8 pcs Hair Clay (komisi Rp 7.500/pcs): **Rp 60.000**
- **Take Home Pay**:
  - Karyawan 1: Rp 4.800.000 + Rp 60.000 = **Rp 4.860.000**
  - Karyawan 2: **Rp 3.600.000**

---

## 🗂️ Rencana Checklist Tugas (TODO Checklist)

### Milestone 1: Arsitektur & Database Modeling
- [ ] Buat migration tabel `branches`:
  - `id`, `code` (`KLASEMAN`, `LEDOK`), `name`, `address`, `is_active`, `created_at`
- [ ] Perbarui tabel `users` / `operators`:
  - Tambahkan `branch_id` (foreign key ke `branches`)
  - Kolom tipe staf (`barberman`, `cashier`, `manager`, `owner`)
- [ ] Perbarui tabel `catalog_items` / `products`:
  - Tambahkan flag `item_type` (`SERVICE` vs `PRODUCT`)
  - Tambahkan `commission_amount` (integer Rp 5.000 - Rp 10.000 untuk item produk)
- [ ] Buat tabel `branch_profit_sharing_rules`:
  - `id`, `branch_id`, `period_month` (format `YYYY-MM`), `owner_percentage`, `unallocated_percentage`, `updated_by` (`ipang`), `updated_at`
- [ ] Buat tabel `employee_profit_sharing_rules`:
  - `id`, `rule_id`, `user_id`, `percentage`, `created_at`
- [ ] Buat tabel `discounts_and_bundles`:
  - `id`, `code`, `name`, `type` (`PERCENTAGE`, `FIXED_AMOUNT`, `BUNDLE`), `value`, `service_allocation_ratio`, `product_allocation_ratio`, `is_active`
- [ ] Perbarui tabel `transactions` & `transaction_items`:
  - Tambahkan `branch_id`
  - Tambahkan `barber_id` / `served_by_id` pada tiap baris item
  - Tambahkan `discount_amount`, `bundle_id`, `commission_earned`

---

### Milestone 2: Core Calculation Engine (Go Backend)
- [ ] Buat paket `internal/backoffice/calculator`:
  - Logika pemisahan omzet Jasa vs Retail
  - Perhitungan persentase bagi hasil bulanan per-cabang
  - Kalkulasi alokasi sisa saldo tidak terpakai
  - Kalkulasi komisi produk per-karyawan berdasarkan item yang dijual
  - Kalkulasi penyesuaian diskon & bundling terhadap pendapatan jasa
- [ ] Buat validasi otorisasi:
  - Hanya user dengan role `superadmin` / email `ipang@example.com` yang diizinkan mengubah konfigurasi persentase bagi hasil
- [ ] Buat unit test lengkap untuk skenario:
  - Pembagian hasil Klaseman (Januari 10jt, 35%-35%-20%-10%)
  - Pembagian hasil Ledok (Januari 12jt, 40%-30%-20%-10%)
  - Pengecualian produk dari bagi hasil dan akurasi komisi produk

---

### Milestone 3: Backoffice UI & Navigasi Terpisah
- [ ] Pisahkan antarmuka Backoffice:
  - Buat layout & route terpisah `/backoffice/*` (atau sub-aplikasi terpisah) yang terlindungi middleware khusus Owner
- [ ] Halaman **Konfigurasi Bagi Hasil Cabang** (`/backoffice/profit-sharing`):
  - Selector Cabang (Klaseman / Ledok) & Periode Bulan
  - Form input persentase Owner, Karyawan 1, Karyawan 2, dsb.
  - Indikator real-time kalkulasi sisa saldo tidak terpakai (Total persentase & nominal saldo cadangan)
- [ ] Halaman **Manajemen Diskon & Bundling** (`/backoffice/discounts`):
  - Input promo diskon & paket bundling
- [ ] Halaman **Katalog & Komisi Produk** (`/backoffice/products`):
  - Pengaturan nilai komisi per produk (Rp 5.000 - Rp 10.000)

---

### Milestone 4: Dashboard & Chart Pendapatan Historis (2 Tahun)
- [ ] Buat query agregasi bulanan untuk 24 bulan ke belakang dari tanggal hari ini (`CURRENT_DATE` trailing window)
- [ ] Endpoint API `/backoffice/api/analytics/trend-24m`:
  - Parameter filter: `branch_id` (`all`, `klaseman`, `ledok`)
  - Output JSON: data 24 bulan (omzet kotor, omzet jasa, omzet produk, diskon, sisa saldo cadangan)
- [ ] Visualisasi Chart di Backoffice Dashboard:
  - Interaktif: beralih tampilan antara Klaseman, Ledok, dan Konsolidasi 2 Cabang
  - Tooltip interaktif rincian omzet jasa vs produk

---

### Milestone 5: Ekspor Laporan Finansial (Excel)
- [ ] Service generator Excel (`excelize` di Go) untuk laporan finansial:
  - Range: Bulan berjalan hingga 24 bulan ke belakang
  - Sheet 1: Rangkuman Konsolidasi 2 Cabang
  - Sheet 2: Rincian Cabang Klaseman (Omzet, Bagi Hasil, Sisa Kas Cadangan)
  - Sheet 3: Rincian Cabang Ledok (Omzet, Bagi Hasil, Sisa Kas Cadangan)
  - Sheet 4: Rincian Penjualan Produk & Komisi
- [ ] Tombol Download di UI Backoffice:
  - Filter rentang bulan/tahun fleksibel

---

### Milestone 6: Generator Slip Gaji Karyawan (Excel Payroll Slip)
- [ ] Service generator Slip Gaji individual format Excel (`.xlsx`):
  - Template resmi Pardis Barbershop (Logo, Nama Cabang, Nama Karyawan, Periode)
  - Bagian Rincian Bagi Hasil Jasa (Omzet Cabang, Persentase, Nominal)
  - Bagian Rincian Komisi Produk (Daftar produk terjual, Qty, Rate komisi, Subtotal)
  - Bagian Ringkasan Gaji Bersih (Take Home Pay) & Tanda Tangan
- [ ] Halaman UI Slip Gaji (`/backoffice/payroll`):
  - Filter Cabang & Bulan
  - Tabel rekap karyawan beserta nominal gaji bersih
  - Tombol **"Download Slip Gaji Excel"** per-karyawan
  - Tombol **"Download Semua Slip Gaji (ZIP / Multi-Sheet Excel)"** untuk seluruh karyawan cabang terkait

---

### Milestone 7: Verifikasi, Pengujian E2E & Dokumentasi
- [ ] Unit Test Go:
  - Pengujian kalkulasi bagi hasil, sisa saldo, dan komisi produk
  - Pengujian validasi otorisasi eksklusif Ipang
  - Pengujian pembuatan file Excel laporan dan slip gaji
- [ ] Automated E2E Test (Playwright):
  - Skenario login Owner -> masuk Backoffice -> ubah persentase bagi hasil Klaseman & Ledok
  - Skenario download laporan Excel 24 bulan
  - Skenario generate slip gaji karyawan
- [ ] Pembaruan panduan deployment dan dokumentasi API Backoffice
