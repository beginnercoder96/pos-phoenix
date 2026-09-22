package httpserver

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mekari/pos-phoenix/internal/auth"
	"github.com/mekari/pos-phoenix/internal/backoffice"
	"github.com/mekari/pos-phoenix/internal/database"
	transactionstore "github.com/mekari/pos-phoenix/internal/transaction"
	"github.com/xuri/excelize/v2"
)

func testServer(t *testing.T) (*sql.DB, http.Handler, auth.Service) {
	t.Helper()
	db, err := database.Open("file:" + t.Name() + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	handler, err := New(db, false, jakarta)
	if err != nil {
		t.Fatal(err)
	}
	return db, handler, auth.Service{DB: db, Now: func() time.Time { return time.Now().UTC() }}
}

func addUser(t *testing.T, db *sql.DB, email, role string) int64 {
	t.Helper()
	hash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO users(email,display_name,password_hash,role) VALUES(?,?,?,?)`, email, email, hash, role)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func requestAs(t *testing.T, handler http.Handler, service auth.Service, userID int64, method, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	token, err := service.CreateSession(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	var body *strings.Reader
	csrf := "01234567890123456789012345678901"
	if form == nil {
		form = url.Values{}
	}
	form.Set("csrf", csrf)
	body = strings.NewReader(form.Encode())
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	req.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestOperatorCannotAccessAdministration(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator@example.com", "operator")
	rr := requestAs(t, handler, service, operatorID, http.MethodGet, "/operators", nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestSuperadminCreatesAndDeactivatesOperator(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "admin@example.com", "superadmin")
	form := url.Values{"email": {"new@example.com"}, "display_name": {"New Operator"}, "password": {"correct horse battery staple"}}
	rr := requestAs(t, handler, service, adminID, http.MethodPost, "/operators", form)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var operatorID int64
	if err := db.QueryRow(`SELECT id FROM users WHERE email='new@example.com'`).Scan(&operatorID); err != nil {
		t.Fatal(err)
	}
	operatorToken, err := service.CreateSession(context.Background(), operatorID)
	if err != nil {
		t.Fatal(err)
	}
	rr = requestAs(t, handler, service, adminID, http.MethodPost, "/operators/"+strconv.FormatInt(operatorID, 10)+"/active", url.Values{"active": {"false"}})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("deactivate status=%d", rr.Code)
	}
	if _, err := service.UserForSession(context.Background(), operatorToken); err == nil {
		t.Fatal("deactivated operator session remains valid")
	}
}

func TestOperatorListsOnlyOwnTransactions(t *testing.T) {
	db, handler, service := testServer(t)
	first := addUser(t, db, "first@example.com", "operator")
	second := addUser(t, db, "second@example.com", "operator")
	now := time.Now().UTC()
	for _, v := range []struct {
		id       int64
		category string
	}{{first, "Mine"}, {second, "Other operator secret"}} {
		if _, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,occurred_at) VALUES(?,'income',100,?,?)`, v.id, v.category, now); err != nil {
			t.Fatal(err)
		}
	}
	rr := requestAs(t, handler, service, first, http.MethodGet, "/", nil)
	body := rr.Body.String()
	if !strings.Contains(body, "Mine") {
		t.Fatal("own transaction missing")
	}
	if strings.Contains(body, "Other operator secret") {
		t.Fatal("cross-operator transaction leaked")
	}
}

func TestBusinessDayRangeJakarta(t *testing.T) {
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	// 17:30 UTC is already the next calendar day in Jakarta.
	from, to := businessDayRange(time.Date(2026, 9, 12, 17, 30, 0, 0, time.UTC), jakarta)
	if got, want := from.UTC(), time.Date(2026, 9, 12, 17, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("from=%s, want %s", got, want)
	}
	if got, want := to.UTC(), time.Date(2026, 9, 13, 17, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("to=%s, want %s", got, want)
	}
}

func TestCustomReportRangeBounds(t *testing.T) {
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, jakarta)
	from, to, err := customReportRange("2024-09-12", "2026-09-12", now, jakarta)
	if err != nil {
		t.Fatal(err)
	}
	if from.Location() != jakarta || to.Format("2006-01-02") != "2026-09-13" {
		t.Fatalf("unexpected range %s to %s", from, to)
	}
	invalid := [][2]string{{"2024-09-11", "2026-09-12"}, {"2026-09-13", "2026-09-13"}, {"2026-09-12", "2026-09-11"}}
	for _, dates := range invalid {
		if _, _, err := customReportRange(dates[0], dates[1], now, jakarta); err == nil {
			t.Errorf("range %s to %s accepted", dates[0], dates[1])
		}
	}
}

func TestOperatorCustomReportIsIgnored(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator-report@example.com", "operator")
	rr := requestAs(t, handler, service, operatorID, http.MethodGet, "/?range=custom&from=2020-01-01&to=2020-01-02", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusOK)
	}
}

func TestCSVExportRequiresSuperadmin(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "csv-operator@example.com", "operator")
	rr := requestAs(t, handler, service, operatorID, http.MethodGet, "/reports.csv", nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestSuperadminCSVExport(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "csv-admin@example.com", "superadmin")
	operatorID := addUser(t, db, "csv-owner@example.com", "operator")
	if _, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,note,occurred_at) VALUES(?,'income',1234,'Food, drink','quoted "note"',?)`, operatorID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	rr := requestAs(t, handler, service, adminID, http.MethodGet, "/reports.csv", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("content type=%q", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"Food, drink"`) || !strings.Contains(body, `"quoted ""note"""`) || !strings.Contains(body, "12.34") {
		t.Fatalf("CSV escaping or amount missing: %q", body)
	}
}

func TestExcelExportRequiresSuperadmin(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "excel-operator@example.com", "operator")
	rr := requestAs(t, handler, service, operatorID, http.MethodGet, "/reports.xlsx", nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestSuperadminExcelExport(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "excel-admin@example.com", "superadmin")
	operatorID := addUser(t, db, "excel-owner@example.com", "operator")
	if _, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,note,occurred_at) VALUES(?,'income',125000,'Haircut Services','VIP Customer',?)`, operatorID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	rr := requestAs(t, handler, service, adminID, http.MethodGet, "/reports.xlsx", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("content type=%q", got)
	}
	excelData := rr.Body.Bytes()
	if len(excelData) == 0 {
		t.Fatal("empty excel data returned")
	}

	zr, err := zip.NewReader(bytes.NewReader(excelData), int64(len(excelData)))
	if err != nil {
		t.Fatalf("failed to open generated xlsx as zip: %v", err)
	}
	expectedFiles := map[string]bool{
		"[Content_Types].xml":        false,
		"_rels/.rels":                false,
		"xl/workbook.xml":            false,
		"xl/_rels/workbook.xml.rels": false,
		"xl/styles.xml":              false,
		"xl/worksheets/sheet1.xml":   false,
	}
	for _, f := range zr.File {
		if _, ok := expectedFiles[f.Name]; ok {
			expectedFiles[f.Name] = true
		}
	}
	for name, found := range expectedFiles {
		if !found {
			t.Fatalf("missing file in xlsx zip package: %s", name)
		}
	}
}

func TestBuildTrend(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 13, 14, 30, 0, 0, loc)

	// Test today
	todayTrend := buildTrend(nil, loc, now, "today")
	if len(todayTrend) != 8 {
		t.Fatalf("today trend points=%d, want 8", len(todayTrend))
	}
	if todayTrend[0].Label != "00:00" || todayTrend[7].Label != "21:00" {
		t.Fatalf("today trend labels start=%s end=%s", todayTrend[0].Label, todayTrend[7].Label)
	}

	// Test week
	weekTrend := buildTrend(nil, loc, now, "week")
	if len(weekTrend) != 7 {
		t.Fatalf("week trend points=%d, want 7", len(weekTrend))
	}
	if weekTrend[6].Label != "13 Sep" {
		t.Fatalf("week trend last label=%s, want 13 Sep", weekTrend[6].Label)
	}

	// Test month
	monthTrend := buildTrend(nil, loc, now, "month")
	if len(monthTrend) != 6 {
		t.Fatalf("month trend points=%d, want 6", len(monthTrend))
	}
	if monthTrend[5].Label != "Sep" {
		t.Fatalf("month trend last label=%s, want Sep", monthTrend[5].Label)
	}

	// Test populated entries
	sampleEntries := []transactionstore.Entry{
		{Kind: "income", AmountCents: 4300000, OccurredAt: now.AddDate(0, 0, -6)},
		{Kind: "expense", AmountCents: 5000000, OccurredAt: now.AddDate(0, 0, -6)},
		{Kind: "income", AmountCents: 10000000, OccurredAt: now},
	}
	populatedWeek := buildTrend(sampleEntries, loc, now, "week")
	if populatedWeek[0].IncomeHeight == 0 || populatedWeek[0].ExpenseHeight == 0 {
		t.Fatalf("expected non-zero heights for day 0: %+v", populatedWeek[0])
	}
	if populatedWeek[6].IncomeHeight == 0 {
		t.Fatalf("expected non-zero income height for today: %+v", populatedWeek[6])
	}

	// Test custom date period within September (13 days)
	customStart := time.Date(2026, 9, 1, 0, 0, 0, 0, loc)
	customEnd := time.Date(2026, 9, 13, 0, 0, 0, 0, loc)
	customTrend := buildTrend(sampleEntries, loc, now, "custom", customStart, customEnd)
	if len(customTrend) != 13 {
		t.Fatalf("custom trend points=%d, want 13", len(customTrend))
	}
	if customTrend[0].Label != "01 Sep" || customTrend[12].Label != "13 Sep" {
		t.Fatalf("custom trend labels start=%s end=%s", customTrend[0].Label, customTrend[12].Label)
	}

	// Test max days limit based on month shown (28, 29, 30, or 31 days)
	// September has 30 days -> requesting 45 days should be clamped to 30
	oversizedEnd := customStart.AddDate(0, 0, 44) // 45 days
	clampedSept := buildTrend(nil, loc, now, "custom", customStart, oversizedEnd)
	if len(clampedSept) != 30 {
		t.Fatalf("September clamped days=%d, want 30", len(clampedSept))
	}

	// February 2026 (non-leap) has 28 days -> requesting 40 days clamped to 28
	feb2026Start := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	clampedFeb2026 := buildTrend(nil, loc, now, "custom", feb2026Start, feb2026Start.AddDate(0, 0, 39))
	if len(clampedFeb2026) != 28 {
		t.Fatalf("Feb 2026 clamped days=%d, want 28", len(clampedFeb2026))
	}

	// February 2024 (leap year) has 29 days -> requesting 40 days clamped to 29
	feb2024Start := time.Date(2024, 2, 1, 0, 0, 0, 0, loc)
	clampedFeb2024 := buildTrend(nil, loc, now, "custom", feb2024Start, feb2024Start.AddDate(0, 0, 39))
	if len(clampedFeb2024) != 29 {
		t.Fatalf("Feb 2024 clamped days=%d, want 29", len(clampedFeb2024))
	}

	// August 2026 has 31 days -> requesting 40 days clamped to 31
	augStart := time.Date(2026, 8, 1, 0, 0, 0, 0, loc)
	clampedAug := buildTrend(nil, loc, now, "custom", augStart, augStart.AddDate(0, 0, 39))
	if len(clampedAug) != 31 {
		t.Fatalf("August clamped days=%d, want 31", len(clampedAug))
	}
}

func TestDashboardChartAndExcelLink(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "dashboard-admin@example.com", "superadmin")

	// 1. Check default week chart, custom selector option, calendar modal, and Excel export link
	rr := requestAs(t, handler, service, adminID, http.MethodGet, "/", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "/reports.xlsx?") {
		t.Fatalf("missing /reports.xlsx export link in dashboard: %s", body)
	}
	if !strings.Contains(body, "Download Excel (.xlsx)") {
		t.Fatalf("missing 'Download Excel (.xlsx)' label in dashboard: %s", body)
	}
	if !strings.Contains(body, "Last 7 Days") {
		t.Fatalf("missing 'Last 7 Days' option in dashboard: %s", body)
	}
	if !strings.Contains(body, "Custom Date") && !strings.Contains(body, "value=\"custom\"") {
		t.Fatalf("missing 'Custom Date' option in dashboard: %s", body)
	}
	if !strings.Contains(body, "chart-inline-calendar-panel") && !strings.Contains(body, "chart-calendar-modal") {
		t.Fatalf("missing calendar panel in dashboard: %s", body)
	}

	// 2. Check today chart
	rrToday := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=today", nil)
	if rrToday.Code != http.StatusOK {
		t.Fatalf("status=%d", rrToday.Code)
	}
	bodyToday := rrToday.Body.String()
	if !strings.Contains(bodyToday, "00:00") || !strings.Contains(bodyToday, "21:00") {
		t.Fatalf("missing hourly labels in today chart: %s", bodyToday)
	}

	// 3. Check month chart
	rrMonth := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=month", nil)
	if rrMonth.Code != http.StatusOK {
		t.Fatalf("status=%d", rrMonth.Code)
	}
	bodyMonth := rrMonth.Body.String()
	if !strings.Contains(bodyMonth, "Monthly") {
		t.Fatalf("missing Monthly chart in body: %s", bodyMonth)
	}

	// 4. Check custom date chart
	rrCustom := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=custom&chart_from=2026-09-01&chart_to=2026-09-13", nil)
	if rrCustom.Code != http.StatusOK {
		t.Fatalf("status=%d", rrCustom.Code)
	}
	bodyCustom := rrCustom.Body.String()
	idx := strings.Index(bodyCustom, "id=\"chart-period-form\"")
	if idx != -1 {
		endIdx := idx + 1000
		if endIdx > len(bodyCustom) {
			endIdx = len(bodyCustom)
		}
		t.Logf("SNIPPET CUSTOM:\n%s", bodyCustom[idx:endIdx])
	}
	if !strings.Contains(bodyCustom, "01 Sep") || !strings.Contains(bodyCustom, "13 Sep") {
		t.Fatalf("missing daily labels in custom chart: %s", bodyCustom)
	}

	// 5. Check live clock rendered on operator dashboard
	operatorID := addUser(t, db, "clock-operator@example.com", "operator")
	rrOp := requestAs(t, handler, service, operatorID, http.MethodGet, "/", nil)
	if rrOp.Code != http.StatusOK {
		t.Fatalf("status=%d", rrOp.Code)
	}
	bodyOp := rrOp.Body.String()
	if !strings.Contains(bodyOp, "data-live-clock") || !strings.Contains(bodyOp, "WIB") {
		t.Fatalf("missing live clock on operator dashboard: %s", bodyOp)
	}
	if strings.Contains(bodyOp, "Business day: WIB (Asia/Jakarta)") {
		t.Fatalf("old 'Business day: WIB (Asia/Jakarta)' still found on operator dashboard: %s", bodyOp)
	}
	if !strings.Contains(bodyOp, "chart-calendar-toggle-btn") {
		t.Fatalf("missing chart-calendar-toggle-btn on operator dashboard: %s", bodyOp)
	}
	if !strings.Contains(bodyOp, "chart-inline-calendar-panel") {
		t.Fatalf("missing chart-inline-calendar-panel on operator dashboard: %s", bodyOp)
	}
	if !strings.Contains(bodyOp, "Custom Date") && !strings.Contains(bodyOp, "value=\"custom\"") {
		t.Fatalf("missing Custom Date option on operator dashboard: %s", bodyOp)
	}

	// 6. Check operator custom chart period works properly
	rrOpCustom := requestAs(t, handler, service, operatorID, http.MethodGet, "/?chart_period=custom&chart_from=2026-09-01&chart_to=2026-09-13", nil)
	if rrOpCustom.Code != http.StatusOK {
		t.Fatalf("status=%d", rrOpCustom.Code)
	}
	bodyOpCustom := rrOpCustom.Body.String()
	if !strings.Contains(bodyOpCustom, "01 Sep") || !strings.Contains(bodyOpCustom, "13 Sep") {
		t.Fatalf("missing daily labels in operator custom chart: %s", bodyOpCustom)
	}
	if !strings.Contains(bodyOpCustom, "chart-inline-calendar-panel") {
		t.Fatalf("missing calendar panel in operator custom chart: %s", bodyOpCustom)
	}
}

func TestCustomChartCalendarMonthLimits(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "cal-admin@example.com", "superadmin")

	// 1. Default load has calendar panel with button
	rrDefault := requestAs(t, handler, service, adminID, http.MethodGet, "/", nil)
	if rrDefault.Code != http.StatusOK {
		t.Fatalf("status=%d", rrDefault.Code)
	}
	bodyDef := rrDefault.Body.String()
	if !strings.Contains(bodyDef, "id=\"chart-inline-calendar-panel\"") {
		t.Fatalf("missing inline calendar panel in dashboard: %s", bodyDef)
	}
	if !strings.Contains(bodyDef, "id=\"chart-calendar-toggle-btn\"") {
		t.Fatalf("missing calendar toggle button in dashboard: %s", bodyDef)
	}
	if !strings.Contains(bodyDef, "Calendar Picker") {
		t.Fatalf("missing 'Calendar Picker' button label: %s", bodyDef)
	}

	// 2. Custom period for September 2026 (30 days)
	rrSept := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=custom&chart_from=2026-09-01&chart_to=2026-09-30", nil)
	if rrSept.Code != http.StatusOK {
		t.Fatalf("status=%d", rrSept.Code)
	}
	bodySept := rrSept.Body.String()
	if !strings.Contains(bodySept, "Maximum selectable: 30 days based on month shown") {
		t.Fatalf("missing 30 days limit badge in Sept: %s", bodySept)
	}
	if !strings.Contains(bodySept, "data-date=\"2026-09-30\"") {
		t.Fatalf("missing Sept 30 in calendar: %s", bodySept)
	}
	if strings.Contains(bodySept, "data-date=\"2026-09-31\"") {
		t.Fatalf("Sept should not have day 31: %s", bodySept)
	}

	// 3. Custom period for February 2026 (28 days)
	rrFeb2026 := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=custom&chart_from=2026-02-01&chart_to=2026-02-28", nil)
	if rrFeb2026.Code != http.StatusOK {
		t.Fatalf("status=%d", rrFeb2026.Code)
	}
	bodyFeb2026 := rrFeb2026.Body.String()
	if !strings.Contains(bodyFeb2026, "Maximum selectable: 28 days based on month shown") {
		t.Fatalf("missing 28 days limit badge in Feb 2026: %s", bodyFeb2026)
	}
	if !strings.Contains(bodyFeb2026, "data-date=\"2026-02-28\"") {
		t.Fatalf("missing Feb 28 in calendar: %s", bodyFeb2026)
	}
	if strings.Contains(bodyFeb2026, "data-date=\"2026-02-29\"") {
		t.Fatalf("Feb 2026 should not have day 29: %s", bodyFeb2026)
	}

	// 4. Custom period for February 2024 leap year (29 days)
	rrFeb2024 := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=custom&chart_from=2024-02-01&chart_to=2024-02-29", nil)
	if rrFeb2024.Code != http.StatusOK {
		t.Fatalf("status=%d", rrFeb2024.Code)
	}
	bodyFeb2024 := rrFeb2024.Body.String()
	if !strings.Contains(bodyFeb2024, "Maximum selectable: 29 days based on month shown") {
		t.Fatalf("missing 29 days limit badge in Feb 2024: %s", bodyFeb2024)
	}
	if !strings.Contains(bodyFeb2024, "data-date=\"2024-02-29\"") {
		t.Fatalf("missing Feb 29 in calendar for leap year: %s", bodyFeb2024)
	}

	// 5. Custom period for August 2026 (31 days)
	rrAug := requestAs(t, handler, service, adminID, http.MethodGet, "/?chart_period=custom&chart_from=2026-08-01&chart_to=2026-08-31", nil)
	if rrAug.Code != http.StatusOK {
		t.Fatalf("status=%d", rrAug.Code)
	}
	bodyAug := rrAug.Body.String()
	if !strings.Contains(bodyAug, "Maximum selectable: 31 days based on month shown") {
		t.Fatalf("missing 31 days limit badge in Aug: %s", bodyAug)
	}
	if !strings.Contains(bodyAug, "data-date=\"2026-08-31\"") {
		t.Fatalf("missing Aug 31 in calendar: %s", bodyAug)
	}
}

func TestLoadingPage(t *testing.T) {
	_, handler, _ := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/loading", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "we prepare something good for you") {
		t.Fatalf("missing loading label in body: %s", body)
	}
	if !strings.Contains(body, "logopardis.jpg") {
		t.Fatalf("missing Pardis logo in loading page: %s", body)
	}
	if !strings.Contains(body, "app-body") {
		t.Fatalf("missing main app-body theme class in loading page: %s", body)
	}
}



func TestDashboardPagination(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "paged@example.com", "operator")
	now := time.Now().UTC()
	for i := 0; i < 26; i++ {
		if _, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,occurred_at) VALUES(?,'income',100,?,?)`, operatorID, "Entry "+strconv.Itoa(i), now.Add(-time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	rr := requestAs(t, handler, service, operatorID, http.MethodGet, "/?page=1", nil)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "26 records · page 1 of 2") || !strings.Contains(rr.Body.String(), "Next") {
		t.Fatalf("first page missing pagination: status=%d body=%s", rr.Code, rr.Body.String())
	}
	rr = requestAs(t, handler, service, operatorID, http.MethodGet, "/?page=2", nil)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "page 2 of 2") || !strings.Contains(rr.Body.String(), "Previous") {
		t.Fatalf("second page missing pagination: status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSuperadminReversesTransactionWithAudit(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "admin-reverse@example.com", "superadmin")
	operatorID := addUser(t, db, "operator-reverse@example.com", "operator")
	result, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,note,occurred_at) VALUES(?,'income',12550,'Sale','original',?)`, operatorID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	originalID, _ := result.LastInsertId()
	path := "/transactions/" + strconv.FormatInt(originalID, 10) + "/reverse"
	rr := requestAs(t, handler, service, adminID, http.MethodPost, path, url.Values{"reason": {"Customer refund"}})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var reversalID, ownerID, amount, reversalOf int64
	var kind, details string
	if err := db.QueryRow(`SELECT id,operator_id,kind,amount_cents,reversal_of_id FROM transactions WHERE reversal_of_id=?`, originalID).Scan(&reversalID, &ownerID, &kind, &amount, &reversalOf); err != nil {
		t.Fatal(err)
	}
	if ownerID != operatorID || kind != "expense" || amount != 12550 || reversalOf != originalID {
		t.Fatalf("invalid reversal: owner=%d kind=%s amount=%d reversal_of=%d", ownerID, kind, amount, reversalOf)
	}
	var actorID, entityID int64
	if err := db.QueryRow(`SELECT actor_id,entity_id,details FROM audit_events WHERE action='transaction.reversed'`).Scan(&actorID, &entityID, &details); err != nil {
		t.Fatal(err)
	}
	if actorID != adminID || entityID != originalID || !strings.Contains(details, "Customer refund") || !strings.Contains(details, strconv.FormatInt(reversalID, 10)) {
		t.Fatalf("invalid audit: actor=%d entity=%d details=%q", actorID, entityID, details)
	}
	rr = requestAs(t, handler, service, adminID, http.MethodPost, path, url.Values{"reason": {"Duplicate attempt"}})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("duplicate status=%d", rr.Code)
	}
}

func TestOperatorCannotReverseTransaction(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator-no-reverse@example.com", "operator")
	result, err := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,occurred_at) VALUES(?,'expense',500,'Stock',?)`, operatorID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	rr := requestAs(t, handler, service, operatorID, http.MethodPost, "/transactions/"+strconv.FormatInt(id, 10)+"/reverse", url.Values{"reason": {"Not allowed"}})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE reversal_of_id=?`, id).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("operator created %d reversal(s)", count)
	}
}

func TestReversalRequiresCSRFAndReason(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "admin-validation@example.com", "superadmin")
	result, _ := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,occurred_at) VALUES(?,'income',100,'Sale',?)`, adminID, time.Now().UTC())
	id, _ := result.LastInsertId()
	token, _ := service.CreateSession(context.Background(), adminID)
	req := httptest.NewRequest(http.MethodPost, "/transactions/"+strconv.FormatInt(id, 10)+"/reverse", strings.NewReader("reason=Valid+reason"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("missing CSRF status=%d", rr.Code)
	}
	rr = requestAs(t, handler, service, adminID, http.MethodPost, "/transactions/"+strconv.FormatInt(id, 10)+"/reverse", url.Values{"reason": {""}})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("missing reason status=%d", rr.Code)
	}
}

func TestFinancialAndAuditRecordsAreImmutable(t *testing.T) {
	db, _, _ := testServer(t)
	userID := addUser(t, db, "immutable@example.com", "operator")
	result, _ := db.Exec(`INSERT INTO transactions(operator_id,kind,amount_cents,category,occurred_at) VALUES(?,'income',100,'Sale',?)`, userID, time.Now().UTC())
	id, _ := result.LastInsertId()
	if _, err := db.Exec(`UPDATE transactions SET amount_cents=1 WHERE id=?`, id); err == nil {
		t.Fatal("transaction update succeeded")
	}
	if _, err := db.Exec(`DELETE FROM transactions WHERE id=?`, id); err == nil {
		t.Fatal("transaction delete succeeded")
	}
	result, err := db.Exec(`INSERT INTO audit_events(actor_id,action,entity_type,entity_id,details) VALUES(?,'test','transaction',?,'')`, userID, id)
	if err != nil {
		t.Fatal(err)
	}
	auditID, _ := result.LastInsertId()
	if _, err := db.Exec(`DELETE FROM audit_events WHERE id=?`, auditID); err == nil {
		t.Fatal("audit delete succeeded")
	}

	itemRes, err := db.Exec(`INSERT INTO transaction_items(transaction_id,category,item_name,amount_cents) VALUES(?,'HAIRCUT SERVICE','Haircut',4000000)`, id)
	if err != nil {
		t.Fatal(err)
	}
	itemID, _ := itemRes.LastInsertId()
	if _, err := db.Exec(`UPDATE transaction_items SET amount_cents=1 WHERE id=?`, itemID); err == nil {
		t.Fatal("transaction_items update succeeded")
	}
	if _, err := db.Exec(`DELETE FROM transaction_items WHERE id=?`, itemID); err == nil {
		t.Fatal("transaction_items delete succeeded")
	}
}

func TestCreateTransactionWithMultipleCategories(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "barber@example.com", "operator")

	today := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02")
	form := url.Values{
		"kind":        {"income"},
		"date":        {today},
		"note":        {"Customer Budi: Haircut, Vitamin, Pomade"},
		"category[]":  {"HAIRCUT SERVICE", "ADD ON", "PRODUCT"},
		"item_name[]": {"Haircut", "Vitamin", "Death Waterbased"},
		"amount[]":    {"40000", "3000", "120000"},
	}

	rr := requestAs(t, handler, service, operatorID, http.MethodPost, "/transactions", form)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}

	var txID int64
	var totalCents int64
	var category string
	if err := db.QueryRow(`SELECT id, amount_cents, category FROM transactions WHERE operator_id=?`, operatorID).Scan(&txID, &totalCents, &category); err != nil {
		t.Fatal(err)
	}

	// 40,000 + 3,000 + 120,000 = 163,000 IDR => 16,300,000 cents
	expectedCents := int64(163000 * 100)
	if totalCents != expectedCents {
		t.Fatalf("totalCents=%d, want %d", totalCents, expectedCents)
	}

	if !strings.Contains(category, "HAIRCUT SERVICE") || !strings.Contains(category, "ADD ON") || !strings.Contains(category, "PRODUCT") {
		t.Fatalf("category summary=%q missing expected categories", category)
	}

	rows, err := db.Query(`SELECT category, item_name, amount_cents FROM transaction_items WHERE transaction_id=? ORDER BY id ASC`, txID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type itemRow struct {
		cat, name string
		cents     int64
	}
	var items []itemRow
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(&it.cat, &it.name, &it.cents); err != nil {
			t.Fatal(err)
		}
		items = append(items, it)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].cat != "HAIRCUT SERVICE" || items[0].name != "Haircut" || items[0].cents != 4000000 {
		t.Fatalf("unexpected item 0: %+v", items[0])
	}
	if items[1].cat != "ADD ON" || items[1].name != "Vitamin" || items[1].cents != 300000 {
		t.Fatalf("unexpected item 1: %+v", items[1])
	}
	if items[2].cat != "PRODUCT" || items[2].name != "Death Waterbased" || items[2].cents != 12000000 {
		t.Fatalf("unexpected item 2: %+v", items[2])
	}

	// Verify dashboard renders the items and total
	rr = requestAs(t, handler, service, operatorID, http.MethodGet, "/", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("dashboard status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "HAIRCUT SERVICE: Haircut") {
		t.Fatalf("dashboard missing Haircut badge: %s", body)
	}
	if !strings.Contains(body, "ADD ON: Vitamin") {
		t.Fatalf("dashboard missing Vitamin badge: %s", body)
	}
	if !strings.Contains(body, "PRODUCT: Death Waterbased") {
		t.Fatalf("dashboard missing Product badge: %s", body)
	}
	if !strings.Contains(body, "Rp 163.000,00") {
		t.Fatalf("dashboard missing formatted total: %s", body)
	}
}

func TestCatalogProductCommissionIntegration(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "barber2@example.com", "operator")

	// 1. Insert product into catalog_items with commission
	_, err := db.Exec(`INSERT INTO catalog_items(name, category, item_type, price_cents, commission_amount, is_active)
		VALUES('Hair Tonic Special', 'Product', 'PRODUCT', 5000000, 500000, 1)`)
	if err != nil {
		t.Fatalf("failed to insert catalog item: %v", err)
	}

	today := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02")
	form := url.Values{
		"kind":        {"income"},
		"date":        {today},
		"note":        {"Customer Toni: Hair Tonic Special"},
		"category[]":  {"Product"},
		"item_name[]": {"Hair Tonic Special"},
		"amount[]":    {"50000"},
	}

	rr := requestAs(t, handler, service, operatorID, http.MethodPost, "/transactions", form)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}

	var itType string
	var barberID, commEarned int64
	err = db.QueryRow(`SELECT item_type, barber_id, commission_earned FROM transaction_items WHERE item_name='Hair Tonic Special'`).
		Scan(&itType, &barberID, &commEarned)
	if err != nil {
		t.Fatalf("failed to query transaction item: %v", err)
	}

	if itType != "PRODUCT" {
		t.Errorf("item_type=%q, want 'PRODUCT'", itType)
	}
	if barberID != operatorID {
		t.Errorf("barber_id=%d, want %d", barberID, operatorID)
	}
	if commEarned != 500000 {
		t.Errorf("commission_earned=%d, want 500000 (Rp 5.000)", commEarned)
	}
}

func TestReversalOfMultiCategoryTransaction(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "admin-multi@example.com", "superadmin")
	operatorID := addUser(t, db, "barber-multi@example.com", "operator")

	today := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02")
	form := url.Values{
		"kind":        {"income"},
		"date":        {today},
		"note":        {"Multi-item service"},
		"category[]":  {"HAIRCUT SERVICE", "ADD ON"},
		"item_name[]": {"Haircut", "Beard Trim"},
		"amount[]":    {"40000", "15000"},
	}

	rr := requestAs(t, handler, service, operatorID, http.MethodPost, "/transactions", form)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status=%d", rr.Code)
	}

	var txID int64
	if err := db.QueryRow(`SELECT id FROM transactions WHERE operator_id=?`, operatorID).Scan(&txID); err != nil {
		t.Fatal(err)
	}

	reversePath := "/transactions/" + strconv.FormatInt(txID, 10) + "/reverse"
	rr = requestAs(t, handler, service, adminID, http.MethodPost, reversePath, url.Values{"reason": {"Customer requested refund"}})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("reversal status=%d body=%s", rr.Code, rr.Body.String())
	}

	var reversalID, revAmount int64
	var revKind string
	if err := db.QueryRow(`SELECT id, amount_cents, kind FROM transactions WHERE reversal_of_id=?`, txID).Scan(&reversalID, &revAmount, &revKind); err != nil {
		t.Fatal(err)
	}
	if revKind != "expense" || revAmount != 5500000 {
		t.Fatalf("reversal mismatch: kind=%s amount=%d", revKind, revAmount)
	}

	// Verify items were replicated to reversal transaction
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM transaction_items WHERE transaction_id=?`, reversalID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 items for reversal transaction, got %d", count)
	}
}

func TestIndonesianDashboardCardLabels(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "admin-id-labels@example.com", "superadmin")

	token, err := service.CreateSession(context.Background(), adminID)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	req.AddCookie(&http.Cookie{Name: "language", Value: "id"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()

	// 1. Top Cashflow report card (superadmin)
	expectedReportLabels := []string{
		"Laporan arus kas",
		"Hari ini",
		"Bulan ini",
		"Dari",
		"Sampai",
		"Terapkan",
		"Unduh Excel (.xlsx)",
	}
	for _, label := range expectedReportLabels {
		if !strings.Contains(body, label) {
			t.Errorf("missing Indonesian report card label %q in body", label)
		}
	}

	// 2. Stat cards
	expectedStatLabels := []string{
		"Pemasukan",
		"Pengeluaran",
		"Saldo",
	}
	for _, label := range expectedStatLabels {
		if !strings.Contains(body, label) {
			t.Errorf("missing Indonesian stat card label %q in body", label)
		}
	}

	// 3. Cash-flow trend card
	expectedTrendLabels := []string{
		"Tren arus kas",
		"Pemasukan dan pengeluaran per periode",
		"Hari ini (Per jam)",
		"7 Hari Terakhir",
		"Bulanan",
		"Tanggal Khusus",
		"Pilih Kalender",
		"Sebulan Penuh",
		"Awal Bulan Hingga Kini",
		"Tutup kalender",
		"Bulan sebelumnya",
		"Bulan berikutnya",
		"Maksimal dapat dipilih",
		"hari berdasarkan bulan yang ditampilkan",
		"Terapkan Tanggal Kalender",
		"Arahkan kursor ke grafik untuk melihat periode",
		"Bersih:",
	}
	for _, label := range expectedTrendLabels {
		if !strings.Contains(body, label) {
			t.Errorf("missing Indonesian trend card label %q in body", label)
		}
	}

	// Weekday headers in Indonesian
	for _, day := range []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"} {
		if !strings.Contains(body, day) {
			t.Errorf("missing Indonesian weekday header %q in body", day)
		}
	}

	// 4. New transaction card
	expectedTxFormLabels := []string{
		"Transaksi baru",
		"Catatan opsional / nama pelanggan",
		"Daftar Harga Barbershop",
		"-- Pilih kategori --",
		"Lainnya / Kustom",
		"-- Pilih layanan / item --",
		"Tambah Kategori / Item",
		"Simpan transaksi",
	}
	for _, label := range expectedTxFormLabels {
		if !strings.Contains(body, label) {
			t.Errorf("missing Indonesian transaction form label %q in body", label)
		}
	}

	// 5. Transactions feed card
	expectedTxFeedLabels := []string{
		"Transaksi",
		"catatan",
		"halaman",
		"dari",
	}
	for _, label := range expectedTxFeedLabels {
		if !strings.Contains(body, label) {
			t.Errorf("missing Indonesian transactions feed label %q in body", label)
		}
	}
}

func TestPayrollSlipViewAndDownload(t *testing.T) {
	db, handler, authSvc := testServer(t)
	superadminID := addUser(t, db, "superadmin@example.com", "superadmin")
	empID := addUser(t, db, "barber@example.com", "operator")

	// Assign branch
	var branchID int64
	_ = db.QueryRow(`SELECT id FROM branches WHERE code='KLASEMAN'`).Scan(&branchID)
	if branchID > 0 {
		_, _ = db.Exec(`UPDATE users SET branch_id=?, staff_type='barberman' WHERE id=?`, branchID, empID)
	}

	token, err := authSvc.CreateSession(context.Background(), superadminID)
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: "session", Value: token}

	// 1. Single Employee Slip HTML View
	req := httptest.NewRequest("GET", "/backoffice/payroll/slip?branch=KLASEMAN&period=2026-01&employee_id="+strconv.FormatInt(empID, 10), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	html := rec.Body.String()
	if !strings.Contains(html, "SLIP GAJI") {
		t.Errorf("expected HTML to contain 'SLIP GAJI'")
	}
	if !strings.Contains(html, "TAKE HOME PAY") {
		t.Errorf("expected HTML to contain 'TAKE HOME PAY'")
	}
	if !strings.Contains(html, "Pendapatan") {
		t.Errorf("expected HTML to contain 'Pendapatan'")
	}

	// 2. Single Employee Slip Excel Download
	reqExcel := httptest.NewRequest("GET", "/backoffice/payroll/slip?branch=KLASEMAN&period=2026-01&employee_id="+strconv.FormatInt(empID, 10)+"&format=excel", nil)
	reqExcel.AddCookie(cookie)
	recExcel := httptest.NewRecorder()
	handler.ServeHTTP(recExcel, reqExcel)

	if recExcel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for excel, got %d", recExcel.Code)
	}
	if !strings.Contains(recExcel.Header().Get("Content-Type"), "spreadsheetml") {
		t.Errorf("expected spreadsheetml content type, got %s", recExcel.Header().Get("Content-Type"))
	}

	// 3. Bulk Slip-All HTML View
	reqAll := httptest.NewRequest("GET", "/backoffice/payroll/slip-all?branch=KLASEMAN&period=2026-01", nil)
	reqAll.AddCookie(cookie)
	recAll := httptest.NewRecorder()
	handler.ServeHTTP(recAll, reqAll)

	if recAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for slip-all HTML, got %d: %s", recAll.Code, recAll.Body.String())
	}
	if !strings.Contains(recAll.Body.String(), "SLIP GAJI") {
		t.Errorf("expected slip-all HTML to contain 'SLIP GAJI'")
	}

	// 4. Bulk Slip-All Excel Download
	reqAllExcel := httptest.NewRequest("GET", "/backoffice/payroll/slip-all?branch=KLASEMAN&period=2026-01&format=excel", nil)
	reqAllExcel.AddCookie(cookie)
	recAllExcel := httptest.NewRecorder()
	handler.ServeHTTP(recAllExcel, reqAllExcel)

	if recAllExcel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for slip-all excel, got %d", recAllExcel.Code)
	}
	if !strings.Contains(recAllExcel.Header().Get("Content-Type"), "spreadsheetml") {
		t.Errorf("expected spreadsheetml content type, got %s", recAllExcel.Header().Get("Content-Type"))
	}
}

func TestOperatorBankCredentialsAndSlipColors(t *testing.T) {
	db, handler, authSvc := testServer(t)
	adminID := addUser(t, db, "admin@example.com", "superadmin")
	empID := addUser(t, db, "barber1@example.com", "operator")

	var branchID int64
	_ = db.QueryRow(`SELECT id FROM branches LIMIT 1`).Scan(&branchID)

	token, _ := authSvc.CreateSession(context.Background(), adminID)
	cookie := &http.Cookie{Name: "session", Value: token}

	// 1. Update operator credentials via POST /operators/{id}/credentials
	form := url.Values{
		"csrf":                {"test-csrf"},
		"display_name":        {"Budi Barberman"},
		"phone_number":        {"081234567890"},
		"bank_name":           {"BCA"},
		"bank_account_number": {"1234567890"},
		"branch_id":           {strconv.FormatInt(branchID, 10)},
		"staff_type":          {"barberman"},
	}
	reqUpdate := httptest.NewRequest("POST", "/operators/"+strconv.FormatInt(empID, 10)+"/credentials", strings.NewReader(form.Encode()))
	reqUpdate.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqUpdate.AddCookie(cookie)
	reqUpdate.AddCookie(&http.Cookie{Name: "csrf", Value: "test-csrf"})
	recUpdate := httptest.NewRecorder()
	handler.ServeHTTP(recUpdate, reqUpdate)

	if recUpdate.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther, got %d: %s", recUpdate.Code, recUpdate.Body.String())
	}

	// Verify in DB
	var bName, bAcc, phone string
	_ = db.QueryRow(`SELECT bank_name, bank_account_number, phone_number FROM users WHERE id=?`, empID).Scan(&bName, &bAcc, &phone)
	if bName != "BCA" || bAcc != "1234567890" || phone != "081234567890" {
		t.Fatalf("credentials not saved properly in DB: bank=%s acc=%s phone=%s", bName, bAcc, phone)
	}

	// 2. Fetch Slip and verify dynamic bank info and red/green colors
	var branchCode string
	_ = db.QueryRow(`SELECT code FROM branches WHERE id=?`, branchID).Scan(&branchCode)

	reqSlip := httptest.NewRequest("GET", fmt.Sprintf("/backoffice/payroll/slip?branch=%s&period=2026-01&employee_id=%d", branchCode, empID), nil)
	reqSlip.AddCookie(cookie)
	recSlip := httptest.NewRecorder()
	handler.ServeHTTP(recSlip, reqSlip)

	if recSlip.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for slip, got %d", recSlip.Code)
	}
	slipHTML := recSlip.Body.String()

	// Verify dynamic bank credentials rendered
	if !strings.Contains(slipHTML, "1234567890 (BCA)") {
		t.Errorf("expected slip to contain '1234567890 (BCA)', got: %s", slipHTML)
	}
	// Verify red header color (#dc2626)
	if !strings.Contains(slipHTML, "#dc2626") {
		t.Errorf("expected slip to contain red header color '#dc2626'")
	}
	// Verify green total color (#86efac)
	if !strings.Contains(slipHTML, "#86efac") {
		t.Errorf("expected slip to contain green total color '#86efac'")
	}
	// Verify deductions font is solid black
	if !strings.Contains(slipHTML, `<span class="text-black">Potongan Telat</span>`) {
		t.Errorf("expected slip to contain black deductions text 'Potongan Telat'")
	}

	// 3. Verify that distinct employees have different seeded bank accounts in users table
	var k1Bank, k1Acc, k2Bank, k2Acc string
	_ = db.QueryRow(`SELECT bank_name, bank_account_number FROM users WHERE email='karyawan1.klaseman@pardis.com'`).Scan(&k1Bank, &k1Acc)
	_ = db.QueryRow(`SELECT bank_name, bank_account_number FROM users WHERE email='karyawan2.klaseman@pardis.com'`).Scan(&k2Bank, &k2Acc)
	if k1Bank != "BCA" || k2Bank != "Mandiri" || k1Acc == k2Acc {
		t.Errorf("expected distinct bank info for sample employees, got K1=%s %s, K2=%s %s", k1Bank, k1Acc, k2Bank, k2Acc)
	}

	// 4. Verify unconfigured employee shows "-" (strip)
	emptyEmpID := addUser(t, db, "empty-bank@example.com", "operator")
	reqEmptySlip := httptest.NewRequest("GET", fmt.Sprintf("/backoffice/payroll/slip?branch=%s&period=2026-01&employee_id=%d", branchCode, emptyEmpID), nil)
	reqEmptySlip.AddCookie(cookie)
	recEmptySlip := httptest.NewRecorder()
	handler.ServeHTTP(recEmptySlip, reqEmptySlip)
	if !strings.Contains(recEmptySlip.Body.String(), "-") {
		t.Errorf("expected empty employee slip to show '-' for bank info")
	}
}

func TestUsernameLoginAndForgotPasswordHTTP(t *testing.T) {
	db, handler, service := testServer(t)

	// Seed admin user
	adminID := addUser(t, db, "admin@example.com", "superadmin")
	_, err := db.Exec(`UPDATE users SET username='admin' WHERE id=?`, adminID)
	if err != nil {
		t.Fatal(err)
	}

	csrf := "01234567890123456789012345678901"

	// 1. Test Login with Username (not email)
	loginForm := url.Values{
		"csrf":     {csrf},
		"username": {"admin"},
		"password": {"correct horse battery staple"},
	}
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginForm.Encode()))
	reqLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogin.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recLogin := httptest.NewRecorder()
	handler.ServeHTTP(recLogin, reqLogin)

	if recLogin.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect on username login, got %d, body: %s", recLogin.Code, recLogin.Body.String())
	}
	if recLogin.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %s", recLogin.Header().Get("Location"))
	}

	// 1b. Test Login with Email entered into the form
	loginFormEmail := url.Values{
		"csrf":     {csrf},
		"username": {"admin@example.com"},
		"password": {"correct horse battery staple"},
	}
	reqLoginEmail := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginFormEmail.Encode()))
	reqLoginEmail.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLoginEmail.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recLoginEmail := httptest.NewRecorder()
	handler.ServeHTTP(recLoginEmail, reqLoginEmail)

	if recLoginEmail.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect on email login, got %d, body: %s", recLoginEmail.Code, recLoginEmail.Body.String())
	}
	if recLoginEmail.Header().Get("Location") != "/" {
		t.Fatalf("expected redirect to '/', got %s", recLoginEmail.Header().Get("Location"))
	}

	// 1c. Test Login with legacy email parameter
	loginFormLegacy := url.Values{
		"csrf":     {csrf},
		"email":    {"admin@example.com"},
		"password": {"correct horse battery staple"},
	}
	reqLoginLegacy := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginFormLegacy.Encode()))
	reqLoginLegacy.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLoginLegacy.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recLoginLegacy := httptest.NewRecorder()
	handler.ServeHTTP(recLoginLegacy, reqLoginLegacy)

	if recLoginLegacy.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect on legacy email login, got %d, body: %s", recLoginLegacy.Code, recLoginLegacy.Body.String())
	}

	// 2. Test GET /forgot-password
	reqForgot := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	recForgot := httptest.NewRecorder()
	handler.ServeHTTP(recForgot, reqForgot)
	if recForgot.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /forgot-password, got %d", recForgot.Code)
	}

	// 3. Test POST /forgot-password
	forgotForm := url.Values{
		"csrf":  {csrf},
		"email": {"admin@example.com"},
	}
	reqForgotPost := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(forgotForm.Encode()))
	reqForgotPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqForgotPost.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recForgotPost := httptest.NewRecorder()
	handler.ServeHTTP(recForgotPost, reqForgotPost)

	if recForgotPost.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for POST /forgot-password, got %d, body: %s", recForgotPost.Code, recForgotPost.Body.String())
	}
	forgotHTML := recForgotPost.Body.String()
	if !strings.Contains(forgotHTML, "/reset-password?token=") {
		t.Fatalf("expected forgot-password response to contain mock reset link, got: %s", forgotHTML)
	}

	// Extract token from mock link in HTML
	tokenIdx := strings.Index(forgotHTML, "/reset-password?token=")
	if tokenIdx < 0 {
		t.Fatal("token not found in response")
	}
	tokenSub := forgotHTML[tokenIdx+len("/reset-password?token="):]
	endToken := strings.IndexAny(tokenSub, `"'> `)
	token := tokenSub[:endToken]

	// 4. Test GET /reset-password?token=...
	reqResetGet := httptest.NewRequest(http.MethodGet, "/reset-password?token="+token, nil)
	recResetGet := httptest.NewRecorder()
	handler.ServeHTTP(recResetGet, reqResetGet)
	if recResetGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /reset-password, got %d", recResetGet.Code)
	}

	// 5. Test POST /reset-password
	newPassword := "brandNewSecurePassword999!"
	resetForm := url.Values{
		"csrf":             {csrf},
		"token":            {token},
		"password":         {newPassword},
		"password_confirm": {newPassword},
	}
	reqResetPost := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(resetForm.Encode()))
	reqResetPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResetPost.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recResetPost := httptest.NewRecorder()
	handler.ServeHTTP(recResetPost, reqResetPost)

	if recResetPost.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for POST /reset-password, got %d, body: %s", recResetPost.Code, recResetPost.Body.String())
	}
	if !strings.Contains(recResetPost.Body.String(), "login") && !strings.Contains(recResetPost.Body.String(), "signIn") {
		t.Fatalf("expected login page returned upon successful reset")
	}

	// 6. Test Login with new password and username
	loginFormNew := url.Values{
		"csrf":     {csrf},
		"username": {"admin"},
		"password": {newPassword},
	}
	reqLoginNew := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginFormNew.Encode()))
	reqLoginNew.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLoginNew.AddCookie(&http.Cookie{Name: "csrf", Value: csrf})
	recLoginNew := httptest.NewRecorder()
	handler.ServeHTTP(recLoginNew, reqLoginNew)

	if recLoginNew.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect on login with new password, got %d", recLoginNew.Code)
	}

	// 7. Test Create Operator with Username
	createOpForm := url.Values{
		"csrf":         {csrf},
		"username":     {"barber_anto"},
		"email":        {"anto@example.com"},
		"display_name": {"Anto Barber"},
		"password":     {"secureanto12345"},
		"staff_type":   {"barberman"},
	}
	recCreateOp := requestAs(t, handler, service, adminID, http.MethodPost, "/operators", createOpForm)
	if recCreateOp.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 on create operator, got %d, body: %s", recCreateOp.Code, recCreateOp.Body.String())
	}

	var savedUname, savedEmail string
	err = db.QueryRow(`SELECT username, email FROM users WHERE username='barber_anto'`).Scan(&savedUname, &savedEmail)
	if err != nil {
		t.Fatalf("operator barber_anto was not saved with username: %v", err)
	}
	if savedUname != "barber_anto" || savedEmail != "anto@example.com" {
		t.Fatalf("unexpected operator data: uname=%s, email=%s", savedUname, savedEmail)
	}
}

func TestCreateTransactionWithBundlingAndDiscount(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "bnd-admin@example.com", "superadmin")

	// 1. Insert an active BUNDLE into discounts_and_bundles
	res, err := db.Exec(`INSERT INTO discounts_and_bundles(code, name, type, value, service_allocation_ratio, product_allocation_ratio, is_active)
		VALUES ('BND-01', 'Paket Ganteng', 'BUNDLE', 7500000, 0.6, 0.4, 1)`)
	if err != nil {
		t.Fatalf("failed to insert bundle: %v", err)
	}
	bundleID, _ := res.LastInsertId()

	csrf := "01234567890123456789012345678901"
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	today := time.Now().In(jakarta).Format("2006-01-02")

	// 2. Submit transaction with bundling item
	form := url.Values{
		"csrf":        {csrf},
		"date":        {today},
		"kind":        {"income"},
		"category[]":  {"Bundling"},
		"item_name[]": {"Paket Ganteng"},
		"amount[]":    {"75000"},
		"bundle_id[]": {strconv.FormatInt(bundleID, 10)},
		"note":        {"Test bundling tx"},
	}

	rec := requestAs(t, handler, service, adminID, http.MethodPost, "/transactions", form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther on bundling transaction, got %d, body: %s", rec.Code, rec.Body.String())
	}

	// 3. Verify transaction_items record
	var savedItemType string
	var savedBundleID int64
	var savedAmountCents int64
	err = db.QueryRow(`SELECT item_type, bundle_id, amount_cents FROM transaction_items WHERE category='Bundling' ORDER BY id DESC LIMIT 1`).
		Scan(&savedItemType, &savedBundleID, &savedAmountCents)
	if err != nil {
		t.Fatalf("failed to query saved transaction item: %v", err)
	}
	if savedBundleID != bundleID {
		t.Fatalf("expected bundle_id=%d, got %d", bundleID, savedBundleID)
	}
	if savedItemType != "SERVICE" {
		t.Fatalf("expected item_type='SERVICE', got %s", savedItemType)
	}
	if savedAmountCents != 7500000 {
		t.Fatalf("expected amount_cents=7500000, got %d", savedAmountCents)
	}

	// 4. Test order-level discount
	discForm := url.Values{
		"csrf":                  {csrf},
		"date":                  {today},
		"kind":                  {"income"},
		"category[]":            {"Haircut"},
		"item_name[]":           {"Haircut Regular"},
		"amount[]":              {"50000"},
		"order_discount_amount": {"10000"},
		"note":                  {"Test order discount tx"},
	}

	recDisc := requestAs(t, handler, service, adminID, http.MethodPost, "/transactions", discForm)
	if recDisc.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther on discount transaction, got %d, body: %s", recDisc.Code, recDisc.Body.String())
	}

	var txAmountCents int64
	err = db.QueryRow(`SELECT amount_cents FROM transactions WHERE note='Test order discount tx' ORDER BY id DESC LIMIT 1`).
		Scan(&txAmountCents)
	if err != nil {
		t.Fatalf("failed to query discount transaction: %v", err)
	}
	if txAmountCents != 4000000 { // 50.000 - 10.000 = 40.000 (in cents = 4000000)
		t.Fatalf("expected net amount_cents=4000000, got %d", txAmountCents)
	}

	// 5. Verify deleting a bundle that was used in an existing transaction item succeeds
	delRec := requestAs(t, handler, service, adminID, http.MethodPost, fmt.Sprintf("/backoffice/discounts/%d/delete", bundleID), url.Values{
		"csrf": {csrf},
	})
	if delRec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther when deleting used bundle, got %d body: %s", delRec.Code, delRec.Body.String())
	}
	var count int
	_ = db.QueryRow(`SELECT COUNT(*) FROM discounts_and_bundles WHERE id=?`, bundleID).Scan(&count)
	if count != 0 {
		t.Fatalf("expected bundle to be deleted from discounts_and_bundles, but still exists")
	}
}

func TestCatalogItemLifecycleAndDashboardIntegration(t *testing.T) {
	db, handler, service := testServer(t)
	adminID := addUser(t, db, "cat-admin@example.com", "superadmin")
	csrf := "01234567890123456789012345678901"

	// 1. Create a new catalog item via POST /backoffice/products (using ordinary Rupiah: 75000 and 5000)
	createForm := url.Values{
		"csrf":            {csrf},
		"name":            {"Pomade Premium Oil"},
		"item_type":       {"PRODUCT"},
		"category_select": {"__NEW__"},
		"category_new":    {"Hair Care"},
		"price":           {"75000"}, // Rp 75.000 (auto-converted to 7500000 cents)
		"commission":      {"5000"},  // Rp 5.000 (auto-converted to 500000 cents)
	}

	recCreate := requestAs(t, handler, service, adminID, http.MethodPost, "/backoffice/products", createForm)
	if recCreate.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther on create product, got %d, body: %s", recCreate.Code, recCreate.Body.String())
	}

	// Verify inserted row in cents
	var itemID int64
	var itemName, itemCategory, itemType string
	var itemPrice, itemComm int64
	var isActive int
	err := db.QueryRow(`SELECT id, name, category, item_type, price_cents, commission_amount, is_active FROM catalog_items WHERE name='Pomade Premium Oil'`).
		Scan(&itemID, &itemName, &itemCategory, &itemType, &itemPrice, &itemComm, &isActive)
	if err != nil {
		t.Fatalf("failed to find created catalog item: %v", err)
	}
	if itemCategory != "Hair Care" || itemType != "PRODUCT" || itemPrice != 7500000 || itemComm != 500000 || isActive != 1 {
		t.Fatalf("unexpected item fields: cat=%s type=%s price=%d comm=%d active=%d", itemCategory, itemType, itemPrice, itemComm, isActive)
	}

	// 2. Verify it appears on Dashboard GET / (via JSON catalog script in dashboard HTML)
	recDash := requestAs(t, handler, service, adminID, http.MethodGet, "/", nil)
	if recDash.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on dashboard, got %d", recDash.Code)
	}
	dashBody := recDash.Body.String()
	if !strings.Contains(dashBody, "Pomade Premium Oil") {
		t.Fatalf("expected dashboard catalog to contain 'Pomade Premium Oil'")
	}
	if !strings.Contains(dashBody, "Hair Care") {
		t.Fatalf("expected dashboard catalog to contain category 'Hair Care'")
	}

	// 3. Edit the catalog item (change price & name using ordinary Rupiah: 80000 and 6000)
	editForm := url.Values{
		"csrf":            {csrf},
		"id":              {strconv.FormatInt(itemID, 10)},
		"name":            {"Pomade Premium Matte"},
		"item_type":       {"PRODUCT"},
		"category_select": {"Hair Care"},
		"price":           {"80000"}, // Rp 80.000
		"commission":      {"6000"},  // Rp 6.000
	}
	recEdit := requestAs(t, handler, service, adminID, http.MethodPost, "/backoffice/products", editForm)
	if recEdit.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 SeeOther on edit product, got %d, body: %s", recEdit.Code, recEdit.Body.String())
	}

	err = db.QueryRow(`SELECT name, price_cents, commission_amount FROM catalog_items WHERE id=?`, itemID).
		Scan(&itemName, &itemPrice, &itemComm)
	if err != nil || itemName != "Pomade Premium Matte" || itemPrice != 8000000 || itemComm != 600000 {
		t.Fatalf("expected updated item: name=%s price=%d comm=%d err=%v", itemName, itemPrice, itemComm, err)
	}

	// 4. Toggle Status (Active -> Inactive)
	toggleRec := requestAs(t, handler, service, adminID, http.MethodPost, fmt.Sprintf("/backoffice/products/%d/toggle", itemID), url.Values{"csrf": {csrf}})
	if toggleRec.Code != http.StatusSeeOther && toggleRec.Code != http.StatusOK {
		t.Fatalf("expected 303 or 200 on toggle, got %d", toggleRec.Code)
	}

	var statusAfterToggle int
	_ = db.QueryRow(`SELECT is_active FROM catalog_items WHERE id=?`, itemID).Scan(&statusAfterToggle)
	if statusAfterToggle != 0 {
		t.Fatalf("expected is_active=0 after toggle, got %d", statusAfterToggle)
	}

	// 5. Inactive items should NOT be included in dashboard catalog
	recDash2 := requestAs(t, handler, service, adminID, http.MethodGet, "/", nil)
	if strings.Contains(recDash2.Body.String(), "Pomade Premium Matte") {
		t.Fatalf("expected inactive item 'Pomade Premium Matte' to NOT appear in dashboard catalog")
	}

	// 6. Delete the catalog item
	delRec := requestAs(t, handler, service, adminID, http.MethodPost, fmt.Sprintf("/backoffice/products/%d/delete", itemID), url.Values{"csrf": {csrf}})
	if delRec.Code != http.StatusSeeOther && delRec.Code != http.StatusOK {
		t.Fatalf("expected 303 or 200 on delete, got %d", delRec.Code)
	}

	var countAfterDel int
	_ = db.QueryRow(`SELECT COUNT(*) FROM catalog_items WHERE id=?`, itemID).Scan(&countAfterDel)
	if countAfterDel != 0 {
		t.Fatalf("expected item to be deleted, found count=%d", countAfterDel)
	}
}

func TestBackofficeProfitSharingAuthorizationAndSave(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator_ps@example.com", "operator")
	adminID := addUser(t, db, "ipang@example.com", "superadmin")

	// 1. Operator cannot access GET /backoffice/profit-sharing
	recOpGet := requestAs(t, handler, service, operatorID, http.MethodGet, "/backoffice/profit-sharing", nil)
	if recOpGet.Code != http.StatusForbidden {
		t.Fatalf("expected operator GET to be 403, got %d", recOpGet.Code)
	}

	// 2. Operator cannot POST /backoffice/profit-sharing
	postForm := url.Values{
		"branch":           {"KLASEMAN"},
		"period":           {"2026-01"},
		"owner_percentage": {"20"},
	}
	recOpPost := requestAs(t, handler, service, operatorID, http.MethodPost, "/backoffice/profit-sharing", postForm)
	if recOpPost.Code != http.StatusForbidden {
		t.Fatalf("expected operator POST to be 403, got %d", recOpPost.Code)
	}

	// 3. Superadmin can access GET /backoffice/profit-sharing
	recAdminGet := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/profit-sharing?branch=KLASEMAN&period=2026-01", nil)
	if recAdminGet.Code != http.StatusOK {
		t.Fatalf("expected admin GET to be 200, got %d", recAdminGet.Code)
	}

	// 4. Superadmin cannot save percentage > 100%
	invalidForm := url.Values{
		"branch":              {"KLASEMAN"},
		"period":              {"2026-01"},
		"owner_percentage":    {"80"},
		"employee_id":         {"1", "2"},
		"employee_percentage": {"30", "30"}, // 80 + 30 + 30 = 140% > 100%
	}
	recInvalid := requestAs(t, handler, service, adminID, http.MethodPost, "/backoffice/profit-sharing", invalidForm)
	if recInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected >100%% to be 400 Bad Request, got %d", recInvalid.Code)
	}

	// 5. Superadmin can save valid profit sharing configuration (20% owner, 35% emp 1, 35% emp 2, 10% reserve)
	validForm := url.Values{
		"branch":              {"KLASEMAN"},
		"period":              {"2026-01"},
		"owner_percentage":    {"20"},
		"employee_id":         {"1", "2"},
		"employee_percentage": {"35", "35"},
	}
	recValid := requestAs(t, handler, service, adminID, http.MethodPost, "/backoffice/profit-sharing", validForm)
	if recValid.Code != http.StatusSeeOther && recValid.Code != http.StatusOK {
		t.Fatalf("expected 303 or 200 on valid save, got %d", recValid.Code)
	}
}

func TestBackofficeAnalytics24mAPI(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator_trend@example.com", "operator")
	adminID := addUser(t, db, "ipang_trend@example.com", "superadmin")

	// 1. Operator cannot access analytics API
	recOp := requestAs(t, handler, service, operatorID, http.MethodGet, "/backoffice/api/analytics/trend-24m", nil)
	if recOp.Code != http.StatusForbidden {
		t.Fatalf("expected operator to receive 403, got %d", recOp.Code)
	}

	// 2. Superadmin accesses all branches
	recAll := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/api/analytics/trend-24m?branch_id=all", nil)
	if recAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", recAll.Code)
	}

	var allData []backoffice.MonthlyAnalytics
	if err := json.Unmarshal(recAll.Body.Bytes(), &allData); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if len(allData) != 24 {
		t.Fatalf("expected exactly 24 months, got %d", len(allData))
	}

	// 3. Superadmin accesses Klaseman branch
	recKlaseman := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/api/analytics/trend-24m?branch_id=KLASEMAN", nil)
	if recKlaseman.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Klaseman, got %d", recKlaseman.Code)
	}
	var klasemanData []backoffice.MonthlyAnalytics
	if err := json.Unmarshal(recKlaseman.Body.Bytes(), &klasemanData); err != nil {
		t.Fatalf("failed to parse json for Klaseman: %v", err)
	}
	if len(klasemanData) != 24 {
		t.Fatalf("expected 24 months for Klaseman, got %d", len(klasemanData))
	}

	// 4. Superadmin accesses Ledok branch
	recLedok := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/api/analytics/trend-24m?branch_id=LEDOK", nil)
	if recLedok.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Ledok, got %d", recLedok.Code)
	}
	var ledokData []backoffice.MonthlyAnalytics
	if err := json.Unmarshal(recLedok.Body.Bytes(), &ledokData); err != nil {
		t.Fatalf("failed to parse json for Ledok: %v", err)
	}
	if len(ledokData) != 24 {
		t.Fatalf("expected 24 months for Ledok, got %d", len(ledokData))
	}
}

func TestBackofficeFinancialReportExcel(t *testing.T) {
	db, handler, service := testServer(t)
	operatorID := addUser(t, db, "operator_rep@example.com", "operator")
	adminID := addUser(t, db, "ipang_rep@example.com", "superadmin")

	// 1. Operator cannot download backoffice financial report
	recOp := requestAs(t, handler, service, operatorID, http.MethodGet, "/backoffice/reports.xlsx", nil)
	if recOp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for operator, got %d", recOp.Code)
	}

	// 2. Superadmin downloads default financial report Excel (24 months)
	recAdmin := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/reports.xlsx", nil)
	if recAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin, got %d", recAdmin.Code)
	}

	contentType := recAdmin.Header().Get("Content-Type")
	if !strings.Contains(contentType, "spreadsheetml.sheet") {
		t.Fatalf("expected Excel spreadsheetml.sheet Content-Type, got %s", contentType)
	}

	// Parse with excelize to verify valid OOXML workbook and sheets
	body := recAdmin.Body.Bytes()
	f, err := excelize.OpenReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to open Excel with excelize: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	expectedSheets := []string{"Konsolidasi", "Pardis Barbershop Klaseman", "Pardis Barbershop Ledok", "Penjualan Produk"}
	for _, expected := range expectedSheets {
		found := false
		for _, s := range sheets {
			if s == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected sheet %s not found in %v", expected, sheets)
		}
	}

	cellA1, err := f.GetCellValue("Konsolidasi", "A1")
	if err != nil || cellA1 != "Rangkuman Konsolidasi 2 Cabang" {
		t.Fatalf("expected title 'Rangkuman Konsolidasi 2 Cabang', got '%s', err: %v", cellA1, err)
	}

	// 3. Superadmin downloads with flexible date filter (?from=2025-01&to=2025-06)
	recFiltered := requestAs(t, handler, service, adminID, http.MethodGet, "/backoffice/reports.xlsx?from=2025-01&to=2025-06", nil)
	if recFiltered.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for filtered report, got %d", recFiltered.Code)
	}

	disp := recFiltered.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "laporan-keuangan-2025-01-sd-2025-06.xlsx") {
		t.Fatalf("expected filename with date range in Content-Disposition, got %s", disp)
	}

	fFiltered, err := excelize.OpenReader(bytes.NewReader(recFiltered.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to open filtered Excel with excelize: %v", err)
	}
	defer fFiltered.Close()

	// Sheet Konsolidasi should contain rows for 6 months (Jan 2025 - Jun 2025)
	rows, err := fFiltered.GetRows("Konsolidasi")
	if err != nil {
		t.Fatalf("failed to get rows: %v", err)
	}
	// Row 1: Title, Row 2: Empty, Row 3: Header, Rows 4-9: 6 months, Row 10: Total -> at least 10 rows
	if len(rows) < 10 {
		t.Fatalf("expected at least 10 rows for 6-month filtered report, got %d", len(rows))
	}
}





