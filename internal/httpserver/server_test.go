package httpserver

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mekari/pos-phoenix/internal/auth"
	"github.com/mekari/pos-phoenix/internal/database"
	transactionstore "github.com/mekari/pos-phoenix/internal/transaction"
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


