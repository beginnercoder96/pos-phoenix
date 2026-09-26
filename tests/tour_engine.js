/**
 * Cinematic Tour & Camera Movement Engine for Playwright
 * Part of Milestone 11: Automated Product Tour Recording
 */

const path = require('path');

const CAPTION_CSS_PATH = path.resolve(__dirname, '../web/static/css/product_tour_caption.css');

/**
 * Injeksi dinamis CSS caption dan transition overlay ke dalam halaman.
 * @param {import('playwright').Page} page
 */
async function injectTourStyles(page) {
  try {
    await page.addStyleTag({ path: CAPTION_CSS_PATH });
  } catch (e) {
    // Abaikan jika stylesheet sudah terpasang
  }
}

/**
 * Transisi Layar: Fade out halus ke warna gelap saat berganti scene/halaman.
 * @param {import('playwright').Page} page
 */
async function sceneTransitionStart(page) {
  await injectTourStyles(page);
  await page.evaluate(() => {
    let overlay = document.getElementById('tour-scene-transition-overlay');
    if (!overlay) {
      overlay = document.createElement('div');
      overlay.id = 'tour-scene-transition-overlay';
      document.body.appendChild(overlay);
    }
    requestAnimationFrame(() => {
      overlay.classList.add('active');
    });
  });
  await page.waitForTimeout(280);
}

/**
 * Transisi Layar: Fade in halus membuka halaman baru.
 * @param {import('playwright').Page} page
 */
async function sceneTransitionEnd(page) {
  await injectTourStyles(page);
  await page.evaluate(() => {
    let overlay = document.getElementById('tour-scene-transition-overlay');
    if (overlay) {
      overlay.classList.remove('active');
      setTimeout(() => {
        if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay);
      }, 300);
    }
  });
  await page.waitForTimeout(300);
}

/**
 * Tampilkan caption subtitle ala presenter/storyteller secara NON-BLOCKING (berjalan bersamaan dengan alur aksi).
 * Posisi adaptif ala subtitle bioskop: 'bottom' (bawah) atau 'top' (atas).
 * @param {import('playwright').Page} page
 * @param {{ step: string, title: string, narrative: string, duration?: number, position?: 'bottom' | 'top' }} options
 */
async function showStorytellerCaption(page, { step = 'DEMO', title = '', narrative = '', duration = 4000, position = 'bottom' }) {
  await injectTourStyles(page);

  await page.evaluate(({ step, title, narrative, duration, position }) => {
    // Bersihkan timer sebelumnya jika ada
    if (window.__tourCaptionTimer) {
      clearTimeout(window.__tourCaptionTimer);
      window.__tourCaptionTimer = null;
    }
    if (window.__tourCaptionExitTimer) {
      clearTimeout(window.__tourCaptionExitTimer);
      window.__tourCaptionExitTimer = null;
    }

    let container = document.getElementById('tour-storyteller-root');
    if (!container) {
      container = document.createElement('div');
      container.id = 'tour-storyteller-root';
      document.body.appendChild(container);
    }

    container.className = `tour-storyteller-container tour-pos-${position}`;

    container.innerHTML = `
      <div class="tour-storyteller-card" id="tour-storyteller-card">
        <div class="tour-storyteller-header">
          <div class="tour-storyteller-badge-group">
            <span class="tour-storyteller-badge">${step}</span>
            <h4 class="tour-storyteller-title">${title}</h4>
          </div>
          <div class="tour-storyteller-live-indicator">
            <span class="tour-storyteller-pulse-dot"></span>
            <span>Live Showcase</span>
          </div>
        </div>
        <p class="tour-storyteller-narrative">${narrative}</p>
      </div>
    `;

    // Tampilkan card dengan animasi masuk halus
    requestAnimationFrame(() => {
      const card = document.getElementById('tour-storyteller-card');
      if (card) {
        card.classList.add('tour-active');
      }
    });

    // Otomatis hilangkan secara halus setelah durasi pembacaan selesai
    window.__tourCaptionTimer = setTimeout(() => {
      const card = document.getElementById('tour-storyteller-card');
      if (card) {
        card.classList.remove('tour-active');
        card.classList.add('tour-exiting');
        window.__tourCaptionExitTimer = setTimeout(() => {
          const root = document.getElementById('tour-storyteller-root');
          if (root) root.innerHTML = '';
        }, 260);
      }
    }, duration);
  }, { step, title, narrative, duration, position });
}

/**
 * Hilangkan caption storyteller secara halus dengan animasi fade-out.
 * @param {import('playwright').Page} page
 * @param {number} fadeDuration - Durasi transisi fade out (ms)
 */
async function dismissStorytellerCaption(page, fadeDuration = 260) {
  await page.evaluate((fadeDuration) => {
    if (window.__tourCaptionTimer) {
      clearTimeout(window.__tourCaptionTimer);
      window.__tourCaptionTimer = null;
    }
    if (window.__tourCaptionExitTimer) {
      clearTimeout(window.__tourCaptionExitTimer);
      window.__tourCaptionExitTimer = null;
    }
    const card = document.getElementById('tour-storyteller-card');
    if (card) {
      card.classList.remove('tour-active');
      card.classList.add('tour-exiting');
      setTimeout(() => {
        const root = document.getElementById('tour-storyteller-root');
        if (root) root.innerHTML = '';
      }, fadeDuration);
    }
  }, fadeDuration);
  await page.waitForTimeout(fadeDuration + 50);
}

/**
 * Zoom-in terarah tipis dan berikan outline glow halus pada elemen target.
 * @param {import('playwright').Page} page
 * @param {string} selector - Selector elemen yang difokuskan
 * @param {number} scale - Pembesaran elemen fokus (default 1.04)
 * @param {number} duration - Durasi transisi zoom (default 400ms)
 */
async function cameraZoomTo(page, selector, scale = 1.04, duration = 400) {
  await injectTourStyles(page);

  await page.evaluate(({ selector, scale, duration }) => {
    const el = document.querySelector(selector);
    if (!el) {
      console.warn(`[Camera] Target tidak ditemukan: ${selector}`);
      return;
    }

    el.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'center' });

    // Lepas target fokus sebelumnya jika ada
    document.querySelectorAll('.tour-focus-target').forEach((prev) => {
      prev.classList.remove('tour-focus-target');
      prev.style.transform = '';
    });

    el.classList.add('tour-focus-target');
    el.style.transform = `scale(${scale})`;
  }, { selector, scale, duration });

  await page.waitForTimeout(duration);
}

/**
 * Reset Zoom: Mengembalikan elemen yang disorot ke ukuran normal.
 * @param {import('playwright').Page} page
 * @param {number} duration - Durasi transisi (default 300ms)
 */
async function cameraResetZoom(page, duration = 300) {
  await page.evaluate(({ duration }) => {
    document.querySelectorAll('.tour-focus-target').forEach((el) => {
      el.style.transform = 'scale(1)';
      setTimeout(() => {
        el.classList.remove('tour-focus-target');
        el.style.transform = '';
      }, duration);
    });
  }, { duration });

  await page.waitForTimeout(duration);
}

module.exports = {
  injectTourStyles,
  sceneTransitionStart,
  sceneTransitionEnd,
  showStorytellerCaption,
  setStorytellerCaption: showStorytellerCaption, // kompatibilitas
  dismissStorytellerCaption,
  cameraZoomTo,
  cameraResetZoom,
};
