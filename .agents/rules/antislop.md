# Anti-Slop Guidelines & Design System Rules (Adopsi dari `posv2`)

Pedoman ini adalah standar wajib dalam perancangan antarmuka (UI), penataan layout, copywriting, dan penulisan kode di **POS Phoenix (Pardis Barbershop)**. Hindari segala bentuk "AI slop" (desain murahan, efek generik, atau kode berlebih yang tidak bernilai guna).

---

## 1. Anti-Slop UI (Desain & Estetika Tajam)
* **Palet Warna Crisp & Gelap Pekat**:
  * Dark mode menggunakan background pekat `#000210` dan surface `#010521` / `#0a1130`, BUKAN abu-abu kusam atau gradien ungu/biru murahan.
  * Border halus dan tajam: gunakan `border-white/10` pada mode gelap dan `border-slate-200` pada mode terang.
  * Aksen fungsional: Emerald (`#10b981` / `#059669`) untuk angka positif/keuntungan, Amber (`#f59e0b`) untuk dana cadangan/peringatan, Rose (`#ef4444`) untuk pembatalan/beban.
* **Tipografi & Angka Finansial**:
  * **Wajib**: Seluruh tampilan uang (Rupiah), nomor invoice, dan persentase **harus menggunakan `font-mono` dan `tabular-nums`**. Hal ini memastikan seluruh angka di baris dan kolom tabel sejajar tegak lurus secara vertikal.
  * Gunakan hierarki teks yang tegas: `font-bold` / `font-black` untuk angka utama, teks sekunder yang mudah dibaca (minimal `text-slate-300` di dark mode, `text-slate-600` di light mode).
* **Elevasi & Shadow**:
  * Hindari drop shadow tebal yang melayang (*slop shadow*).
  * Gunakan shadow tipis berkelas: `0 1px 2px 0 rgba(0, 0, 0, 0.05)` atau `shadow-sm`.

---

## 2. Anti-Slop Layout & Mobile Responsiveness
* **Sidebar Desktop Standar (`.pos-sidebar`)**:
  * Lebar desktop wajib **288px (`lg:w-72`)**, memberikan ruang nafas yang nyaman bagi navigasi, identitas cabang, dan tombol sign out.
* **Target Sentuh Mobile & Tablet (Touch Target Compliance)**:
  * Seluruh tombol interaktif, link navigasi, dan input form wajib memiliki tinggi dan lebar sentuh minimal **44x44 piksel (`min-h-11`)** agar nyaman digunakan pada handheld kasir (Android 10 / Chrome 83).
* **Adaptasi Multi-Resolusi**:
  * Dukung resolusi layar tablet/iPad Pro (1024x768 & 1920x1080) tanpa elemen yang berhimpitan atau teks yang terpotong secara canggung.

---

## 3. Anti-Slop Copywriting & Human-Centric Tone
* Hindari teks template generik (misal: "Lorem ipsum", "Manage your awesome business here", "Action button").
* Gunakan bahasa operasional nyata barbershop: "Pardis Barber Shop", "Bagi Hasil Jasa", "Komisi Produk", "Take Home Pay", "Dana Cadangan Kas", "Cetak Thermal", "Simpan PDF".

---

## 4. Anti-Slop Code & Performance
* Gunakan HTML5 semantik (`<main>`, `<aside>`, `<header>`, `<article>`, `<table>`).
* Tidak menambahkan pustaka JavaScript eksternal yang membengkak jika bisa diselesaikan dengan Tailwind dan native browser.
* Semua perubahan wajib lolos verifikasi automated testing Playwright dan Go test.
