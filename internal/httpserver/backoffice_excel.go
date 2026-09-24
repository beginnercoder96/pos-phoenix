package httpserver

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/mekari/pos-phoenix/internal/backoffice"
	"github.com/xuri/excelize/v2"
)

type payrollSlipData struct {
	BranchName        string
	EmployeeID        int64
	EmployeeName      string
	StaffType         string
	PhoneNumber       string
	BankName          string
	BankAccountNumber string
	Period            string
	PrintDate         string
	NetServiceRev     int64
	SharePercentage   float64
	ServiceShare      int64
	ProductComm       int64
	TakeHomePay       int64
	ProductSales      []backoffice.EmployeeProductSaleItem
}

type branchReportData struct {
	Branch    backoffice.Branch
	Analytics []backoffice.MonthlyAnalytics
}

func formatRupiahExcel(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole := cents / 100
	fraction := cents % 100
	digits := fmt.Sprintf("%d", whole)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "." + digits[i:]
	}
	result := "Rp " + digits + "," + fmt.Sprintf("%02d", fraction)
	if negative {
		return "-" + result
	}
	return result
}

func formatBankInfo(bankName, bankAccountNumber string) string {
	bankName = strings.TrimSpace(bankName)
	bankAccountNumber = strings.TrimSpace(bankAccountNumber)
	if bankName != "" && bankAccountNumber != "" {
		return fmt.Sprintf("%s - %s", bankName, bankAccountNumber)
	}
	if bankAccountNumber != "" {
		return bankAccountNumber
	}
	if bankName != "" {
		return bankName
	}
	return "-"
}

// generatePayrollSlipExcel creates a single-sheet payroll slip Excel file.
func generatePayrollSlipExcel(data payrollSlipData) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	writeContentTypes(zw, 1)
	writeRels(zw)
	writeWorkbookRels(zw, 1)
	writeWorkbook(zw, []string{"Slip Gaji"})
	writePayrollStyles(zw)

	w, err := zw.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <cols>
    <col min="1" max="1" width="34" customWidth="1"/>
    <col min="2" max="2" width="28" customWidth="1"/>
    <col min="3" max="3" width="22" customWidth="1"/>
    <col min="4" max="4" width="24" customWidth="1"/>
  </cols>
  <sheetData>`)

	// Header
	sb.WriteString(fmt.Sprintf(`
    <row r="1"><c r="A1" s="1" t="inlineStr"><is><t>SLIP GAJI KARYAWAN</t></is></c><c r="B1" s="1" t="inlineStr"><is><t></t></is></c><c r="C1" s="1" t="inlineStr"><is><t></t></is></c><c r="D1" s="1" t="inlineStr"><is><t></t></is></c></row>
    <row r="2"><c r="A2" s="1" t="inlineStr"><is><t>PARDIS BARBERSHOP</t></is></c><c r="B2" s="1" t="inlineStr"><is><t></t></is></c><c r="C2" s="1" t="inlineStr"><is><t></t></is></c><c r="D2" s="1" t="inlineStr"><is><t></t></is></c></row>
    <row r="3"><c r="A3" s="6" t="inlineStr"><is><t>Cabang</t></is></c><c r="B3" s="5" t="inlineStr"><is><t>%s</t></is></c><c r="C3" s="6" t="inlineStr"><is><t>Tanggal Cetak</t></is></c><c r="D3" s="5" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="4"><c r="A4" s="6" t="inlineStr"><is><t>Nama Karyawan</t></is></c><c r="B4" s="5" t="inlineStr"><is><t>%s</t></is></c><c r="C4" s="6" t="inlineStr"><is><t>Jabatan / Peran</t></is></c><c r="D4" s="5" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="5"><c r="A5" s="6" t="inlineStr"><is><t>Periode Gaji</t></is></c><c r="B5" s="5" t="inlineStr"><is><t>%s</t></is></c><c r="C5" s="6" t="inlineStr"><is><t>Rekening Pembayaran</t></is></c><c r="D5" s="5" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="6"></row>`,
		xmlEscape(data.BranchName),
		xmlEscape(data.PrintDate),
		xmlEscape(data.EmployeeName),
		xmlEscape(strings.ToUpper(data.StaffType)),
		xmlEscape(data.Period),
		xmlEscape(formatBankInfo(data.BankName, data.BankAccountNumber))))

	// Profit Sharing Section
	sb.WriteString(fmt.Sprintf(`
    <row r="7"><c r="A7" s="2" t="inlineStr"><is><t>RINCIAN BAGI HASIL JASA</t></is></c><c r="B7" s="2" t="inlineStr"><is><t></t></is></c><c r="C7" s="2" t="inlineStr"><is><t></t></is></c><c r="D7" s="2" t="inlineStr"><is><t></t></is></c></row>
    <row r="8"><c r="A8" s="5" t="inlineStr"><is><t>Omzet Jasa Cabang</t></is></c><c r="B8" s="7" t="inlineStr"><is><t>%s</t></is></c><c r="C8" s="5" t="inlineStr"><is><t></t></is></c><c r="D8" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="9"><c r="A9" s="5" t="inlineStr"><is><t>Persentase Bagi Hasil</t></is></c><c r="B9" s="7" t="inlineStr"><is><t>%.1f%%</t></is></c><c r="C9" s="5" t="inlineStr"><is><t></t></is></c><c r="D9" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="10"><c r="A10" s="6" t="inlineStr"><is><t>Nominal Bagi Hasil Jasa</t></is></c><c r="B10" s="8" t="inlineStr"><is><t>%s</t></is></c><c r="C10" s="5" t="inlineStr"><is><t></t></is></c><c r="D10" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="11"></row>`,
		formatRupiahExcel(data.NetServiceRev),
		data.SharePercentage,
		formatRupiahExcel(data.ServiceShare)))

	// Product Commission Section with Itemized details
	row := 12
	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="2" t="inlineStr"><is><t>RINCIAN KOMISI PENJUALAN PRODUK</t></is></c><c r="B%d" s="2" t="inlineStr"><is><t></t></is></c><c r="C%d" s="2" t="inlineStr"><is><t></t></is></c><c r="D%d" s="2" t="inlineStr"><is><t></t></is></c></row>`, row, row, row, row, row))
	row++

	if len(data.ProductSales) > 0 {
		sb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="A%d" s="3" t="inlineStr"><is><t>Nama Produk</t></is></c>
      <c r="B%d" s="3" t="inlineStr"><is><t>Qty Terjual</t></is></c>
      <c r="C%d" s="3" t="inlineStr"><is><t>Komisi / Pcs</t></is></c>
      <c r="D%d" s="3" t="inlineStr"><is><t>Subtotal Komisi</t></is></c>
    </row>`, row, row, row, row, row))
		row++
		for _, ps := range data.ProductSales {
			sb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="A%d" s="5" t="inlineStr"><is><t>%s</t></is></c>
      <c r="B%d" s="7" t="inlineStr"><is><t>%d pcs</t></is></c>
      <c r="C%d" s="7" t="inlineStr"><is><t>%s</t></is></c>
      <c r="D%d" s="7" t="inlineStr"><is><t>%s</t></is></c>
    </row>`, row, row, xmlEscape(ps.ProductName), row, ps.Quantity, row, formatRupiahExcel(ps.CommissionRate), row, formatRupiahExcel(ps.TotalCommission)))
			row++
		}
	} else {
		sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="5" t="inlineStr"><is><t>(Tidak ada penjualan produk pada periode ini)</t></is></c><c r="B%d" s="5" t="inlineStr"><is><t>-</t></is></c><c r="C%d" s="5" t="inlineStr"><is><t>-</t></is></c><c r="D%d" s="5" t="inlineStr"><is><t>-</t></is></c></row>`, row, row, row, row, row))
		row++
	}

	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="6" t="inlineStr"><is><t>Total Komisi Produk</t></is></c><c r="B%d" s="8" t="inlineStr"><is><t>%s</t></is></c><c r="C%d" s="5" t="inlineStr"><is><t></t></is></c><c r="D%d" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="%d"></row>`, row, row, row, formatRupiahExcel(data.ProductComm), row, row, row+1))
	row += 2

	// Take Home Pay Summary & Signature Block
	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="2" t="inlineStr"><is><t>RINGKASAN GAJI BERSIH (TAKE HOME PAY)</t></is></c><c r="B%d" s="2" t="inlineStr"><is><t></t></is></c><c r="C%d" s="2" t="inlineStr"><is><t></t></is></c><c r="D%d" s="2" t="inlineStr"><is><t></t></is></c></row>
    <row r="%d"><c r="A%d" s="5" t="inlineStr"><is><t>Bagi Hasil Jasa</t></is></c><c r="B%d" s="7" t="inlineStr"><is><t>%s</t></is></c><c r="C%d" s="5" t="inlineStr"><is><t></t></is></c><c r="D%d" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="%d"><c r="A%d" s="5" t="inlineStr"><is><t>Komisi Penjualan Produk</t></is></c><c r="B%d" s="7" t="inlineStr"><is><t>%s</t></is></c><c r="C%d" s="5" t="inlineStr"><is><t></t></is></c><c r="D%d" s="5" t="inlineStr"><is><t></t></is></c></row>
    <row r="%d"><c r="A%d" s="4" t="inlineStr"><is><t>TOTAL GAJI BERSIH</t></is></c><c r="B%d" s="4" t="inlineStr"><is><t>%s</t></is></c><c r="C%d" s="4" t="inlineStr"><is><t></t></is></c><c r="D%d" s="4" t="inlineStr"><is><t></t></is></c></row>
    <row r="%d"></row>
    <row r="%d"><c r="A%d" s="0" t="inlineStr"><is><t>Tanda Tangan Penerima:</t></is></c><c r="C%d" s="0" t="inlineStr"><is><t>Tanda Tangan Pengelola:</t></is></c></row>
    <row r="%d"></row>
    <row r="%d"></row>
    <row r="%d"><c r="A%d" s="0" t="inlineStr"><is><t>________________________</t></is></c><c r="C%d" s="0" t="inlineStr"><is><t>________________________</t></is></c></row>
    <row r="%d"><c r="A%d" s="6" t="inlineStr"><is><t>%s</t></is></c><c r="C%d" s="6" t="inlineStr"><is><t>Ipang (Owner)</t></is></c></row>`,
		row, row, row, row, row,
		row+1, row+1, row+1, formatRupiahExcel(data.ServiceShare), row+1, row+1,
		row+2, row+2, row+2, formatRupiahExcel(data.ProductComm), row+2, row+2,
		row+3, row+3, row+3, formatRupiahExcel(data.TakeHomePay), row+3, row+3,
		row+4,
		row+5, row+5, row+5,
		row+6,
		row+7,
		row+8, row+8, row+8,
		row+9, row+9, xmlEscape(data.EmployeeName), row+9))

	sb.WriteString(`
  </sheetData>
</worksheet>`)

	io.WriteString(w, sb.String())

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generatePayrollSlipAllExcel deprecated, use generatePayrollSlipsZip instead.
func generatePayrollSlipAllExcel(slips []payrollSlipData) ([]byte, error) {
	return generatePayrollSlipsZip(slips)
}

// generatePayrollSlipsZip packages individual Excel files for each employee into a single ZIP archive.
func generatePayrollSlipsZip(slips []payrollSlipData) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, slip := range slips {
		excelBytes, err := generatePayrollSlipExcel(slip)
		if err != nil {
			return nil, err
		}

		empName := strings.ReplaceAll(slip.EmployeeName, " ", "_")
		branchName := strings.ReplaceAll(slip.BranchName, " ", "_")
		filename := fmt.Sprintf("Slip_Gaji_%s_%s_%s.xlsx", branchName, empName, slip.Period)
		filename = strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
				return '_'
			}
			return r
		}, filename)

		w, err := zw.Create(filename)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(excelBytes); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generateFinancialReportExcel creates a multi-sheet financial report using excelize.
func generateFinancialReportExcel(consolidated []backoffice.MonthlyAnalytics, branchData []branchReportData, productSales []backoffice.ProductSaleRecord) ([]byte, error) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	// Sheet 1: Konsolidasi
	const sheet1 = "Konsolidasi"
	f.SetSheetName("Sheet1", sheet1)
	if err := populateAnalyticsSheetExcelize(f, sheet1, "Rangkuman Konsolidasi 2 Cabang", consolidated); err != nil {
		return nil, err
	}

	// Per-branch sheets
	for _, bd := range branchData {
		sheetName := bd.Branch.Name
		f.NewSheet(sheetName)
		if err := populateAnalyticsSheetExcelize(f, sheetName, "Rincian "+bd.Branch.Name, bd.Analytics); err != nil {
			return nil, err
		}
	}

	// Sheet: Product Sales & Commissions
	const prodSheet = "Penjualan Produk"
	f.NewSheet(prodSheet)
	if err := populateProductSalesSheetExcelize(f, prodSheet, productSales); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func populateAnalyticsSheetExcelize(f *excelize.File, sheetName, title string, analytics []backoffice.MonthlyAnalytics) error {
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "0F172A"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{`"Rp "#,##0.00`}[0],
		Alignment:    &excelize.Alignment{Horizontal: "right"},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true},
		CustomNumFmt: &[]string{`"Rp "#,##0.00`}[0],
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 6, Color: "000000"},
		},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	totalLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 6, Color: "000000"},
		},
	})

	_ = f.SetColWidth(sheetName, "A", "A", 15)
	_ = f.SetColWidth(sheetName, "B", "D", 22)
	_ = f.SetColWidth(sheetName, "E", "E", 18)
	_ = f.SetColWidth(sheetName, "F", "F", 24)

	_ = f.SetCellValue(sheetName, "A1", title)
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	headers := []any{"Bulan", "Omzet Kotor", "Omzet Jasa", "Omzet Produk", "Diskon", "Sisa Kas Cadangan"}
	_ = f.SetSheetRow(sheetName, "A3", &headers)
	_ = f.SetCellStyle(sheetName, "A3", "F3", headerStyle)

	row := 4
	var totalGross, totalService, totalProduct, totalDiscount, totalReserve float64
	for _, a := range analytics {
		gross := float64(a.GrossRevenueCents) / 100.0
		service := float64(a.ServiceRevenueCents) / 100.0
		product := float64(a.ProductRevenueCents) / 100.0
		discount := float64(a.DiscountCents) / 100.0
		reserve := float64(a.ReserveCents) / 100.0

		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), a.Month)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), gross)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), service)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), product)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), discount)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), reserve)

		_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("F%d", row), currencyStyle)

		totalGross += gross
		totalService += service
		totalProduct += product
		totalDiscount += discount
		totalReserve += reserve
		row++
	}

	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL")
	_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), totalGross)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), totalService)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalProduct)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), totalDiscount)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), totalReserve)

	_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), totalLabelStyle)
	_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("F%d", row), totalStyle)

	return nil
}

func populateProductSalesSheetExcelize(f *excelize.File, sheetName string, sales []backoffice.ProductSaleRecord) error {
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "0F172A"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: &[]string{`"Rp "#,##0.00`}[0],
		Alignment:    &excelize.Alignment{Horizontal: "right"},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true},
		CustomNumFmt: &[]string{`"Rp "#,##0.00`}[0],
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 6, Color: "000000"},
		},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	totalLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 6, Color: "000000"},
		},
	})

	_ = f.SetColWidth(sheetName, "A", "A", 14)
	_ = f.SetColWidth(sheetName, "B", "B", 25)
	_ = f.SetColWidth(sheetName, "C", "C", 28)
	_ = f.SetColWidth(sheetName, "D", "D", 22)
	_ = f.SetColWidth(sheetName, "E", "E", 10)
	_ = f.SetColWidth(sheetName, "F", "G", 18)
	_ = f.SetColWidth(sheetName, "H", "I", 22)

	_ = f.SetCellValue(sheetName, "A1", "RINCIAN PENJUALAN PRODUK & KOMISI")
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	headers := []any{"Bulan", "Cabang", "Nama Produk", "Barberman", "Qty", "Harga Satuan", "Komisi / Pcs", "Total Omzet", "Total Komisi"}
	_ = f.SetSheetRow(sheetName, "A3", &headers)
	_ = f.SetCellStyle(sheetName, "A3", "I3", headerStyle)

	row := 4
	var totalRev, totalComm float64
	for _, s := range sales {
		price := float64(s.PriceCents) / 100.0
		commRate := float64(s.CommissionRate) / 100.0
		rev := float64(s.TotalRevenue) / 100.0
		comm := float64(s.TotalCommission) / 100.0

		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), s.PeriodMonth)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.BranchName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.ProductName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), s.BarberName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), s.Quantity)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), price)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), commRate)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), rev)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), comm)

		_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("I%d", row), currencyStyle)

		totalRev += rev
		totalComm += comm
		row++
	}

	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL")
	_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), totalRev)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), totalComm)

	_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), totalLabelStyle)
	_ = f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), totalStyle)

	return nil
}

// OOXML helper functions for multi-sheet Excel

func writeContentTypes(zw *zip.Writer, sheetCount int) {
	w, _ := zw.Create("[Content_Types].xml")
	var sheets strings.Builder
	for i := 1; i <= sheetCount; i++ {
		sheets.WriteString(fmt.Sprintf(`  <Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
`, i))
	}
	io.WriteString(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
%s  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`, sheets.String()))
}

func writeRels(zw *zip.Writer) {
	w, _ := zw.Create("_rels/.rels")
	io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`)
}

func writeWorkbookRels(zw *zip.Writer, sheetCount int) {
	w, _ := zw.Create("xl/_rels/workbook.xml.rels")
	var rels strings.Builder
	for i := 1; i <= sheetCount; i++ {
		rels.WriteString(fmt.Sprintf(`  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>
`, i, i))
	}
	rels.WriteString(fmt.Sprintf(`  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
`, sheetCount+1))
	io.WriteString(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
%s</Relationships>`, rels.String()))
}

func writeWorkbook(zw *zip.Writer, sheetNames []string) {
	w, _ := zw.Create("xl/workbook.xml")
	var sheets strings.Builder
	for i, name := range sheetNames {
		sheets.WriteString(fmt.Sprintf(`    <sheet name="%s" sheetId="%d" r:id="rId%d"/>
`, xmlEscape(name), i+1, i+1))
	}
	io.WriteString(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
%s  </sheets>
</workbook>`, sheets.String()))
}

func writePayrollStyles(zw *zip.Writer) {
	w, _ := zw.Create("xl/styles.xml")
	io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="7">
    <!-- 0: Regular 11 -->
    <font><name val="Calibri"/><sz val="11"/></font>
    <!-- 1: Bold White 12 (Main Title) -->
    <font><b/><color rgb="FFFFFFFF"/><name val="Calibri"/><sz val="12"/></font>
    <!-- 2: Bold White 11 (Section Header) -->
    <font><b/><color rgb="FFFFFFFF"/><name val="Calibri"/><sz val="11"/></font>
    <!-- 3: Bold Dark Slate 10 (Table Column Header) -->
    <font><b/><color rgb="FF0F172A"/><name val="Calibri"/><sz val="10"/></font>
    <!-- 4: Bold Dark Green 12 (Take Home Pay) -->
    <font><b/><color rgb="FF064E3B"/><name val="Calibri"/><sz val="12"/></font>
    <!-- 5: Bold Regular 11 -->
    <font><b/><name val="Calibri"/><sz val="11"/></font>
    <!-- 6: Muted Gray 10 -->
    <font><color rgb="FF64748B"/><name val="Calibri"/><sz val="10"/></font>
  </fonts>
  <fills count="7">
    <!-- 0: none -->
    <fill><patternFill patternType="none"/></fill>
    <!-- 1: gray125 -->
    <fill><patternFill patternType="gray125"/></fill>
    <!-- 2: Pardis Brand Red (FFDC2626) -->
    <fill><patternFill patternType="solid"><fgColor rgb="FFDC2626"/></patternFill></fill>
    <!-- 3: Dark Slate (FF1E293B) -->
    <fill><patternFill patternType="solid"><fgColor rgb="FF1E293B"/></patternFill></fill>
    <!-- 4: Table Header Soft Gray (FFF1F5F9) -->
    <fill><patternFill patternType="solid"><fgColor rgb="FFF1F5F9"/></patternFill></fill>
    <!-- 5: Pastel Green Highlight (FF86EFAC) -->
    <fill><patternFill patternType="solid"><fgColor rgb="FF86EFAC"/></patternFill></fill>
    <!-- 6: Soft Ice Blue/Gray (FFF8FAFC) -->
    <fill><patternFill patternType="solid"><fgColor rgb="FFF8FAFC"/></patternFill></fill>
  </fills>
  <borders count="2">
    <!-- 0: None -->
    <border><left/><right/><top/><bottom/></border>
    <!-- 1: Light Gray Thin Border -->
    <border>
      <left style="thin"><color rgb="FFE2E8F0"/></left>
      <right style="thin"><color rgb="FFE2E8F0"/></right>
      <top style="thin"><color rgb="FFE2E8F0"/></top>
      <bottom style="thin"><color rgb="FFE2E8F0"/></bottom>
    </border>
  </borders>
  <cellStyleXfs count="1">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0"/>
  </cellStyleXfs>
  <cellXfs count="9">
    <!-- 0: Default Normal (no border) -->
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <!-- 1: Brand Title Header (Red, Bold White 12, thin border) -->
    <xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"/>
    <!-- 2: Section Header (Dark Slate, Bold White 11, thin border) -->
    <xf numFmtId="0" fontId="2" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"/>
    <!-- 3: Table Column Header (Soft Gray, Bold Dark 10, thin border) -->
    <xf numFmtId="0" fontId="3" fillId="4" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"/>
    <!-- 4: Take Home Pay Highlight (Pastel Green, Dark Green Bold 12, thin border) -->
    <xf numFmtId="0" fontId="4" fillId="5" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"/>
    <!-- 5: Regular Cell with Thin Border -->
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"/>
    <!-- 6: Bold Cell with Thin Border -->
    <xf numFmtId="0" fontId="5" fillId="0" borderId="1" xfId="0" applyFont="1" applyBorder="1"/>
    <!-- 7: Number/Currency Cell with Thin Border (Right Aligned) -->
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1"><alignment horizontal="right"/></xf>
    <!-- 8: Bold Currency Cell with Thin Border (Right Aligned) -->
    <xf numFmtId="0" fontId="5" fillId="0" borderId="1" xfId="0" applyFont="1" applyBorder="1"><alignment horizontal="right"/></xf>
  </cellXfs>
</styleSheet>`)
}
