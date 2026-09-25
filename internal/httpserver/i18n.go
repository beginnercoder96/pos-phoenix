package httpserver

import (
	"fmt"
	"strconv"
	"time"
)

func templateFuncs(location *time.Location) map[string]any {
	return map[string]any{
		"money":        func(cents int64) string { return formatRupiah(cents) },
		"moneyCompact": func(cents int64) string { return formatRupiahCompact(cents) },
		"moneyNumber":  func(cents int64) string { return formatRupiahNumber(cents) },
		"date":         func(t time.Time) string { return t.In(location).Format("02 Jan 2006 15:04") },
		"pct":          func(ratio float64) string { return fmt.Sprintf("%.0f%%", ratio*100) },
		"sub":          func(a, b int64) int64 { return a - b },
		"add": func(a, b any) int64 {
			var aInt, bInt int64
			switch v := a.(type) {
			case int:
				aInt = int64(v)
			case int64:
				aInt = v
			}
			switch v := b.(type) {
			case int:
				bInt = int64(v)
			case int64:
				bInt = v
			}
			return aInt + bInt
		},
		"barClass": func(height int) string {
			if height <= 0 {
				return "bar-height-0"
			}
			bucket := (height + 9) / 10
			if bucket > 10 {
				bucket = 10
			}
			return "bar-height-" + strconv.Itoa(bucket)
		},
	}
}

func formatRupiah(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole, fraction := cents/100, cents%100
	digits := strconv.FormatInt(whole, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "." + digits[i:]
	}
	result := "Rp " + digits + "," + fmt.Sprintf("%02d", fraction)
	if negative {
		return "-" + result
	}
	return result
}

func formatRupiahCompact(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole, fraction := cents/100, cents%100
	digits := strconv.FormatInt(whole, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "." + digits[i:]
	}
	result := "Rp " + digits
	if fraction != 0 {
		result += "," + fmt.Sprintf("%02d", fraction)
	}
	if negative {
		return "-" + result
	}
	return result
}

func formatRupiahNumber(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole, fraction := cents/100, cents%100
	digits := strconv.FormatInt(whole, 10)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "." + digits[i:]
	}
	result := digits
	if fraction != 0 {
		result += "," + fmt.Sprintf("%02d", fraction)
	}
	if negative {
		return "-" + result
	}
	return result
}

func translate(language, key string) string {
	en := map[string]string{
		"language": "Language", "theme": "Theme", "english": "English", "indonesian": "Indonesian", "light": "Light", "dark": "Dark", "save": "Save", "currency": "Indonesian Rupiah", "appName": "POS Phoenix", "welcome": "Welcome back", "secureCashflow": "A place where the nice haircut trim come from us!", "email": "Email", "password": "Password", "signIn": "Sign in", "cashflowReport": "Cash-flow report", "today": "Today", "thisMonth": "This month", "from": "From", "to": "To", "apply": "Apply", "downloadCSV": "Download CSV", "income": "Income", "expense": "Expense", "balance": "Balance", "newTransaction": "New transaction", "type": "Type", "amount": "Amount", "category": "Category", "note": "Note", "saveTransaction": "Save transaction", "transactions": "Transactions", "records": "records", "page": "page", "previous": "Previous", "next": "Next", "noTransactions": "No transactions in this period.", "reverseReason": "Reversal reason", "reverse": "Reverse", "reversedBy": "Cancelled by reversal transaction", "operatorAdministration": "Operator administration", "dashboard": "Dashboard", "createOperator": "Create operator", "displayName": "Display name", "temporaryPassword": "Temporary password", "passwordHint": "12–128 characters. Share it securely.", "operators": "Operators", "active": "Active", "inactive": "Inactive", "deactivate": "Deactivate", "activate": "Activate", "noOperators": "No operators created yet.", "signOut": "Sign out", "reversal": "Reversal", "customerRefund": "Customer refund",
		"username": "Username", "usernamePlaceholder": "Enter username", "usernameExample": "e.g. yogi", "usernameHint": "3–50 characters, letters, numbers, dot, dash, underscore.",
		"usernameOrEmail": "Username or Email", "usernameOrEmailPlaceholder": "Enter username or email",
		"forgotPassword": "Forgot password?", "resetPassword": "Reset Password", "sendResetLink": "Send Reset Link",
		"newPassword": "New Password", "confirmNewPassword": "Confirm New Password",
		"passwordResetSuccess": "Password has been successfully updated. You can now sign in.",
		"backToSignIn": "Back to Sign In",
		"enterRegisteredEmail": "Enter your registered email address to receive password reset instructions.",
		"resetLinkMockNotice": "Mock Mode: Reset link generated below (simulated email delivery):",
		"invalidOrExpiredToken": "The password reset link is invalid or has expired.",
		"passwordsDoNotMatch": "New password and confirmation do not match.",
		"resetEmailSentMessage": "If that email is registered in our system, a password reset link has been generated.",
		"clickToResetPassword": "Click here to reset your password",
		"addCategory": "Add Category / Item", "item": "Item / Service", "total": "Total", "remove": "Remove", "selectCategory": "Select category", "selectItem": "Select service / item", "customItem": "Custom item", "items": "Items", "downloadExcel": "Download Excel (.xlsx)",
		"customDate":    "Custom Date",
		"cashflowTrend": "Cash-flow trend", "trendSubtitle": "Income and expense by period",
		"chartPeriod": "Chart period", "todayHourly": "Today (Hourly)", "last7Days": "Last 7 Days", "monthly": "Monthly",
		"calendarPicker": "Calendar Picker", "calendarOpen": "Calendar (Open)",
		"entireMonth": "Entire Month", "monthToDate": "Month to Date", "closeCalendar": "Close calendar",
		"prevMonth": "Previous month", "nextMonth": "Next month",
		"maxSelectable": "Maximum selectable", "daysBasedOnMonth": "days based on month shown",
		"applyCalendarDates": "Apply Calendar Dates", "selected": "Selected",
		"hoverInspectPrompt": "Hover over chart to inspect period", "net": "Net",
		"notePlaceholder": "Optional note / customer name", "pricelistTitle": "Barbershop Pricelist",
		"otherCustom": "Other / Custom", "customItemPlaceholder": "Custom Item...",
		"of": "of",
		"transactionsMenu": "Transactions",
		"openTransactionsCTA": "Open Cashier / Transactions",
		"viewReportsCTA": "Full Financial Reports",
		"shiftStatus": "Shift & Active Cashier",
		"activeCashier": "Active Cashier",
		"activeBranch": "Active Branch",
		"reserveBalance": "Reserve Balance",
		// Backoffice keys
		"boBackoffice": "Backoffice", "boMainPOS": "← Main POS", "boSignOut": "Sign Out",
		"boDashboard": "Dashboard", "boProfitSharing": "Profit Sharing", "boDiscounts": "Discounts & Bundling",
		"boProducts": "Catalog & Commission", "boPayroll": "Payroll",
		"boDownloadReport": "Download Financial Report (Excel)",
		"boDownloadReportModalTitle": "Export Financial Report (Excel)",
		"boDownloadReportModalDesc": "Select custom month range or choose quick preset.",
		"boFromMonth": "From Month", "boToMonth": "To Month",
		"boPreset24M": "24 Months (Default)", "boPreset12M": "Last 12 Months", "boPresetYTD": "Year to Date (YTD)",
		"boDownloadXLSX": "Download .xlsx", "boCancel": "Cancel",
		"boBranchFilter":   "Filter Branch:", "boAllBranches": "All Branches (Consolidated)",
		"boChartTitle": "24-Month Revenue Trend", "boChartSubtitle": "Service Revenue vs Product Retail per month",
		"boServiceRevenue": "Service Revenue", "boProductRevenue": "Product Revenue",
		"boGrossRevenue": "Gross Revenue", "boMonth": "Month", "boCadangan": "Reserve Balance",
		"boTotalGross": "Total Gross Revenue (24 Mo)", "boTotalService": "Total Service Revenue",
		"boTotalProduct": "Total Product Revenue", "boTotalReserve": "Total Reserve Balance",
		"boTotalMonths":    "Total Months",
		"boDetailTitle":    "Monthly Revenue Detail (24 Months)",
		"boDiscount":       "Discount",
		"boChartHoverHint": "Hover over a bar to see monthly details.",
		"boNoData":         "No revenue data to display yet.",
		"boBranch":         "Branch", "boPeriod": "Period (Month)", "boShow": "Show",
		"boConfigTitle": "Profit Sharing Configuration", "boPercentageTitle": "Percentage Settings",
		"boOwnerShare": "Owner Share (Ipang)", "boOwnerShareDesc": "Owner percentage of service revenue",
		"boEmpShareDesc": "Share amount:",
		"boUnallocated":  "Unallocated Reserve Balance", "boUnallocatedDesc": "Automatically calculated from remaining percentage",
		"boTotalAlloc":  "Total Allocation",
		"boSaveConfig":  "Save Profit Sharing Configuration",
		"boNoEmployees": "No employees assigned to this branch. Add employees via the Operators page.",
		"boNetRevenue":  "Branch Net Service Revenue", "boReserveBalance": "Unallocated Reserve",
		"boReserveRemainder": "remaining reserve",
		"boPayrollTitle":     "Employee Payroll", "boPayrollRecap": "Payroll Summary",
		"boEmpName": "Employee Name", "boProfitSharePct": "Share (%)",
		"boProfitShareAmt": "Service Share", "boProductComm": "Product Commission",
		"boTakeHome": "Net Take-Home Pay", "boAction": "Action",
		"boDownloadSlip": "Download Slip", "boDownloadAll": "Download All Payroll Slips (PDF)",
		"boNoPayroll":      "No employees assigned to this branch, or no profit sharing config for this period.",
		"boDiscountsTitle": "Discount & Bundling Management", "boAddDiscount": "Add New Discount / Bundling",
		"boProductsTitle": "Product Catalog & Commission Settings", "boAddProduct": "Add / Edit Catalog Item",
		// Discounts form
		"boDiscCode": "Code", "boDiscName": "Name", "boDiscNamePlaceholder": "e.g. New Year Discount", "boDiscType": "Type", "boDiscValue": "Value", "boDiscStatus": "Status",
		"boDiscTypePct": "Percentage (%)", "boDiscTypeFixed": "Fixed Amount (Rp)", "boDiscTypeBundle": "Bundle Package",
		"boDiscActive": "Active", "boDiscInactive": "Inactive",
		"boDiscServiceRatio": "Service Allocation Ratio", "boDiscProductRatio": "Product Allocation Ratio",
		"boDiscServiceShort": "Service", "boDiscProductShort": "Product", "boDiscBundleAllocation": "Bundle Allocation",
		"boDiscSave": "Save Discount", "boDiscList": "Discount & Bundling List",
		"boDiscTypePctLabel": "Percentage", "boDiscTypeFixedLabel": "Fixed", "boDiscTypeBundleLabel": "Bundle",
		"boDiscActionDelete": "Delete", "boDiscConfirmDelete": "Delete this discount?",
		"boDiscToggleStatus": "Click to toggle active status",
		"boDiscEmpty": "No discounts or bundles yet.",
		// Products form
		"boProdName": "Item Name", "boProdCategory": "Category", "boProdType": "Item Type",
		"boProdTypeService": "Service (SERVICE)", "boProdTypeProduct": "Product (PRODUCT)",
		"boProdPrice": "Price (Rupiah)", "boProdPriceHint": "e.g. 40000 for Rp 40,000",
		"boProdComm": "Commission per Item (Rupiah)", "boProdCommHint": "e.g. 5000 for Rp 5,000 (Products only)",
		"boProdSave": "Save Item", "boProdUpdate": "Update Item", "boProdCancelEdit": "Cancel Edit", "boProdList": "Catalog List",
		"boProdColName": "Name", "boProdColCategory": "Category", "boProdColType": "Type",
		"boProdColPrice": "Price (Rp)", "boProdColComm": "Commission / Item (Rp)",
		"boProdColCommLine1": "Commission", "boProdColCommLine2": "/ Item (Rp)", "boProdColStatus": "Status",
		"boProdLabelProduct": "PRODUCT", "boProdLabelService": "SERVICE",
		"boProdActionEdit": "Edit", "boProdActionDelete": "Delete", "boProdConfirmDelete": "Delete this catalog item?",
		"boProdCategorySelect": "-- Select Category --", "boProdCategoryNew": "+ New Category...", "boProdCategoryNewPlaceholder": "Enter new category name",
		"boProdToggleStatus": "Click to toggle active status", "boProdEmpty": "No catalog items yet.",
		"boProdSavedToast": "Item saved successfully", "boProdDeletedToast": "Item deleted successfully", "boProdStatusUpdated": "Item status updated",
		// Thermal Printer keys (EN)
		"printReceipt": "Print Receipt", "thermalPrinter": "Thermal Printer", "printerConnected": "Printer Connected",
		"printerDisconnected": "Printer Disconnected", "connectPrinter": "Connect Bluetooth Printer",
		"disconnectPrinter": "Disconnect Printer", "testPrint": "Test Print", "printViaBluetooth": "Print via Bluetooth",
		"printViaBrowser": "Print via Browser (PDF)", "receiptPreview": "Receipt Preview",
		"printReceiptQuestion": "Would you like to print the receipt for this transaction?",
		"printReceiptSuccess": "Receipt printed successfully.", "skip": "Skip",
		"printerModalTitle": "Thermal Printer Settings (Okay 58D)",
		"printerModalHint": "Connect via Web Bluetooth to standard 58mm thermal printers (Okay 58D) or use browser print dialog as fallback.",
		"storeName": "Pardis Barber Shop", "receiptFooter": "Thank you for your visit!",
		// Operator Credential keys (EN)
		"phoneNumber": "Phone / WhatsApp Number", "phonePlaceholder": "e.g. 08123456789",
		"branch": "Branch", "selectBranchOptional": "-- Select Branch (Optional) --", "selectBranch": "-- Select Branch --",
		"staffType": "Staff Type", "barberman": "Barberman", "cashier": "Cashier", "manager": "Manager",
		"bankName": "Bank Name / Code", "bankNamePlaceholder": "e.g. BCA, Mandiri, BRI, BNI",
		"bankAccount": "Bank Account Number", "bankAccountPlaceholder": "e.g. 001 1234567",
		"editEmployeeCredentials": "⚙️ Edit Employee Data & Bank Credentials",
		"fullName": "Full Name", "accountNumber": "Account Number", "accountNumberPlaceholder": "Account number",
		"cancel": "Cancel", "saveCredentialChanges": "Save Credential Changes",
		"confirmChangesTitle": "Confirm Changes", "confirmChangesDesc": "Please ensure employee data is correct before saving.",
		"credentialChangeDetails": "Credential Change Details", "credentialChangeNotice": "Changes will be saved and apply immediately to operations and profit sharing.",
		"yesSaveChanges": "Yes, Save Changes", "saving": "Saving...", "accountShort": "Acc",
		"dataSavedSuccess": "Data Saved Successfully", "dataSavedSuccessMsg": "Employee data changes have been successfully saved.",
		"configSavedSuccess": "Configuration Saved Successfully", "configSavedSuccessMsg": "Profit sharing configuration has been saved successfully.",
		"discSavedSuccess": "Discount Saved Successfully", "discSavedSuccessMsg": "Discount has been added to the catalog.",
		"discDeletedSuccess": "Discount Deleted Successfully", "discDeletedSuccessMsg": "Discount has been successfully removed.",
		"confirmDeleteDisc": "Delete Discount", "confirmDeleteDiscDesc": "Are you sure you want to delete this discount or bundling package?",
		"yesDeleteDiscount": "Yes, Delete Discount", "deleting": "Deleting...",
		"noDiscount": "No Discount (Normal Price)", "subtotal": "Subtotal", "discount": "Discount",
		"shiftTransactions": "Shift Transactions", "cashierIncome": "Shift Income", "cashierExpense": "Shift Expense", "cashDrawerBalance": "Drawer Cash Balance",
		"cashierSopTitle": "Cashier SOP & Guidelines",
		"cashierSop1Title": "Verify QRIS & Transfers", "cashierSop1Desc": "Ensure successful payment status is verified before releasing customer receipt.",
		"cashierSop2Title": "Real-time Recording", "cashierSop2Desc": "Record every haircut service, treatment, and grooming product immediately upon payment.",
		"cashierSop3Title": "Shift Handover & Reconciliation", "cashierSop3Desc": "Physically count drawer cash to ensure exact match with recorded balance before end of shift.",
		"printerStatusTitle": "Thermal Receipt Hardware", "printerStatusDesc": "Connect Okay 58D thermal printer via Bluetooth or use standard system print dialog.",
		// 2FA / TOTP keys (EN)
		"loginSetup2FATitle": "Set Up Two-Factor Authentication (2FA)",
		"loginSetup2FASubtitle": "Protect your POS account with Google Authenticator or any RFC 6238 TOTP app.",
		"scanQRCodeStep": "1. Scan QR Code",
		"scanQRCodeDesc": "Open Google Authenticator on your phone, tap +, and scan the QR code below:",
		"orEnterManualKey": "Or enter this secret key manually if you cannot scan the QR code:",
		"copySecretKey": "Copy Secret",
		"secretCopied": "Copied to clipboard!",
		"verifyCodeStep": "2. Enter 6-Digit Code",
		"verifyCodeDesc": "Enter the 6-digit verification code generated by your authenticator app:",
		"otpPlaceholder": "6-digit code (e.g. 123456)",
		"verifyAndActivate": "Verify & Activate 2FA",
		"verifyOTPTitle": "Two-Factor Authentication (2FA)",
		"verifyOTPSubtitle": "Enter the 6-digit code from Google Authenticator to continue.",
		"verifyOTPButton": "Verify & Sign In",
		"invalidOTPCode": "Invalid or expired verification code. Please try again.",
		"rateLimitedLocked": "Too many failed attempts. Account temporarily locked for %d seconds.",
		"totpBadge": "TOTP RFC 6238 Secured",
		"devBypassNotice": "Development Mode: Enter master code 123456 or 000000 to bypass phone scanning.",
		"devBypassButton": "⚡ Instant Dev Bypass",
	}
	if language == "id" {
		id := map[string]string{
			"language": "Bahasa", "theme": "Tema", "english": "Inggris", "indonesian": "Indonesia", "light": "Terang", "dark": "Gelap", "save": "Simpan", "currency": "Rupiah Indonesia", "appName": "POS Phoenix", "welcome": "Selamat datang kembali", "secureCashflow": "Tempat untuk mendapatkan potongan rambut terbaik Anda!", "email": "Email", "password": "Kata sandi", "signIn": "Masuk", "cashflowReport": "Laporan arus kas", "today": "Hari ini", "thisMonth": "Bulan ini", "from": "Dari", "to": "Sampai", "apply": "Terapkan", "downloadCSV": "Unduh CSV", "income": "Pemasukan", "expense": "Pengeluaran", "balance": "Saldo", "newTransaction": "Transaksi baru", "type": "Jenis", "amount": "Jumlah", "category": "Kategori", "note": "Catatan", "saveTransaction": "Simpan transaksi", "transactions": "Transaksi", "records": "catatan", "page": "halaman", "previous": "Sebelumnya", "next": "Berikutnya", "noTransactions": "Tidak ada transaksi pada periode ini.", "reverseReason": "Alasan pembatalan", "reverse": "Batalkan", "reversedBy": "Dibatalkan oleh transaksi pembalik", "operatorAdministration": "Administrasi operator", "dashboard": "Dasbor", "createOperator": "Buat operator", "displayName": "Nama tampilan", "temporaryPassword": "Kata sandi sementara", "passwordHint": "12–128 karakter. Bagikan dengan aman.", "operators": "Operator", "active": "Aktif", "inactive": "Tidak aktif", "deactivate": "Nonaktifkan", "activate": "Aktifkan", "noOperators": "Belum ada operator.", "signOut": "Keluar", "reversal": "Pembatalan", "customerRefund": "Pengembalian dana pelanggan",
			"username": "Username", "usernamePlaceholder": "Masukkan username", "usernameExample": "Contoh: yogi", "usernameHint": "3–50 karakter, huruf, angka, titik, strip, garis bawah.",
			"usernameOrEmail": "Username atau Email", "usernameOrEmailPlaceholder": "Masukkan username atau email",
			"forgotPassword": "Lupa kata sandi?", "resetPassword": "Atur Ulang Kata Sandi", "sendResetLink": "Kirim Link Atur Ulang",
			"newPassword": "Kata Sandi Baru", "confirmNewPassword": "Konfirmasi Kata Sandi Baru",
			"passwordResetSuccess": "Kata sandi berhasil diperbarui. Anda sekarang dapat masuk kembali.",
			"backToSignIn": "Kembali ke Halaman Masuk",
			"enterRegisteredEmail": "Masukkan email akun Anda untuk menerima tautan atur ulang kata sandi.",
			"resetLinkMockNotice": "Mode Mock / Dev: Tautan atur ulang berhasil dibuat (simulasi pengiriman email):",
			"invalidOrExpiredToken": "Tautan atur ulang kata sandi tidak valid atau telah kedaluwarsa.",
			"passwordsDoNotMatch": "Kata sandi baru dan konfirmasi kata sandi tidak cocok.",
			"resetEmailSentMessage": "Jika email terdaftar di sistem, tautan atur ulang kata sandi telah berhasil dibuat.",
			"clickToResetPassword": "Klik tautan ini untuk mengatur ulang kata sandi Anda",
			"addCategory": "Tambah Kategori / Item", "item": "Item / Layanan", "total": "Total", "remove": "Hapus", "selectCategory": "Pilih kategori", "selectItem": "Pilih layanan / item", "customItem": "Item lainnya", "items": "Item", "downloadExcel": "Unduh Excel (.xlsx)",
			"customDate":    "Tanggal Khusus",
			"cashflowTrend": "Tren arus kas", "trendSubtitle": "Pemasukan dan pengeluaran per periode",
			"chartPeriod": "Periode grafik", "todayHourly": "Hari ini (Per jam)", "last7Days": "7 Hari Terakhir", "monthly": "Bulanan",
			"calendarPicker": "Pilih Kalender", "calendarOpen": "Kalender (Buka)",
			"entireMonth": "Sebulan Penuh", "monthToDate": "Awal Bulan Hingga Kini", "closeCalendar": "Tutup kalender",
			"prevMonth": "Bulan sebelumnya", "nextMonth": "Bulan berikutnya",
			"maxSelectable": "Maksimal dapat dipilih", "daysBasedOnMonth": "hari berdasarkan bulan yang ditampilkan",
			"applyCalendarDates": "Terapkan Tanggal Kalender", "selected": "Dipilih",
			"hoverInspectPrompt": "Arahkan kursor ke grafik untuk melihat periode", "net": "Bersih",
			"notePlaceholder": "Catatan opsional / nama pelanggan", "pricelistTitle": "Daftar Harga Barbershop",
			"otherCustom": "Lainnya / Kustom", "customItemPlaceholder": "Item Lainnya...",
			"of": "dari",
			"transactionsMenu": "Transaksi / Kasir",
			"openTransactionsCTA": "Buka Menu Transaksi Kasir",
			"viewReportsCTA": "Laporan Lengkap",
			"shiftStatus": "Status Shift & Kasir",
			"activeCashier": "Kasir Bertugas",
			"activeBranch": "Cabang Aktif",
			"reserveBalance": "Kas Cadangan",
			// Backoffice keys (ID)
			"boBackoffice": "Backoffice", "boMainPOS": "← Kasir Utama", "boSignOut": "Keluar",
			"boDashboard": "Dasbor", "boProfitSharing": "Bagi Hasil", "boDiscounts": "Diskon & Bundling",
			"boProducts": "Katalog & Komisi", "boPayroll": "Slip Gaji",
			"boDownloadReport": "Unduh Laporan Keuangan (Excel)",
			"boDownloadReportModalTitle": "Ekspor Laporan Finansial (Excel)",
			"boDownloadReportModalDesc": "Pilih rentang bulan laporan keuangan atau pilih preset cepat.",
			"boFromMonth": "Dari Bulan", "boToMonth": "Sampai Bulan",
			"boPreset24M": "24 Bulan (Default)", "boPreset12M": "12 Bulan Terakhir", "boPresetYTD": "Tahun Berjalan (YTD)",
			"boDownloadXLSX": "Unduh .xlsx", "boCancel": "Batal",
			"boBranchFilter":   "Filter Cabang:", "boAllBranches": "Semua Cabang (Konsolidasi)",
			"boChartTitle": "Tren Pendapatan 24 Bulan", "boChartSubtitle": "Perbandingan Omzet Jasa vs Retail Produk tiap bulan",
			"boServiceRevenue": "Omzet Jasa", "boProductRevenue": "Omzet Produk",
			"boGrossRevenue": "Omzet Kotor", "boMonth": "Bulan", "boCadangan": "Kas Cadangan",
			"boTotalGross": "Total Omzet Kotor (24 Bln)", "boTotalService": "Total Omzet Jasa",
			"boTotalProduct": "Total Omzet Produk", "boTotalReserve": "Total Sisa Kas Cadangan",
			"boTotalMonths":    "Jumlah Periode Bulan",
			"boDetailTitle":    "Detail Pendapatan Bulanan (24 Bulan)",
			"boDiscount":       "Diskon",
			"boChartHoverHint": "Arahkan kursor ke salah satu bar bulan untuk melihat detail.",
			"boNoData":         "Belum ada data pendapatan untuk ditampilkan.",
			"boBranch":         "Cabang", "boPeriod": "Periode Bulan", "boShow": "Tampilkan",
			"boConfigTitle": "Konfigurasi Bagi Hasil Cabang", "boPercentageTitle": "Pengaturan Persentase",
			"boOwnerShare": "Porsi Owner (Ipang)", "boOwnerShareDesc": "Persentase bagian owner dari omzet jasa",
			"boEmpShareDesc": "Bagi hasil:",
			"boUnallocated":  "Sisa Saldo Tidak Terpakai (Cadangan)", "boUnallocatedDesc": "Otomatis dihitung dari sisa persentase",
			"boTotalAlloc":  "Total Alokasi",
			"boSaveConfig":  "Simpan Konfigurasi Bagi Hasil",
			"boNoEmployees": "Belum ada karyawan yang ditugaskan ke cabang ini. Tambahkan karyawan melalui halaman Operators.",
			"boNetRevenue":  "Net Omzet Jasa Cabang", "boReserveBalance": "Saldo Cadangan (Tidak Terpakai)",
			"boReserveRemainder": "sisa saldo",
			"boPayrollTitle":     "Slip Gaji Karyawan", "boPayrollRecap": "Rekap Gaji",
			"boEmpName": "Nama Karyawan", "boProfitSharePct": "Bagi Hasil (%)",
			"boProfitShareAmt": "Bagi Hasil Jasa", "boProductComm": "Komisi Produk",
			"boTakeHome": "Total Gaji Bersih", "boAction": "Aksi",
			"boDownloadSlip": "Download Slip", "boDownloadAll": "Download Semua Slip Gaji (PDF)",
			"boNoPayroll":      "Belum ada karyawan yang ditugaskan ke cabang ini, atau belum ada konfigurasi bagi hasil untuk periode ini.",
			"boDiscountsTitle": "Manajemen Diskon & Bundling", "boAddDiscount": "Tambah Diskon / Bundling Baru",
			"boProductsTitle": "Katalog Produk & Pengaturan Komisi", "boAddProduct": "Tambah / Edit Item Katalog",
			// Discounts form (ID)
			"boDiscCode": "Kode", "boDiscName": "Nama", "boDiscNamePlaceholder": "Contoh: Diskon Tahun Baru", "boDiscType": "Tipe", "boDiscValue": "Nilai", "boDiscStatus": "Status",
			"boDiscTypePct": "Persentase (%)", "boDiscTypeFixed": "Nominal Tetap (Rp)", "boDiscTypeBundle": "Paket Bundling",
			"boDiscActive": "Aktif", "boDiscInactive": "Tidak Aktif",
			"boDiscServiceRatio": "Rasio Alokasi Jasa", "boDiscProductRatio": "Rasio Alokasi Produk",
			"boDiscServiceShort": "Jasa", "boDiscProductShort": "Produk", "boDiscBundleAllocation": "Alokasi Bundling",
			"boDiscSave": "Simpan Diskon", "boDiscList": "Daftar Diskon & Bundling",
			"boDiscTypePctLabel": "Persentase", "boDiscTypeFixedLabel": "Nominal", "boDiscTypeBundleLabel": "Bundling",
			"boDiscActionDelete": "Hapus", "boDiscConfirmDelete": "Hapus diskon ini?",
			"boDiscToggleStatus": "Klik untuk ubah status aktif/tidak aktif",
			"boDiscEmpty": "Belum ada diskon atau bundling.",
			// Products form (ID)
			"boProdName": "Nama Item", "boProdCategory": "Kategori", "boProdType": "Tipe Item",
			"boProdTypeService": "Jasa (SERVICE)", "boProdTypeProduct": "Produk (PRODUCT)",
			"boProdPrice": "Harga (Rupiah)", "boProdPriceHint": "Contoh: 40000 untuk Rp 40.000",
			"boProdComm": "Komisi per-Item (Rupiah)", "boProdCommHint": "Contoh: 5000 untuk Rp 5.000 (Khusus produk)",
			"boProdSave": "Simpan Item", "boProdUpdate": "Perbarui Item", "boProdCancelEdit": "Batal Edit", "boProdList": "Daftar Katalog",
			"boProdColName": "Nama", "boProdColCategory": "Kategori", "boProdColType": "Tipe",
			"boProdColPrice": "Harga (Rp)", "boProdColComm": "Komisi / Item (Rp)",
			"boProdColCommLine1": "Komisi", "boProdColCommLine2": "/ Item (Rp)", "boProdColStatus": "Status",
			"boProdLabelProduct": "PRODUK", "boProdLabelService": "JASA",
			"boProdActionEdit": "Edit", "boProdActionDelete": "Hapus", "boProdConfirmDelete": "Hapus item katalog ini?",
			"boProdCategorySelect": "-- Pilih Kategori --", "boProdCategoryNew": "+ Kategori Baru...", "boProdCategoryNewPlaceholder": "Ketik nama kategori baru",
			"boProdToggleStatus": "Klik untuk ubah status aktif/tidak aktif",
			"boProdEmpty": "Belum ada item dalam katalog.",
			"boProdSavedToast": "Item berhasil disimpan", "boProdDeletedToast": "Item berhasil dihapus", "boProdStatusUpdated": "Status item diperbarui",
			// Thermal Printer keys (ID)
			"printReceipt": "Cetak Struk", "thermalPrinter": "Printer Thermal", "printerConnected": "Printer Terhubung",
			"printerDisconnected": "Printer Belum Terhubung", "connectPrinter": "Hubungkan Printer Bluetooth",
			"disconnectPrinter": "Putuskan Printer", "testPrint": "Cetak Uji Coba", "printViaBluetooth": "Cetak via Bluetooth",
			"printViaBrowser": "Cetak via Browser (PDF)", "receiptPreview": "Pratinjau Struk",
			"printReceiptQuestion": "Apakah Anda ingin langsung mencetak struk transaksi ini?",
			"printReceiptSuccess": "Struk berhasil dicetak.", "skip": "Lewati",
			"printerModalTitle": "Pengaturan Printer Thermal (Okay 58D)",
			"printerModalHint": "Hubungkan ke printer thermal 58mm via Web Bluetooth (Okay 58D) atau gunakan dialog cetak browser jika printer belum terhubung.",
			"storeName": "Pardis Barber Shop", "receiptFooter": "Terima Kasih Atas Kunjungan Anda!",
			// Operator Credential keys (ID)
			"phoneNumber": "Nomor HP / WhatsApp", "phonePlaceholder": "Contoh: 08123456789",
			"branch": "Cabang", "selectBranchOptional": "-- Pilih Cabang (Opsional) --", "selectBranch": "-- Pilih Cabang --",
			"staffType": "Tipe Staf", "barberman": "Barberman", "cashier": "Kasir", "manager": "Manager",
			"bankName": "Nama / Kode Bank", "bankNamePlaceholder": "Contoh: BCA, Mandiri, BRI, BNI",
			"bankAccount": "Nomor Rekening Bank", "bankAccountPlaceholder": "Contoh: 001 1234567",
			"editEmployeeCredentials": "⚙️ Edit Data Karyawan & Kredensial Bank",
			"fullName": "Nama Lengkap", "accountNumber": "Nomor Rekening", "accountNumberPlaceholder": "Nomor rekening",
			"cancel": "Batal", "saveCredentialChanges": "Simpan Perubahan Kredensial",
			"confirmChangesTitle": "Konfirmasi Perubahan", "confirmChangesDesc": "Pastikan data karyawan sudah benar sebelum disimpan.",
			"credentialChangeDetails": "Detail Perubahan Kredensial", "credentialChangeNotice": "Perubahan akan disimpan dan langsung berlaku untuk operasional serta bagi hasil.",
			"yesSaveChanges": "Ya, Simpan Perubahan", "saving": "Menyimpan...", "accountShort": "Rek",
			"dataSavedSuccess": "Data Berhasil Disimpan", "dataSavedSuccessMsg": "Perubahan data karyawan telah berhasil disimpan.",
			"configSavedSuccess": "Konfigurasi Berhasil Disimpan", "configSavedSuccessMsg": "Pengaturan persentase bagi hasil berhasil diperbarui.",
			"discSavedSuccess": "Diskon Berhasil Disimpan", "discSavedSuccessMsg": "Diskon telah berhasil ditambahkan ke katalog.",
			"discDeletedSuccess": "Diskon Berhasil Dihapus", "discDeletedSuccessMsg": "Diskon telah berhasil dihapus dari sistem.",
			"confirmDeleteDisc": "Hapus Diskon", "confirmDeleteDiscDesc": "Apakah Anda yakin ingin menghapus diskon atau paket bundling ini?",
			"yesDeleteDiscount": "Ya, Hapus Diskon", "deleting": "Menghapus...",
			"noDiscount": "Tanpa Diskon (Harga Normal)", "subtotal": "Subtotal", "discount": "Diskon",
			"shiftTransactions": "Transaksi Shift", "cashierIncome": "Pemasukan Shift", "cashierExpense": "Pengeluaran Shift", "cashDrawerBalance": "Saldo Kas Laci",
			"cashierSopTitle": "SOP & Panduan Kasir Bertugas",
			"cashierSop1Title": "Verifikasi QRIS & Transfer", "cashierSop1Desc": "Pastikan notifikasi dana masuk m-Banking telah dicek sebelum mencetak dan menyerahkan struk.",
			"cashierSop2Title": "Pencatatan Real-time", "cashierSop2Desc": "Catat setiap jasa pangkas rambut, treatment, dan produk grooming segera setelah pembayaran.",
			"cashierSop3Title": "Serah Terima & Hitung Laci", "cashierSop3Desc": "Hitung fisik uang tunai di laci kasir agar sesuai dengan saldo tercatat sebelum pergantian shift.",
			"printerStatusTitle": "Perangkat Printer Struk", "printerStatusDesc": "Koneksikan printer thermal Okay 58D via Bluetooth atau gunakan dialog cetak browser bawaan.",
			// 2FA / TOTP keys (ID)
			"loginSetup2FATitle": "Aktivasi Two-Factor Authentication (2FA)",
			"loginSetup2FASubtitle": "Tingkatkan keamanan akun POS Anda dengan Google Authenticator atau aplikasi TOTP RFC 6238 lainnya.",
			"scanQRCodeStep": "1. Pindai Kode QR",
			"scanQRCodeDesc": "Buka aplikasi Google Authenticator di smartphone Anda, ketuk tanda +, dan pindai kode QR di bawah ini:",
			"orEnterManualKey": "Atau masukkan kode rahasia secara manual jika Anda tidak dapat memindai QR:",
			"copySecretKey": "Salin Kode Rahasia",
			"secretCopied": "Kode berhasil disalin!",
			"verifyCodeStep": "2. Masukkan 6 Digit Kode Verifikasi",
			"verifyCodeDesc": "Masukkan 6 digit kode yang tampil di aplikasi authenticator untuk menyelesaikan aktivasi:",
			"otpPlaceholder": "6 digit kode (contoh: 123456)",
			"verifyAndActivate": "Verifikasi & Aktifkan 2FA",
			"verifyOTPTitle": "Verifikasi Two-Factor Authentication (2FA)",
			"verifyOTPSubtitle": "Masukkan 6 digit kode dari Google Authenticator untuk melanjutkan masuk ke POS.",
			"verifyOTPButton": "Verifikasi & Masuk",
			"invalidOTPCode": "Kode verifikasi salah atau telah kedaluwarsa. Silakan periksa kembali aplikasi authenticator Anda.",
			"rateLimitedLocked": "Terlalu banyak percobaan salah. Akun dikunci sementara selama %d detik demi keamanan.",
			"totpBadge": "Dilindungi TOTP RFC 6238",
			"devBypassNotice": "Mode Development: Masukkan master kode 123456 atau 000000 untuk masuk tanpa scan HP.",
			"devBypassButton": "⚡ Masuk Instan (Dev Mode)",
		}
		if value, ok := id[key]; ok {
			return value
		}
	}
	return en[key]
}
