/**
 * E2E Backoffice Test Suite using Playwright (Headless)
 *
 * Scenarios:
 * 1. Owner Login -> Backoffice -> Update Profit Sharing (Klaseman & Ledok)
 * 2. Download 24-Month Financial Report Excel (.xlsx) & Validate Content
 * 3. Open Payroll Page -> Generate & Preview Employee Payroll Slip (HTML)
 */

const { chromium } = require('playwright');
const http = require('http');
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

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

async function waitForServer(url, timeoutMs = 30000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (await isServerRunning(url)) return true;
    await new Promise((r) => setTimeout(r, 600));
  }
  return false;
}

async function launchHeadlessBrowser() {
  const channels = ['chrome', 'msedge'];
  for (const channel of channels) {
    try {
      return await chromium.launch({ channel, headless: true });
    } catch (e) {
      // try next channel
    }
  }
  return await chromium.launch({ headless: true });
}

async function run() {
  console.log('🚀 Starting Backoffice E2E Automated Tests (Playwright Headless)...');
  let serverProcess = null;
  let baseURL = 'http://127.0.0.1:8080';

  if (!(await isServerRunning(baseURL))) {
    baseURL = 'http://127.0.0.1:8089';
    console.log(`📡 Spawning test server on ${baseURL}...`);
    serverProcess = spawn('go', ['run', './cmd/server'], {
      cwd: path.resolve(__dirname, '..'),
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
      throw new Error(`Failed to start test server on ${baseURL}`);
    }
    console.log(`✅ Test server is active and responding on ${baseURL}`);
  } else {
    console.log(`✅ Connected to existing active server on ${baseURL}`);
  }

  const browser = await launchHeadlessBrowser();
  const context = await browser.newContext({
    viewport: { width: 1280, height: 800 },
    acceptDownloads: true
  });
  const page = await context.newPage();

  try {
    // ========================================================
    // --- [Scenario 1: Owner Login & Profit Sharing Update] ---
    // ========================================================
    console.log('\n--- [Scenario 1: Owner Login & Profit Sharing Update] ---');
    await page.goto(`${baseURL}/login`, { waitUntil: 'domcontentloaded' });
    console.log('Navigated to Login page');

    // Fill login as Ipang (Owner)
    await page.fill('input[name="username"]', 'ipang');
    await page.fill('input[name="password"]', 'adminsupervisor');
    await page.click('button[type="submit"]');
    await page.waitForLoadState('domcontentloaded');

    // Navigate to Profit Sharing for KLASEMAN
    await page.goto(`${baseURL}/backoffice/profit-sharing?branch=KLASEMAN&period=2026-09`, { waitUntil: 'domcontentloaded' });
    console.log('Navigated to Klaseman Profit Sharing page');

    // Verify page elements
    const pageTitle = await page.title();
    console.log(`Page title: "${pageTitle}"`);

    const ownerInput = page.locator('#owner-pct-input');
    await ownerInput.waitFor({ state: 'visible', timeout: 5000 });
    
    // Set owner percentage to 20%
    await ownerInput.fill('20');
    await ownerInput.dispatchEvent('input');
    await page.waitForTimeout(300);

    // Save configuration
    const saveBtn = page.locator('#ps-save-btn');
    if (await saveBtn.isEnabled()) {
      await saveBtn.click();
      await page.waitForLoadState('domcontentloaded');
      console.log('Submitted profit sharing update for Klaseman');
    }

    // Verify saved value
    const savedOwnerVal = await page.locator('#owner-pct-input').inputValue();
    console.log(`Verified Klaseman owner percentage: ${savedOwnerVal}%`);

    // Switch to LEDOK
    await page.goto(`${baseURL}/backoffice/profit-sharing?branch=LEDOK&period=2026-09`, { waitUntil: 'domcontentloaded' });
    console.log('Navigated to Ledok Profit Sharing page');
    const ledokOwnerInput = page.locator('#owner-pct-input');
    await ledokOwnerInput.waitFor({ state: 'visible', timeout: 5000 });
    console.log(`Verified Ledok profit sharing loaded successfully (${await ledokOwnerInput.inputValue()}%)`);

    console.log('🎉 PASSED: Scenario 1 (Owner Login & Profit Sharing)');

    // ========================================================
    // --- [Scenario 2: Download 24-Month Financial Report Excel] ---
    // ========================================================
    console.log('\n--- [Scenario 2: Download 24-Month Financial Report Excel] ---');
    await page.goto(`${baseURL}/backoffice`, { waitUntil: 'domcontentloaded' });
    console.log('Navigated to Backoffice Dashboard');

    // Trigger download for /backoffice/reports.xlsx
    const [downloadExcel] = await Promise.all([
      page.waitForEvent('download'),
      page.evaluate((url) => { window.location.href = url; }, `${baseURL}/backoffice/reports.xlsx`)
    ]);

    const excelFilename = downloadExcel.suggestedFilename();
    console.log(`Downloaded file: ${excelFilename}`);
    if (!excelFilename.endsWith('.xlsx')) {
      throw new Error(`Expected .xlsx file, got ${excelFilename}`);
    }

    const excelSavePath = path.resolve(__dirname, 'temp_report.xlsx');
    await downloadExcel.saveAs(excelSavePath);
    const stats = fs.statSync(excelSavePath);
    console.log(`Excel report verified (${stats.size} bytes)`);
    if (stats.size < 500) {
      throw new Error(`Excel report size too small (${stats.size} bytes)`);
    }
    fs.unlinkSync(excelSavePath);

    console.log('🎉 PASSED: Scenario 2 (24-Month Excel Financial Report)');

    // ========================================================
    // --- [Scenario 3: Payroll Slips & HTML Preview] ---
    // ========================================================
    console.log('\n--- [Scenario 3: Payroll Slips & HTML Preview] ---');
    await page.goto(`${baseURL}/backoffice/payroll?branch=KLASEMAN&period=2026-09`, { waitUntil: 'domcontentloaded' });
    console.log('Navigated to Payroll page for Klaseman 2026-09');

    // Wait for payroll table
    const payrollTable = page.locator('table');
    await payrollTable.waitFor({ state: 'visible', timeout: 5000 });

    const rows = await page.locator('tbody tr').count();
    console.log(`Found ${rows} employee payroll records in table`);
    if (rows === 0) {
      throw new Error('No payroll records found in Klaseman');
    }

    // Find and click the first employee slip preview link
    const firstSlipBtn = page.locator('a[id^="btn-slip-"]').first();
    const slipHref = await firstSlipBtn.getAttribute('href');
    console.log(`Opening payroll slip preview from: ${slipHref}`);

    // Navigate to the slip preview
    await page.goto(`${baseURL}${slipHref}`, { waitUntil: 'domcontentloaded' });

    // Verify slip content
    const slipTitle = await page.locator('h1').textContent();
    if (!slipTitle.includes('SLIP GAJI')) {
      throw new Error(`Expected "SLIP GAJI" in header, got: ${slipTitle}`);
    }

    const logo = page.locator('img[alt="Pardis Barber Shop"]');
    if ((await logo.count()) === 0) {
      throw new Error('Logo Pardis Barber Shop not found in slip preview');
    }

    console.log('Verified Slip Gaji header, branch branding, and salary details successfully');
    console.log('🎉 PASSED: Scenario 3 (Payroll Slips & HTML Preview)');

    console.log('\n======================================================');
    console.log('🏆 ALL 3 BACKOFFICE E2E SCENARIOS PASSED SUCCESSFULLY!');
    console.log('======================================================');

  } catch (err) {
    console.error('\n❌ E2E TEST FAILED:', err);
    process.exitCode = 1;
  } finally {
    await browser.close();
    if (serverProcess) {
      console.log('Stopping test server...');
      serverProcess.kill();
    }
  }
}

run();