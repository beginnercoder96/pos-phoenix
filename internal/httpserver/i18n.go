package httpserver

import (
	"fmt"
	"strconv"
	"time"
)

func templateFuncs(location *time.Location) map[string]any {
	return map[string]any{
		"money": func(cents int64) string { return formatRupiah(cents) },
		"date":  func(t time.Time) string { return t.In(location).Format("02 Jan 2006 15:04") },
		"sub":   func(a, b int64) int64 { return a - b },
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

func translate(language, key string) string {
	en := map[string]string{
		"language": "Language", "theme": "Theme", "english": "English", "indonesian": "Indonesian", "light": "Light", "dark": "Dark", "save": "Save", "currency": "Indonesian Rupiah", "appName": "POS Phoenix", "welcome": "Welcome back", "secureCashflow": "A place where the nice haircut trim come from us!", "email": "Email", "password": "Password", "signIn": "Sign in", "cashflowReport": "Cash-flow report", "today": "Today", "thisMonth": "This month", "from": "From", "to": "To", "apply": "Apply", "downloadCSV": "Download CSV", "income": "Income", "expense": "Expense", "balance": "Balance", "newTransaction": "New transaction", "type": "Type", "amount": "Amount", "category": "Category", "note": "Note", "saveTransaction": "Save transaction", "transactions": "Transactions", "records": "records", "page": "page", "previous": "Previous", "next": "Next", "noTransactions": "No transactions in this period.", "reverseReason": "Reversal reason", "reverse": "Reverse", "reversedBy": "Reversed by transaction", "operatorAdministration": "Operator administration", "dashboard": "Dashboard", "createOperator": "Create operator", "displayName": "Display name", "temporaryPassword": "Temporary password", "passwordHint": "12–128 characters. Share it securely.", "operators": "Operators", "active": "Active", "inactive": "Inactive", "deactivate": "Deactivate", "activate": "Activate", "noOperators": "No operators created yet.", "signOut": "Sign out", "reversal": "Reversal", "customerRefund": "Customer refund",
		"username": "Username", "usernamePlaceholder": "Enter username", "usernameHint": "3–50 characters, letters, numbers, dot, dash, underscore.",
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
		// Backoffice keys
		"boBackoffice": "Backoffice", "boMainPOS": "← Main POS", "boSignOut": "Sign Out",
		"boDashboard": "Dashboard", "boProfitSharing": "Profit Sharing", "boDiscounts": "Discounts & Bundling",
		"boProducts": "Catalog & Commission", "boPayroll": "Payroll",
		"boDownloadReport": "Download Financial Report (Excel)",
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
		"boSaveConfig":  "💾 Save Profit Sharing Configuration",
		"boNoEmployees": "No employees assigned to this branch. Add employees via the Operators page.",
		"boNetRevenue":  "Branch Net Service Revenue", "boReserveBalance": "Unallocated Reserve",
		"boReserveRemainder": "remaining reserve",
		"boPayrollTitle":     "Employee Payroll", "boPayrollRecap": "Payroll Summary",
		"boEmpName": "Employee Name", "boProfitSharePct": "Share (%)",
		"boProfitShareAmt": "Service Share", "boProductComm": "Product Commission",
		"boTakeHome": "Net Take-Home Pay", "boAction": "Action",
		"boDownloadSlip": "📄 Download Slip", "boDownloadAll": "📥 Download All Payroll Slips (PDF)",
		"boNoPayroll":      "No employees assigned to this branch, or no profit sharing config for this period.",
		"boDiscountsTitle": "Discount & Bundling Management", "boAddDiscount": "Add New Discount / Bundling",
		"boProductsTitle": "Product Catalog & Commission Settings", "boAddProduct": "Add / Edit Catalog Item",
		// Discounts form
		"boDiscCode": "Code", "boDiscName": "Name", "boDiscType": "Type", "boDiscValue": "Value", "boDiscStatus": "Status",
		"boDiscTypePct": "Percentage (%)", "boDiscTypeFixed": "Fixed Amount (Rp)", "boDiscTypeBundle": "Bundle Package",
		"boDiscActive": "Active", "boDiscInactive": "Inactive",
		"boDiscServiceRatio": "Service Allocation Ratio", "boDiscProductRatio": "Product Allocation Ratio",
		"boDiscSave": "💾 Save Discount", "boDiscList": "Discount & Bundling List",
		"boDiscTypePctLabel": "Percentage", "boDiscTypeFixedLabel": "Fixed", "boDiscTypeBundleLabel": "Bundle",
		"boDiscActionDelete": "🗑️ Delete", "boDiscConfirmDelete": "Delete this discount?",
		"boDiscEmpty": "No discounts or bundles yet.",
		// Products form
		"boProdName": "Item Name", "boProdCategory": "Category", "boProdType": "Item Type",
		"boProdTypeService": "Service (SERVICE)", "boProdTypeProduct": "Product (PRODUCT)",
		"boProdPrice": "Price (Cents)", "boProdPriceHint": "Example: Rp 120,000 = 12000000",
		"boProdComm": "Commission per Item (Cents)", "boProdCommHint": "Rp 5,000 = 500000, Rp 10,000 = 1000000",
		"boProdSave": "💾 Save Item", "boProdList": "Catalog List",
		"boProdColName": "Name", "boProdColCategory": "Category", "boProdColType": "Type",
		"boProdColPrice": "Price", "boProdColComm": "Commission / Item", "boProdColStatus": "Status",
		"boProdLabelProduct": "PRODUCT", "boProdLabelService": "SERVICE",
		"boProdEmpty": "No items in catalog yet.",
	}
	if language == "id" {
		id := map[string]string{
			"language": "Bahasa", "theme": "Tema", "english": "Inggris", "indonesian": "Indonesia", "light": "Terang", "dark": "Gelap", "save": "Simpan", "currency": "Rupiah Indonesia", "appName": "POS Phoenix", "welcome": "Selamat datang kembali", "secureCashflow": "Tempat untuk mendapatkan potongan rambut terbaik Anda!", "email": "Email", "password": "Kata sandi", "signIn": "Masuk", "cashflowReport": "Laporan arus kas", "today": "Hari ini", "thisMonth": "Bulan ini", "from": "Dari", "to": "Sampai", "apply": "Terapkan", "downloadCSV": "Unduh CSV", "income": "Pemasukan", "expense": "Pengeluaran", "balance": "Saldo", "newTransaction": "Transaksi baru", "type": "Jenis", "amount": "Jumlah", "category": "Kategori", "note": "Catatan", "saveTransaction": "Simpan transaksi", "transactions": "Transaksi", "records": "catatan", "page": "halaman", "previous": "Sebelumnya", "next": "Berikutnya", "noTransactions": "Tidak ada transaksi pada periode ini.", "reverseReason": "Alasan pembatalan", "reverse": "Batalkan", "reversedBy": "Dibatalkan oleh transaksi", "operatorAdministration": "Administrasi operator", "dashboard": "Dasbor", "createOperator": "Buat operator", "displayName": "Nama tampilan", "temporaryPassword": "Kata sandi sementara", "passwordHint": "12–128 karakter. Bagikan dengan aman.", "operators": "Operator", "active": "Aktif", "inactive": "Tidak aktif", "deactivate": "Nonaktifkan", "activate": "Aktifkan", "noOperators": "Belum ada operator.", "signOut": "Keluar", "reversal": "Pembatalan", "customerRefund": "Pengembalian dana pelanggan",
			"username": "Username", "usernamePlaceholder": "Masukkan username", "usernameHint": "3–50 karakter, huruf, angka, titik, strip, garis bawah.",
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
			// Backoffice keys (ID)
			"boBackoffice": "Backoffice", "boMainPOS": "← Kasir Utama", "boSignOut": "Keluar",
			"boDashboard": "Dasbor", "boProfitSharing": "Bagi Hasil", "boDiscounts": "Diskon & Bundling",
			"boProducts": "Katalog & Komisi", "boPayroll": "Slip Gaji",
			"boDownloadReport": "Unduh Laporan Keuangan (Excel)",
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
			"boSaveConfig":  "💾 Simpan Konfigurasi Bagi Hasil",
			"boNoEmployees": "Belum ada karyawan yang ditugaskan ke cabang ini. Tambahkan karyawan melalui halaman Operators.",
			"boNetRevenue":  "Net Omzet Jasa Cabang", "boReserveBalance": "Saldo Cadangan (Tidak Terpakai)",
			"boReserveRemainder": "sisa saldo",
			"boPayrollTitle":     "Slip Gaji Karyawan", "boPayrollRecap": "Rekap Gaji",
			"boEmpName": "Nama Karyawan", "boProfitSharePct": "Bagi Hasil (%)",
			"boProfitShareAmt": "Bagi Hasil Jasa", "boProductComm": "Komisi Produk",
			"boTakeHome": "Total Gaji Bersih", "boAction": "Aksi",
			"boDownloadSlip": "📄 Download Slip", "boDownloadAll": "📥 Download Semua Slip Gaji (PDF)",
			"boNoPayroll":      "Belum ada karyawan yang ditugaskan ke cabang ini, atau belum ada konfigurasi bagi hasil untuk periode ini.",
			"boDiscountsTitle": "Manajemen Diskon & Bundling", "boAddDiscount": "Tambah Diskon / Bundling Baru",
			"boProductsTitle": "Katalog Produk & Pengaturan Komisi", "boAddProduct": "Tambah / Edit Item Katalog",
			// Discounts form (ID)
			"boDiscCode": "Kode", "boDiscName": "Nama", "boDiscType": "Tipe", "boDiscValue": "Nilai", "boDiscStatus": "Status",
			"boDiscTypePct": "Persentase (%)", "boDiscTypeFixed": "Nominal Tetap (Rp)", "boDiscTypeBundle": "Paket Bundling",
			"boDiscActive": "Aktif", "boDiscInactive": "Tidak Aktif",
			"boDiscServiceRatio": "Rasio Alokasi Jasa", "boDiscProductRatio": "Rasio Alokasi Produk",
			"boDiscSave": "💾 Simpan Diskon", "boDiscList": "Daftar Diskon & Bundling",
			"boDiscTypePctLabel": "Persentase", "boDiscTypeFixedLabel": "Nominal", "boDiscTypeBundleLabel": "Bundling",
			"boDiscActionDelete": "🗑️ Hapus", "boDiscConfirmDelete": "Hapus diskon ini?",
			"boDiscEmpty": "Belum ada diskon atau bundling.",
			// Products form (ID)
			"boProdName": "Nama Item", "boProdCategory": "Kategori", "boProdType": "Tipe Item",
			"boProdTypeService": "Jasa (SERVICE)", "boProdTypeProduct": "Produk (PRODUCT)",
			"boProdPrice": "Harga (Cents)", "boProdPriceHint": "Contoh: Rp 120.000 = 12000000",
			"boProdComm": "Komisi per-Item (Cents)", "boProdCommHint": "Rp 5.000 = 500000, Rp 10.000 = 1000000",
			"boProdSave": "💾 Simpan Item", "boProdList": "Daftar Katalog",
			"boProdColName": "Nama", "boProdColCategory": "Kategori", "boProdColType": "Tipe",
			"boProdColPrice": "Harga", "boProdColComm": "Komisi / Item", "boProdColStatus": "Status",
			"boProdLabelProduct": "PRODUK", "boProdLabelService": "JASA",
			"boProdEmpty": "Belum ada item dalam katalog.",
		}
		if value, ok := id[key]; ok {
			return value
		}
	}
	return en[key]
}
