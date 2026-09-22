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
    <col min="1" max="1" width="32" customWidth="1"/>
    <col min="2" max="2" width="25" customWidth="1"/>
    <col min="3" max="3" width="22" customWidth="1"/>
    <col min="4" max="4" width="22" customWidth="1"/>
  </cols>
  <sheetData>`)

	// Header
	sb.WriteString(fmt.Sprintf(`
    <row r="1"><c r="A1" s="1" t="inlineStr"><is><t>SLIP GAJI KARYAWAN</t></is></c></row>
    <row r="2"><c r="A2" s="1" t="inlineStr"><is><t>PARDIS BARBERSHOP</t></is></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>Cabang</t></is></c><c r="B3" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="4"><c r="A4" t="inlineStr"><is><t>Nama Karyawan</t></is></c><c r="B4" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="5"><c r="A5" t="inlineStr"><is><t>Periode</t></is></c><c r="B5" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="6"></row>`,
		xmlEscape(data.BranchName),
		xmlEscape(data.EmployeeName),
		xmlEscape(data.Period)))

	// Profit Sharing Section
	sb.WriteString(fmt.Sprintf(`
    <row r="7"><c r="A7" s="1" t="inlineStr"><is><t>RINCIAN BAGI HASIL JASA</t></is></c></row>
    <row r="8"><c r="A8" t="inlineStr"><is><t>Omzet Jasa Cabang</t></is></c><c r="B8" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="9"><c r="A9" t="inlineStr"><is><t>Persentase Bagi Hasil</t></is></c><c r="B9" t="inlineStr"><is><t>%.1f%%</t></is></c></row>
    <row r="10"><c r="A10" s="3" t="inlineStr"><is><t>Nominal Bagi Hasil</t></is></c><c r="B10" s="3" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="11"></row>`,
		formatRupiahExcel(data.NetServiceRev),
		data.SharePercentage,
		formatRupiahExcel(data.ServiceShare)))

	// Product Commission Section with Itemized details
	row := 12
	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="1" t="inlineStr"><is><t>RINCIAN KOMISI PENJUALAN PRODUK</t></is></c></row>`, row, row))
	row++

	if len(data.ProductSales) > 0 {
		sb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="A%d" s="1" t="inlineStr"><is><t>Nama Produk</t></is></c>
      <c r="B%d" s="1" t="inlineStr"><is><t>Qty Terjual</t></is></c>
      <c r="C%d" s="1" t="inlineStr"><is><t>Komisi / Pcs</t></is></c>
      <c r="D%d" s="1" t="inlineStr"><is><t>Subtotal Komisi</t></is></c>
    </row>`, row, row, row, row, row))
		row++
		for _, ps := range data.ProductSales {
			sb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="A%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="B%d" s="2"><v>%d</v></c>
      <c r="C%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="D%d" s="2"><v>%.2f</v></c>
    </row>`, row, row, xmlEscape(ps.ProductName), row, ps.Quantity, row, formatRupiahExcel(ps.CommissionRate), row, float64(ps.TotalCommission)/100.0))
			row++
		}
	} else {
		sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>(Tidak ada penjualan produk pada periode ini)</t></is></c></row>`, row, row))
		row++
	}

	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="3" t="inlineStr"><is><t>Total Komisi Produk</t></is></c><c r="B%d" s="3" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="%d"></row>`, row, row, row, formatRupiahExcel(data.ProductComm), row+1))
	row += 2

	// Take Home Pay Summary & Signature Block
	sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="1" t="inlineStr"><is><t>RINGKASAN GAJI BERSIH (TAKE HOME PAY)</t></is></c></row>
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>Bagi Hasil Jasa</t></is></c><c r="B%d" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>Komisi Penjualan Produk</t></is></c><c r="B%d" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="%d"><c r="A%d" s="3" t="inlineStr"><is><t>TOTAL GAJI BERSIH</t></is></c><c r="B%d" s="3" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="%d"></row>
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>Tanda Tangan:</t></is></c></row>
    <row r="%d"></row>
    <row r="%d"></row>
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>_________________</t></is></c><c r="B%d" t="inlineStr"><is><t>_________________</t></is></c></row>
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>Karyawan</t></is></c><c r="B%d" t="inlineStr"><is><t>Owner (Ipang)</t></is></c></row>`,
		row, row,
		row+1, row+1, row+1, formatRupiahExcel(data.ServiceShare),
		row+2, row+2, row+2, formatRupiahExcel(data.ProductComm),
		row+3, row+3, row+3, formatRupiahExcel(data.TakeHomePay),
		row+4,
		row+5, row+5,
		row+6,
		row+7,
		row+8, row+8, row+8,
		row+9, row+9, row+9))

	sb.WriteString(`
  </sheetData>
</worksheet>`)

	io.WriteString(w, sb.String())

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// generatePayrollSlipAllExcel creates a multi-sheet Excel with one sheet per employee.
func generatePayrollSlipAllExcel(slips []payrollSlipData) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	seenNames := make(map[string]int)
	sheetNames := make([]string, len(slips))
	for i, s := range slips {
		name := strings.TrimSpace(s.EmployeeName)
		if name == "" {
			name = fmt.Sprintf("Karyawan %d", i+1)
		}
		for _, ch := range []string{"\\", "/", "?", "*", ":", "[", "]"} {
			name = strings.ReplaceAll(name, ch, "_")
		}
		if len(name) > 25 {
			name = name[:25]
		}
		seenNames[name]++
		if seenNames[name] > 1 {
			name = fmt.Sprintf("%s (%d)", name, seenNames[name])
		}
		sheetNames[i] = name
	}

	writeContentTypes(zw, len(slips))
	writeRels(zw)
	writeWorkbookRels(zw, len(slips))
	writeWorkbook(zw, sheetNames)
	writePayrollStyles(zw)

	for i, data := range slips {
		w, err := zw.Create(fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1))
		if err != nil {
			return nil, err
		}

		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <cols>
    <col min="1" max="1" width="32" customWidth="1"/>
    <col min="2" max="2" width="25" customWidth="1"/>
    <col min="3" max="3" width="22" customWidth="1"/>
    <col min="4" max="4" width="22" customWidth="1"/>
  </cols>
  <sheetData>`)

		sb.WriteString(fmt.Sprintf(`
    <row r="1"><c r="A1" s="1" t="inlineStr"><is><t>SLIP GAJI - %s</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>Cabang: %s</t></is></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>Periode: %s</t></is></c></row>
    <row r="4"></row>
    <row r="5"><c r="A5" t="inlineStr"><is><t>Bagi Hasil Jasa (%.1f%%)</t></is></c><c r="B5" t="inlineStr"><is><t>%s</t></is></c></row>`,
			xmlEscape(data.EmployeeName),
			xmlEscape(data.BranchName),
			xmlEscape(data.Period),
			data.SharePercentage,
			formatRupiahExcel(data.ServiceShare)))

		row := 6
		if len(data.ProductSales) > 0 {
			sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" s="1" t="inlineStr"><is><t>Rincian Penjualan Produk</t></is></c></row>`, row, row))
			row++
			for _, ps := range data.ProductSales {
				sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>%s (x%d)</t></is></c><c r="B%d" t="inlineStr"><is><t>%s</t></is></c></row>`,
					row, row, xmlEscape(ps.ProductName), ps.Quantity, row, formatRupiahExcel(ps.TotalCommission)))
				row++
			}
		}

		sb.WriteString(fmt.Sprintf(`
    <row r="%d"><c r="A%d" t="inlineStr"><is><t>Total Komisi Produk</t></is></c><c r="B%d" t="inlineStr"><is><t>%s</t></is></c></row>
    <row r="%d"><c r="A%d" s="3" t="inlineStr"><is><t>TOTAL GAJI BERSIH</t></is></c><c r="B%d" s="3" t="inlineStr"><is><t>%s</t></is></c></row>`,
			row, row, row, formatRupiahExcel(data.ProductComm),
			row+1, row+1, row+1, formatRupiahExcel(data.TakeHomePay)))

		sb.WriteString(`
  </sheetData>
</worksheet>`)
		io.WriteString(w, sb.String())
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
  <fonts count="3">
    <font><name val="Calibri"/><sz val="11"/></font>
    <font><b/><color rgb="FFFFFFFF"/><name val="Calibri"/><sz val="11"/></font>
    <font><b/><name val="Calibri"/><sz val="11"/></font>
  </fonts>
  <fills count="3">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF172033"/></patternFill></fill>
  </fills>
  <borders count="1">
    <border><left/><right/><top/><bottom/></border>
  </borders>
  <cellStyleXfs count="1">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0"/>
  </cellStyleXfs>
  <cellXfs count="4">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>
    <xf numFmtId="3" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>
    <xf numFmtId="3" fontId="2" fillId="0" borderId="0" xfId="0" applyFont="1" applyNumberFormat="1"/>
  </cellXfs>
</styleSheet>`)
}
