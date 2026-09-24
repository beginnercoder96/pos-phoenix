# 📋 POS Phoenix — Master TODO & Roadmap

Dokumen ini merangkum seluruh rencana pengembangan, arsitektur, checklist fitur Backoffice, serta peningkatan desain UI dan pemisahan modul transaksi.

---

## 🚀 Fitur Utama & Prioritas Saat Ini

### 1. UI Revamp Menggunakan `taste-skill` & `antislop` (Adopsi dari `posv2`)
- **Sumber Referensi**: `/Users/mekari/Documents/newlearn/posv2` (`.agents/skills/taste-skill`, `.agents/rules/antislop.md`, `antislop-ui`, `antislop-layoutmobile`, `antislop-human`, dll.)
- **Tujuan**:
  - Tampilan visual modern, bersih, profesional, dan bebas dari AI visual slop (tidak ada border tebal blur, tidak ada teks abu-abu rendah kontras, tidak ada metric palsu).
  - Tipografi monospaced/tabular (`tabular-nums`) untuk semua angka uang Rupiah, kuantiti, dan persentase.
  - Micro-interactions yang cepat (150ms-250ms) dan responsif.
  - Layout optimal untuk iPad Pro (1920x1080 & 1024x768) serta target sentuh minimal 44x44px pada Mobile/Android Chrome.
  - Rasio kontras tinggi sesuai standar WCAG AAA/AA.

### 2. Pemisahan Menu Transaksi dari Dashboard Utama
- **Tujuan**:
  - **Dashboard (`/`)**: Difokuskan murni untuk ringkasan eksekutif (Executive Overview) — Kartu KPI (Pemasukan, Pengeluaran, Saldo Kas), jam operasional/shift harian, grafik tren, dan shortcut cepat.
  - **Menu Transaksi / Kasir (`/transactions`)**: Menu terpisah khusus kasir — Form entri multi-kategori layanan/produk, pemilihan barberman bertugas, pemilihan diskon/bundling, kalkulasi instan, receipt preview/print, serta tabel mutasi riwayat transaksi harian dan fitur reversal.
  - Navigasi sidebar diperbarui:
    - 📊 **Dashboard** (`/`)
    - 💳 **Transaksi / Kasir** (`/transactions`)
    - 👥 **Operator** (`/operators` - Superadmin)
    - 🏢 **Backoffice** (`/backoffice` - Superadmin)

### 3. Sistem Backoffice, Multi-Cabang & Bagi Hasil Fleksibel
- **Multi-Cabang**: Cabang Klaseman & Cabang Ledok.
- **Akses Khusus Owner (Ipang)**: Hanya owner yang dapat mengonfigurasi persentase bagi hasil bulanan dan melihat laporan konsolidasi kedua cabang.
- **Formula Bagi Hasil**:
  - Jasa dibagi hasilkan berdasarkan persentase (Owner + Karyawan).
  - Sisa saldo bulanan otomatis dicatat terpisah sebagai **Saldo Kas Tidak Terpakai / Cadangan Cabang**.
  - Penjualan produk tidak masuk bagi hasil omzet.
  - Komisi penjualan produk (Rp 5.000 – Rp 10.000/item) langsung dialokasikan ke karyawan yang menjual.
- **Dashboard & Analitik 2 Tahun**: Chart tren historis 24 bulan ke belakang trailing window.
- **Ekspor Excel & Slip Gaji**: Download report finansial bulanan/tahunan dan slip gaji karyawan individual (Excel `.xlsx`).

### 4. Simplifikasi Login & 2FA Google Authenticator (TOTP RFC 6238)
- **Login Sederhana**: Form login rutin kasir/operator hanya meminta **Username** dan **Password**.
- **Email Khusus Initial Setup**: Input email hanya diwajibkan saat bootstrapping akun superadmin pertama kali di sistem POS.
- **OTP Google Authenticator per User**:
  - Setelah username & password terverifikasi, sistem meminta 6 digit kode OTP dari Google Authenticator.
  - Setiap user memiliki secret key unik masing-masing di database, sehingga kode OTP antar user berbeda dan independen.
  - Alur aktivasi 2FA: Pengguna pertama kali login diarahkan ke halaman setup yang menampilkan QR Code unik + manual secret key untuk di-scan dengan Google Authenticator, lalu memverifikasi 1 kode awal.

### 5. Sistem Product Tour Sinematik (Storyteller Live Caption & Camera Movement)
- **CSS Caption Independen**: File CSS terpisah (`product_tour_caption.css`) yang hanya dipanggil saat proses pembuatan product tour dan **sama sekali tidak disertakan** di bundle CSS production maupun testing rutin.
- **Storyteller Live Caption**: Teks narasi dinamis yang mendeskripsikan konteks fitur dan alur kerja aplikasi secara manusiawi (seperti presenter/storyteller menjelaskan langsung).
- **Efek Gerakan Kamera Dinamis**: Efek *camera movement* (zoom-in terarah pada elemen fokus dan zoom-out kembali ke full view) untuk memberikan pengalaman visual yang lebih profesional dan menarik.

---

## 🗂️ Detailed Roadmap & Checklist

Lihat spesifikasi lengkap, formula matematis, dan simulasi riil di file [BACKOFFICE_TODO.md](BACKOFFICE_TODO.md).

### Milestone 1: Arsitektur & Database Modeling
- [x] Buat migration tabel `branches` (`id`, `code`, `name`, `address`, `is_active`)
- [x] Perbarui tabel `users`: tambah `branch_id` dan `staff_type` (`barberman`, `cashier`, `manager`, `owner`)
- [x] Perbarui tabel `catalog_items` / `products`: tambah `item_type` (`SERVICE` vs `PRODUCT`) dan `commission_amount`
- [x] Buat tabel `branch_profit_sharing_rules` & `employee_profit_sharing_rules`
- [x] Buat tabel `discounts_and_bundles`
- [x] Perbarui tabel `transactions` & `transaction_items`: tambah `branch_id`, `barber_id`, `discount_amount`, `bundle_id`, `commission_earned`

### Milestone 2: Core Calculation Engine (Go Backend)
- [x] Paket `internal/backoffice/calculator` (Pemisahan Jasa vs Produk, Bagi Hasil, Sisa Cadangan, Komisi)
- [x] Validasi otorisasi eksklusif Superadmin / Owner (`ipang`)
- [x] Unit test perhitungan skenario riil Klaseman & Ledok

### Milestone 3: Backoffice UI & Navigasi
- [x] Route terpisah `/backoffice/*`
- [x] Halaman Bagi Hasil Cabang (`/backoffice/profit-sharing`)
- [x] Halaman Diskon & Bundling (`/backoffice/discounts`)
- [x] Halaman Produk & Komisi (`/backoffice/products`)

### Milestone 4: Dashboard & Chart Pendapatan Historis (2 Tahun)
- [x] Query agregasi bulanan 24 bulan trailing window
- [x] Endpoint API `/backoffice/api/analytics/trend-24m`
- [x] Visualisasi Chart interaktif multi-cabang & konsolidasi

### Milestone 5: Ekspor Laporan Finansial (Excel)
- [x] Service generator Excel (`excelize`) laporan konsolidasi & per-cabang
- [x] Tombol download & filter periode di UI

### Milestone 6: Generator Slip Gaji Karyawan (Excel Payroll Slip)
- [x] Service template resmi Slip Gaji Excel per-karyawan
- [x] Halaman UI `/backoffice/payroll` & download single/bulk ZIP

### Milestone 7: Verifikasi & Testing
- [ ] Unit Test Go & E2E Playwright test
- [ ] Dokumentasi panduan operasional

### Milestone 8: UI Revamp Menggunakan Taste-Skill & Anti-Slop (Adopsi dari `posv2`)
- [ ] Audit & sinkronisasi pedoman `taste-skill` dan `antislop` dari `posv2` (`.agents/skills/taste-skill`, `.agents/rules/antislop.md`)
- [ ] Sinkronisasi styling `tailwind.config.js` dan `web/static/css/input.css` dengan token desain `posv2`
- [ ] Perbarui sidebar desktop `pos-sidebar`, header, card surfaces, dan badge status
- [ ] Optimasi responsivitas iPad Pro dan touch target mobile (min 44x44px)
- [ ] Audit aksesibilitas rasio kontras warna (WCAG compliance)

### Milestone 9: Pemisahan Menu Transaksi dari Dashboard Utama
- [ ] Buat route `GET /transactions` dan template `web/templates/transactions.html`
- [ ] Pindahkan form entri kasir, kalkulator, dan riwayat mutasi dari dashboard ke menu transaksi
- [ ] Sederhanakan `dashboard.html` khusus untuk metrik eksekutif, status shift, dan tren
- [ ] Update navigasi sidebar dan mobile header: Dashboard, Transaksi / Kasir, Operator, Backoffice
- [ ] Tambahkan unit test HTTP untuk `/transactions` dan update Playwright product tour script

### Milestone 10: Simplifikasi Login & 2FA Google Authenticator (TOTP RFC 6238)
- [ ] Database migration: tambahkan `totp_secret` (TEXT, Base32) dan `totp_enabled` (INTEGER DEFAULT 0) pada tabel `users`
- [ ] Implementasi TOTP engine di Go (`internal/auth/totp.go`): generate secret unik, buat URI `otpauth://totp/...`, dan validasi kode 6 digit berbasis waktu (30s time step ±1 skew)
- [ ] Alur login 2 langkah:
  - Step 1: Input Username & Password
  - Step 2: Jika `totp_enabled == 0`, tampilkan QR Code untuk di-scan di Google Authenticator & aktivasi kode pertama. Jika `totp_enabled == 1`, minta 6 digit OTP.
- [ ] Rate limiting: maksimal 5 kali percobaan OTP salah berturut-turut untuk mencegah brute-force
- [ ] Template UI: `login_otp.html` (6 digit code input) dan `login_setup_2fa.html` (QR code & petunjuk)
- [ ] Pastikan email hanya wajib saat inisialisasi awal sistem POS (bootstrap superadmin)
- [ ] Unit test: validasi OTP valid, penolakan OTP salah/expired, dan independensi OTP antar-user

### Milestone 11: Sistem Product Tour Sinematik (Storyteller Live Caption & Camera Movement)
- [ ] Buat file stylesheet independen `web/static/css/product_tour_caption.css` khusus untuk product tour (terisolasi penuh dari bundle CSS production dan testing).
- [ ] Desain storyteller caption modern: lower-third glassmorphism, badge step, judul, dan narasi perilaku produk layaknya presenter manusia.
- [ ] Injeksi CSS secara dinamis via Playwright (`page.addStyleTag`) hanya pada skrip `record_product_tour.js` dan `record_ipad_intro.js`.
- [ ] Implementasi camera movement di Playwright: helper `cameraZoomTo(selector, scale, duration)` dan `cameraResetZoom(duration)` untuk zoom-in terarah pada interaksi penting dan zoom-out saat transisi layar.
- [ ] Pastikan perlindungan isolasi (guard) agar aset tour tidak pernah termuat pada build production.



