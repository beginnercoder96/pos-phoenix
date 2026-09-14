package httpserver

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	transactionstore "github.com/mekari/pos-phoenix/internal/transaction"
)

func xmlEscape(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&apos;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func generateExcelReport(entries []transactionstore.Entry, loc *time.Location, from, to time.Time) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// 1. [Content_Types].xml
	w, err := zw.Create("[Content_Types].xml")
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`)

	// 2. _rels/.rels
	w, err = zw.Create("_rels/.rels")
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`)

	// 3. xl/_rels/workbook.xml.rels
	w, err = zw.Create("xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`)

	// 4. xl/workbook.xml
	w, err = zw.Create("xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Transactions" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`)

	// 5. xl/styles.xml
	w, err = zw.Create("xl/styles.xml")
	if err != nil {
		return nil, err
	}
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
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

	// 6. xl/worksheets/sheet1.xml
	w, err = zw.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return nil, err
	}

	var sheetSb strings.Builder
	sheetSb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <cols>
    <col min="1" max="1" width="8" customWidth="1"/>
    <col min="2" max="2" width="20" customWidth="1"/>
    <col min="3" max="3" width="20" customWidth="1"/>
    <col min="4" max="4" width="12" customWidth="1"/>
    <col min="5" max="5" width="18" customWidth="1"/>
    <col min="6" max="6" width="25" customWidth="1"/>
    <col min="7" max="7" width="38" customWidth="1"/>
    <col min="8" max="8" width="28" customWidth="1"/>
    <col min="9" max="9" width="16" customWidth="1"/>
  </cols>
  <sheetData>
    <row r="1">
      <c r="A1" s="1" t="inlineStr"><is><t>ID</t></is></c>
      <c r="B1" s="1" t="inlineStr"><is><t>Date (WIB)</t></is></c>
      <c r="C1" s="1" t="inlineStr"><is><t>Operator</t></is></c>
      <c r="D1" s="1" t="inlineStr"><is><t>Type</t></is></c>
      <c r="E1" s="1" t="inlineStr"><is><t>Amount (IDR)</t></is></c>
      <c r="F1" s="1" t="inlineStr"><is><t>Category</t></is></c>
      <c r="G1" s="1" t="inlineStr"><is><t>Items / Services</t></is></c>
      <c r="H1" s="1" t="inlineStr"><is><t>Note</t></is></c>
      <c r="I1" s="1" t="inlineStr"><is><t>Status</t></is></c>
    </row>`)

	var totalIncomeCents, totalExpenseCents int64
	rowIdx := 2
	for _, entry := range entries {
		var status string
		if entry.ReversalOfID.Valid {
			status = fmt.Sprintf("Reversal of #%d", entry.ReversalOfID.Int64)
		} else if entry.ReversedByID.Valid {
			status = fmt.Sprintf("Reversed by #%d", entry.ReversedByID.Int64)
		} else {
			status = "Active"
		}

		if entry.Kind == "income" {
			totalIncomeCents += entry.AmountCents
		} else {
			totalExpenseCents += entry.AmountCents
		}

		var itemsSummary []string
		for _, item := range entry.Items {
			if item.ItemName != "" {
				itemsSummary = append(itemsSummary, fmt.Sprintf("%s: %s (%s)", item.Category, item.ItemName, transactionstore.FormatCents(item.AmountCents)))
			} else {
				itemsSummary = append(itemsSummary, fmt.Sprintf("%s (%s)", item.Category, transactionstore.FormatCents(item.AmountCents)))
			}
		}
		itemsStr := strings.Join(itemsSummary, "; ")

		amountDecimal := float64(entry.AmountCents) / 100.0

		sheetSb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="A%d"><v>%d</v></c>
      <c r="B%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="C%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="D%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="E%d" s="2"><v>%.2f</v></c>
      <c r="F%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="G%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="H%d" t="inlineStr"><is><t>%s</t></is></c>
      <c r="I%d" t="inlineStr"><is><t>%s</t></is></c>
    </row>`,
			rowIdx,
			rowIdx, entry.ID,
			rowIdx, xmlEscape(entry.OccurredAt.In(loc).Format("2006-01-02 15:04:05")),
			rowIdx, xmlEscape(entry.OperatorName),
			rowIdx, xmlEscape(entry.Kind),
			rowIdx, amountDecimal,
			rowIdx, xmlEscape(entry.Category),
			rowIdx, xmlEscape(itemsStr),
			rowIdx, xmlEscape(entry.Note),
			rowIdx, xmlEscape(status),
		))
		rowIdx++
	}

	// Summary rows
	sheetSb.WriteString(fmt.Sprintf(`
    <row r="%d">
      <c r="D%d" s="3" t="inlineStr"><is><t>Total Income</t></is></c>
      <c r="E%d" s="3"><v>%.2f</v></c>
    </row>
    <row r="%d">
      <c r="D%d" s="3" t="inlineStr"><is><t>Total Expense</t></is></c>
      <c r="E%d" s="3"><v>%.2f</v></c>
    </row>
    <row r="%d">
      <c r="D%d" s="3" t="inlineStr"><is><t>Net Balance</t></is></c>
      <c r="E%d" s="3"><v>%.2f</v></c>
    </row>`,
		rowIdx, rowIdx, rowIdx, float64(totalIncomeCents)/100.0,
		rowIdx+1, rowIdx+1, rowIdx+1, float64(totalExpenseCents)/100.0,
		rowIdx+2, rowIdx+2, rowIdx+2, float64(totalIncomeCents-totalExpenseCents)/100.0,
	))

	sheetSb.WriteString(`
  </sheetData>
</worksheet>`)

	_, _ = io.WriteString(w, sheetSb.String())

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Server) exportExcel(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	from, to, _, err := s.reportRange(r, u, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entries, err := s.transactions.All(r.Context(), u.ID, true, from, to)
	if err != nil {
		http.Error(w, "unable to export transactions", http.StatusInternalServerError)
		return
	}

	excelBytes, err := generateExcelReport(entries, s.location, from, to)
	if err != nil {
		http.Error(w, "unable to generate excel report", http.StatusInternalServerError)
		return
	}

	filename := "transactions-" + from.In(s.location).Format("20060102") + "-" + to.In(s.location).AddDate(0, 0, -1).Format("20060102") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(excelBytes)))
	_, _ = w.Write(excelBytes)
}

