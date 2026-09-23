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

    var totalDuration = duration || 3000;
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
    var d = duration || 3000;
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

  window.queueGlobalToast = function (title, message, duration, type) {
    var d = duration || 3000;
    try {
      sessionStorage.setItem(TOAST_KEY, JSON.stringify({
        title: title,
        message: message,
        startedAt: Date.now(),
        duration: d,
        type: type || 'info'
      }));
    } catch (e) {}
  };

  function checkToastOnLoad() {
    try {
      var params = new URLSearchParams(window.location.search);
      var savedParam = params.get("saved");
      if (savedParam) {
        params.delete("saved");
        var newSearch = params.toString() ? "?" + params.toString() : "";
        window.history.replaceState({}, document.title, window.location.pathname + newSearch);
        var isId = (document.documentElement.lang || "id") === "id";
        if (savedParam === "config") {
          window.showGlobalToast(
            isId ? "Konfigurasi Berhasil Disimpan" : "Configuration Saved Successfully",
            isId ? "Pengaturan persentase bagi hasil berhasil diperbarui." : "Profit sharing configuration has been saved successfully.",
            3000,
            "success"
          );
        } else {
          window.showGlobalToast(
            isId ? "Data Berhasil Disimpan" : "Data Saved Successfully",
            isId ? ('Perubahan data karyawan "' + decodeURIComponent(savedParam) + '" telah berhasil disimpan.') : ('Employee data changes for "' + decodeURIComponent(savedParam) + '" have been saved successfully.'),
            3000,
            "success"
          );
        }
        return;
      }

      var savedDisc = params.get("saved_disc");
      if (savedDisc) {
        params.delete("saved_disc");
        var newSearch = params.toString() ? "?" + params.toString() : "";
        window.history.replaceState({}, document.title, window.location.pathname + newSearch);
        var isId = (document.documentElement.lang || "id") === "id";
        window.showGlobalToast(
          isId ? "Diskon Berhasil Disimpan" : "Discount Saved Successfully",
          isId ? ('Diskon "' + decodeURIComponent(savedDisc) + '" telah berhasil ditambahkan ke katalog.') : ('Discount "' + decodeURIComponent(savedDisc) + '" has been added to the catalog.'),
          3000,
          "success"
        );
        return;
      }

      var deletedDisc = params.get("deleted_disc");
      if (deletedDisc) {
        params.delete("deleted_disc");
        var newSearch = params.toString() ? "?" + params.toString() : "";
        window.history.replaceState({}, document.title, window.location.pathname + newSearch);
        var isId = (document.documentElement.lang || "id") === "id";
        window.showGlobalToast(
          isId ? "Diskon Berhasil Dihapus" : "Discount Deleted Successfully",
          isId ? "Diskon telah berhasil dihapus dari sistem." : "Discount has been successfully removed.",
          3000,
          "success"
        );
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
  var discountSection = document.getElementById("transaction-discount-section");
  var discountSelect = document.getElementById("transaction-discount-select");
  var discountOrderAmountInput = document.getElementById("order-discount-amount-input");
  var discountBadge = document.getElementById("discount-applied-badge");
  var summarySubtotalRow = document.getElementById("summary-subtotal-row");
  var summaryDiscountRow = document.getElementById("summary-discount-row");
  var summaryDiscountLabel = document.getElementById("summary-discount-label");
  var subtotalDisplay = document.getElementById("transaction-subtotal-display");
  var discountDisplay = document.getElementById("transaction-discount-display");

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
  window.formatIDR = formatIDR;

  function updateTotal() {
    if (!totalDisplay) return;
    var subtotal = 0;
    if (container) {
      container.querySelectorAll(".amount-input").forEach(function (input) {
        var rawVal = input.value.replace(/[^0-9.]/g, "");
        var val = parseFloat(rawVal);
        if (!isNaN(val) && val > 0) {
          subtotal += val;
        }
      });
    }

    var kind = kindSelect ? kindSelect.value : "income";
    var discountAmt = 0;
    var promoName = "";

    if (kind === "income") {
      if (discountSection) discountSection.classList.remove("hidden");
      if (discountSelect && discountSelect.value) {
        var opt = discountSelect.options[discountSelect.selectedIndex];
        if (opt) {
          var dType = opt.dataset.type;
          var dVal = parseFloat(opt.dataset.value) || 0;
          promoName = opt.dataset.name || "";
          if (dType === "PERCENTAGE") {
            discountAmt = Math.round(subtotal * dVal / 100);
          } else if (dType === "FIXED_AMOUNT") {
            var fixedUnits = (dVal >= 100000) ? Math.round(dVal / 100) : dVal;
            discountAmt = Math.min(subtotal, fixedUnits);
          }
        }
      }
    } else {
      if (discountSection) discountSection.classList.add("hidden");
      if (discountSelect) discountSelect.value = "";
    }

    if (discountOrderAmountInput) {
      discountOrderAmountInput.value = discountAmt;
    }

    var netTotal = Math.max(0, subtotal - discountAmt);

    if (subtotalDisplay) {
      subtotalDisplay.textContent = formatIDR(subtotal);
    }
    if (discountDisplay) {
      discountDisplay.textContent = "-" + formatIDR(discountAmt);
    }
    if (discountBadge) {
      if (discountAmt > 0) {
        discountBadge.textContent = "-" + formatIDR(discountAmt);
        discountBadge.classList.remove("hidden");
      } else {
        discountBadge.classList.add("hidden");
      }
    }

    var isId = (document.documentElement.lang === "id");
    if (summaryDiscountLabel) {
      summaryDiscountLabel.textContent = (isId ? "Diskon" : "Discount") + (promoName ? " (" + promoName + "):" : ":");
    }

    if (summarySubtotalRow && summaryDiscountRow) {
      if (discountAmt > 0) {
        summarySubtotalRow.classList.remove("hidden");
        summaryDiscountRow.classList.remove("hidden");
      } else {
        summarySubtotalRow.classList.add("hidden");
        summaryDiscountRow.classList.add("hidden");
      }
    }

    totalDisplay.textContent = formatIDR(netTotal);
  }

  function findCategory(catName) {
    if (!catName) return null;
    var lower = catName.toLowerCase().trim();
    for (var i = 0; i < catalog.length; i++) {
      var cName = catalog[i].name.toLowerCase();
      if (cName === lower) return catalog[i];
      // Flexible matching (e.g. "haircut" matches "haircut services")
      if (lower.indexOf("bundle") !== -1 && cName.indexOf("bundle") !== -1) return catalog[i];
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
      itemSelect.required = true;
      foundCat.items.forEach(function (item) {
        var opt = document.createElement("option");
        opt.value = item.name;
        opt.textContent = item.name + (item.amount_label ? " (" + item.amount_label + ")" : "");
        opt.dataset.amount = item.amount;
        if (item.bundle_id) {
          opt.dataset.bundleId = item.bundle_id;
        }
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
        var bundleInput = row.querySelector(".bundle-id-input");
        if (bundleInput) {
          bundleInput.value = it.bundle_id || "0";
        }
      } else if (!selectedItemName) {
        if (amountInput) amountInput.value = "";
        var bundleInput = row.querySelector(".bundle-id-input");
        if (bundleInput) bundleInput.value = "0";
      }
    } else if (catVal) {
      itemSelect.required = (catVal !== "Other");
      var customOpt = document.createElement("option");
      customOpt.value = "Custom";
      customOpt.textContent = isId ? "Item Lainnya..." : "Custom Item...";
      customOpt.dataset.amount = "0";
      customOpt.selected = true;
      itemSelect.appendChild(customOpt);
      if (!selectedItemName) {
        if (amountInput) amountInput.value = "";
        var bundleInput = row.querySelector(".bundle-id-input");
        if (bundleInput) bundleInput.value = "0";
      }
    } else {
      itemSelect.required = false;
      if (!selectedItemName) {
        if (amountInput) amountInput.value = "";
        var bundleInput = row.querySelector(".bundle-id-input");
        if (bundleInput) bundleInput.value = "0";
      }
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
        '<input type="hidden" name="bundle_id[]" class="bundle-id-input" value="0">' +
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
        var amountInput = row.querySelector(".amount-input");
        var bundleInput = row.querySelector(".bundle-id-input");
        if (!e.target.value) {
          if (amountInput) amountInput.value = "";
          if (bundleInput) bundleInput.value = "0";
        }
        updateTotal();
      }
    }
    if (e.target && e.target.classList.contains("item-select")) {
      var row = e.target.closest(".category-row");
      if (row) {
        e.target.setCustomValidity("");
        var opt = e.target.options[e.target.selectedIndex];
        var amountInput = row.querySelector(".amount-input");
        var bundleInput = row.querySelector(".bundle-id-input");
        if (opt && opt.value) {
          if (opt.dataset && opt.dataset.amount && parseFloat(opt.dataset.amount) > 0) {
            if (amountInput) {
              amountInput.value = opt.dataset.amount;
            }
          }
          if (bundleInput) {
            bundleInput.value = (opt.dataset && opt.dataset.bundleId) ? opt.dataset.bundleId : "0";
          }
        } else {
          // User chose "-- Pilih layanan / item --", clear the amount!
          if (amountInput) {
            amountInput.value = "";
          }
          if (bundleInput) {
            bundleInput.value = "0";
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

  // Listener for Discount Dropdown
  if (discountSelect) {
    discountSelect.addEventListener("change", function () {
      updateTotal();
    });
  }

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
          var bundleInput = row.querySelector(".bundle-id-input");
          if (catSel) catSel.value = "";
          if (itemSel) itemSel.innerHTML = '<option value="">-- Pilih layanan / item --</option>';
          if (amtInput) amtInput.value = "";
          if (bundleInput) bundleInput.value = "0";
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

// Global Logout Confirmation Modal (Option A - Modern, elegant, non-intrusive)
(function () {
  var activeLogoutForm = null;

  function createLogoutModal() {
    var isId = (document.documentElement.lang || "id") === "id";

    var title = isId ? "Konfirmasi Keluar" : "Confirm Sign Out";
    var subtitle = isId ? "Sesi kasir Anda akan diakhiri." : "Your cashier session will be ended.";
    var message = isId
      ? "Apakah Anda yakin ingin keluar dari sistem POS Phoenix? Pastikan seluruh transaksi aktif Anda telah tersimpan."
      : "Are you sure you want to sign out from POS Phoenix? Please make sure all active transactions have been saved.";
    var cancelText = isId ? "Batal" : "Cancel";
    var confirmText = isId ? "Ya, Keluar" : "Yes, Sign Out";

    var modal = document.getElementById("pos-logout-modal");
    if (modal) {
      // Update text in case language was switched
      var titleEl = document.getElementById("pos-logout-title");
      var subEl = document.getElementById("pos-logout-subtitle");
      var msgEl = document.getElementById("pos-logout-message");
      var cancelEl = document.getElementById("pos-logout-cancel");
      var confirmEl = document.getElementById("pos-logout-confirm");
      if (titleEl) titleEl.textContent = title;
      if (subEl) subEl.textContent = subtitle;
      if (msgEl) msgEl.textContent = message;
      if (cancelEl) cancelEl.textContent = cancelText;
      if (confirmEl) {
        confirmEl.innerHTML =
          '<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>' +
          confirmText;
      }
      return modal;
    }

    modal = document.createElement("div");
    modal.id = "pos-logout-modal";
    modal.className = "fixed inset-0 z-[999999] hidden items-center justify-center p-4 sm:p-6";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "none";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.68)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    modal.innerHTML =
      '<div class="relative w-full max-w-md rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-6 sm:p-7 text-slate-800 dark:text-slate-100 transition-all" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="flex items-center gap-3.5 pb-4 border-b border-slate-100 dark:border-slate-800/80">' +
          '<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-rose-50 dark:bg-rose-950/70 border border-rose-100 dark:border-rose-900/60 text-rose-600 dark:text-rose-400 shadow-sm">' +
            '<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>' +
          '</div>' +
          '<div class="min-w-0">' +
            '<h3 id="pos-logout-title" class="font-black text-base sm:text-lg tracking-tight text-slate-900 dark:text-white">' + title + '</h3>' +
            '<p id="pos-logout-subtitle" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">' + subtitle + '</p>' +
          '</div>' +
        '</div>' +
        '<div class="mt-5 mb-6">' +
          '<p id="pos-logout-message" class="text-xs sm:text-sm text-slate-600 dark:text-slate-300 leading-relaxed">' + message + '</p>' +
        '</div>' +
        '<div class="pt-4 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-end gap-3">' +
          '<button id="pos-logout-cancel" class="btn-secondary !min-h-11 px-5 text-xs sm:text-sm font-semibold transition cursor-pointer" type="button">' +
            cancelText +
          '</button>' +
          '<button id="pos-logout-confirm" class="inline-flex min-h-11 items-center justify-center text-center rounded-xl px-6 font-bold shadow-lg transition active:scale-95 bg-rose-600 hover:bg-rose-700 text-white shadow-rose-600/25 text-xs sm:text-sm cursor-pointer" type="button">' +
            '<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>' +
            confirmText +
          '</button>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.style.display = "none";
      modal.classList.add("hidden");
      modal.classList.remove("flex");
      activeLogoutForm = null;
    }

    document.getElementById("pos-logout-cancel").addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && modal.style.display !== "none" && !modal.classList.contains("hidden")) {
        closeModal();
      }
    });

    document.getElementById("pos-logout-confirm").addEventListener("click", function () {
      if (activeLogoutForm) {
        var formToSubmit = activeLogoutForm;
        activeLogoutForm = null;
        formToSubmit.dataset.posLogoutConfirmed = "true";
        closeModal();
        try {
          sessionStorage.removeItem("pos_admin_selected_branch");
          localStorage.removeItem("pos_admin_selected_branch");
        } catch (err) {}
        if (typeof HTMLFormElement.prototype.submit === "function") {
          HTMLFormElement.prototype.submit.call(formToSubmit);
        } else {
          formToSubmit.submit();
        }
      }
    });

    return modal;
  }

  function handleLogoutIntercept(e, form) {
    if (!form) return;
    var action = form.getAttribute("action") || "";
    if (action.indexOf("/logout") === -1) return;
    try {
      sessionStorage.removeItem("pos_admin_selected_branch");
      localStorage.removeItem("pos_admin_selected_branch");
    } catch (err) {}
    if (form.dataset.posLogoutConfirmed === "true") return;

    if (e) {
      e.preventDefault();
      e.stopPropagation();
      if (typeof e.stopImmediatePropagation === "function") {
        e.stopImmediatePropagation();
      }
    }
    activeLogoutForm = form;
    var modal = createLogoutModal();
    modal.style.display = "flex";
    modal.classList.remove("hidden");
    modal.classList.add("flex");
  }

  function initLogoutInterception() {
    // 1. Intercept clicks on any element inside a logout form (Capture phase)
    document.addEventListener("click", function (e) {
      var target = e.target;
      if (!target) return;
      var btn = target.closest ? target.closest("button, a, input[type='submit']") : null;
      if (!btn) return;
      // Do not intercept if clicking buttons inside the confirmation modal itself
      if (btn.id === "pos-logout-confirm" || btn.id === "pos-logout-cancel") return;

      var form = btn.closest ? btn.closest("form") : null;
      if (form) {
        var action = form.getAttribute("action") || "";
        if (action.indexOf("/logout") !== -1) {
          handleLogoutIntercept(e, form);
        }
      }
    }, true);

    // 2. Intercept form submit event (Capture phase)
    document.addEventListener("submit", function (e) {
      var form = e.target;
      if (form) {
        var action = form.getAttribute("action") || "";
        if (action.indexOf("/logout") !== -1) {
          handleLogoutIntercept(e, form);
        }
      }
    }, true);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initLogoutInterception);
  } else {
    initLogoutInterception();
  }
})();

// Global Transaction Confirmation Modals (Save & Reverse)
(function () {
  var activeSaveTxForm = null;
  var activeReverseTxForm = null;

  function getFormatIDR(amt) {
    if (typeof window.formatIDR === "function") return window.formatIDR(amt);
    return "Rp " + amt.toLocaleString("id-ID") + ",00";
  }

  // --- 1. MODAL SIMPAN TRANSAKSI ---
  function showSaveTxModal(form) {
    var isId = (document.documentElement.lang || "id") === "id";
    activeSaveTxForm = form;

    var kindSelect = form.querySelector('select[name="kind"]');
    var isIncome = !kindSelect || kindSelect.value === "income";
    var kindLabel = isIncome ? (isId ? "Pemasukan" : "Income") : (isId ? "Pengeluaran" : "Expense");
    var kindBadgeClass = isIncome 
      ? "bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-900/60 text-emerald-700 dark:text-emerald-400"
      : "bg-rose-50 dark:bg-rose-950/60 border border-rose-200 dark:border-rose-900/60 text-rose-700 dark:text-rose-400";

    var noteInput = form.querySelector('[name="note"]');
    var noteVal = (noteInput && noteInput.value.trim()) || "";

    // Calculate total & list of items
    var container = document.getElementById("category-rows-container");
    var itemsList = [];
    var calculatedTotal = 0;

    if (container) {
      container.querySelectorAll(".category-row").forEach(function (row) {
        var catSel = row.querySelector(".category-select");
        var itemSel = row.querySelector(".item-select");
        var amtInput = row.querySelector(".amount-input");

        var catVal = catSel ? catSel.value.trim() : "";
        var itemVal = itemSel && itemSel.value ? itemSel.value.trim() : "";
        var rawVal = amtInput ? amtInput.value.replace(/[^0-9.]/g, "") : "";
        var amt = parseFloat(rawVal) || 0;

        if (catVal || itemVal || amt > 0) {
          calculatedTotal += amt;
          var title = catVal + (itemVal ? ": " + itemVal : "");
          if (!title) title = isId ? "Item Layanan" : "Service Item";
          itemsList.push({ title: title, amountStr: getFormatIDR(amt) });
        }
      });
    }

    var totalDisplay = document.getElementById("transaction-total-display");
    var totalFormatted = (totalDisplay && totalDisplay.textContent.trim()) || getFormatIDR(calculatedTotal);

    var discountOrderInput = document.getElementById("order-discount-amount-input");
    var discountVal = discountOrderInput ? (parseFloat(discountOrderInput.value) || 0) : 0;
    var discountSelectEl = document.getElementById("transaction-discount-select");
    var selectedDiscountOpt = discountSelectEl && discountSelectEl.selectedIndex > 0 ? discountSelectEl.options[discountSelectEl.selectedIndex] : null;
    var discountNameStr = selectedDiscountOpt ? (selectedDiscountOpt.dataset.name || selectedDiscountOpt.textContent.trim()) : "";

    var existing = document.getElementById("pos-save-tx-modal");
    if (existing) existing.remove();

    var modal = document.createElement("div");
    modal.id = "pos-save-tx-modal";
    modal.className = "fixed inset-0 z-[999999] flex items-center justify-center p-4 sm:p-6";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "flex";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.68)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    var itemsHtml = "";
    if (itemsList.length > 0) {
      itemsHtml = '<div class="mt-4 max-h-40 overflow-y-auto space-y-1.5 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 p-3 text-xs">';
      itemsList.forEach(function (it) {
        itemsHtml += '<div class="flex items-center justify-between gap-2">' +
          '<span class="truncate font-medium text-slate-700 dark:text-slate-300">• ' + it.title + '</span>' +
          '<span class="shrink-0 font-bold text-slate-900 dark:text-white">' + it.amountStr + '</span>' +
        '</div>';
      });
      itemsHtml += '</div>';
    }

    var noteHtml = "";
    if (noteVal) {
      noteHtml = '<div class="mt-2.5 text-xs text-slate-500 dark:text-slate-400 italic break-words">' +
        (isId ? 'Catatan: "' : 'Note: "') + noteVal + '"' +
      '</div>';
    }

    modal.innerHTML =
      '<div class="relative w-full max-w-md rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-6 sm:p-7 text-slate-800 dark:text-slate-100 transition-all" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="flex items-center gap-3.5 pb-4 border-b border-slate-100 dark:border-slate-800/80">' +
          '<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-blue-50 dark:bg-blue-950/70 border border-blue-100 dark:border-blue-900/60 text-blue-600 dark:text-blue-400 shadow-sm">' +
            '<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>' +
          '</div>' +
          '<div class="min-w-0">' +
            '<h3 class="font-black text-base sm:text-lg tracking-tight text-slate-900 dark:text-white">' + (isId ? "Konfirmasi Simpan Transaksi" : "Confirm Save Transaction") + '</h3>' +
            '<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">' + (isId ? "Pastikan rincian transaksi sudah benar." : "Please review the transaction details.") + '</p>' +
          '</div>' +
        '</div>' +
        '<div class="mt-4 p-4 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 text-center">' +
          '<span class="text-[11px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-400">' + (isId ? "Total Pembayaran" : "Total Amount") + '</span>' +
          '<p class="text-2xl sm:text-3xl font-black ' + (isIncome ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400') + ' mt-1 tracking-tight">' + totalFormatted + '</p>' +
          (discountVal > 0 ? (
            '<div class="mt-2.5 pt-2 border-t border-slate-200/70 dark:border-slate-800 flex items-center justify-between text-xs px-1">' +
              '<span class="text-slate-500 dark:text-slate-400">' + (isId ? "Subtotal: " : "Subtotal: ") + '<b class="text-slate-700 dark:text-slate-300">' + getFormatIDR(calculatedTotal) + '</b></span>' +
              '<span class="font-bold text-emerald-600 dark:text-emerald-400">-' + getFormatIDR(discountVal) + '</span>' +
            '</div>'
          ) : '') +
          '<div class="mt-2 flex justify-center">' +
            '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold ' + kindBadgeClass + '">' + kindLabel + '</span>' +
          '</div>' +
        '</div>' +
        itemsHtml +
        noteHtml +
        '<div class="mt-5 pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-2.5">' +
          '<button id="pos-tx-save-print" class="w-full inline-flex min-h-12 items-center justify-center text-center rounded-xl px-5 font-bold shadow-lg transition active:scale-95 bg-blue-600 hover:bg-blue-700 text-white shadow-blue-600/25 text-sm cursor-pointer" type="button">' +
            '<svg class="w-4 h-4 mr-2 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>' +
            '<span>' + (isId ? "Simpan & Cetak Struk" : "Save & Print Receipt") + '</span>' +
          '</button>' +
          '<div class="grid grid-cols-2 gap-2.5">' +
            '<button id="pos-tx-cancel" class="btn-secondary !min-h-11 w-full px-4 text-xs sm:text-sm font-semibold transition cursor-pointer" type="button">' +
              (isId ? "Batal" : "Cancel") +
            '</button>' +
            '<button id="pos-tx-save-only" class="btn-secondary !min-h-11 w-full px-4 text-xs sm:text-sm font-semibold transition cursor-pointer" type="button">' +
              (isId ? "💾 Simpan Saja" : "💾 Save Only") +
            '</button>' +
          '</div>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.remove();
      activeSaveTxForm = null;
    }

    document.getElementById("pos-tx-cancel").addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && document.getElementById("pos-save-tx-modal")) {
        closeModal();
      }
    });

    function submitTransaction(shouldPrint) {
      if (!activeSaveTxForm) return;
      var formToSubmit = activeSaveTxForm;
      activeSaveTxForm = null;
      formToSubmit.dataset.posTxConfirmed = "true";

      // Store transaction details in sessionStorage for thermal receipt
      try {
        var opEl = document.querySelector(".topbar-inner p.muted") || document.querySelector("aside p.text-slate-300");
        var opName = opEl ? opEl.textContent.split("·")[0].trim() : "Kasir";
        var txDetails = {
          id: "",
          amount: totalFormatted,
          subtotal: discountVal > 0 ? getFormatIDR(calculatedTotal) : "",
          discount: discountVal > 0 ? getFormatIDR(discountVal) : "",
          discountName: discountNameStr,
          kind: isIncome ? "income" : "expense",
          items: itemsList.map(function (it) {
            return { name: it.title, amount: it.amountStr };
          }),
          note: noteVal,
          operator: opName,
          date: new Date().toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" }) + " " +
                new Date().toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" })
        };
        if (shouldPrint) {
          sessionStorage.setItem("pendingTxPrintPrompt", JSON.stringify(txDetails));
          sessionStorage.setItem("autoOpenThermalPrint", "true");
        } else {
          sessionStorage.removeItem("pendingTxPrintPrompt");
          sessionStorage.removeItem("autoOpenThermalPrint");
          sessionStorage.setItem("pos_active_toast", JSON.stringify({
            title: isId ? "Transaksi Berhasil Disimpan" : "Transaction Saved",
            message: isId ? "Transaksi sebesar " + totalFormatted + " telah berhasil dicatat." : "Transaction of " + totalFormatted + " has been successfully recorded.",
            startedAt: Date.now(),
            duration: 3500,
            type: "success"
          }));
        }
      } catch (e) {}

      closeModal();
      if (typeof HTMLFormElement.prototype.submit === "function") {
        HTMLFormElement.prototype.submit.call(formToSubmit);
      } else {
        formToSubmit.submit();
      }
    }

    document.getElementById("pos-tx-save-only").addEventListener("click", function () {
      submitTransaction(false);
    });

    document.getElementById("pos-tx-save-print").addEventListener("click", function () {
      submitTransaction(true);
    });
  }

  // --- 2. MODAL BATALKAN TRANSAKSI ---
  function showReverseTxModal(form) {
    var isId = (document.documentElement.lang || "id") === "id";
    activeReverseTxForm = form;

    var reasonInput = form.querySelector('[name="reason"]');
    var reason = reasonInput ? reasonInput.value.trim() : "";
    var txId = form.dataset.txId || (form.getAttribute("action") || "").replace(/[^0-9]/g, "") || "";

    var existing = document.getElementById("pos-reverse-tx-modal");
    if (existing) existing.remove();

    var modal = document.createElement("div");
    modal.id = "pos-reverse-tx-modal";
    modal.className = "fixed inset-0 z-[999999] flex items-center justify-center p-4 sm:p-6";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "flex";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.68)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    modal.innerHTML =
      '<div class="relative w-full max-w-md rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-6 sm:p-7 text-slate-800 dark:text-slate-100 transition-all" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="flex items-center gap-3.5 pb-4 border-b border-slate-100 dark:border-slate-800/80">' +
          '<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-rose-50 dark:bg-rose-950/70 border border-rose-100 dark:border-rose-900/60 text-rose-600 dark:text-rose-400 shadow-sm">' +
            '<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>' +
          '</div>' +
          '<div class="min-w-0">' +
            '<h3 class="font-black text-base sm:text-lg tracking-tight text-slate-900 dark:text-white">' + (isId ? "Konfirmasi Batalkan Transaksi" : "Confirm Transaction Reversal") + '</h3>' +
            '<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">' + (isId ? ("Transaksi #" + (txId ? txId : "") + " akan dibatalkan.") : ("Transaction #" + (txId ? txId : "") + " will be reversed.")) + '</p>' +
          '</div>' +
        '</div>' +
        '<div class="mt-4 p-4 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800">' +
          '<span class="text-xs font-bold text-slate-500 dark:text-slate-400">' + (isId ? "Alasan Pembatalan:" : "Reversal Reason:") + '</span>' +
          '<p class="text-sm font-semibold text-slate-800 dark:text-slate-200 mt-1 italic break-words">"' + reason + '"</p>' +
        '</div>' +
        '<p class="mt-3 text-xs text-rose-600 dark:text-rose-400 leading-relaxed font-medium">' +
          (isId ? "Tindakan ini akan membuat transaksi pembalik dan tercatat secara permanen di riwayat audit." : "This action will create a reversal transaction and be permanently recorded in the audit log.") +
        '</p>' +
        '<div class="mt-5 pt-4 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-end gap-3">' +
          '<button id="pos-reverse-cancel" class="btn-secondary !min-h-11 px-5 text-xs sm:text-sm font-semibold transition cursor-pointer" type="button">' +
            (isId ? "Batal" : "Cancel") +
          '</button>' +
          '<button id="pos-reverse-confirm" class="inline-flex min-h-11 items-center justify-center text-center rounded-xl px-6 font-bold shadow-lg transition active:scale-95 bg-rose-600 hover:bg-rose-700 text-white shadow-rose-600/25 text-xs sm:text-sm cursor-pointer" type="button">' +
            '<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>' +
            (isId ? "Ya, Batalkan" : "Yes, Reverse") +
          '</button>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.remove();
      activeReverseTxForm = null;
    }

    document.getElementById("pos-reverse-cancel").addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && document.getElementById("pos-reverse-tx-modal")) {
        closeModal();
      }
    });

    document.getElementById("pos-reverse-confirm").addEventListener("click", function () {
      if (activeReverseTxForm) {
        var formToSubmit = activeReverseTxForm;
        activeReverseTxForm = null;
        formToSubmit.dataset.posReverseConfirmed = "true";

        if (window.showGlobalToast) {
          window.showGlobalToast(
            isId ? "Transaksi Dibatalkan" : "Transaction Reversed",
            isId ? "Transaksi #" + (txId ? txId : "") + " telah berhasil dibatalkan." : "Transaction #" + (txId ? txId : "") + " has been successfully reversed.",
            3000,
            "success"
          );
        }

        closeModal();
        if (typeof HTMLFormElement.prototype.submit === "function") {
          HTMLFormElement.prototype.submit.call(formToSubmit);
        } else {
          formToSubmit.submit();
        }
      }
    });
  }

  // Intercept form submit event in capture phase
  document.addEventListener("submit", function (e) {
    var form = e.target;
    if (!form) return;

    // A. Intercept new transaction creation form
    if (form.id === "transaction-form" || form.getAttribute("action") === "/transactions") {
      if (form.dataset.posTxConfirmed === "true") return;

      // Check native validation first
      if (form.reportValidity && !form.reportValidity()) {
        return;
      }

      // Check each category row to ensure items are selected
      var isId = (document.documentElement.lang || "id") === "id";
      var rows = form.querySelectorAll(".category-row");
      var hasValidRow = false;
      var totalAmount = 0;

      for (var r = 0; r < rows.length; r++) {
        var row = rows[r];
        var catSel = row.querySelector(".category-select");
        var itemSel = row.querySelector(".item-select");
        var amtInput = row.querySelector(".amount-input");

        var catVal = catSel ? catSel.value.trim() : "";
        var itemVal = itemSel ? itemSel.value.trim() : "";
        var rawAmt = amtInput ? amtInput.value.replace(/[^0-9.]/g, "") : "";
        var amt = parseFloat(rawAmt) || 0;

        // If category is selected but item is omitted (and not 'Other')
        if (catVal && catVal !== "Other" && !itemVal) {
          if (itemSel) {
            itemSel.focus();
            if (itemSel.reportValidity) {
              itemSel.setCustomValidity(isId ? "Silakan pilih layanan/item terlebih dahulu." : "Please select a service/item.");
              itemSel.reportValidity();
            }
          }
          if (window.showGlobalToast) {
            window.showGlobalToast(
              isId ? "Layanan Belum Dipilih" : "Item Not Selected",
              isId ? "Silakan pilih layanan/item untuk kategori " + catVal + " sebelum menyimpan." : "Please select a service/item for category " + catVal + ".",
              3000,
              "warning"
            );
          }
          e.preventDefault();
          e.stopPropagation();
          return;
        }

        if (catVal && amt > 0) {
          hasValidRow = true;
          totalAmount += amt;
        }
      }

      if (!hasValidRow || totalAmount <= 0) {
        if (window.showGlobalToast) {
          window.showGlobalToast(
            isId ? "Rincian Transaksi Kosong" : "Empty Transaction Details",
            isId ? "Pastikan layanan telah dipilih dan nominal pembayaran lebih dari Rp 0." : "Please make sure an item is selected and amount is greater than Rp 0.",
            3000,
            "warning"
          );
        }
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      e.preventDefault();
      e.stopPropagation();
      if (typeof e.stopImmediatePropagation === "function") {
        e.stopImmediatePropagation();
      }
      showSaveTxModal(form);
      return;
    }

    // B. Intercept transaction reversal form
    var action = form.getAttribute("action") || "";
    if (action.indexOf("/transactions/") !== -1 && action.indexOf("/reverse") !== -1) {
      if (form.dataset.posReverseConfirmed === "true") return;

      var reasonInput = form.querySelector('[name="reason"]');
      var reason = reasonInput ? reasonInput.value.trim() : "";
      var isId = (document.documentElement.lang || "id") === "id";

      if (reason.length < 5) {
        if (reasonInput) {
          reasonInput.focus();
          if (reasonInput.reportValidity) reasonInput.reportValidity();
        }
        if (window.showGlobalToast) {
          window.showGlobalToast(
            isId ? "Alasan Kurang Lengkap" : "Reason Too Short",
            isId ? "Alasan pembatalan minimal harus 5 karakter." : "Reversal reason must be at least 5 characters.",
            2500,
            "warning"
          );
        }
        e.preventDefault();
        e.stopPropagation();
        return;
      }

      e.preventDefault();
      e.stopPropagation();
      if (typeof e.stopImmediatePropagation === "function") {
        e.stopImmediatePropagation();
      }
      showReverseTxModal(form);
      return;
    }
  }, true);

  // =========================================================================
  // --- 3. MODUL THERMAL PRINTER 58mm (OKAY 58D & BROWSER PRINT FALLBACK) ---
  // =========================================================================
  var bluetoothDevice = null;
  var bluetoothCharacteristic = null;

  var ThermalPrinter = {
    isSupported: function () {
      return typeof navigator !== "undefined" && !!(navigator.bluetooth && navigator.bluetooth.requestDevice);
    },

    getStoreName: function () {
      return localStorage.getItem("pos_thermal_store_name") || "PARDIS BARBER SHOP";
    },

    getFooterNote: function () {
      return localStorage.getItem("pos_thermal_footer_note") || "Terima Kasih Atas Kunjungan Anda!";
    },

    formatLine32: function (left, right, width) {
      width = width || 32;
      left = (left || "").toString();
      right = (right || "").toString();
      if (left.length + right.length + 1 > width) {
        var availableForLeft = width - right.length - 1;
        if (availableForLeft > 0) {
          left = left.substring(0, availableForLeft);
        }
      }
      var spaces = width - left.length - right.length;
      if (spaces < 1) spaces = 1;
      return left + " ".repeat(spaces) + right;
    },

    centerText32: function (text, width) {
      width = width || 32;
      text = (text || "").toString().trim();
      if (text.length >= width) return text.substring(0, width);
      var totalSpaces = width - text.length;
      var leftSpaces = Math.floor(totalSpaces / 2);
      var rightSpaces = totalSpaces - leftSpaces;
      return " ".repeat(leftSpaces) + text + " ".repeat(rightSpaces);
    },

    generateReceiptLines: function (txData) {
      var lines = [];
      var isId = (document.documentElement.lang || "id") === "id";
      var storeName = this.getStoreName();
      var footerNote = this.getFooterNote();

      lines.push("================================");
      lines.push(this.centerText32(storeName));
      lines.push(this.centerText32("Pardis Barbershop"));
      lines.push("================================");

      var txNum = txData.id ? ("#" + txData.id) : (isId ? "BARU" : "NEW");
      lines.push(this.formatLine32(isId ? "No. Trx" : "Trx ID", txNum));
      lines.push(this.formatLine32(isId ? "Waktu" : "Date", txData.date || "-"));
      lines.push(this.formatLine32(isId ? "Kasir" : "Cashier", txData.operator || "-"));
      if (txData.kind && txData.kind !== "income") {
        var kindLabel = isId ? "Pengeluaran" : "Expense";
        lines.push(this.formatLine32(isId ? "Tipe" : "Type", kindLabel));
      }
      lines.push("--------------------------------");
      lines.push(this.formatLine32(isId ? "ITEM / LAYANAN" : "ITEM / SERVICE", isId ? "HARGA" : "PRICE"));
      lines.push("--------------------------------");

      var items = txData.items || [];
      if (items.length === 0) {
        lines.push(this.formatLine32(txData.category || (isId ? "Transaksi" : "Transaction"), txData.amount || "-"));
      } else {
        for (var i = 0; i < items.length; i++) {
          var it = items[i];
          var name = it.name || it.category || "Item";
          var amt = it.amount || "-";
          if (name.length > 19) {
            lines.push(name);
            lines.push(ThermalPrinter.formatLine32("", amt));
          } else {
            lines.push(ThermalPrinter.formatLine32(name, amt));
          }
        }
      }

      lines.push("--------------------------------");
      if (txData.discount && txData.subtotal) {
        lines.push(this.formatLine32(isId ? "Subtotal" : "Subtotal", txData.subtotal));
        var discLabel = isId ? "Diskon" : "Discount";
        lines.push(this.formatLine32(discLabel, "-" + txData.discount));
      }
      lines.push(this.formatLine32("TOTAL", txData.amount || "-"));
      if (txData.note) {
        lines.push(this.formatLine32(isId ? "Catatan" : "Note", txData.note));
      }
      lines.push("================================");
      lines.push(this.centerText32(footerNote));
      lines.push("================================");
      return lines;
    },

    renderReceiptHtml: function (txData) {
      var isId = (document.documentElement.lang || "id") === "id";
      var storeName = this.escapeHtml(this.getStoreName());
      var footerNote = this.escapeHtml(this.getFooterNote());
      var txNum = txData.id ? ("#" + this.escapeHtml(txData.id)) : (isId ? "BARU" : "NEW");
      var txDate = this.escapeHtml(txData.date || "-");
      var txOperator = this.escapeHtml(txData.operator || "-");
      var kindLabel = txData.kind === "income" ? (isId ? "Pemasukan" : "Income") : (isId ? "Pengeluaran" : "Expense");
      var totalAmt = this.escapeHtml(txData.amount || "-");

      var items = txData.items || [];
      var itemsHtml = "";
      if (items.length === 0) {
        var cat = this.escapeHtml(txData.category || (isId ? "Transaksi" : "Transaction"));
        itemsHtml = '<div class="flex justify-between items-start text-[11px] leading-snug w-full py-0.5">' +
          '<span class="text-left">' + cat + '</span>' +
          '<span class="font-bold text-right shrink-0 ml-2">' + totalAmt + '</span>' +
        '</div>';
      } else {
        for (var i = 0; i < items.length; i++) {
          var it = items[i];
          var itName = this.escapeHtml(it.name || it.category || "Item");
          var itAmt = this.escapeHtml(it.amount || "-");
          itemsHtml += '<div class="flex justify-between items-start text-[11px] leading-snug w-full py-0.5">' +
            '<span class="text-left pr-2">' + itName + '</span>' +
            '<span class="font-bold text-right shrink-0">' + itAmt + '</span>' +
          '</div>';
        }
      }

      var noteHtml = "";
      if (txData.note) {
        noteHtml = '<div class="text-[11px] text-slate-700 italic py-1 text-left break-words">' +
          (isId ? "Catatan: " : "Note: ") + this.escapeHtml(txData.note) +
        '</div>';
      }

      var discountBreakdownHtml = "";
      if (txData.discount && txData.subtotal) {
        discountBreakdownHtml =
          '<div class="flex justify-between items-center text-[11px] w-full py-0.5" style="color: #4b5563 !important;">' +
            '<span>' + (isId ? "Subtotal" : "Subtotal") + '</span>' +
            '<span class="text-right">' + this.escapeHtml(txData.subtotal) + '</span>' +
          '</div>' +
          '<div class="flex justify-between items-center text-[11px] w-full py-0.5 font-bold" style="color: #059669 !important;">' +
            '<span>' + (isId ? "Diskon" : "Discount") + '</span>' +
            '<span class="text-right">-' + this.escapeHtml(txData.discount) + '</span>' +
          '</div>';
      }

      var html =
        '<div class="font-mono text-[11px] leading-tight w-full select-none" style="color: #111827 !important; background-color: #ffffff !important;">' +
          '<!-- Centered Header -->' +
          '<div class="text-center pb-2">' +
            '<h4 class="font-black text-xs sm:text-sm tracking-wider uppercase" style="color: #000000 !important;">' + storeName + '</h4>' +
            '<p class="text-[10px] mt-0.5" style="color: #4b5563 !important;">Pardis Barbershop</p>' +
          '</div>' +

          '<!-- Edge-to-edge dashed Divider -->' +
          '<div style="border-bottom: 1px dashed #111827 !important; width: 100%; margin: 6px 0;"></div>' +

          '<!-- Meta info (Left Label, Right Value) -->' +
          '<div class="space-y-0.5 w-full py-1 text-[11px]">' +
            '<div class="flex justify-between items-center w-full">' +
              '<span style="color: #4b5563 !important;">' + (isId ? "No. Trx" : "Trx ID") + '</span>' +
              '<span class="font-semibold text-right" style="color: #111827 !important;">' + txNum + '</span>' +
            '</div>' +
            '<div class="flex justify-between items-center w-full">' +
              '<span style="color: #4b5563 !important;">' + (isId ? "Waktu" : "Date") + '</span>' +
              '<span class="text-right" style="color: #111827 !important;">' + txDate + '</span>' +
            '</div>' +
            '<div class="flex justify-between items-center w-full">' +
              '<span style="color: #4b5563 !important;">' + (isId ? "Kasir" : "Cashier") + '</span>' +
              '<span class="font-medium text-right" style="color: #111827 !important;">' + txOperator + '</span>' +
            '</div>' +
            (txData.kind && txData.kind !== "income" ? (
              '<div class="flex justify-between items-center w-full">' +
                '<span style="color: #4b5563 !important;">' + (isId ? "Tipe" : "Type") + '</span>' +
                '<span class="font-medium text-right" style="color: #111827 !important;">' + (isId ? "Pengeluaran" : "Expense") + '</span>' +
              '</div>'
            ) : '') +
          '</div>' +

          '<!-- Items Header Divider -->' +
          '<div style="border-bottom: 1px dashed #111827 !important; width: 100%; margin: 6px 0;"></div>' +
          '<div class="flex justify-between items-center text-[10px] font-bold uppercase tracking-wider w-full py-0.5" style="color: #4b5563 !important;">' +
            '<span>' + (isId ? "Item / Layanan" : "Item / Service") + '</span>' +
            '<span>' + (isId ? "Harga" : "Price") + '</span>' +
          '</div>' +
          '<div style="border-bottom: 1px dashed #9ca3af !important; width: 100%; margin: 4px 0;"></div>' +

          '<!-- Items List -->' +
          '<div class="py-1 space-y-1 w-full">' +
            itemsHtml +
          '</div>' +

          '<!-- Total Divider -->' +
          '<div style="border-bottom: 1px dashed #111827 !important; width: 100%; margin: 6px 0;"></div>' +
          discountBreakdownHtml +

          '<!-- Total (Flush Left & Right) -->' +
          '<div class="flex justify-between items-center text-xs font-black w-full py-1" style="color: #000000 !important;">' +
            '<span>TOTAL</span>' +
            '<span class="text-right">' + totalAmt + '</span>' +
          '</div>' +
          noteHtml +

          '<!-- Footer Divider -->' +
          '<div style="border-bottom: 1px dashed #111827 !important; width: 100%; margin: 8px 0 6px 0;"></div>' +

          '<!-- Centered Footer -->' +
          '<div class="text-center pt-1 pb-1 text-[10px] leading-snug" style="color: #374151 !important;">' +
            '<p class="font-medium">' + footerNote + '</p>' +
          '</div>' +
        '</div>';

      return html;
    },

    generateEscPosBytes: function (lines) {
      var bytes = [];
      // Init printer (ESC @)
      bytes.push(0x1B, 0x40);
      // Code page 0 (PC437)
      bytes.push(0x1B, 0x74, 0x00);

      var textContent = lines.join("\n") + "\n\n\n\n";
      var encoder = new TextEncoder();
      var encoded = encoder.encode(textContent);
      for (var i = 0; i < encoded.length; i++) {
        bytes.push(encoded[i]);
      }
      // Partial cut paper (GS V A 3)
      bytes.push(0x1D, 0x56, 0x41, 0x03);

      return new Uint8Array(bytes);
    },

    connectBluetooth: function (onSuccess, onError) {
      var isId = (document.documentElement.lang || "id") === "id";
      if (!this.isSupported()) {
        var notSupportedMsg = isId
          ? "Web Bluetooth tidak didukung pada browser ini. Silakan gunakan Google Chrome atau Microsoft Edge, atau gunakan fitur Cetak via Browser (PDF)."
          : "Web Bluetooth is not supported in this browser. Please use Chrome or Edge, or use Print via Browser (PDF).";
        if (window.showGlobalToast) window.showGlobalToast(isId ? "Bluetooth Tidak Didukung" : "Bluetooth Not Supported", notSupportedMsg, 4000, "warning");
        if (onError) onError(new Error(notSupportedMsg));
        return;
      }

      var serviceUUIDs = [
        "000018f0-0000-1000-8000-00805f9b34fb",
        "0000ffe0-0000-1000-8000-00805f9b34fb",
        "49535343-fe7d-4ae5-8fa9-9fafd205e455",
        "e7810a71-73ae-499d-8c15-faa9aef0c3f2"
      ];

      navigator.bluetooth.requestDevice({
        acceptAllDevices: true,
        optionalServices: serviceUUIDs
      })
      .then(function (device) {
        bluetoothDevice = device;
        device.addEventListener("gattserverdisconnected", function () {
          bluetoothCharacteristic = null;
          ThermalPrinter.updateStatusUI(false);
          if (window.showGlobalToast) {
            window.showGlobalToast(
              isId ? "Printer Terputus" : "Printer Disconnected",
              isId ? "Koneksi printer Bluetooth terputus." : "Bluetooth printer disconnected.",
              3000,
              "warning"
            );
          }
        });
        return device.gatt.connect();
      })
      .then(function (server) {
        return server.getPrimaryServices();
      })
      .then(function (services) {
        if (!services || services.length === 0) {
          throw new Error("Layanan printer BLE tidak ditemukan.");
        }
        var searchChar = function (idx) {
          if (idx >= services.length) {
            throw new Error("Karakteristik penulisan ESC/POS tidak ditemukan.");
          }
          return services[idx].getCharacteristics().then(function (chars) {
            for (var c = 0; c < chars.length; c++) {
              var props = chars[c].properties;
              if (props.write || props.writeWithoutResponse) {
                return chars[c];
              }
            }
            return searchChar(idx + 1);
          });
        };
        return searchChar(0);
      })
      .then(function (characteristic) {
        bluetoothCharacteristic = characteristic;
        var devName = (bluetoothDevice && bluetoothDevice.name) ? bluetoothDevice.name : "Okay 58D Thermal";
        localStorage.setItem("pos_thermal_device_name", devName);
        ThermalPrinter.updateStatusUI(true, devName);

        if (window.showGlobalToast) {
          window.showGlobalToast(
            isId ? "Printer Terhubung" : "Printer Connected",
            isId ? "Berhasil terhubung ke " + devName : "Connected to " + devName,
            3500,
            "success"
          );
        }
        if (onSuccess) onSuccess(devName);
      })
      .catch(function (err) {
        if (err.name === "NotFoundError") return; // User closed Bluetooth picker
        if (onError) onError(err);
        if (window.showGlobalToast) {
          window.showGlobalToast(
            isId ? "Gagal Menghubungkan" : "Connection Failed",
            err.message || (isId ? "Gagal pairing ke printer Bluetooth." : "Failed to connect to Bluetooth printer."),
            4000,
            "warning"
          );
        }
      });
    },

    disconnectBluetooth: function () {
      if (bluetoothDevice && bluetoothDevice.gatt && bluetoothDevice.gatt.connected) {
        bluetoothDevice.gatt.disconnect();
      }
      bluetoothDevice = null;
      bluetoothCharacteristic = null;
      localStorage.removeItem("pos_thermal_device_name");
      this.updateStatusUI(false);
    },

    printBluetooth: function (txData, onComplete, onError) {
      var isId = (document.documentElement.lang || "id") === "id";
      if (!bluetoothCharacteristic) {
        if (onError) onError(new Error("Printer belum terhubung via Bluetooth."));
        return;
      }

      var lines = this.generateReceiptLines(txData);
      var bytes = this.generateEscPosBytes(lines);

      // Send chunks of 100 bytes to prevent buffer overflow
      var chunkSize = 100;
      var offset = 0;

      function sendNextChunk() {
        if (offset >= bytes.length) {
          if (window.showGlobalToast) {
            window.showGlobalToast(
              isId ? "Struk Berhasil Dicetak" : "Receipt Printed",
              isId ? "Data struk telah dikirim ke printer thermal." : "Receipt data sent to thermal printer.",
              3000,
              "success"
            );
          }
          if (onComplete) onComplete();
          return;
        }
        var chunk = bytes.slice(offset, offset + chunkSize);
        offset += chunkSize;

        var promise = bluetoothCharacteristic.writeValueWithoutResponse
          ? bluetoothCharacteristic.writeValueWithoutResponse(chunk)
          : bluetoothCharacteristic.writeValue(chunk);

        promise
          .then(function () {
            setTimeout(sendNextChunk, 30);
          })
          .catch(function (err) {
            if (onError) onError(err);
          });
      }

      sendNextChunk();
    },

    printBrowser: function (txData) {
      var container = document.getElementById("thermal-print-container");
      if (!container) {
        container = document.createElement("div");
        container.id = "thermal-print-container";
        container.style.display = "none";
        document.body.appendChild(container);
      }

      container.innerHTML = '<div class="thermal-receipt-paper" style="font-family: \'Courier New\', Courier, monospace; font-size: 11px; line-height: 1.25; color: #000000; background: #ffffff; width: 58mm; max-width: 58mm; margin: 0 auto; padding: 2mm 0;">' +
        ThermalPrinter.renderReceiptHtml(txData) +
      '</div>';

      document.body.classList.add("printing-thermal");

      var cleanup = function () {
        document.body.classList.remove("printing-thermal");
        window.removeEventListener("afterprint", cleanup);
      };
      window.addEventListener("afterprint", cleanup);

      setTimeout(function () {
        window.print();
        setTimeout(cleanup, 2000);
      }, 50);
    },

    escapeHtml: function (str) {
      return (str || "")
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
    },

    updateStatusUI: function (isConnected, devName) {
      var pingEl = document.getElementById("printer-status-ping");
      var dotEl = document.getElementById("printer-status-dot");
      var labelEl = document.getElementById("printer-btn-label");
      var isId = (document.documentElement.lang || "id") === "id";

      if (isConnected) {
        if (pingEl) pingEl.classList.remove("hidden");
        if (dotEl) dotEl.className = "relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500";
        if (labelEl) labelEl.textContent = devName ? ("🖨️ " + devName) : (isId ? "Printer Terhubung" : "Printer Connected");
      } else {
        if (pingEl) pingEl.classList.add("hidden");
        if (dotEl) dotEl.className = "relative inline-flex rounded-full h-2.5 w-2.5 bg-slate-400";
        if (labelEl) labelEl.textContent = isId ? "Printer Thermal" : "Thermal Printer";
      }
    }
  };

  // --- Virtual Receipt Preview Modal ---
  function showReceiptPreviewModal(txData) {
    var isId = (document.documentElement.lang || "id") === "id";
    var existing = document.getElementById("pos-receipt-preview-modal");
    if (existing) existing.remove();

    var lines = ThermalPrinter.generateReceiptLines(txData);
    var isBtConnected = !!(bluetoothCharacteristic);
    var isBtSupported = ThermalPrinter.isSupported();

    var modal = document.createElement("div");
    modal.id = "pos-receipt-preview-modal";
    modal.className = "fixed inset-0 z-[999999] flex items-center justify-center p-4 sm:p-6 overflow-y-auto";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "flex";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.72)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    var receiptTextHtml = lines.map(function (l) {
      return ThermalPrinter.escapeHtml(l);
    }).join("\n");

    modal.innerHTML =
      '<div class="relative w-full max-w-sm rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-5 sm:p-6 text-slate-800 dark:text-slate-100 transition-all max-h-[90vh] flex flex-col" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800/80 shrink-0">' +
          '<div class="flex items-center gap-2.5">' +
            '<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-50 dark:bg-blue-950/70 border border-blue-100 dark:border-blue-900/60 text-blue-600 dark:text-blue-400">' +
              '<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>' +
            '</div>' +
            '<div>' +
              '<h3 class="font-black text-sm sm:text-base tracking-tight text-slate-900 dark:text-white">' + (isId ? "Pratinjau Struk (58mm)" : "Receipt Preview (58mm)") + '</h3>' +
              '<p class="text-[11px] text-slate-500 dark:text-slate-400">' + (isId ? "Standar Thermal 32 Karakter" : "32 Columns Thermal Standard") + '</p>' +
            '</div>' +
          '</div>' +
          '<button id="receipt-close-x" type="button" class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer">' +
            '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>' +
          '</button>' +
        '</div>' +

        '<!-- Receipt Paper Visual (Accurate 58mm Layout) -->' +
        '<div class="my-3 overflow-y-auto max-h-[56vh] p-3 sm:p-4 flex justify-center bg-slate-200/90 dark:bg-slate-950/80 rounded-2xl border border-slate-300 dark:border-slate-800/80 shadow-inner [scrollbar-width:thin]">' +
          '<div class="thermal-receipt-paper rounded-t-lg pt-4 px-4 pb-2 w-full max-w-[270px] mx-auto shadow-2xl font-mono" style="background-color: #ffffff !important; color: #111827 !important; border: 1px solid #cbd5e1;">' +
            ThermalPrinter.renderReceiptHtml(txData) +
            '<div class="thermal-receipt-paper-cut mt-3"></div>' +
          '</div>' +
        '</div>' +

        '<!-- Action Buttons -->' +
        '<div class="pt-2 border-t border-slate-100 dark:border-slate-800/80 space-y-2 shrink-0">' +
          '<button id="btn-print-bt" type="button" class="w-full min-h-12 inline-flex items-center justify-center text-center rounded-xl px-5 font-bold shadow-lg transition active:scale-95 bg-blue-600 hover:bg-blue-700 text-white shadow-blue-600/25 text-sm cursor-pointer">' +
            '<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>' +
            '<span>' + (isBtConnected ? (isId ? "Cetak ke Printer Thermal" : "Print to Thermal Printer") : (isId ? "Hubungkan & Cetak Thermal" : "Connect & Print Thermal")) + '</span>' +
          '</button>' +
          '<div class="flex items-center justify-between gap-3 pt-1">' +
            '<button id="btn-print-browser" type="button" class="text-xs text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 font-semibold underline cursor-pointer">' +
              '📄 ' + (isId ? "Cetak via Dialog Browser / PDF" : "Browser / PDF Print") +
            '</button>' +
            '<button id="btn-receipt-cancel" type="button" class="btn-secondary !min-h-9 px-4 text-xs font-semibold cursor-pointer">' +
              (isId ? "Tutup" : "Close") +
            '</button>' +
          '</div>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.remove();
    }

    document.getElementById("receipt-close-x").addEventListener("click", closeModal);
    document.getElementById("btn-receipt-cancel").addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });

    document.getElementById("btn-print-browser").addEventListener("click", function () {
      ThermalPrinter.printBrowser(txData);
    });

    document.getElementById("btn-print-bt").addEventListener("click", function () {
      if (bluetoothCharacteristic) {
        ThermalPrinter.printBluetooth(txData, function () {
          closeModal();
        }, function (err) {
          if (window.showGlobalToast) {
            window.showGlobalToast(
              isId ? "Gagal Cetak Bluetooth" : "Bluetooth Print Failed",
              err.message || (isId ? "Terjadi kesalahan saat mencetak." : "Error sending data to printer."),
              4000,
              "warning"
            );
          }
        });
      } else {
        // Connect first, then print
        ThermalPrinter.connectBluetooth(function () {
          ThermalPrinter.printBluetooth(txData, function () {
            closeModal();
          });
        });
      }
    });
  }

  // --- Thermal Printer Settings Modal ---
  function showPrinterSettingsModal() {
    var isId = (document.documentElement.lang || "id") === "id";
    var existing = document.getElementById("pos-printer-settings-modal");
    if (existing) existing.remove();

    var isBtConnected = !!(bluetoothCharacteristic);
    var devName = localStorage.getItem("pos_thermal_device_name") || "Okay 58D";
    var isBtSupported = ThermalPrinter.isSupported();
    var currentStoreName = ThermalPrinter.getStoreName();
    var currentFooterNote = ThermalPrinter.getFooterNote();
    var autoPrompt = localStorage.getItem("pos_thermal_auto_prompt") !== "false";

    var modal = document.createElement("div");
    modal.id = "pos-printer-settings-modal";
    modal.className = "fixed inset-0 z-[999999] flex items-center justify-center p-4 sm:p-6 overflow-y-auto";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "flex";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.72)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    modal.innerHTML =
      '<div class="relative w-full max-w-md rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-6 sm:p-7 text-slate-800 dark:text-slate-100 transition-all max-h-[90vh] overflow-y-auto" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="flex items-center justify-between pb-4 border-b border-slate-100 dark:border-slate-800/80">' +
          '<div class="flex items-center gap-3">' +
            '<div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-blue-50 dark:bg-blue-950/70 border border-blue-100 dark:border-blue-900/60 text-blue-600 dark:text-blue-400 shadow-sm">' +
              '<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>' +
            '</div>' +
            '<div>' +
              '<h3 class="font-black text-base sm:text-lg tracking-tight text-slate-900 dark:text-white">' + (isId ? "Pengaturan Printer Thermal" : "Thermal Printer Settings") + '</h3>' +
              '<p class="text-xs text-slate-500 dark:text-slate-400">Okay 58D / Standard 58mm</p>' +
            '</div>' +
          '</div>' +
          '<button id="settings-close-x" type="button" class="p-2 rounded-xl text-slate-400 hover:text-slate-600 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer">' +
            '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>' +
          '</button>' +
        '</div>' +

        '<!-- Connection Status Card -->' +
        '<div class="mt-4 p-4 rounded-2xl border ' + (isBtConnected ? 'bg-emerald-50/70 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-900/60' : 'bg-slate-50 dark:bg-slate-900/60 border-slate-200/80 dark:border-slate-800') + '">' +
          '<div class="flex items-center justify-between">' +
            '<div class="flex items-center gap-2.5">' +
              '<span class="flex h-3 w-3 relative">' +
                (isBtConnected ? '<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>' : '') +
                '<span class="relative inline-flex rounded-full h-3 w-3 ' + (isBtConnected ? 'bg-emerald-500' : 'bg-slate-400') + '"></span>' +
              '</span>' +
              '<div>' +
                '<p class="text-xs font-bold text-slate-900 dark:text-white">' +
                  (isBtConnected ? ((isId ? "Terhubung ke: " : "Connected to: ") + devName) : (isId ? "Printer Belum Terhubung" : "Printer Not Connected")) +
                '</p>' +
                '<p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">' +
                  (isBtConnected ? (isId ? "Siap mencetak langsung via Web Bluetooth." : "Ready to print via Web Bluetooth.") : (isId ? "Gunakan tombol di bawah untuk menghubungkan Bluetooth." : "Click below to connect via Bluetooth.")) +
                '</p>' +
              '</div>' +
            '</div>' +
          '</div>' +

          '<div class="mt-3.5 flex flex-wrap gap-2">' +
            (isBtConnected ?
              '<button id="btn-bt-disconnect" type="button" class="btn-secondary !min-h-10 px-4 text-xs font-semibold !text-rose-600 !border-rose-300 dark:!border-rose-900/60 hover:!bg-rose-50 dark:hover:!bg-rose-950/40 transition cursor-pointer">' +
                (isId ? "Putuskan Koneksi" : "Disconnect") +
              '</button>' :
              '<button id="btn-bt-connect" type="button" class="inline-flex min-h-10 items-center justify-center rounded-xl px-4 font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-sm text-xs transition active:scale-95 cursor-pointer">' +
                '<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>' +
                (isId ? "Hubungkan Bluetooth (Okay 58D)" : "Connect Bluetooth (Okay 58D)") +
              '</button>'
            ) +
            '<button id="btn-test-print" type="button" class="btn-secondary !min-h-10 px-4 text-xs font-semibold transition cursor-pointer">' +
              '🖨️ ' + (isId ? "Cetak Uji Coba (Test Print)" : "Test Print") +
            '</button>' +
          '</div>' +
        '</div>' +

        (!isBtSupported ?
          '<div class="mt-3 p-3 rounded-xl bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900/60 text-[11px] text-amber-800 dark:text-amber-300 leading-relaxed">' +
            'ℹ️ ' + (isId ? "Browser ini belum mendukung Web Bluetooth secara langsung. Namun Anda tetap dapat mencetak struk dengan rapi menggunakan tombol <strong>Cetak Browser / PDF</strong>." : "This browser does not support Web Bluetooth natively. You can still print seamlessly using <strong>Browser / PDF Print</strong>.") +
          '</div>' : ''
        ) +

        '<!-- Receipt Text Customization -->' +
        '<div class="mt-5 space-y-3 pt-4 border-t border-slate-100 dark:border-slate-800/80">' +
          '<h4 class="text-xs font-black uppercase tracking-wider text-slate-400">' + (isId ? "Pengaturan Format Struk" : "Receipt Format Settings") + '</h4>' +
          '<div>' +
            '<label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1" for="input-store-name">' + (isId ? "Nama Toko / Barbershop (Header):" : "Store Name (Header):") + '</label>' +
            '<input id="input-store-name" class="field h-10 min-h-10 w-full text-xs" value="' + ThermalPrinter.escapeHtml(currentStoreName) + '">' +
          '</div>' +
          '<div>' +
            '<label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1" for="input-footer-note">' + (isId ? "Catatan Kaki (Footer):" : "Footer Note:") + '</label>' +
            '<input id="input-footer-note" class="field h-10 min-h-10 w-full text-xs" value="' + ThermalPrinter.escapeHtml(currentFooterNote) + '">' +
          '</div>' +
          '<div class="pt-1">' +
            '<label class="inline-flex items-center gap-2 cursor-pointer text-xs font-semibold text-slate-700 dark:text-slate-300">' +
              '<input id="check-auto-prompt" type="checkbox" class="rounded border-slate-300 text-blue-600 focus:ring-blue-500 h-4 w-4" ' + (autoPrompt ? 'checked' : '') + '>' +
              '<span>' + (isId ? "Tampilkan tawaran cetak otomatis setelah simpan transaksi" : "Offer print automatically after saving transaction") + '</span>' +
            '</label>' +
          '</div>' +
        '</div>' +

        '<div class="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800/80 flex flex-wrap items-center justify-between gap-3">' +
          '<div id="settings-dirty-status" class="text-xs text-slate-600 dark:text-slate-400 font-medium italic">' +
            (isId ? "Tidak ada perubahan" : "No changes") +
          '</div>' +
          '<div class="flex items-center gap-2">' +
            '<button id="btn-settings-cancel" type="button" class="btn-secondary !min-h-10 px-4 text-xs sm:text-sm font-semibold cursor-pointer">' +
              (isId ? "Tutup" : "Close") +
            '</button>' +
            '<button id="btn-save-settings" type="button" disabled class="!min-h-10 px-5 text-xs sm:text-sm font-bold rounded-xl transition duration-150 border border-slate-300 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-500 opacity-60 cursor-not-allowed shadow-none">' +
              '💾 ' + (isId ? "Simpan Pengaturan" : "Save Settings") +
            '</button>' +
          '</div>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.remove();
    }

    document.getElementById("settings-close-x").addEventListener("click", closeModal);
    var cancelBtn = document.getElementById("btn-settings-cancel");
    if (cancelBtn) cancelBtn.addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });

    var connectBtn = document.getElementById("btn-bt-connect");
    if (connectBtn) {
      connectBtn.addEventListener("click", function () {
        ThermalPrinter.connectBluetooth(function () {
          closeModal();
          showPrinterSettingsModal();
        });
      });
    }

    var disconnectBtn = document.getElementById("btn-bt-disconnect");
    if (disconnectBtn) {
      disconnectBtn.addEventListener("click", function () {
        ThermalPrinter.disconnectBluetooth();
        closeModal();
        showPrinterSettingsModal();
      });
    }

    document.getElementById("btn-test-print").addEventListener("click", function () {
      var sampleTx = {
        id: "TEST",
        amount: "Rp 45.000,00",
        kind: "income",
        operator: "Yoga (Dev)",
        date: new Date().toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" }) + " 12:00",
        note: "Uji coba printer thermal 58mm",
        items: [
          { name: "Haircut Gent", amount: "Rp 35.000,00" },
          { name: "Pomade Waterbased", amount: "Rp 10.000,00" }
        ]
      };
      showReceiptPreviewModal(sampleTx);
    });

    var storeNameInput = document.getElementById("input-store-name");
    var footerNoteInput = document.getElementById("input-footer-note");
    var autoPromptCheck = document.getElementById("check-auto-prompt");
    var saveBtn = document.getElementById("btn-save-settings");
    var dirtyStatusEl = document.getElementById("settings-dirty-status");

    function checkDirty() {
      var sVal = storeNameInput ? storeNameInput.value.trim() : "";
      var fVal = footerNoteInput ? footerNoteInput.value.trim() : "";
      var aVal = autoPromptCheck ? autoPromptCheck.checked : true;

      var isDirty = (sVal !== currentStoreName.trim()) ||
                    (fVal !== currentFooterNote.trim()) ||
                    (aVal !== autoPrompt);

      if (saveBtn) {
        if (isDirty) {
          saveBtn.disabled = false;
          saveBtn.className = "btn-primary !min-h-10 px-5 text-xs sm:text-sm font-bold shadow-md shadow-blue-500/20 cursor-pointer transition active:scale-95";
          if (dirtyStatusEl) {
            dirtyStatusEl.innerHTML = '<span class="inline-flex items-center text-amber-600 dark:text-amber-400 font-semibold">● ' + (isId ? "Ada perubahan belum disimpan" : "Unsaved changes") + '</span>';
          }
        } else {
          saveBtn.disabled = true;
          saveBtn.className = "!min-h-10 px-5 text-xs sm:text-sm font-bold rounded-xl transition duration-150 border border-slate-300 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-400 dark:text-slate-500 opacity-60 cursor-not-allowed shadow-none";
          if (dirtyStatusEl) {
            dirtyStatusEl.textContent = isId ? "Tidak ada perubahan" : "No changes";
            dirtyStatusEl.className = "text-xs text-slate-600 dark:text-slate-400 font-medium italic";
          }
        }
      }
      return isDirty;
    }

    if (storeNameInput) storeNameInput.addEventListener("input", checkDirty);
    if (footerNoteInput) footerNoteInput.addEventListener("input", checkDirty);
    if (autoPromptCheck) autoPromptCheck.addEventListener("change", checkDirty);

    saveBtn.addEventListener("click", function () {
      if (!checkDirty()) return;

      var sName = storeNameInput ? storeNameInput.value.trim() : "";
      var fNote = footerNoteInput ? footerNoteInput.value.trim() : "";
      var autoP = autoPromptCheck ? autoPromptCheck.checked : true;

      if (!sName) sName = "PARDIS BARBER SHOP";
      if (!fNote) fNote = "Terima Kasih Atas Kunjungan Anda!";

      localStorage.setItem("pos_thermal_store_name", sName);
      localStorage.setItem("pos_thermal_footer_note", fNote);
      localStorage.setItem("pos_thermal_auto_prompt", autoP ? "true" : "false");

      if (window.showGlobalToast) {
        window.showGlobalToast(
          isId ? "Pengaturan Disimpan" : "Settings Saved",
          isId ? "Pengaturan printer thermal berhasil diperbarui." : "Thermal printer settings updated.",
          2500,
          "success"
        );
      }
      closeModal();
    });
  }

  // --- Post-Save Prompt Modal ---
  function showPostSavePromptModal(txData) {
    var isId = (document.documentElement.lang || "id") === "id";
    var existing = document.getElementById("pos-post-save-prompt-modal");
    if (existing) existing.remove();

    var modal = document.createElement("div");
    modal.id = "pos-post-save-prompt-modal";
    modal.className = "fixed inset-0 z-[999999] flex items-center justify-center p-4 sm:p-6";
    modal.style.position = "fixed";
    modal.style.top = "0";
    modal.style.left = "0";
    modal.style.width = "100vw";
    modal.style.height = "100vh";
    modal.style.zIndex = "999999";
    modal.style.display = "flex";
    modal.style.alignItems = "center";
    modal.style.justifyContent = "center";
    modal.style.background = "rgba(10, 16, 35, 0.68)";
    modal.style.backdropFilter = "blur(8px)";
    modal.style.webkitBackdropFilter = "blur(8px)";
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");

    modal.innerHTML =
      '<div class="relative w-full max-w-md rounded-3xl bg-white dark:bg-[#070d24] border border-slate-200/90 dark:border-slate-800 shadow-2xl p-6 sm:p-7 text-slate-800 dark:text-slate-100 transition-all text-center" style="animation: modalPopIn .25s cubic-bezier(.16,1,.3,1) forwards;">' +
        '<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-emerald-50 dark:bg-emerald-950/70 border border-emerald-100 dark:border-emerald-900/60 text-emerald-600 dark:text-emerald-400 shadow-sm mb-4">' +
          '<svg class="h-7 w-7" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>' +
        '</div>' +
        '<h3 class="font-black text-lg sm:text-xl tracking-tight text-slate-900 dark:text-white">' + (isId ? "Transaksi Berhasil Disimpan!" : "Transaction Saved Successfully!") + '</h3>' +
        '<p class="text-sm font-bold text-emerald-600 dark:text-emerald-400 mt-1">' + (txData.amount || "") + '</p>' +
        '<p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-2 leading-relaxed">' +
          (isId ? "Apakah Anda ingin mencetak struk transaksi ini ke printer thermal?" : "Would you like to print the receipt to the thermal printer now?") +
        '</p>' +
        '<div class="mt-6 flex items-center justify-center gap-3">' +
          '<button id="post-save-skip" class="btn-secondary !min-h-11 px-5 text-xs sm:text-sm font-semibold transition cursor-pointer" type="button">' +
            (isId ? "Selesai" : "Done") +
          '</button>' +
          '<button id="post-save-print" class="inline-flex min-h-11 items-center justify-center text-center rounded-xl px-6 font-bold shadow-lg transition active:scale-95 bg-blue-600 hover:bg-blue-700 text-white shadow-blue-600/25 text-xs sm:text-sm cursor-pointer" type="button">' +
            '<svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/></svg>' +
            (isId ? "🖨️ Cetak Struk Thermal" : "🖨️ Print Thermal Receipt") +
          '</button>' +
        '</div>' +
      '</div>';

    document.body.appendChild(modal);

    function closeModal() {
      modal.remove();
    }

    document.getElementById("post-save-skip").addEventListener("click", closeModal);
    modal.addEventListener("click", function (e) {
      if (e.target === modal) closeModal();
    });

    document.getElementById("post-save-print").addEventListener("click", function () {
      closeModal();
      showReceiptPreviewModal(txData);
    });
  }

  // Check for auto-open or post-save prompt on page reload
  try {
    var shouldAutoOpen = sessionStorage.getItem("autoOpenThermalPrint");
    var pendingPrompt = sessionStorage.getItem("pendingTxPrintPrompt");
    if (shouldAutoOpen === "true" && pendingPrompt) {
      sessionStorage.removeItem("autoOpenThermalPrint");
      sessionStorage.removeItem("pendingTxPrintPrompt");
      var autoTxData = JSON.parse(pendingPrompt);
      setTimeout(function () {
        showReceiptPreviewModal(autoTxData);
      }, 350);
    } else if (pendingPrompt) {
      sessionStorage.removeItem("pendingTxPrintPrompt");
      var promptData = JSON.parse(pendingPrompt);
      setTimeout(function () {
        showPostSavePromptModal(promptData);
      }, 350);
    }
  } catch (e) {}

  // Check saved Bluetooth device name in localStorage to update topbar UI on load
  var savedDeviceName = localStorage.getItem("pos_thermal_device_name");
  if (savedDeviceName) {
    ThermalPrinter.updateStatusUI(false); // Default disconnected until user connects, or shows label
  }

  // Attach Topbar Printer Settings Button
  var btnSettings = document.getElementById("btn-thermal-printer-settings");
  if (btnSettings) {
    btnSettings.addEventListener("click", function (e) {
      e.preventDefault();
      showPrinterSettingsModal();
    });
  }

  // Delegate click for Reprint Receipt buttons on transaction history table
  document.addEventListener("click", function (e) {
    var btn = e.target.closest(".btn-print-receipt");
    if (!btn) return;
    e.preventDefault();

    var items = [];
    try {
      var rawItems = btn.getAttribute("data-tx-items");
      if (rawItems) items = JSON.parse(rawItems);
    } catch (err) {
      items = [];
    }

    var txData = {
      id: btn.getAttribute("data-tx-id") || "",
      kind: btn.getAttribute("data-tx-kind") || "income",
      amount: btn.getAttribute("data-tx-amount") || "",
      subtotal: btn.getAttribute("data-tx-subtotal") || "",
      discount: btn.getAttribute("data-tx-discount") || "",
      operator: btn.getAttribute("data-tx-operator") || "",
      date: btn.getAttribute("data-tx-date") || "",
      note: btn.getAttribute("data-tx-note") || "",
      items: items
    };

    showReceiptPreviewModal(txData);
  });
})();

// Admin Branch Selection Persistence (stays on selected branch across transactions, resets on logout)
(function initAdminBranchPersistence() {
  function clearBranchStorage() {
    try {
      sessionStorage.removeItem("pos_admin_selected_branch");
      localStorage.removeItem("pos_admin_selected_branch");
    } catch (e) {}
  }

  function setup() {
    // If on login page, clear any stored branch so fresh login starts fresh
    if (window.location.pathname.indexOf("/login") !== -1) {
      clearBranchStorage();
      return;
    }

    var branchSelect = document.getElementById("admin-branch-select") || document.querySelector('#transaction-form select[name="branch_id"]');
    if (!branchSelect) return;

    // Restore previously selected branch
    try {
      var saved = sessionStorage.getItem("pos_admin_selected_branch") || localStorage.getItem("pos_admin_selected_branch");
      if (saved) {
        var opt = branchSelect.querySelector('option[value="' + saved + '"]');
        if (opt) {
          branchSelect.value = saved;
        }
      }
    } catch (e) {}

    // Save on manual user selection
    branchSelect.addEventListener("change", function () {
      try {
        sessionStorage.setItem("pos_admin_selected_branch", this.value);
        localStorage.setItem("pos_admin_selected_branch", this.value);
      } catch (e) {}
    });

    // Also persist on transaction form submission so it stays selected after reload
    var form = document.getElementById("transaction-form");
    if (form) {
      form.addEventListener("submit", function () {
        try {
          if (branchSelect.value) {
            sessionStorage.setItem("pos_admin_selected_branch", branchSelect.value);
            localStorage.setItem("pos_admin_selected_branch", branchSelect.value);
          }
        } catch (e) {}
      });
    }

    // Clear when clicking any logout button or submitting any logout form
    document.addEventListener("submit", function (e) {
      var f = e.target;
      if (f && (f.getAttribute("action") || "").indexOf("/logout") !== -1) {
        clearBranchStorage();
      }
    }, true);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", setup);
  } else {
    setup();
  }
})();

