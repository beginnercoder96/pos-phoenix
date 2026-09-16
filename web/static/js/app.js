// Global Cross-Page Toast System with sessionStorage persistence
(function () {
  var TOAST_KEY = "pos_active_toast";

  function renderToast(title, message, duration, remainingTime, type) {
    var old = document.getElementById("global-pos-toast");
    if (old) old.remove();

    var isWarning = type === 'warning' || (title && (title.indexOf('Tidak Ada') !== -1 || title.indexOf('Peringatan') !== -1));
    var isError = type === 'error' || (title && (title.indexOf('Gagal') !== -1 || title.indexOf('Error') !== -1));

    var badgeBg = isWarning ? 'bg-amber-500 shadow-amber-500/30' : (isError ? 'bg-rose-600 shadow-rose-500/30' : 'bg-blue-600 shadow-blue-500/30');
    var progressBg = isWarning ? 'bg-amber-500' : (isError ? 'bg-rose-600' : 'bg-blue-600');
    var borderColor = isWarning ? 'border-amber-200/90 dark:border-amber-900' : (isError ? 'border-rose-200/90 dark:border-rose-900' : 'border-blue-200/90 dark:border-blue-900');
    var iconSvg = isWarning 
      ? '<svg style="width:18px;height:18px" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>'
      : (isError 
        ? '<svg style="width:18px;height:18px" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/></svg>'
        : '<svg style="width:18px;height:18px" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/></svg>');

    var toast = document.createElement("div");
    toast.id = "global-pos-toast";
    toast.style.position = "fixed";
    toast.style.top = "20px";
    toast.style.right = "20px";
    toast.style.zIndex = "999999";
    toast.style.minWidth = "300px";
    toast.style.maxWidth = "420px";
    toast.style.boxShadow = "0 20px 35px -8px rgba(15,23,42,0.22), 0 8px 16px -4px rgba(37,99,235,0.18)";
    toast.style.animation = "posToastSlideIn .35s cubic-bezier(.16,1,.3,1) forwards";
    toast.className = "flex flex-col rounded-2xl border " + borderColor + " bg-white/95 dark:bg-[#060b24]/95 backdrop-blur-md p-4 text-slate-800 dark:text-slate-100 transition-all";

    toast.innerHTML =
      '<div class="flex items-start gap-3">' +
        '<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ' + badgeBg + ' text-white shadow-md font-black text-sm">' +
          iconSvg +
        '</div>' +
        '<div class="min-w-0 flex-1 pr-1">' +
          '<p class="font-black text-sm tracking-tight text-slate-900 dark:text-white">' + (title || 'Informasi') + '</p>' +
          '<p class="text-xs text-slate-600 dark:text-slate-300 mt-0.5 leading-relaxed break-words">' + (message || '') + '</p>' +
        '</div>' +
        '<button type="button" id="global-pos-toast-close" class="shrink-0 rounded-lg p-1 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition" aria-label="Tutup">' +
          '<svg style="width:16px;height:16px" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>' +
        '</button>' +
      '</div>' +
      '<div class="mt-3 h-1 w-full overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">' +
        '<div id="global-pos-toast-progress" class="h-full ' + progressBg + ' transition-all ease-linear" style="width:100%"></div>' +
      '</div>';

    document.body.appendChild(toast);

    var closeBtn = document.getElementById("global-pos-toast-close");
    var progressBar = document.getElementById("global-pos-toast-progress");

    function dismissToast() {
      sessionStorage.removeItem(TOAST_KEY);
      toast.style.transition = "opacity .3s, transform .3s";
      toast.style.opacity = "0";
      toast.style.transform = "translateX(50px) scale(.95)";
      setTimeout(function () {
        if (toast.parentNode) toast.remove();
      }, 320);
    }

    if (closeBtn) {
      closeBtn.addEventListener("click", dismissToast);
    }

    var totalDuration = duration || 4500;
    var currentRemaining = remainingTime != null ? remainingTime : totalDuration;
    var startPercent = Math.max(0, Math.min(100, (currentRemaining / totalDuration) * 100));
    if (progressBar) progressBar.style.width = startPercent + "%";

    var startTime = Date.now();
    var interval = setInterval(function () {
      var passed = Date.now() - startTime;
      var left = currentRemaining - passed;
      if (left <= 0) {
        clearInterval(interval);
        dismissToast();
      } else if (progressBar) {
        var pct = (left / totalDuration) * 100;
        progressBar.style.width = pct + "%";
      }
    }, 50);
  }

  window.showGlobalToast = function (title, message, duration, type) {
    var d = duration || 4500;
    try {
      sessionStorage.setItem(TOAST_KEY, JSON.stringify({
        title: title,
        message: message,
        startedAt: Date.now(),
        duration: d,
        type: type || 'info'
      }));
    } catch (e) {}
    renderToast(title, message, d, d, type);
  };

  function checkToastOnLoad() {
    try {
      var params = new URLSearchParams(window.location.search);
      var savedParam = params.get("saved");
      if (savedParam) {
        params.delete("saved");
        var newSearch = params.toString() ? "?" + params.toString() : "";
        window.history.replaceState({}, document.title, window.location.pathname + newSearch);
        window.showGlobalToast("Data Berhasil Disimpan", 'Perubahan data karyawan "' + decodeURIComponent(savedParam) + '" telah berhasil disimpan.', 4500, 'success');
        return;
      }

      var raw = sessionStorage.getItem(TOAST_KEY);
      if (!raw) return;
      var data = JSON.parse(raw);
      var elapsed = Date.now() - data.startedAt;
      if (elapsed < data.duration) {
        var remaining = data.duration - elapsed;
        renderToast(data.title, data.message, data.duration, remaining, data.type);
      } else {
        sessionStorage.removeItem(TOAST_KEY);
      }
    } catch (e) {}
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", checkToastOnLoad);
  } else {
    checkToastOnLoad();
  }
})();

document.addEventListener("htmx:responseError", function () {
  if (window.showGlobalToast) {
    window.showGlobalToast("Gagal", "Permintaan gagal diproses. Silakan coba lagi.");
  }
});

document.querySelectorAll("[data-auto-submit] select").forEach(function (select) {
  select.addEventListener("change", function () {
    select.form.submit();
  });
});

(function () {
  // Built-in catalog matching the pricelist exactly
  var defaultCatalog = [
    {
      name: "Haircut Services",
      kind: "income",
      items: [
        { name: "Haircut", amount: 40000, amount_label: "40K" }
      ]
    },
    {
      name: "Add Ons",
      kind: "income",
      items: [
        { name: "Vitamin", amount: 3000, amount_label: "3K" },
        { name: "Shampoo & Blow Dry", amount: 20000, amount_label: "20K" },
        { name: "Beard Trim", amount: 15000, amount_label: "15K" }
      ]
    },
    {
      name: "Chemical Services",
      kind: "income",
      items: [
        { name: "Color Basic", amount: 60000, amount_label: "60K" },
        { name: "Color Fashion", amount: 220000, amount_label: "220K" },
        { name: "Smoothing", amount: 200000, amount_label: "200K" },
        { name: "Perm", amount: 200000, amount_label: "200K" },
        { name: "Root Lift & Down", amount: 200000, amount_label: "200K" }
      ]
    },
    {
      name: "Hair Treatment",
      kind: "income",
      items: [
        { name: "Hair & Face Mask", amount: 30000, amount_label: "30K" },
        { name: "Smooth Keratin", amount: 200000, amount_label: "200K" }
      ]
    },
    {
      name: "Product",
      kind: "income",
      items: [
        { name: "Death Waterbased", amount: 120000, amount_label: "120K" },
        { name: "Death Clay", amount: 120000, amount_label: "120K" },
        { name: "Death Powder", amount: 80000, amount_label: "80K" },
        { name: "Serum Pack", amount: 50000, amount_label: "50K" },
        { name: "Hair Tonic", amount: 20000, amount_label: "20K" }
      ]
    },
    {
      name: "Expense",
      kind: "expense",
      items: [
        { name: "Supplies & Inventory", amount: 0, amount_label: "Custom" },
        { name: "Rent", amount: 0, amount_label: "Custom" },
        { name: "Utilities", amount: 0, amount_label: "Custom" },
        { name: "Salary / Commission", amount: 0, amount_label: "Custom" },
        { name: "Operational / Maintenance", amount: 0, amount_label: "Custom" },
        { name: "Other Expense", amount: 0, amount_label: "Custom" }
      ]
    }
  ];

  var catalog = window.POS_CATALOG;
  if (!Array.isArray(catalog)) {
    try {
      var el = document.getElementById("catalog-data");
      if (el && el.textContent) {
        var parsed = JSON.parse(el.textContent.trim());
        if (typeof parsed === "string") {
          parsed = JSON.parse(parsed);
        }
        if (Array.isArray(parsed) && parsed.length > 0) {
          catalog = parsed;
        }
      }
    } catch (e) {
      console.warn("Could not parse catalog element, using fallback catalog", e);
    }
  }
  if (!Array.isArray(catalog) || catalog.length === 0) {
    catalog = defaultCatalog;
  }

  var form = document.getElementById("transaction-form");
  if (!form) return;

  var container = document.getElementById("category-rows-container");
  var totalDisplay = document.getElementById("transaction-total-display");
  var kindSelect = form.querySelector('select[name="kind"]');

  function formatIDR(amount) {
    var negative = amount < 0;
    if (negative) amount = -amount;
    var whole = Math.floor(amount);
    var parts = whole.toString().split("");
    var res = "";
    for (var i = parts.length - 1, count = 0; i >= 0; i--, count++) {
      if (count > 0 && count % 3 === 0) {
        res = "." + res;
      }
      res = parts[i] + res;
    }
    return (negative ? "-" : "") + "Rp " + (res || "0") + ",00";
  }

  function updateTotal() {
    if (!totalDisplay) return;
    var total = 0;
    if (container) {
      container.querySelectorAll(".amount-input").forEach(function (input) {
        var rawVal = input.value.replace(/[^0-9.]/g, "");
        var val = parseFloat(rawVal);
        if (!isNaN(val) && val > 0) {
          total += val;
        }
      });
    }
    totalDisplay.textContent = formatIDR(total);
  }

  function findCategory(catName) {
    if (!catName) return null;
    var lower = catName.toLowerCase().trim();
    for (var i = 0; i < catalog.length; i++) {
      var cName = catalog[i].name.toLowerCase();
      if (cName === lower) return catalog[i];
      // Flexible matching (e.g. "haircut" matches "haircut services")
      if (lower.indexOf("haircut") !== -1 && cName.indexOf("haircut") !== -1) return catalog[i];
      if (lower.indexOf("add on") !== -1 && cName.indexOf("add on") !== -1) return catalog[i];
      if (lower.indexOf("chemical") !== -1 && cName.indexOf("chemical") !== -1) return catalog[i];
      if (lower.indexOf("treatment") !== -1 && cName.indexOf("treatment") !== -1) return catalog[i];
      if (lower.indexOf("product") !== -1 && cName.indexOf("product") !== -1) return catalog[i];
      if (lower.indexOf("expense") !== -1 && cName.indexOf("expense") !== -1) return catalog[i];
    }
    return null;
  }

  function populateCategories(categorySelect, selectedCategory) {
    var isId = (document.documentElement.lang === "id");
    var kind = kindSelect ? kindSelect.value : "income";
    categorySelect.innerHTML = '<option value="">-- ' + (isId ? 'Pilih kategori' : 'Select category') + ' --</option>';

    catalog.forEach(function (cat) {
      if ((kind === "expense" && cat.kind === "expense") || (kind === "income" && cat.kind !== "expense")) {
        var opt = document.createElement("option");
        opt.value = cat.name;
        opt.textContent = cat.name;
        if (selectedCategory && cat.name.toLowerCase() === selectedCategory.toLowerCase()) {
          opt.selected = true;
        }
        categorySelect.appendChild(opt);
      }
    });

    var customOpt = document.createElement("option");
    customOpt.value = "Other";
    customOpt.textContent = isId ? "Lainnya / Kustom" : "Other / Custom";
    if (selectedCategory === "Other") customOpt.selected = true;
    categorySelect.appendChild(customOpt);
  }

  function populateItems(row, selectedItemName) {
    var isId = (document.documentElement.lang === "id");
    var categorySelect = row.querySelector(".category-select");
    var itemSelect = row.querySelector(".item-select");
    var amountInput = row.querySelector(".amount-input");
    if (!categorySelect || !itemSelect) return;

    var catVal = categorySelect.value;
    itemSelect.innerHTML = '<option value="">-- ' + (isId ? 'Pilih layanan / item' : 'Select service / item') + ' --</option>';

    var foundCat = findCategory(catVal);

    if (foundCat && foundCat.items && foundCat.items.length > 0) {
      foundCat.items.forEach(function (item) {
        var opt = document.createElement("option");
        opt.value = item.name;
        opt.textContent = item.name + (item.amount_label ? " (" + item.amount_label + ")" : "");
        opt.dataset.amount = item.amount;
        if (selectedItemName && item.name.toLowerCase() === selectedItemName.toLowerCase()) {
          opt.selected = true;
        }
        itemSelect.appendChild(opt);
      });

      var customOpt = document.createElement("option");
      customOpt.value = "Custom";
      customOpt.textContent = isId ? "Item Lainnya..." : "Custom Item...";
      customOpt.dataset.amount = "0";
      itemSelect.appendChild(customOpt);

      // If only 1 item in category, auto-select it and set amount
      if (foundCat.items.length === 1 && !selectedItemName) {
        itemSelect.selectedIndex = 1;
        var it = foundCat.items[0];
        if (it.amount > 0 && amountInput) {
          amountInput.value = it.amount;
        }
      }
    } else if (catVal) {
      var customOpt = document.createElement("option");
      customOpt.value = "Custom";
      customOpt.textContent = isId ? "Item Lainnya..." : "Custom Item...";
      customOpt.dataset.amount = "0";
      customOpt.selected = true;
      itemSelect.appendChild(customOpt);
    }
  }

  function addRow() {
    if (!container) return;
    var isId = (document.documentElement.lang === "id");
    var newRow = document.createElement("div");
    newRow.className = "category-row rounded-xl border border-slate-200 dark:border-slate-700 p-3 surface-soft space-y-2 sm:space-y-0 sm:grid sm:grid-cols-[1fr_1fr_130px_auto] sm:gap-2 items-center";
    newRow.innerHTML =
      '<div>' +
        '<label class="sr-only">' + (isId ? 'Kategori' : 'Category') + '</label>' +
        '<select name="category[]" class="field h-11 min-h-11 category-select" required>' +
        '</select>' +
      '</div>' +
      '<div>' +
        '<label class="sr-only">' + (isId ? 'Item / Layanan' : 'Item / Service') + '</label>' +
        '<select name="item_name[]" class="field h-11 min-h-11 item-select">' +
          '<option value="">-- ' + (isId ? 'Pilih layanan / item' : 'Select service / item') + ' --</option>' +
        '</select>' +
      '</div>' +
      '<div>' +
        '<label class="sr-only">' + (isId ? 'Jumlah' : 'Amount') + '</label>' +
        '<input type="text" name="amount[]" class="field h-11 min-h-11 text-right font-bold amount-input" placeholder="0" required inputmode="numeric">' +
      '</div>' +
      '<div class="flex justify-end sm:justify-center">' +
        '<button type="button" class="btn-secondary h-11 min-h-11 w-11 p-0 text-rose-600 font-black remove-row-btn" title="' + (isId ? 'Hapus' : 'Remove') + '" aria-label="' + (isId ? 'Hapus' : 'Remove') + '">✕</button>' +
      '</div>';

    container.appendChild(newRow);
    var catSelect = newRow.querySelector(".category-select");
    populateCategories(catSelect);
    updateTotal();
  }

  // Delegated event listener for Category changes
  document.addEventListener("change", function (e) {
    if (e.target && e.target.classList.contains("category-select")) {
      var row = e.target.closest(".category-row");
      if (row) {
        populateItems(row);
        updateTotal();
      }
    }
    if (e.target && e.target.classList.contains("item-select")) {
      var row = e.target.closest(".category-row");
      if (row) {
        var opt = e.target.options[e.target.selectedIndex];
        var amountInput = row.querySelector(".amount-input");
        if (opt && opt.dataset && opt.dataset.amount && parseFloat(opt.dataset.amount) > 0) {
          if (amountInput) {
            amountInput.value = opt.dataset.amount;
          }
        }
        updateTotal();
      }
    }
    if (e.target === kindSelect) {
      if (container) {
        container.querySelectorAll(".category-select").forEach(function (sel) {
          populateCategories(sel, sel.value);
        });
        container.querySelectorAll(".category-row").forEach(function (r) {
          populateItems(r);
        });
      }
      updateTotal();
    }
  });

  // Delegated input event for amount changes
  document.addEventListener("input", function (e) {
    if (e.target && e.target.classList.contains("amount-input")) {
      updateTotal();
    }
  });

  // Delegated click event for Delete Row button
  document.addEventListener("click", function (e) {
    var removeBtn = e.target.closest(".remove-row-btn");
    if (removeBtn) {
      e.preventDefault();
      var row = removeBtn.closest(".category-row");
      if (row && container) {
        var rows = container.querySelectorAll(".category-row");
        if (rows.length > 1) {
          row.remove();
        } else {
          // Clear row if it's the only one left
          var catSel = row.querySelector(".category-select");
          var itemSel = row.querySelector(".item-select");
          var amtInput = row.querySelector(".amount-input");
          if (catSel) catSel.value = "";
          if (itemSel) itemSel.innerHTML = '<option value="">-- Pilih layanan / item --</option>';
          if (amtInput) amtInput.value = "";
        }
        updateTotal();
      }
      return;
    }

    var addBtn = e.target.closest("#add-category-btn");
    if (addBtn) {
      e.preventDefault();
      addRow();
      return;
    }
  });

  // Initialize existing rows on page load
  if (container) {
    var initialRows = container.querySelectorAll(".category-row");
    initialRows.forEach(function (row) {
      var catSelect = row.querySelector(".category-select");
      var currentCat = catSelect ? catSelect.value : "";
      if (catSelect) {
        populateCategories(catSelect, currentCat);
      }
      populateItems(row);
    });
  }
  updateTotal();
})();

// Live Clock: Formats and refreshes elements with [data-live-clock] every minute
(function () {
  function formatLiveClock(date, lang) {
    var locale = (lang === "id") ? "id-ID" : "en-US";
    try {
      var parts = new Intl.DateTimeFormat(locale, {
        timeZone: "Asia/Jakarta",
        weekday: "long",
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        hour12: false
      }).formatToParts(date);

      var getPart = function (t) {
        var p = parts.find(function (x) { return x.type === t; });
        return p ? p.value : "";
      };

      return getPart("weekday") + ", " + getPart("day") + " " + getPart("month") + " " + getPart("year") + " " + getPart("hour") + ":" + getPart("minute") + " WIB";
    } catch (e) {
      var daysEn = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
      var daysId = ["Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"];
      var monthsEn = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      var monthsId = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];
      var pad = function (n) { return n < 10 ? "0" + n : "" + n; };
      var days = (lang === "id") ? daysId : daysEn;
      var months = (lang === "id") ? monthsId : monthsEn;
      return days[date.getDay()] + ", " + pad(date.getDate()) + " " + months[date.getMonth()] + " " + date.getFullYear() + " " + pad(date.getHours()) + ":" + pad(date.getMinutes()) + " WIB";
    }
  }

  function updateLiveClocks() {
    var elements = document.querySelectorAll("[data-live-clock]");
    if (!elements || elements.length === 0) return;
    var now = new Date();
    var lang = document.documentElement.lang || "en";
    var formatted = formatLiveClock(now, lang);
    elements.forEach(function (el) {
      el.textContent = formatted;
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", updateLiveClocks);
  } else {
    updateLiveClocks();
  }

  var now = new Date();
  var msUntilNextMinute = (60 - now.getSeconds()) * 1000 - now.getMilliseconds();
  setTimeout(function () {
    updateLiveClocks();
    setInterval(updateLiveClocks, 60000);
  }, Math.max(msUntilNextMinute, 500));
})();

// Interactive Cash-flow Trend Hover Function
(function initChartHover() {
  function setup() {
    var container = document.getElementById("chart-bars-container");
    var displayPeriod = document.getElementById("chart-hover-period");
    var displayIncome = document.getElementById("chart-hover-income");
    var displayExpense = document.getElementById("chart-hover-expense");
    var displayNet = document.getElementById("chart-hover-net");
    var incomeWrap = document.getElementById("chart-hover-income-wrap");
    var expenseWrap = document.getElementById("chart-hover-expense-wrap");
    var hoverDot = document.getElementById("chart-hover-dot");

    if (!container || !displayPeriod) return;

    var cols = container.querySelectorAll(".chart-col");
    if (!cols.length) return;

    // Calculate total summary across all chart bars
    var totalIncomeCents = 0;
    var totalExpenseCents = 0;
    cols.forEach(function (col) {
      totalIncomeCents += parseInt(col.getAttribute("data-income-cents") || "0", 10);
      totalExpenseCents += parseInt(col.getAttribute("data-expense-cents") || "0", 10);
    });

    function formatRupiahCents(cents) {
      var negative = cents < 0;
      if (negative) cents = -cents;
      var whole = Math.floor(cents / 100);
      var fraction = cents % 100;
      var strWhole = whole.toString();
      for (var i = strWhole.length - 3; i > 0; i -= 3) {
        strWhole = strWhole.slice(0, i) + "." + strWhole.slice(i);
      }
      var strFrac = (fraction < 10 ? "0" : "") + fraction;
      return (negative ? "-Rp " : "Rp ") + strWhole + "," + strFrac;
    }

    function resetToTotal() {
      var isId = (document.documentElement.lang === "id");
      displayPeriod.textContent = isId
        ? "Total Periode (" + cols.length + " titik) · Arahkan kursor ke grafik untuk melihat"
        : "Period Total (" + cols.length + " points) · Hover any bar to inspect";
      displayIncome.textContent = formatRupiahCents(totalIncomeCents);
      displayExpense.textContent = formatRupiahCents(totalExpenseCents);
      var net = totalIncomeCents - totalExpenseCents;
      displayNet.textContent = (net >= 0 ? "+" : "") + formatRupiahCents(net);
      displayNet.className = "font-black " + (net >= 0 ? "chart-inspect-income" : "chart-inspect-expense");
      if (hoverDot) hoverDot.className = "inline-block h-2.5 w-2.5 shrink-0 rounded-full bg-emerald-500 transition-colors";
      if (incomeWrap) {
        incomeWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all opacity-100";
      }
      if (expenseWrap) {
        expenseWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all opacity-100";
      }
    }

    function showSelected(col, specificType) {
      var isId = (document.documentElement.lang === "id");
      var period = col.getAttribute("data-period") || "";
      var incomeStr = col.getAttribute("data-income") || "Rp 0,00";
      var expenseStr = col.getAttribute("data-expense") || "Rp 0,00";
      var incCents = parseInt(col.getAttribute("data-income-cents") || "0", 10);
      var expCents = parseInt(col.getAttribute("data-expense-cents") || "0", 10);
      var netCents = incCents - expCents;

      displayIncome.textContent = incomeStr;
      displayExpense.textContent = expenseStr;
      displayNet.textContent = (netCents >= 0 ? "+" : "") + formatRupiahCents(netCents);
      displayNet.className = "font-black " + (netCents >= 0 ? "chart-inspect-income" : "chart-inspect-expense");

      if (specificType === "income") {
        displayPeriod.textContent = period + (isId ? " · Dipilih: Pemasukan" : " · Selected: Income");
        if (hoverDot) hoverDot.className = "inline-block h-2.5 w-2.5 shrink-0 rounded-full bg-emerald-500 ring-2 ring-emerald-400 transition-colors";
        if (incomeWrap) incomeWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all chart-badge-selected-income";
        if (expenseWrap) expenseWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all chart-dimmed";
      } else if (specificType === "expense") {
        displayPeriod.textContent = period + (isId ? " · Dipilih: Pengeluaran" : " · Selected: Expense");
        if (hoverDot) hoverDot.className = "inline-block h-2.5 w-2.5 shrink-0 rounded-full bg-rose-500 ring-2 ring-rose-400 transition-colors";
        if (expenseWrap) expenseWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all chart-badge-selected-expense";
        if (incomeWrap) incomeWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all chart-dimmed";
      } else {
        displayPeriod.textContent = period;
        if (hoverDot) hoverDot.className = "inline-block h-2.5 w-2.5 shrink-0 rounded-full bg-emerald-500 transition-colors";
        if (incomeWrap) incomeWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all opacity-100";
        if (expenseWrap) expenseWrap.className = "flex items-center gap-1.5 rounded-lg px-2 py-1 transition-all opacity-100";
      }
    }

    // Set initial display to period totals
    resetToTotal();

    // Attach listeners to each column and its specific bars
    cols.forEach(function (col) {
      col.addEventListener("mouseenter", function () {
        showSelected(col);
      });

      var incomeBar = col.querySelector(".chart-bar-income");
      if (incomeBar) {
        incomeBar.addEventListener("mouseenter", function (e) {
          e.stopPropagation();
          showSelected(col, "income");
        });
        incomeBar.addEventListener("mouseleave", function () {
          showSelected(col);
        });
      }

      var expenseBar = col.querySelector(".chart-bar-expense");
      if (expenseBar) {
        expenseBar.addEventListener("mouseenter", function (e) {
          e.stopPropagation();
          showSelected(col, "expense");
        });
        expenseBar.addEventListener("mouseleave", function () {
          showSelected(col);
        });
      }

      // Support touch on mobile/tablets
      col.addEventListener("touchstart", function () {
        showSelected(col);
      }, { passive: true });
    });

    container.addEventListener("mouseleave", function () {
      resetToTotal();
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", setup);
  } else {
    setup();
  }
})();

// Global Loading Screen Overlay Manager
(function initLoadingScreen() {
  function getOrCreateOverlay() {
    var overlay = document.getElementById("global-loading-overlay");
    if (!overlay) {
      overlay = document.createElement("div");
      overlay.id = "global-loading-overlay";
      overlay.className = "fixed inset-0 z-50 hidden flex-col items-center justify-center p-5 app-body backdrop-blur-sm transition-opacity duration-300";
      overlay.innerHTML = '<div class="flex flex-col items-center text-center max-w-sm">' +
        '<div class="relative mb-6 flex items-center justify-center">' +
        '<div class="absolute -inset-3 rounded-3xl bg-emerald-500/20 blur-xl animate-pulse"></div>' +
        '<div class="relative flex h-24 w-24 items-center justify-center rounded-2xl bg-emerald-500 p-1 shadow-2xl ring-4 ring-emerald-500/30">' +
        '<img src="/static/img/logopardis.jpg" alt="Pardis Barber Shop" class="h-full w-full rounded-xl object-cover shadow-inner">' +
        '</div>' +
        '<div class="pointer-events-none absolute -inset-4 rounded-3xl border-2 border-emerald-500/40 border-t-transparent animate-spin"></div>' +
        '</div>' +
        '<h2 id="global-loading-label" class="text-2xl sm:text-3xl font-black tracking-tight text-slate-800 dark:text-slate-100">' +
        'we prepare something good for you' +
        '</h2>' +
        '<p class="mt-2.5 text-xs font-semibold uppercase tracking-[.18em] text-emerald-600 dark:text-emerald-400">' +
        'Pardis Barber Shop · POS Phoenix' +
        '</p>' +
        '<div class="mt-7 flex items-center gap-2">' +
        '<span class="h-2.5 w-2.5 rounded-full bg-emerald-500 animate-bounce [animation-delay:-0.3s]"></span>' +
        '<span class="h-2.5 w-2.5 rounded-full bg-emerald-500 animate-bounce [animation-delay:-0.15s]"></span>' +
        '<span class="h-2.5 w-2.5 rounded-full bg-emerald-500 animate-bounce"></span>' +
        '</div>' +
        '</div>';
      document.body.appendChild(overlay);
    }
    return overlay;
  }

  window.showLoading = function (message) {
    var overlay = getOrCreateOverlay();
    var labelEl = document.getElementById("global-loading-label");
    if (labelEl) {
      labelEl.textContent = message || "we prepare something good for you";
    }
    overlay.classList.remove("hidden");
    overlay.classList.add("flex");
  };

  window.hideLoading = function () {
    var overlay = document.getElementById("global-loading-overlay");
    if (overlay) {
      overlay.classList.remove("flex");
      overlay.classList.add("hidden");
    }
  };

  window.addEventListener("pageshow", function () {
    window.hideLoading();
  });

  document.addEventListener("htmx:configRequest", function () {
    window.showLoading();
  });
  document.addEventListener("htmx:afterOnLoad", function () {
    window.hideLoading();
  });
  document.addEventListener("htmx:responseError", function () {
    window.hideLoading();
  });
})();

// Custom Date Chart Calendar Controller
window.handleChartPeriodChange = function (select) {
  if (select.value === "custom") {
    var trigger = document.getElementById("chart-calendar-trigger");
    if (trigger) trigger.classList.remove("hidden");
    if (typeof window.openChartCalendar === "function") {
      window.openChartCalendar();
    }
  } else {
    var trigger = document.getElementById("chart-calendar-trigger");
    if (trigger) trigger.classList.add("hidden");
    select.form.submit();
  }
};

(function initCustomChartCalendar() {
  function setupCalendar() {
    var panel = document.getElementById("chart-inline-calendar-panel") || document.getElementById("chart-calendar-modal");
    if (!panel) return;

    var isId = (document.documentElement.lang === "id");
    var monthNames = isId ? [
      "Januari", "Februari", "Maret", "April", "Mei", "Juni",
      "Juli", "Agustus", "September", "Oktober", "November", "Desember"
    ] : [
      "January", "February", "March", "April", "May", "June",
      "July", "August", "September", "October", "November", "December"
    ];

    var fromInputHidden = document.getElementById("chart-input-from") || document.getElementById("chart-form-from");
    var toInputHidden = document.getElementById("chart-input-to") || document.getElementById("chart-form-to");
    var directFromInput = document.getElementById("calendar-direct-from");
    var directToInput = document.getElementById("calendar-direct-to");
    var monthTitleEl = document.getElementById("calendar-month-title");
    var limitBadgeEl = document.getElementById("calendar-month-limit-badge");
    var daysGridEl = document.getElementById("calendar-days-grid");
    var summaryEl = document.getElementById("calendar-selection-summary");
    var formEl = document.getElementById("chart-period-form");

    var initialFrom = (fromInputHidden && fromInputHidden.value) ? fromInputHidden.value : "";
    var initialTo = (toInputHidden && toInputHidden.value) ? toInputHidden.value : "";

    var now = new Date();
    var viewYear = now.getFullYear();
    var viewMonth = now.getMonth();

    if (initialFrom) {
      var parts = initialFrom.split("-");
      if (parts.length === 3) {
        viewYear = parseInt(parts[0], 10);
        viewMonth = parseInt(parts[1], 10) - 1;
      }
    }

    var selFrom = initialFrom;
    var selTo = initialTo;
    var clickStep = 0; // 0 = start click, 1 = end click

    function formatYMD(y, m, d) {
      var mm = (m + 1) < 10 ? "0" + (m + 1) : "" + (m + 1);
      var dd = d < 10 ? "0" + d : "" + d;
      return y + "-" + mm + "-" + dd;
    }

    function getDaysInMonth(year, month) {
      return new Date(year, month + 1, 0).getDate();
    }

    function renderCalendar() {
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      if (monthTitleEl) {
        monthTitleEl.textContent = monthNames[viewMonth] + " " + viewYear;
      }
      if (limitBadgeEl) {
        limitBadgeEl.textContent = monthNames[viewMonth] + " " + viewYear + " · " + (isId ? ("Maksimal dapat dipilih: " + maxDays + " hari berdasarkan bulan yang ditampilkan") : ("Maximum selectable: " + maxDays + " days based on month shown"));
      }

      var monthStartStr = formatYMD(viewYear, viewMonth, 1);
      var monthEndStr = formatYMD(viewYear, viewMonth, maxDays);

      if (!selFrom) {
        selFrom = monthStartStr;
        selTo = monthStartStr;
      }

      if (directFromInput) {
        directFromInput.value = selFrom;
        directFromInput.min = monthStartStr;
        directFromInput.max = monthEndStr;
      }
      if (directToInput) {
        directToInput.value = selTo;
        directToInput.min = selFrom || monthStartStr;
        directToInput.max = monthEndStr;
      }

      updateSummary();

      if (!daysGridEl) return;
      daysGridEl.innerHTML = "";
      var firstDayOfWeek = new Date(viewYear, viewMonth, 1).getDay();

      for (var i = 0; i < firstDayOfWeek; i++) {
        var blank = document.createElement("div");
        blank.className = "h-8 w-full";
        daysGridEl.appendChild(blank);
      }

      for (var d = 1; d <= maxDays; d++) {
        var dateStr = formatYMD(viewYear, viewMonth, d);
        var btn = document.createElement("button");
        btn.type = "button";
        btn.textContent = d;
        btn.setAttribute("data-date", dateStr);
        btn.className = "flex h-8 w-full items-center justify-center rounded-lg text-xs font-semibold transition-all hover:bg-emerald-100 dark:hover:bg-emerald-950/60";

        var isStart = (dateStr === selFrom);
        var isEnd = (dateStr === selTo);
        var inRange = (selFrom && selTo && dateStr >= selFrom && dateStr <= selTo);

        if (isStart || isEnd) {
          btn.className += " bg-emerald-500 text-white font-black shadow-md ring-2 ring-emerald-500/40";
        } else if (inRange) {
          btn.className += " bg-emerald-100 text-emerald-900 font-bold dark:bg-emerald-950/80 dark:text-emerald-300";
        } else {
          btn.className += " text-slate-700 dark:text-slate-200";
        }

        (function (targetDate) {
          btn.addEventListener("click", function () {
            handleDayClick(targetDate);
          });
        })(dateStr);

        daysGridEl.appendChild(btn);
      }
    }

    function updateSummary() {
      if (!summaryEl) return;
      if (!selFrom || !selTo) {
        summaryEl.textContent = isId ? "Klik tanggal untuk memilih tanggal mulai" : "Click a day to select start date";
        return;
      }
      var d1 = new Date(selFrom);
      var d2 = new Date(selTo);
      var diffDays = Math.round(Math.abs((d2 - d1) / (24 * 60 * 60 * 1000))) + 1;
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      if (isId) {
        summaryEl.textContent = "Dipilih: " + diffDays + " hari (" + selFrom + " sampai " + selTo + ") · Maks: " + maxDays + " hari";
      } else {
        summaryEl.textContent = "Selected: " + diffDays + " day" + (diffDays > 1 ? "s" : "") + " (" + selFrom + " to " + selTo + ") · Max: " + maxDays + " days";
      }
    }

    function handleDayClick(dateStr) {
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      if (clickStep === 0 || !selFrom) {
        selFrom = dateStr;
        selTo = dateStr;
        clickStep = 1;
      } else {
        if (dateStr < selFrom) {
          selTo = selFrom;
          selFrom = dateStr;
        } else {
          selTo = dateStr;
        }
        var d1 = new Date(selFrom);
        var d2 = new Date(selTo);
        var diffDays = Math.round((d2 - d1) / (24 * 60 * 60 * 1000)) + 1;
        if (diffDays > maxDays) {
          var capped = new Date(d1);
          capped.setDate(capped.getDate() + maxDays - 1);
          selTo = formatYMD(capped.getFullYear(), capped.getMonth(), capped.getDate());
        }
        clickStep = 0;
      }

      if (directFromInput) directFromInput.value = selFrom;
      if (directToInput) directToInput.value = selTo;

      renderCalendar();
    }

    window.handleCalendarDayClick = function (dateStr) {
      handleDayClick(dateStr);
    };

    window.openChartCalendar = function () {
      var fromEl = document.getElementById("chart-form-from") || document.getElementById("chart-input-from");
      var toEl = document.getElementById("chart-form-to") || document.getElementById("chart-input-to");
      if (fromEl && fromEl.value) {
        selFrom = fromEl.value;
        var parts = selFrom.split("-");
        if (parts.length === 3) {
          viewYear = parseInt(parts[0], 10);
          viewMonth = parseInt(parts[1], 10) - 1;
        }
      }
      if (toEl && toEl.value) {
        selTo = toEl.value;
      }
      if (panel) {
        panel.style.display = "block";
        panel.classList.remove("hidden");
        if (panel.id === "chart-calendar-modal") {
          panel.classList.add("flex");
        }
      }
      var btnLabel = document.getElementById("chart-cal-btn-label");
      if (btnLabel) btnLabel.textContent = isId ? "Kalender (Buka)" : "Calendar (Open)";
      var periodSelect = document.getElementById("chart-period");
      if (periodSelect) periodSelect.value = "custom";
      clickStep = 0;
      renderCalendar();
    };

    window.closeChartCalendar = function () {
      if (panel) {
        panel.style.display = "none";
        panel.classList.add("hidden");
        if (panel.id === "chart-calendar-modal") {
          panel.classList.remove("flex");
        }
      }
      var btnLabel = document.getElementById("chart-cal-btn-label");
      var periodSelect = document.getElementById("chart-period");
      if (btnLabel) {
        if (periodSelect && periodSelect.value === "custom" && selFrom && selTo) {
          btnLabel.textContent = selFrom + " - " + selTo;
        } else {
          btnLabel.textContent = isId ? "Pilih Kalender" : "Calendar Picker";
        }
      }
    };

    window.toggleCalendarPanel = function (forceOpen) {
      if (!panel) return;
      var isHidden = (panel.style.display === "none") || panel.classList.contains("hidden");
      if (forceOpen === true || (forceOpen === undefined && isHidden)) {
        window.openChartCalendar();
      } else if (forceOpen === false || (forceOpen === undefined && !isHidden)) {
        window.closeChartCalendar();
      }
    };

    window.renderChartCalendar = function () {
      renderCalendar();
    };

    window.toggleChartCustomPeriod = function (select) {
      if (select.value === "custom") {
        window.openChartCalendar();
      } else {
        window.closeChartCalendar();
        select.form.submit();
      }
    };

    window.prevCalendarMonth = function () {
      viewMonth--;
      if (viewMonth < 0) {
        viewMonth = 11;
        viewYear--;
      }
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      selFrom = formatYMD(viewYear, viewMonth, 1);
      selTo = formatYMD(viewYear, viewMonth, maxDays);
      clickStep = 0;
      renderCalendar();
    };

    window.nextCalendarMonth = function () {
      viewMonth++;
      if (viewMonth > 11) {
        viewMonth = 0;
        viewYear++;
      }
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      selFrom = formatYMD(viewYear, viewMonth, 1);
      selTo = formatYMD(viewYear, viewMonth, maxDays);
      clickStep = 0;
      renderCalendar();
    };

    window.selectFullMonthRange = function () {
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      selFrom = formatYMD(viewYear, viewMonth, 1);
      selTo = formatYMD(viewYear, viewMonth, maxDays);
      clickStep = 0;
      renderCalendar();
    };

    window.selectMonthToDateRange = function () {
      var maxDays = getDaysInMonth(viewYear, viewMonth);
      selFrom = formatYMD(viewYear, viewMonth, 1);
      var today = new Date();
      if (today.getFullYear() === viewYear && today.getMonth() === viewMonth) {
        selTo = formatYMD(viewYear, viewMonth, today.getDate());
      } else {
        selTo = formatYMD(viewYear, viewMonth, maxDays);
      }
      clickStep = 0;
      renderCalendar();
    };

    window.onDirectDateChange = function () {
      if (directFromInput && directFromInput.value) {
        selFrom = directFromInput.value;
        var fDate = new Date(selFrom);
        viewYear = fDate.getFullYear();
        viewMonth = fDate.getMonth();
      }
      if (directToInput && directToInput.value) {
        selTo = directToInput.value;
      }
      if (selFrom && selTo) {
        var d1 = new Date(selFrom);
        var d2 = new Date(selTo);
        if (d2 < d1) {
          selTo = selFrom;
        }
        var maxDays = getDaysInMonth(viewYear, viewMonth);
        var diffDays = Math.round((d2 - d1) / (24 * 60 * 60 * 1000)) + 1;
        if (diffDays > maxDays) {
          var capped = new Date(d1);
          capped.setDate(capped.getDate() + maxDays - 1);
          selTo = formatYMD(capped.getFullYear(), capped.getMonth(), capped.getDate());
          if (directToInput) directToInput.value = selTo;
        }
      }
      clickStep = 0;
      renderCalendar();
    };

    window.applyChartCalendar = function () {
      var fromEl = document.getElementById("chart-form-from") || document.getElementById("chart-input-from");
      var toEl = document.getElementById("chart-form-to") || document.getElementById("chart-input-to");
      if (fromEl) fromEl.value = selFrom;
      if (toEl) toEl.value = selTo;
      var periodSelect = document.getElementById("chart-period");
      if (periodSelect) periodSelect.value = "custom";
      if (typeof syncChartDates === "function") {
        syncChartDates();
      }
      var form = document.getElementById("chart-period-form");
      if (form) {
        form.submit();
      } else if (periodSelect && periodSelect.form) {
        periodSelect.form.submit();
      }
    };

    if (panel) {
      panel.addEventListener("click", function (e) {
        e.stopPropagation();
        if (panel.id === "chart-calendar-modal" && e.target === panel) {
          window.closeChartCalendar();
        }
      });
    }

    renderCalendar();
    var initBtnLabel = document.getElementById("chart-cal-btn-label");
    var initPeriodSelect = document.getElementById("chart-period");
    if (initBtnLabel && initPeriodSelect && initPeriodSelect.value === "custom" && selFrom && selTo) {
      initBtnLabel.textContent = selFrom + " - " + selTo;
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", setupCalendar);
  } else {
    setupCalendar();
  }
})();




