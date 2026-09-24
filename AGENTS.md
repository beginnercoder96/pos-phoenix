# AGENTS.md — Developer & AI Agent Operating Instructions

Selamat datang di repositori **POS Phoenix (Pardis Barbershop Management & POS)**.
Dokumen ini menetapkan standar wajib pengerjaan fitur, arsitektur, dan pedoman desain antarmuka bagi seluruh developer dan AI agent.

---

## 🎨 Taste-Skill & Anti-Slop UI Design System (Standar Adopsi `posv2`)

Setiap modifikasi tampilan web (HTML, CSS, Tailwind) wajib mematuhi standar desain anti-slop:

1. **Palet Warna & Surface**:
   - Dark Mode: Surface utama `#000210`, card/panel `#010521` atau `#0a1130`.
   - Border: Selalu gunakan garis tipis yang renyah (`border-white/10` untuk dark mode, `border-slate-200` untuk light mode).
   - Aksen: Emerald `#10b981` (pemasukan/omzet), Amber `#f59e0b` (dana cadangan), Rose `#ef4444` (pembatalan/beban).
   - Hindari: Gradien neon ungu/pink murahan, border tebal kartun, atau sudut melengkung berlebihan.

2. **Tipografi Finansial (`font-mono tabular-nums`)**:
   - **Wajib mutlak**: Seluruh tampilan uang Rupiah (`Rp ...`), persentase bagi hasil (`...%`), dan nomor invoice/transaksi harus menggunakan kelas `font-mono tabular-nums`.
   - Angka harus sejajar secara vertikal pada kolom tabel.

3. **Komponen Layout Standar**:
   - **Sidebar Desktop (`.pos-sidebar`)**: Lebar desktop wajib **288px (`lg:w-72`)**.
   - **Target Sentuh Mobile**: Tombol, link navigasi, dan input wajib memiliki tinggi minimal **44px (`min-h-11`)** untuk kepatuhan Android 10 / Chrome 83.
   - **Tablet / iPad Pro**: Layout grid harus responsif pada resolusi 1024x768 dan 1920x1080.

4. **Aksesibilitas (WCAG AA)**:
   - Teks sekunder/muted wajib memiliki rasio kontras minimal 4.5:1 terhadap background (gunakan minimal `text-slate-300` pada dark mode dan `text-slate-600` pada light mode).

---

## 🧪 Quality Gate & Automated Testing

Sebelum membuat commit atau push:
1. Pastikan seluruh unit test Go lolos:
   ```bash
   go test ./...
   ```
2. Pastikan rangkaian automated E2E Playwright test lolos:
   ```bash
   npm run test:e2e
   ```
3. Compile ulang stylesheet Tailwind CSS:
   ```bash
   npm run css:build
   ```
