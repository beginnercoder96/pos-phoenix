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
8. **Revamp UI Menggunakan `taste-skill` & `antislop` (Adopsi dari `posv2`)**:
   - Merecheck implementasi dari project `/Users/mekari/Documents/newlearn/posv2` (`.agents/skills/taste-skill`, `.agents/rules/antislop.md`, `antislop-ui`, `antislop-layoutmobile`, `antislop-human`, dll.).
   - Menerapkan desain antarmuka modern yang tajam, kontras tinggi (WCAG AAA/AA), tipografi angka tabular monospaced (`tabular-nums`), tata letak proporsional iPad Pro & Mobile, micro-interactions 150ms-250ms, dan mengeliminasi visual slop (border tebal blur, teks abu-abu rendah kontras, fake metrics).
9. **Pemisahan Menu Transaksi dari Dashboard Utama (Dedicated Transaction Menu)**:
   - Memisahkan form entri kasir dan riwayat transaksi dari Dashboard (`/`).
   - Dashboard (`/`) difokuskan murni untuk ringkasan eksekutif (Executive Overview): kartu metrik KPI (Pemasukan, Pengeluaran, Saldo), jam operasional harian, dan grafik tren.
   - Menu Transaksi Baru / Kasir dipindahkan ke route & tampilan tersendiri (`/transactions`), lengkap dengan form transaksi multi-kategori, pemilihan barberman, diskon/bundling, kalkulator pembayaran, dan struk kasir.
10. **Simplifikasi Login & Two-Factor Authentication (OTP Google Authenticator)**:
    - **Form Login Bersih**: Hanya meminta `username` dan `password` untuk operasional harian.
    - **Email Mandatory Khusus Initial Setup**: Input email hanya diwajibkan saat bootstrapping/setup awal akun superadmin di sistem POS.
    - **OTP Google Authenticator per User**:
      - Setelah input username & password valid, sistem meminta 6 digit kode OTP dari Google Authenticator (TOTP RFC 6238).
      - Tiap user memiliki secret key unik masing-masing di database, sehingga kode OTP antar user berbeda dan independen.
      - Alur onboarding 2FA: Pengguna diarahkan ke halaman aktivasi menampilkan QR code unik + manual secret key untuk di-scan dengan aplikasi Google Authenticator, kemudian memasukkan 1 kode awal untuk mengaktifkan 2FA.
11. **Sistem Product Tour Sinematik (Storyteller Live Caption & Camera Movement)**:
    - **CSS Caption Terpisah & Independen**:
      - Dibuat file stylesheet khusus untuk product tour (misal `web/static/css/product_tour_caption.css`).
      - Hanya dimuat/diinjeksi secara on-demand saat proses perekaman product tour berjalan via Playwright, dan **sama sekali terpisah** dari bundle CSS aplikasi produksi (`app.css`) maupun pengujian fungsional reguler.
    - **Perilaku Live Caption / Human Storyteller**:
      - Saat layar atau fitur tertentu disorot, muncul visual caption dinamis (lower-third / dynamic card) yang bernarasikan konteks penggunaan aplikasi seperti seorang presenter/storyteller menjelaskan alur bisnis (konteks fitur, aksi kasir, dan hasil kalkulasi).
    - **Efek Gerakan Kamera Dinamis (Camera Movement)**:
      - Transisi zoom-in halus (scale 1.2x - 1.4x dengan focal point terarah) saat menyorot interaksi penting (input transaksi, pemilihan diskon, visualisasi kalkulator, toggle dark mode, switch bahasa).
      - Transisi zoom-out kembali ke pandangan penuh (full-page view) saat beralih antar halaman/menu untuk memberikan orientasi layout yang menyeluruh.

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
- [x] Buat migration tabel `branches`:
  - `id`, `code` (`KLASEMAN`, `LEDOK`), `name`, `address`, `is_active`, `created_at`
- [x] Perbarui tabel `users` / `operators`:
  - Tambahkan `branch_id` (foreign key ke `branches`)
  - Kolom tipe staf (`barberman`, `cashier`, `manager`, `owner`)
- [x] Perbarui tabel `catalog_items` / `products`:
  - Tambahkan flag `item_type` (`SERVICE` vs `PRODUCT`)
  - Tambahkan `commission_amount` (integer Rp 5.000 - Rp 10.000 untuk item produk)
- [x] Buat tabel `branch_profit_sharing_rules`:
  - `id`, `branch_id`, `period_month` (format `YYYY-MM`), `owner_percentage`, `unallocated_percentage`, `updated_by` (`ipang`), `updated_at`
- [x] Buat tabel `employee_profit_sharing_rules`:
  - `id`, `rule_id`, `user_id`, `percentage`, `created_at`
- [x] Buat tabel `discounts_and_bundles`:
  - `id`, `code`, `name`, `type` (`PERCENTAGE`, `FIXED_AMOUNT`, `BUNDLE`), `value`, `service_allocation_ratio`, `product_allocation_ratio`, `is_active`
- [x] Perbarui tabel `transactions` & `transaction_items`:
  - Tambahkan `branch_id`
  - Tambahkan `barber_id` / `served_by_id` pada tiap baris item
  - Tambahkan `discount_amount`, `bundle_id`, `commission_earned`

---

### Milestone 2: Core Calculation Engine (Go Backend)
- [x] Buat paket `internal/backoffice/calculator`:
  - Logika pemisahan omzet Jasa vs Retail
  - Perhitungan persentase bagi hasil bulanan per-cabang
  - Kalkulasi alokasi sisa saldo tidak terpakai
  - Kalkulasi komisi produk per-karyawan berdasarkan item yang dijual
  - Kalkulasi penyesuaian diskon & bundling terhadap pendapatan jasa
- [x] Buat validasi otorisasi:
  - Hanya user dengan role `superadmin` / email `ipang@example.com` yang diizinkan mengubah konfigurasi persentase bagi hasil
- [x] Buat unit test lengkap untuk skenario:
  - Pembagian hasil Klaseman (Januari 10jt, 35%-35%-20%-10%)
  - Pembagian hasil Ledok (Januari 12jt, 40%-30%-20%-10%)
  - Pengecualian produk dari bagi hasil dan akurasi komisi produk

---

### Milestone 3: Backoffice UI & Navigasi Terpisah
- [x] Pisahkan antarmuka Backoffice:
  - Buat layout & route terpisah `/backoffice/*` (atau sub-aplikasi terpisah) yang terlindungi middleware khusus Owner
- [x] Halaman **Konfigurasi Bagi Hasil Cabang** (`/backoffice/profit-sharing`):
  - Selector Cabang (Klaseman / Ledok) & Periode Bulan
  - Form input persentase Owner, Karyawan 1, Karyawan 2, dsb.
  - Indikator real-time kalkulasi sisa saldo tidak terpakai (Total persentase & nominal saldo cadangan)
- [x] Halaman **Manajemen Diskon & Bundling** (`/backoffice/discounts`):
  - Input promo diskon & paket bundling
- [x] Halaman **Katalog & Komisi Produk** (`/backoffice/products`):
  - Pengaturan nilai komisi per produk (Rp 5.000 - Rp 10.000)

---

### Milestone 4: Dashboard & Chart Pendapatan Historis (2 Tahun)
- [x] Buat query agregasi bulanan untuk 24 bulan ke belakang dari tanggal hari ini (`CURRENT_DATE` trailing window)
- [x] Endpoint API `/backoffice/api/analytics/trend-24m`:
  - Parameter filter: `branch_id` (`all`, `klaseman`, `ledok`)
  - Output JSON: data 24 bulan (omzet kotor, omzet jasa, omzet produk, diskon, sisa saldo cadangan)
- [x] Visualisasi Chart di Backoffice Dashboard:
  - Interaktif: beralih tampilan antara Klaseman, Ledok, dan Konsolidasi 2 Cabang
  - Tooltip interaktif rincian omzet jasa vs produk

---

### Milestone 5: Ekspor Laporan Finansial (Excel)
- [x] Service generator Excel (`excelize` di Go) untuk laporan finansial:
  - Range: Bulan berjalan hingga 24 bulan ke belakang
  - Sheet 1: Rangkuman Konsolidasi 2 Cabang
  - Sheet 2: Rincian Cabang Klaseman (Omzet, Bagi Hasil, Sisa Kas Cadangan)
  - Sheet 3: Rincian Cabang Ledok (Omzet, Bagi Hasil, Sisa Kas Cadangan)
  - Sheet 4: Rincian Penjualan Produk & Komisi
- [x] Tombol Download di UI Backoffice:
  - Filter rentang bulan/tahun fleksibel

---

### Milestone 6: Generator Slip Gaji Karyawan (Excel Payroll Slip)
- [x] Service generator Slip Gaji individual format Excel (`.xlsx`):
  - Template resmi Pardis Barbershop (Logo, Nama Cabang, Nama Karyawan, Periode)
  - Bagian Rincian Bagi Hasil Jasa (Omzet Cabang, Persentase, Nominal)
  - Bagian Rincian Komisi Produk (Daftar produk terjual, Qty, Rate komisi, Subtotal)
  - Bagian Ringkasan Gaji Bersih (Take Home Pay) & Tanda Tangan
- [x] Halaman UI Slip Gaji (`/backoffice/payroll`):
  - Filter Cabang & Bulan
  - Tabel rekap karyawan beserta nominal gaji bersih
  - Tombol **"Download Slip Gaji Excel"** per-karyawan
  - Tombol **"Download Semua Slip Gaji (ZIP / Multi-Sheet Excel)"** untuk seluruh karyawan cabang terkait

---

### Milestone 7: Verifikasi, Pengujian E2E & Dokumentasi
- [x] Unit Test Go:
  - Pengujian kalkulasi bagi hasil, sisa saldo, dan komisi produk
  - Pengujian validasi otorisasi eksklusif Ipang
  - Pengujian pembuatan file Excel laporan dan slip gaji
- [x] Automated E2E Test (Playwright):
  - Skenario login Owner -> masuk Backoffice -> ubah persentase bagi hasil Klaseman & Ledok
  - Skenario download laporan Excel 24 bulan
  - Skenario generate slip gaji karyawan
- [x] Pembaruan panduan deployment dan dokumentasi API Backoffice

---

### Milestone 8: UI Revamp Menggunakan Taste-Skill & Anti-Slop (Adopsi dari `posv2`)
- [x] **Audit & Sinkronisasi Pedoman dari `posv2`**:
  - Salin/integrasikan rules dan skills dari `/Users/mekari/Documents/newlearn/posv2/.agents` (`antislop.md`, `taste-skill`, `antislop-ui`, `antislop-layoutmobile`, `antislop-human`, `antislop-copywriting`, `antislop-code`).
  - Tambahkan panduan anti-slop ke project instructions / `AGENTS.md`.
- [x] **Modernisasi Desain & Token Tailwind**:
  - Sinkronkan `tailwind.config.js` dan `web/static/css/input.css` dengan token desain dari `posv2` (palet warna tajam, shadow tipis, border crisp `border-white/10`, aksen emerald).
  - Terapkan `font-mono` / `tabular-nums` untuk seluruh tampilan nilai uang (Rupiah), nomor invoice, dan persentase.
- [x] **Penyempurnaan Komponen UI & Layout**:
  - Terapkan layout sidebar desktop `pos-sidebar` (lebar 288px / `lg:w-72`), badge status kontras tinggi, card surfaces dengan elevasi halus, dan hover states yang responsif.
  - Perbaiki responsivitas tablet/iPad Pro (1920x1080 & 1024x768) serta target sentuh mobile (minimal 44x44px untuk Android Chrome).
- [x] **Audit Aksesibilitas & Anti-Slop (Quality Gate)**:
  - Jalankan pengecekan rasio kontras warna (WCAG compliance) pada teks dan tombol di mode terang & gelap.
  - Pastikan tidak ada copy boilerplate atau elemen dekoratif berlebih yang tidak fungsional.

---

### Milestone 9: Pemisahan Menu Transaksi dari Dashboard Utama (Dedicated Transaction View)
- [x] **Pembuatan Route & Template Transaksi Baru**:
  - Daftarkan route baru di `internal/httpserver/server.go`: `GET /transactions` (handler `s.transactionsPage`) dan `POST /transactions`.
  - Buat template baru `web/templates/transactions.html` khusus untuk alur kasir dan transaksi.
- [x] **Migrasi Form Kasir & Riwayat Transaksi**:
  - Pindahkan form input transaksi (kategori jasa/produk, pilih barberman, diskon/bundling, catatan, total real-time) dari `dashboard.html` ke `transactions.html`.
  - Pindahkan tabel mutasi transaksi harian, filter tanggal, dan tombol reversal dari dashboard ke halaman transaksi.
- [x] **Refactor & Penyederhanaan Dashboard (`/`)**:
  - Bersihkan `dashboard.html` agar fokus pada metrik eksekutif:
    - Kartu KPI Utama: Total Pemasukan, Pengeluaran, Saldo Bersih, Saldo Cadangan.
    - Jam shift harian dan status kasir bertugas.
    - Grafik tren pendapatan (visualisasi ringkas).
    - Tombol aksi cepat: "Buka Menu Transaksi Baru" (`/transactions`) dan "Laporan Lengkap".
- [x] **Pembaruan Navigasi & Sidebar**:
  - Tambahkan link menu navigasi yang jelas di desktop sidebar dan mobile navigation bar:
    - 📊 **Dashboard** (`/`)
    - 💳 **Transaksi / Kasir** (`/transactions`)
    - 👥 **Operator** (`/operators` - role superadmin)
    - 🏢 **Backoffice** (`/backoffice` - role superadmin)
- [x] **Regression Testing & Update Otomasi E2E**:
  - Tambahkan unit test HTTP di `internal/httpserver/server_test.go` untuk route `GET /transactions`.
  - Perbarui script tur Playwright (`tests/e2e_backoffice.js`) agar menavigasi ke menu Transaksi terpisah saat mendemonstrasikan proses kasir.

---

### Milestone 10: Simplifikasi Login & 2FA Google Authenticator (TOTP RFC 6238)
- [x] **Database Migration untuk 2FA**:
  - Tambahkan kolom pada tabel `users`: `totp_secret` (TEXT, base32 encoded secret key) dan `totp_enabled` (INTEGER DEFAULT 0).
  - Pastikan setiap pengguna baru atau pengguna yang ada mendapatkan secret key acak unik.
- [x] **Implementasi TOTP Engine di Go (`internal/auth/totp.go`)**:
  - Pembangkitan secret key acak (16–20 bytes cryptographically secure random, Base32).
  - Pembuatan URI standar: `otpauth://totp/POS%20Phoenix:{username}?secret={secret}&issuer=POS%20Phoenix`.
  - Verifikasi kode 6 digit berbasis waktu (time-step 30 detik) dengan toleransi time-skew (±1 step / 30 detik) menggunakan HMAC-SHA1.
- [x] **Alur Autentikasi 2 Langkah (Two-Step Authentication Flow)**:
  - **Langkah 1**: Form login hanya meminta `Username` dan `Password`.
  - **Langkah 2**: Jika kredensial username & password valid:
    - Jika `totp_enabled == 0`: Pengguna diarahkan ke halaman Setup 2FA (`/login/setup-2fa`) menampilkan QR Code (SVG/Data URL) + teks secret key manual untuk di-scan dengan Google Authenticator, serta input 6 digit untuk verifikasi dan aktivasi.
    - Jika `totp_enabled == 1`: Pengguna diarahkan ke halaman verifikasi OTP (`/login/verify-otp`) untuk memasukkan 6 digit kode dari Google Authenticator.
  - Gunakan token pre-auth sementara (berumur pendek, misal 5 menit) di cookie/session sebelum session cookie permanen diterbitkan.
- [x] **Rate Limiting & Perlindungan Brute-Force**:
  - Batasi percobaan input OTP salah (maksimal 5 kali salah berturut-turut) sebelum cooldown sementara.
- [x] **Template UI untuk OTP & Setup 2FA**:
  - `web/templates/login_otp.html`: Input 6 digit angka dengan auto-focus, paste support, dan timer countdown.
  - `web/templates/login_setup_2fa.html`: Tampilan QR code, instruksi download Google Authenticator, dan petunjuk aktivasi.
- [x] **Email Khusus Initial Setup**:
  - Pastikan `email` hanya diwajibkan pada proses inisialisasi awal sistem POS (bootstrap superadmin) atau pemulihan darurat, tidak lagi diminta pada login rutin kasir/operator.
- [x] **Unit Test & Regression Testing**:
  - Test validasi kode OTP yang benar dan penolakan kode yang salah/kedaluwarsa.
  - Test skenario user A dan user B menghasilkan dan memvalidasi OTP yang berbeda secara independen.
  - Test flow setup awal superadmin tetap mewajibkan email, sementara login harian hanya username.

---

### Milestone 11: Sistem Product Tour Sinematik (Storyteller Live Caption & Camera Movement)
- [ ] **Isolasi & Pembuatan CSS Caption Independen**:
  - Buat file stylesheet independen: `web/static/css/product_tour_caption.css`.
  - Pastikan file ini **terpisah penuh** dan tidak di-import di `input.css`, `app.css`, maupun template HTML aplikasi produksi.
  - Desain elemen *cinematic storyteller caption*:
    - Overlay lower-third modern dengan aksen glassmorphism (`backdrop-blur-md`, border halus semitransparan).
    - Tipografi bercerita (storyteller): badge penanda langkah (`STEP 01`), judul fitur tebal, dan teks narasi perilaku sistem layaknya narator manusia yang ramah dan jelas.
    - Animasi transisi masuk & keluar (subtle slide-up & fade-in) 250ms.
- [ ] **Injeksi Dinamis Playwright (`record_product_tour.js` & `record_ipad_intro.js`)**:
  - Panggil CSS caption secara on-demand hanya saat proses perekaman berjalan menggunakan `page.addStyleTag({ path: './web/static/css/product_tour_caption.css' })`.
  - Buat helper function `setStorytellerCaption({ step, title, narrative, duration })` di konteks Playwright.
- [ ] **Implementasi Engine Gerakan Kamera (Dynamic Camera Movement)**:
  - Buat fungsi kontrol kamera di Playwright:
    - `cameraZoomTo(selector, scale = 1.3, duration = 800)`: Menghitung posisi bounding rect elemen target, menggeser transform-origin, dan menganimasikan zoom-in terarah secara mulus (*ease-in-out*).
    - `cameraResetZoom(duration = 600)`: Mengembalikan skala tampilan ke 1.0 (zoom-out penuh) sebelum aksi navigasi atau pergantian adegan besar berikutnya.
  - Skenario kamera pada perekaman:
    - Zoom-in ke kartu KPI omzet dan Jam Operasional outlet saat pembukaan tur.
    - Zoom-in ke form transaksi saat memilih kategori layanan barberman dan memasukkan diskon.
    - Zoom-in ke modal cetak struk kasir saat transaksi berhasil disimpan.
    - Zoom-out ke tampilan halaman penuh sebelum berpindah antar menu/halaman.
- [ ] **Pencegahan Kebocoran Aset (Production Separation Guard)**:
  - Pastikan tidak ada build script npm/Tailwind atau asset pipeline produksi yang menyertakan stylesheet tur ini ke bundle production.



