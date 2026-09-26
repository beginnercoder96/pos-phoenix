/**
 * Cinematic iPad & Tablet Showcase Recorder (iPad Landscape 1024x768 - English Edition)
 * Uses Playwright + Storyteller Live Caption + Clean Focus Camera Engine
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
  sceneTransitionStart,
  sceneTransitionEnd,
  setStorytellerCaption,
  cameraZoomTo,
  cameraResetZoom
} = require('./tests/tour_engine');

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
    } catch (e) {}
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
      } catch (e) {}
    }
  } catch (e) {}
}

async function recordIpadIntro() {
  console.log('📱 [iPad Intro] Starting iPad & Tablet Showcase Recorder (1024x768, English Edition)...');

  const recordingsDir = path.resolve(__dirname, 'recordings');
  if (!fs.existsSync(recordingsDir)) {
    fs.mkdirSync(recordingsDir, { recursive: true });
  }

  let serverProcess = null;
  let baseURL = 'http://127.0.0.1:8080';

  if (!(await isServerRunning(baseURL))) {
    baseURL = 'http://127.0.0.1:8089';
    console.log(`📡 Spawning background server on ${baseURL}...`);
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
      throw new Error(`Failed to start server on ${baseURL}`);
    }
    console.log(`✅ Server ready on ${baseURL}`);
  } else {
    console.log(`✅ Using active server on ${baseURL}`);
  }

  const browser = await launchBrowser();
  const context = await browser.newContext({
    viewport: { width: 1024, height: 768 },
    deviceScaleFactor: 1,
    isMobile: false,
    hasTouch: true,
    recordVideo: {
      dir: recordingsDir,
      size: { width: 1024, height: 768 }
    }
  });

  const page = await context.newPage();

  try {
    // -------------------------------------------------------------
    // SCENE 0: TABLET LOGIN
    // -------------------------------------------------------------
    console.log('📍 [iPad] Navigating to Login Page...');
    await page.goto(`${baseURL}/login`, { waitUntil: 'domcontentloaded' });
    await injectTourStyles(page);

    // Jeda awal 1.4 detik agar penonton melihat tampilan tablet terlebih dahulu
    await page.waitForTimeout(1400);

    await setStorytellerCaption(page, {
      step: 'TABLET',
      title: 'iPad & Tablet Touch-Optimized Layout',
      narrative: 'Responsive design engineered for salon counter workflows using iPad, Android tablets, and touch-screen registers.',
      duration: 5000,
      position: 'bottom'
    });

    await page.fill('input[name="username"]', 'ipang');
    await page.waitForTimeout(400);
    await page.fill('input[name="password"]', 'adminsupervisor');
    await page.waitForTimeout(500);
    await page.click('button[type="submit"]');
    await page.waitForLoadState('domcontentloaded');

    // Handle 2FA
    let currentURL = page.url();
    if (currentURL.includes('/login/setup-2fa')) {
      await injectTourStyles(page);
      const secretInput = page.locator('#totp-secret-input');
      await secretInput.waitFor({ state: 'attached', timeout: 5000 });
      const rawSecret = await secretInput.getAttribute('data-raw-secret');
      const otpCode = generateTOTP(rawSecret);
      await page.locator('#otp-input').type(otpCode, { delay: 90 });
      await page.waitForTimeout(300);
      await page.click('button[type="submit"]');
      await page.waitForLoadState('domcontentloaded');
    } else if (currentURL.includes('/login/verify-otp')) {
      await injectTourStyles(page);
      const otpInput = page.locator('#otp-input');
      await otpInput.waitFor({ state: 'visible', timeout: 5000 });
      await otpInput.type('123456', { delay: 100 });
      await page.waitForTimeout(400);
      await page.click('#verify-submit-btn');
      await page.waitForLoadState('domcontentloaded');
    }

    await sceneTransitionStart(page);
    await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
    await sceneTransitionEnd(page);
    await page.waitForTimeout(500);

    // -------------------------------------------------------------
    // SCENE 1: TABLET TOUCH TARGETS & RESPONSIVENESS
    // -------------------------------------------------------------
    console.log('📍 [iPad] Scene 1: Dashboard Tablet & Touch Target');
    await injectTourStyles(page);

    await setStorytellerCaption(page, {
      step: 'STEP 01',
      title: 'Industry-Standard 44px Touch Targets',
      narrative: 'All navigation links, input controls, and buttons meet WCAG AA standards with minimum 44px touch targets.',
      duration: 5200,
      position: 'bottom'
    });

    console.log('🔍 Subtle focus on Navigation Buttons...');
    await page.waitForTimeout(400);
    await cameraZoomTo(page, '.pos-sidebar, nav', 1.04, 450);
    await page.waitForTimeout(2600);

    await cameraResetZoom(page, 350);
    await page.waitForTimeout(600);

    await sceneTransitionStart(page);
    await page.goto(`${baseURL}/transactions`, { waitUntil: 'domcontentloaded' });
    await sceneTransitionEnd(page);
    await page.waitForTimeout(500);

    // -------------------------------------------------------------
    // SCENE 2: FAST TABLET CHECKOUT
    // -------------------------------------------------------------
    console.log('📍 [iPad] Scene 2: Transaksi Tablet');
    await injectTourStyles(page);

    await setStorytellerCaption(page, {
      step: 'STEP 02',
      title: 'Fast Counter Checkout on Tablet',
      narrative: 'Quick barber selection, hair care add-ons, and instant payment settlement in just a few taps.',
      duration: 5200,
      position: 'top'
    });

    await page.waitForTimeout(2800);

    // -------------------------------------------------------------
    // IPAD OUTRO
    // -------------------------------------------------------------
    await setStorytellerCaption(page, {
      step: 'READY',
      title: 'Multi-Device Responsive Experience',
      narrative: 'Seamless performance across smartphones, counter tablets, and backoffice desktop monitors.',
      duration: 4200,
      position: 'bottom'
    });
    await page.waitForTimeout(2000);

    console.log('✅ [iPad Intro] Successfully completed iPad showcase.');
  } catch (err) {
    console.error('❌ [iPad Intro] Error encountered:', err);
  } finally {
    const videoPage = page.video();
    try {
      await context.close();
      await browser.close();
    } catch (e) {}

    if (serverProcess) {
      serverProcess.kill();
    }

    if (videoPage) {
      try {
        const rawPath = await videoPage.path();
        const finalPath = path.resolve(recordingsDir, 'ipad_intro_tablet.webm');
        if (fs.existsSync(finalPath)) {
          fs.unlinkSync(finalPath);
        }
        fs.renameSync(rawPath, finalPath);
        console.log(`\n🎉 [Video Saved] iPad showcase video ready at:\n   📁 ${finalPath}\n`);
        openVideoInIDE(finalPath);
      } catch (e) {
        console.log('\n🎉 [Video Saved] iPad showcase video saved.\n');
      }
    }
  }
}

if (require.main === module) {
  recordIpadIntro().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}

module.exports = { recordIpadIntro };
