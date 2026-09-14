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
		"addCategory": "Add Category / Item", "item": "Item / Service", "total": "Total", "remove": "Remove", "selectCategory": "Select category", "selectItem": "Select service / item", "customItem": "Custom item", "items": "Items", "downloadExcel": "Download Excel (.xlsx)",
		"customDate": "Custom Date",
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
	}
	if language == "id" {
		id := map[string]string{
			"language": "Bahasa", "theme": "Tema", "english": "Inggris", "indonesian": "Indonesia", "light": "Terang", "dark": "Gelap", "save": "Simpan", "currency": "Rupiah Indonesia", "appName": "POS Phoenix", "welcome": "Selamat datang kembali", "secureCashflow": "Tempat untuk mendapatkan potongan rambut terbaik Anda!", "email": "Email", "password": "Kata sandi", "signIn": "Masuk", "cashflowReport": "Laporan arus kas", "today": "Hari ini", "thisMonth": "Bulan ini", "from": "Dari", "to": "Sampai", "apply": "Terapkan", "downloadCSV": "Unduh CSV", "income": "Pemasukan", "expense": "Pengeluaran", "balance": "Saldo", "newTransaction": "Transaksi baru", "type": "Jenis", "amount": "Jumlah", "category": "Kategori", "note": "Catatan", "saveTransaction": "Simpan transaksi", "transactions": "Transaksi", "records": "catatan", "page": "halaman", "previous": "Sebelumnya", "next": "Berikutnya", "noTransactions": "Tidak ada transaksi pada periode ini.", "reverseReason": "Alasan pembatalan", "reverse": "Batalkan", "reversedBy": "Dibatalkan oleh transaksi", "operatorAdministration": "Administrasi operator", "dashboard": "Dasbor", "createOperator": "Buat operator", "displayName": "Nama tampilan", "temporaryPassword": "Kata sandi sementara", "passwordHint": "12–128 karakter. Bagikan dengan aman.", "operators": "Operator", "active": "Aktif", "inactive": "Tidak aktif", "deactivate": "Nonaktifkan", "activate": "Aktifkan", "noOperators": "Belum ada operator.", "signOut": "Keluar", "reversal": "Pembatalan", "customerRefund": "Pengembalian dana pelanggan",
			"addCategory": "Tambah Kategori / Item", "item": "Item / Layanan", "total": "Total", "remove": "Hapus", "selectCategory": "Pilih kategori", "selectItem": "Pilih layanan / item", "customItem": "Item lainnya", "items": "Item", "downloadExcel": "Unduh Excel (.xlsx)",
			"customDate": "Tanggal Khusus",
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
		}
		if value, ok := id[key]; ok {
			return value
		}
	}
	return en[key]
}
