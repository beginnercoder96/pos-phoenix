/**
 * Cinematic Product Tour Recorder (Desktop 1440x900 - Smooth 60fps GPU Edition)
 * Uses Playwright + GPU-Accelerated Subtitles + Precise Timing + Auto File Reveal
 * Part of Milestone 11
 */

const { chromium } = require('playwright');
const http = require('http');
const { spawn, exec } = require('child_process');
const path = require('path');
const fs = require('fs');
const crypto = require('crypto');
const {
  injectTourStyles,
  showStorytellerCaption,
  dismissStorytellerCaption,
  cameraZoomTo,
  cameraResetZoom
} = require('./tests/tour_engine');

// =========================================================================
// ⏱️ PENGATURAN DURASI SCENE & CAPTION (Ubah angka di sini dalam milidetik)
// Catatan: 1000 ms = 1 detik (contoh: 6000 = 6 detik)
// =========================================================================
const TOUR_CONFIG = {
  // Jeda di awal video sebelum caption intro pertama muncul
  INITIAL_DELAY_MS: 1000,

  // Scene 0: Halaman Login & OTP
  SCENE_0_INTRO_MS: 6200,         // Durasi caption intro aplikasi POS
  SCENE_0_OTP_MS: 6000,           // Durasi caption verifikasi kode OTP (2FA)

  // Scene 1: Executive Dashboard & Kartu Omzet
  SCENE_1_DASHBOARD_MS: 6800,     // Durasi caption ringkasan bisnis & omzet
  SCENE_1_ZOOM_HOLD_MS: 3600,     // Lama waktu kamera zoom di kartu omzet

  // Scene 2: Transaksi Kasir & Checkout
  SCENE_2_CASHIER_MS: 8000,       // Durasi caption proses transaksi kasir

  // Scene 3: Struk Thermal 58mm
  SCENE_3_RECEIPT_MS: 6800,       // Durasi caption preview struk thermal
  SCENE_3_ZOOM_HOLD_MS: 4000,     // Lama waktu kamera zoom di kertas struk

  // Scene 4: Backoffice Payroll & Slip Gaji Karyawan
  SCENE_4_PAYROLL_TABLE_MS: 6000, // Durasi caption tabel payroll
  SCENE_4_SLIP_GAJI_MS: 6200,     // Durasi caption slip gaji resmi

  // Scene 5: Personalisasi Tema Gelap / Terang
  SCENE_5_THEMES_MS: 6800,        // Durasi caption ganti tema & bahasa

  // Scene 6: Kesimpulan / Penutup Tour
  SCENE_6_CONCLUSION_MS: 4500,    // Durasi kartu penutup tour
};

function base32Decode(str) {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let cleaned = str.toUpperCase().replace(/=+$/, '');
  let bits = '';
  for (let i = 0; i < cleaned.length; i++) {
    const val = alphabet.indexOf(cleaned[i]);
    if (val === -1) continue;
    bits += val.toString(2).padStart(5, '0');
  }
  const bytes = [];
  for (let i = 0; i + 8 <= bits.length; i += 8) {
    bytes.push(parseInt(bits.substr(i, 8), 2));
  }
  return Buffer.from(bytes);
}

function generateTOTP(secret, timeStepSeconds = 30) {
  const key = base32Decode(secret);
  const epoch = Math.floor(Date.now() / 1000);
  const counter = Math.floor(epoch / timeStepSeconds);
  const buf = Buffer.alloc(8);
  buf.writeBigInt64BE(BigInt(counter));
  const hmac = crypto.createHmac('sha1', key).update(buf).digest();
  const offset = hmac[hmac.length - 1] & 0x0f;
  const code = (hmac.readUInt32BE(offset) & 0x7fffffff) % 1000000;
  return code.toString().padStart(6, '0');
}

async function isServerRunning(url) {
  return new Promise((resolve) => {
    const req = http.get(url + '/login', (res) => {
      resolve(res.statusCode < 500);
    });
    req.on('error', () => resolve(false));
    req.setTimeout(1200, () => {
      req.destroy();
      resolve(false);
    });
  });
}

async function waitForServer(url, timeoutMs = 60000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (await isServerRunning(url)) return true;
    await new Promise((r) => setTimeout(r, 600));
  }
  return false;
}

async function launchBrowser() {
  const args = [
    '--disable-background-timer-throttling',
    '--disable-backgrounding-occluded-windows',
    '--disable-renderer-backgrounding',
    '--disable-frame-rate-limit',
    '--force-device-scale-factor=1',
    '--force-color-profile=srgb',
    '--no-sandbox',
    '--enable-features=VaapiVideoDecoder',
    '--enable-gpu-rasterization'
  ];

  const channels = ['chrome', 'msedge'];
  for (const channel of channels) {
    try {
      return await chromium.launch({ channel, headless: true, args });
    } catch (e) { }
  }
  return await chromium.launch({ headless: true, args });
}

function openVideoInIDE(filePath) {
  try {
    const editors = process.platform === 'win32'
      ? ['antigravity.cmd', 'code.cmd', 'cursor.cmd']
      : ['antigravity', 'code', 'cursor'];

    for (const cmd of editors) {
      try {
        exec(`${cmd} -r "${filePath}"`, (err) => {
          if (!err) {
            console.log(`🎬 [IDE] Automatically opened ${path.basename(filePath)} in ${cmd}.`);
          }
        });
        return;
      } catch (e) { }
    }
  } catch (e) { }
}

async function waitRemaining(startTime, totalDurationMs) {
  const elapsed = Date.now() - startTime;
  const remaining = totalDurationMs - elapsed;
  if (remaining > 0) {
    await new Promise((resolve) => setTimeout(resolve, remaining));
  }
}

async function recordProductTour() {
  console.log('🎬 [Product Tour] Starting Cinematic Product Tour Recorder (Smooth 60fps GPU Edition)...');

  const recordingsDir = path.resolve(__dirname, 'recordings');
  if (!fs.existsSync(recordingsDir)) {
    fs.mkdirSync(recordingsDir, { recursive: true });
  }

  let serverProcess = null;
  let baseURL = 'http://127.0.0.1:8080';

  if (!(await isServerRunning(baseURL))) {
    baseURL = 'http://127.0.0.1:8089';
    console.log(`📡 Spawning application server in background (${baseURL})...`);
    serverProcess = spawn('go', ['run', './cmd/server'], {
      cwd: __dirname,
      env: {
        ...process.env,
        ADDR: ':8089',
        DATABASE_PATH: 'pos.db',
        SESSION_SECURE: 'false',
        BUSINESS_TIMEZONE: 'Asia/Jakarta'
      },
      stdio: 'ignore'
    });

    const ready = await waitForServer(baseURL);
    if (!ready) {
      if (serverProcess) serverProcess.kill();
      throw new Error(`Failed to connect to application server on ${baseURL}`);
    }
    console.log(`✅ Server is active on ${baseURL}`);
  } else {
    console.log(`✅ Using active server on ${baseURL}`);
  }

  const browser = await launchBrowser();
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    recordVideo: {
      dir: recordingsDir,
      size: { width: 1440, height: 900 }
    }
  });

  const page = await context.newPage();

  try {
    // -------------------------------------------------------------
    // SCENE 0: LOGIN & TWO-FACTOR AUTHENTICATION (2FA)
    // -------------------------------------------------------------
    console.log('📍 [Scene 0] Sign-in & Authentication Page...');
    await page.goto(`${baseURL}/login`, { waitUntil: 'domcontentloaded' });
    await injectTourStyles(page);

    // Berikan jeda awal video sebelum caption pertama muncul
    await page.waitForTimeout(TOUR_CONFIG.INITIAL_DELAY_MS);

    const scene0IntroDuration = TOUR_CONFIG.SCENE_0_INTRO_MS;
    const t0 = Date.now();
    showStorytellerCaption(page, {
      step: 'INTRO',
      title: 'Pardis Barbershop Management & POS',
      narrative: 'An integrated platform for barbershop point-of-sale, inventory tracking, profit sharing, and secure 2FA authentication.',
      duration: scene0IntroDuration,
      position: 'bottom'
    });

    // Pengetikan tenang dan manusiawi
    await page.waitForTimeout(600);
    await page.locator('input[name="username"]').type('ipang', { delay: 110 });
    await page.waitForTimeout(400);
    await page.locator('input[name="password"]').type('adminsupervisor', { delay: 110 });
    await page.waitForTimeout(500);
    await waitRemaining(t0, scene0IntroDuration);

    await page.click('button[type="submit"]');
    await page.waitForLoadState('domcontentloaded');

    // Handle 2FA Verification
    let currentURL = page.url();
    if (currentURL.includes('/login/setup-2fa')) {
      console.log('📍 Handling initial 2FA setup...');
      await injectTourStyles(page);
      const secretInput = page.locator('#totp-secret-input');
      await secretInput.waitFor({ state: 'attached', timeout: 5000 });
      const rawSecret = await secretInput.getAttribute('data-raw-secret');
      const otpCode = generateTOTP(rawSecret);
      await page.waitForTimeout(400);
      await page.locator('#otp-input').type(otpCode, { delay: 110 });
      await page.waitForTimeout(400);
      await page.click('button[type="submit"]');
      await page.waitForLoadState('domcontentloaded');
    } else if (currentURL.includes('/login/verify-otp')) {
      console.log('📍 Verifying 6-digit TOTP code...');
      await injectTourStyles(page);

      showStorytellerCaption(page, {
        step: 'SECURITY',
        title: 'Time-Based OTP Verification (2FA)',
        narrative: 'Verifying access with a 6-digit dynamic code from Google Authenticator to guarantee cashier and backoffice security.',
        duration: TOUR_CONFIG.SCENE_0_OTP_MS,
        position: 'bottom'
      });

      const otpInput = page.locator('#otp-input');
      await otpInput.waitFor({ state: 'visible', timeout: 5000 });
      await page.waitForTimeout(400);
      await otpInput.type('123456', { delay: 110 });
      await page.waitForTimeout(400);

      // Klik verifikasi agar alur mengalir mulus ke Dashboard tanpa bengong
      await page.click('#verify-submit-btn');
      await page.waitForLoadState('domcontentloaded');
    }

    // Pastikan halaman berada di Dashboard
    if (!page.url().endsWith('/') && !page.url().endsWith(':8089/') && !page.url().endsWith(':8080/')) {
      await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
    }
    await page.waitForTimeout(600);

    // -------------------------------------------------------------
    // SCENE 1: EXECUTIVE DASHBOARD & REVENUE CARD HIGHLIGHT
    // -------------------------------------------------------------
    console.log('📍 [Scene 1] Executive Dashboard & Revenue Card (Parallel)...');
    await injectTourStyles(page);

    const scene1Duration = TOUR_CONFIG.SCENE_1_DASHBOARD_MS;
    const t1 = Date.now();
    showStorytellerCaption(page, {
      step: 'STEP 01',
      title: 'Real-Time Business & Revenue Dashboard',
      narrative: 'Monitoring key daily performance metrics: gross revenue, kapster commissions, active transactions, and cash reserves.',
      duration: scene1Duration,
      position: 'bottom'
    });

    console.log('🔍 Subtle focus zoom into Gross Income / Revenue card...');
    const omzetCard = '.stat-card.border-emerald-200, .stat-card:first-child';
    await page.waitForTimeout(500);
    await cameraZoomTo(page, omzetCard, 1.05, 450);
    await page.waitForTimeout(TOUR_CONFIG.SCENE_1_ZOOM_HOLD_MS);

    console.log('🔍 Reset Zoom...');
    await cameraResetZoom(page, 350);

    // Tunggu durasi membaca scene 1 selesai dengan santai
    await waitRemaining(t1, scene1Duration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(300);

    // Pindah ke Transaksi Kasir
    await page.goto(`${baseURL}/transactions`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(600);

    // -------------------------------------------------------------
    // SCENE 2: REAL CASHIER CHECKOUT TRANSACTION
    // -------------------------------------------------------------
    console.log('📍 [Scene 2] Cashier Checkout...');
    await injectTourStyles(page);

    const scene2Duration = TOUR_CONFIG.SCENE_2_CASHIER_MS;
    const t2 = Date.now();
    showStorytellerCaption(page, {
      step: 'STEP 02',
      title: 'Lightning-Fast Cashier Checkout',
      narrative: 'Easily record client haircuts, assign barber specialists, select grooming products, and apply promotional discounts.',
      duration: scene2Duration,
      position: 'top'
    });

    await page.waitForTimeout(600);

    // Isi Catatan / Nama Pelanggan secara alami
    const noteInput = page.locator('input[name="note"]').first();
    if (await noteInput.isVisible()) {
      await noteInput.type('Budi Santoso - Haircut Regular', { delay: 90 });
      await page.waitForTimeout(400);
    }

    // Pilih Kategori Layanan
    const categorySelect = page.locator('select.category-select').first();
    if (await categorySelect.isVisible()) {
      await categorySelect.selectOption({ index: 1 });
      await page.waitForTimeout(400);
    }

    // Pilih Item Layanan
    const itemSelect = page.locator('select.item-select').first();
    if (await itemSelect.isVisible()) {
      const optionCount = await itemSelect.locator('option').count();
      if (optionCount > 1) {
        await itemSelect.selectOption({ index: 1 });
        await page.waitForTimeout(400);
      }
    }

    // Masukkan nominal jika amount input masih kosong
    const amountInput = page.locator('input.amount-input').first();
    if (await amountInput.isVisible()) {
      const currentVal = await amountInput.inputValue();
      if (!currentVal || currentVal === '0') {
        await amountInput.type('40000', { delay: 80 });
        await page.waitForTimeout(400);
      }
    }

    await page.waitForTimeout(500);

    // Submit transaksi baru
    const submitBtn = page.locator('#transaction-form button[type="submit"]').first();
    if (await submitBtn.isVisible()) {
      console.log('💾 Submitting cashier transaction...');
      await submitBtn.click();
      await page.waitForTimeout(600);

      // Handle confirmation modal
      const confirmModal = page.locator('#pos-save-tx-modal');
      if (await confirmModal.isVisible()) {
        console.log('🔍 Confirmation modal opened...');
        await page.waitForTimeout(900);

        // Klik "Simpan & Cetak Struk"
        const savePrintBtn = page.locator('#pos-tx-save-print');
        if (await savePrintBtn.isVisible()) {
          console.log('🖨️ Clicking "Save & Print Receipt"...');
          await savePrintBtn.click();
          await page.waitForTimeout(900);
        } else {
          const saveOnlyBtn = page.locator('#pos-tx-save-only');
          if (await saveOnlyBtn.isVisible()) {
            await saveOnlyBtn.click();
            await page.waitForTimeout(900);
          }
        }
      }
    }

    // Tunggu durasi membaca scene 2 selesai dengan santai
    await waitRemaining(t2, scene2Duration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(300);

    // -------------------------------------------------------------
    // SCENE 3: 58MM THERMAL RECEIPT PREVIEW
    // -------------------------------------------------------------
    console.log('📍 [Scene 3] Thermal Receipt Preview...');
    await injectTourStyles(page);

    // Pastikan modal struk thermal terbuka
    let modalReceipt = page.locator('#pos-receipt-preview-modal');
    if (!(await modalReceipt.isVisible())) {
      const latestPrintBtn = page.locator('.btn-print-receipt').first();
      if (await latestPrintBtn.isVisible()) {
        console.log('🖨️ Opening thermal receipt modal from recent transactions...');
        await latestPrintBtn.click();
        await page.waitForTimeout(700);
      }
    }

    if (await modalReceipt.isVisible()) {
      const scene3Duration = TOUR_CONFIG.SCENE_3_RECEIPT_MS;
      const t3 = Date.now();
      showStorytellerCaption(page, {
        step: 'STEP 03',
        title: '58mm Thermal Receipt Preview',
        narrative: 'Automated thermal receipts formatted to standard 58mm/80mm printers, featuring outlet branding in Salatiga and fee breakdowns.',
        duration: scene3Duration,
        position: 'bottom'
      });

      console.log('🔍 Subtle focus on 58mm thermal receipt paper...');
      await page.waitForTimeout(500);
      await cameraZoomTo(page, '.thermal-receipt-paper', 1.04, 450);
      await page.waitForTimeout(TOUR_CONFIG.SCENE_3_ZOOM_HOLD_MS);

      await cameraResetZoom(page, 350);

      // Tunggu durasi membaca scene 3 selesai dengan santai
      await waitRemaining(t3, scene3Duration);
      await dismissStorytellerCaption(page, 280);
      await page.waitForTimeout(300);

      // Tutup modal struk
      const closeReceiptBtn = page.locator('#btn-receipt-cancel, #receipt-close-x').first();
      if (await closeReceiptBtn.isVisible()) {
        await closeReceiptBtn.click();
        await page.waitForTimeout(500);
      }
    }

    // Pindah ke Backoffice Payroll
    await page.goto(`${baseURL}/backoffice/payroll`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(600);

    // -------------------------------------------------------------
    // SCENE 4: BACKOFFICE PAYROLL & EMPLOYEE SALARY SLIP
    // -------------------------------------------------------------
    console.log('📍 [Scene 4] Backoffice Payroll & Employee Slip...');
    await injectTourStyles(page);

    const scene4Duration = TOUR_CONFIG.SCENE_4_PAYROLL_TABLE_MS;
    const t4 = Date.now();
    showStorytellerCaption(page, {
      step: 'STEP 04',
      title: 'Automated Payroll & Profit Sharing',
      narrative: 'Streamlined commission calculations, outlet revenue splits, and official printable PDF salary slips per barber.',
      duration: scene4Duration,
      position: 'top'
    });

    console.log('🔍 Subtle focus on Payroll table...');
    await page.waitForTimeout(500);
    await cameraZoomTo(page, '#payroll-table, table', 1.03, 400);
    await page.waitForTimeout(3600);

    await cameraResetZoom(page, 350);
    await waitRemaining(t4, scene4Duration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(300);

    // Buka Halaman Slip Gaji Karyawan Resmi
    console.log('📍 Opening official employee salary slip layout...');
    await page.goto(`${baseURL}/backoffice/payroll/slip?branch=KLASEMAN&period=2026-09&employee_id=1`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(600);
    await injectTourStyles(page);

    const slipDuration = TOUR_CONFIG.SCENE_4_SLIP_GAJI_MS;
    const tSlip = Date.now();
    showStorytellerCaption(page, {
      step: 'STEP 05',
      title: 'Official Employee Payroll Slip',
      narrative: 'Official salary slip layout with authorized branch letterhead, transparent service commissions, and employee signature lines.',
      duration: slipDuration,
      position: 'bottom'
    });

    console.log('🔍 Focus on Salary Slip card...');
    await page.waitForTimeout(500);
    const slipCard = '.slip-card, main > div, #payroll-slip-container';
    await cameraZoomTo(page, slipCard, 1.03, 400);
    await page.waitForTimeout(3800);

    await cameraResetZoom(page, 350);
    await waitRemaining(tSlip, slipDuration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(300);

    // Pindah kembali ke Dashboard untuk preferensi tema
    await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(600);

    // -------------------------------------------------------------
    // SCENE 5: DYNAMIC THEME & LANGUAGE PERSONALIZATION
    // -------------------------------------------------------------
    console.log('📍 [Scene 5] Theme & Language Personalization (Comfortable Pacing)...');
    await injectTourStyles(page);

    const scene5Duration = TOUR_CONFIG.SCENE_5_THEMES_MS;
    const t5 = Date.now();
    showStorytellerCaption(page, {
      step: 'STEP 06',
      title: 'Bilingual Support & Dark/Light Themes',
      narrative: 'Seamless language switching (English & Indonesian) and smooth theme transitions without losing cashier form drafts.',
      duration: scene5Duration,
      position: 'bottom'
    });

    // Switch Theme to Light Mode ☀️
    const themeSelect = page.locator('select[name="theme"]').first();
    if (await themeSelect.isVisible()) {
      await page.waitForTimeout(600);
      console.log('☀️ Switching theme to Light Mode...');
      await themeSelect.selectOption('light');
      await page.waitForTimeout(2000);

      console.log('🌙 Returning theme to Dark Mode...');
      const themeSelectDark = page.locator('select[name="theme"]').first();
      if (await themeSelectDark.isVisible()) {
        await themeSelectDark.selectOption('dark');
        await page.waitForTimeout(2000);
      }
    }

    // Pastikan durasi Scene 5 selesai dengan nyaman
    await waitRemaining(t5, scene5Duration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(300);

    // -------------------------------------------------------------
    // SCENE 6: CINEMATIC TOUR CONCLUSION
    // -------------------------------------------------------------
    console.log('📍 [Scene 6] Conclusion...');
    const scene6Duration = TOUR_CONFIG.SCENE_6_CONCLUSION_MS;
    const t6 = Date.now();
    showStorytellerCaption(page, {
      step: 'FINISH',
      title: 'Pardis POS Phoenix Ready for Production',
      narrative: 'A modern, resilient, and enterprise-grade barbershop POS crafted to accelerate your salon business growth.',
      duration: scene6Duration,
      position: 'bottom'
    });

    await waitRemaining(t6, scene6Duration);
    await dismissStorytellerCaption(page, 280);
    await page.waitForTimeout(400);

    console.log('✅ [Product Tour] All tour scenes successfully completed.');
  } catch (err) {
    console.error('❌ [Product Tour] Error encountered:', err);
  } finally {
    const videoPage = page.video();
    try {
      await context.close();
      await browser.close();
    } catch (e) { }

    if (serverProcess) {
      serverProcess.kill();
    }

    if (videoPage) {
      try {
        const rawPath = await videoPage.path();
        const finalPath = path.resolve(recordingsDir, 'product_tour_desktop.webm');
        if (fs.existsSync(finalPath)) {
          fs.unlinkSync(finalPath);
        }
        fs.renameSync(rawPath, finalPath);
        console.log(`\n🎉 [Video Saved] Product Tour Video ready at:\n   📁 ${finalPath}\n`);

        openVideoInIDE(finalPath);
      } catch (e) {
        console.log('\n🎉 [Video Saved] Product Tour Video ready.\n');
      }
    }
  }
}

if (require.main === module) {
  recordProductTour().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}

module.exports = { recordProductTour };
