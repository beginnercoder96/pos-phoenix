package httpserver

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mekari/pos-phoenix/internal/auth"
	"github.com/mekari/pos-phoenix/internal/backoffice"
	transactionstore "github.com/mekari/pos-phoenix/internal/transaction"
)

type contextKey string

const userKey contextKey = "user"

type Server struct {
	auth         auth.Service
	transactions transactionstore.Repository
	backoffice   *backoffice.Repository
	templates    *template.Template
	secure       bool
	location     *time.Location
	now          func() time.Time
}
type trendPoint struct {
	Label                       string
	PeriodRange                 string
	IncomeCents, ExpenseCents   int64
	IncomeHeight, ExpenseHeight int
}

type CalendarDay struct {
	Day        int
	DateStr    string
	IsSelected bool
	IsInRange  bool
}

type pageData struct {
	User                        auth.User
	Entries                     []transactionstore.Entry
	Operators                   []auth.User
	Summary                     transactionstore.Summary
	Trend                       []trendPoint
	CSRF                        string
	Error                       string
	Today                       string
	From                        string
	To                          string
	Range                       string
	Timezone                    string
	CurrentDateTime             string
	Page, Pages, Total          int
	PrevURL, NextURL, ExportURL string
	CurrentURL                  string
	Theme                       string
	Language                    string
	ChartRange                  string
	ChartPeriod                 string
	ChartFrom                   string
	ChartTo                     string
	CalendarMonthTitle          string
	CalendarMonthLimit          int
	CalendarBlanks              []int
	CalendarDays                []CalendarDay
	Catalog                     []CatalogCategory
	CatalogJSON                 template.HTML
	CatalogJS                   template.JS
	Greeting                    string
	RedirectURL                 string
	Branches                    []backoffice.Branch
	SavedName                   string
	Success                     string
	ResetToken                  string
	ResetLink                   string
}

func New(db *sql.DB, secure bool, location *time.Location) (http.Handler, error) {
	if location == nil {
		location = time.UTC
	}
	templatePattern := filepath.Join("web", "templates", "*.html")
	if matches, _ := filepath.Glob(templatePattern); len(matches) == 0 {
		templatePattern = filepath.Join("..", "..", "web", "templates", "*.html")
	}
	funcs := template.FuncMap(templateFuncs(location))
	funcs["label"] = preferenceLabel
	t, err := template.New("").Funcs(funcs).ParseGlob(templatePattern)
	if err != nil {
		return nil, err
	}
	s := &Server{auth: auth.Service{DB: db}, transactions: transactionstore.Repository{DB: db}, backoffice: &backoffice.Repository{DB: db}, templates: t, secure: secure, location: location, now: time.Now}
	mux := http.NewServeMux()
	staticFS := http.StripPrefix("/static/", http.FileServer(safeDir("web/static")))
	mux.HandleFunc("GET /static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		staticFS.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /login", s.loginPage)
	mux.HandleFunc("POST /login", s.login)
	mux.HandleFunc("GET /forgot-password", s.forgotPasswordPage)
	mux.HandleFunc("POST /forgot-password", s.requestPasswordReset)
	mux.HandleFunc("GET /reset-password", s.resetPasswordPage)
	mux.HandleFunc("POST /reset-password", s.submitPasswordReset)
	mux.HandleFunc("POST /logout", s.withUser(s.logout))
	mux.HandleFunc("GET /loading", s.loadingPage)
	mux.HandleFunc("GET /", s.withUser(s.dashboard))
	mux.HandleFunc("GET /reports.xlsx", s.withUser(s.adminOnly(s.exportExcel)))
	mux.HandleFunc("GET /reports.csv", s.withUser(s.adminOnly(s.exportCSV)))
	mux.HandleFunc("POST /transactions", s.withUser(s.createTransaction))
	mux.HandleFunc("POST /transactions/{id}/reverse", s.withUser(s.adminOnly(s.reverseTransaction)))
	mux.HandleFunc("POST /preferences", s.setPreferences)
	mux.HandleFunc("GET /operators", s.withUser(s.adminOnly(s.operatorsPage)))
	mux.HandleFunc("POST /operators", s.withUser(s.adminOnly(s.createOperator)))
	mux.HandleFunc("POST /operators/{id}/active", s.withUser(s.adminOnly(s.setOperatorActive)))
	mux.HandleFunc("POST /operators/{id}/credentials", s.withUser(s.adminOnly(s.updateOperatorCredentials)))
	// Backoffice routes (admin-only)
	mux.HandleFunc("GET /backoffice", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/backoffice/", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /backoffice/", s.withUser(s.adminOnly(s.backofficeDashboard)))
	mux.HandleFunc("GET /backoffice/profit-sharing", s.withUser(s.adminOnly(s.backofficeProfitSharing)))
	mux.HandleFunc("POST /backoffice/profit-sharing", s.withUser(s.adminOnly(s.backofficeSaveProfitSharing)))
	mux.HandleFunc("GET /backoffice/discounts", s.withUser(s.adminOnly(s.backofficeDiscounts)))
	mux.HandleFunc("POST /backoffice/discounts", s.withUser(s.adminOnly(s.backofficeSaveDiscount)))
	mux.HandleFunc("POST /backoffice/discounts/{id}/delete", s.withUser(s.adminOnly(s.backofficeDeleteDiscount)))
	mux.HandleFunc("GET /backoffice/products", s.withUser(s.adminOnly(s.backofficeProducts)))
	mux.HandleFunc("POST /backoffice/products", s.withUser(s.adminOnly(s.backofficeSaveProduct)))
	mux.HandleFunc("GET /backoffice/payroll", s.withUser(s.adminOnly(s.backofficePayroll)))
	mux.HandleFunc("GET /backoffice/payroll/slip", s.withUser(s.adminOnly(s.backofficePayrollSlip)))
	mux.HandleFunc("GET /backoffice/payroll/slip-all", s.withUser(s.adminOnly(s.backofficePayrollSlipAll)))
	mux.HandleFunc("GET /backoffice/reports.xlsx", s.withUser(s.adminOnly(s.backofficeFinancialReport)))
	mux.HandleFunc("GET /backoffice/api/analytics/trend-24m", s.withUser(s.adminOnly(s.backofficeAnalyticsAPI)))
	return s.securityHeaders(mux), nil
}

func (s *Server) loadingPage(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("return_to")
	if redirect != "" && (!strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//")) {
		redirect = "/"
	}
	s.render(w, "loading.html", localizedData(r, pageData{
		RedirectURL: redirect,
	}))
}

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "login.html", localizedData(r, pageData{CSRF: s.csrf(w, r)}))
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	if username == "" || strings.Contains(username, "@") {
		s.renderStatus(w, "login.html", localizedData(r, pageData{CSRF: s.csrf(w, r), Error: "Username or password is incorrect."}), http.StatusUnauthorized)
		return
	}
	u, err := s.auth.Authenticate(r.Context(), username, r.FormValue("password"))
	if err != nil {
		s.renderStatus(w, "login.html", localizedData(r, pageData{CSRF: s.csrf(w, r), Error: "Username or password is incorrect."}), http.StatusUnauthorized)
		return
	}
	token, err := s.auth.CreateSession(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "unable to sign in", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: token, Path: "/", HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode, MaxAge: 43200})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) forgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "forgot_password.html", localizedData(r, pageData{CSRF: s.csrf(w, r)}))
}

func (s *Server) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		s.renderStatus(w, "forgot_password.html", localizedData(r, pageData{
			CSRF:  s.csrf(w, r),
			Error: "Please enter your registered email address.",
		}), http.StatusBadRequest)
		return
	}

	token, resetLink, err := s.auth.CreatePasswordResetToken(r.Context(), email)
	if err != nil {
		s.renderStatus(w, "forgot_password.html", localizedData(r, pageData{
			CSRF:  s.csrf(w, r),
			Error: err.Error(),
		}), http.StatusBadRequest)
		return
	}

	s.render(w, "forgot_password.html", localizedData(r, pageData{
		CSRF:       s.csrf(w, r),
		Success:    "resetEmailSentMessage",
		ResetToken: token,
		ResetLink:  resetLink,
	}))
}

func (s *Server) resetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}
	_, err := s.auth.ValidatePasswordResetToken(r.Context(), token)
	if err != nil {
		s.renderStatus(w, "reset_password.html", localizedData(r, pageData{
			CSRF:       s.csrf(w, r),
			Error:      "invalidOrExpiredToken",
			ResetToken: token,
		}), http.StatusBadRequest)
		return
	}
	s.render(w, "reset_password.html", localizedData(r, pageData{
		CSRF:       s.csrf(w, r),
		ResetToken: token,
	}))
}

func (s *Server) submitPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	token := strings.TrimSpace(r.FormValue("token"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	if password != confirm {
		s.renderStatus(w, "reset_password.html", localizedData(r, pageData{
			CSRF:       s.csrf(w, r),
			Error:      "passwordsDoNotMatch",
			ResetToken: token,
		}), http.StatusBadRequest)
		return
	}

	if err := s.auth.ResetPasswordWithToken(r.Context(), token, password); err != nil {
		s.renderStatus(w, "reset_password.html", localizedData(r, pageData{
			CSRF:       s.csrf(w, r),
			Error:      err.Error(),
			ResetToken: token,
		}), http.StatusBadRequest)
		return
	}

	s.render(w, "login.html", localizedData(r, pageData{
		CSRF:    s.csrf(w, r),
		Success: "passwordResetSuccess",
	}))
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", 403)
		return
	}
	if c, err := r.Cookie("session"); err == nil {
		_ = s.auth.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Path: "/", HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	from, to, rangeName, err := s.reportRange(r, u, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageNumber := positiveInt(r.URL.Query().Get("page"), 1)
	result, err := s.transactions.List(r.Context(), u.ID, u.Role == "superadmin", from, to, pageNumber, 25)
	if err != nil {
		http.Error(w, "unable to load transactions", http.StatusInternalServerError)
		return
	}
	chartPeriod := r.URL.Query().Get("chart_period")
	if chartPeriod != "month" && chartPeriod != "today" && chartPeriod != "custom" {
		chartPeriod = "week"
	}

	var chartFrom, chartTo time.Time
	var chartFromStr, chartToStr string
	nowInLoc := now.In(s.location)
	todayStart := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 0, 0, 0, 0, s.location)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	switch chartPeriod {
	case "today":
		chartFrom = todayStart
		chartTo = tomorrowStart
		chartFromStr = todayStart.Format("2006-01-02")
		chartToStr = todayStart.Format("2006-01-02")
	case "month":
		chartFrom = time.Date(nowInLoc.Year(), nowInLoc.Month()-5, 1, 0, 0, 0, 0, s.location)
		chartTo = tomorrowStart
		chartFromStr = chartFrom.Format("2006-01-02")
		chartToStr = todayStart.Format("2006-01-02")
	case "custom":
		cfVal := r.URL.Query().Get("chart_from")
		ctVal := r.URL.Query().Get("chart_to")
		var cf, ct time.Time
		var parseErr error
		if cfVal != "" {
			cf, parseErr = time.ParseInLocation("2006-01-02", cfVal, s.location)
		}
		if parseErr != nil || cf.IsZero() {
			// default to 1st of current month
			cf = time.Date(nowInLoc.Year(), nowInLoc.Month(), 1, 0, 0, 0, 0, s.location)
		}
		if ctVal != "" {
			ct, parseErr = time.ParseInLocation("2006-01-02", ctVal, s.location)
		}
		if parseErr != nil || ct.IsZero() {
			// default to today
			ct = todayStart
		}
		if ct.Before(cf) {
			ct = cf
		}
		// Enforce maximum days = days in month of cf (28, 29, 30, or 31) based on available dates on month shown
		daysInMonth := time.Date(cf.Year(), cf.Month()+1, 0, 0, 0, 0, 0, s.location).Day()
		numDays := int(ct.Sub(cf).Hours()/24) + 1
		if numDays > daysInMonth {
			ct = cf.AddDate(0, 0, daysInMonth-1)
		}
		chartFrom = cf
		chartTo = ct.AddDate(0, 0, 1) // inclusive end of day for database query
		chartFromStr = cf.Format("2006-01-02")
		chartToStr = ct.Format("2006-01-02")
	default:
		chartFrom = todayStart.AddDate(0, 0, -6)
		chartTo = tomorrowStart
		chartFromStr = chartFrom.Format("2006-01-02")
		chartToStr = todayStart.Format("2006-01-02")
	}

	chartEntries, err := s.transactions.All(r.Context(), u.ID, true, chartFrom, chartTo)
	if err != nil {
		http.Error(w, "unable to load chart data", http.StatusInternalServerError)
		return
	}
	trend := buildTrend(chartEntries, s.location, nowInLoc, chartPeriod, chartFrom, chartTo.AddDate(0, 0, -1))
	query := r.URL.Query()
	query.Del("page")
	baseURL := "/?" + query.Encode()
	if len(query) == 0 {
		baseURL = "/"
	}
	pref := readPreferences(r)
	hour := now.Hour()
	var greeting string
	if pref.Language == "id" {
		switch {
		case hour >= 5 && hour < 12:
			greeting = "Selamat Pagi"
		case hour >= 12 && hour < 15:
			greeting = "Selamat Siang"
		case hour >= 15 && hour < 18:
			greeting = "Selamat Sore"
		default:
			greeting = "Selamat Malam"
		}
	} else {
		switch {
		case hour >= 5 && hour < 12:
			greeting = "Good Morning"
		case hour >= 12 && hour < 18:
			greeting = "Good Afternoon"
		default:
			greeting = "Good Night"
		}
	}
	// Generate pre-rendered calendar grid data for the month shown (chartFrom's month)
	calYear := chartFrom.In(s.location).Year()
	calMonth := chartFrom.In(s.location).Month()
	calDaysInMonth := time.Date(calYear, calMonth+1, 0, 0, 0, 0, 0, s.location).Day()
	calFirstWeekday := int(time.Date(calYear, calMonth, 1, 0, 0, 0, 0, s.location).Weekday())

	calBlanks := make([]int, calFirstWeekday)
	calDays := make([]CalendarDay, calDaysInMonth)
	for d := 1; d <= calDaysInMonth; d++ {
		dStr := fmt.Sprintf("%04d-%02d-%02d", calYear, calMonth, d)
		calDays[d-1] = CalendarDay{
			Day:        d,
			DateStr:    dStr,
			IsSelected: (dStr == chartFromStr || dStr == chartToStr),
			IsInRange:  (dStr >= chartFromStr && dStr <= chartToStr),
		}
	}
	calMonthTitle := fmt.Sprintf("%s %d", calMonth.String(), calYear)
	if pref.Language == "id" {
		monthsId := map[time.Month]string{
			time.January: "Januari", time.February: "Februari", time.March: "Maret", time.April: "April",
			time.May: "Mei", time.June: "Juni", time.July: "Juli", time.August: "Agustus",
			time.September: "September", time.October: "Oktober", time.November: "November", time.December: "Desember",
		}
		calMonthTitle = fmt.Sprintf("%s %d", monthsId[calMonth], calYear)
	}

	currentDateTime := formatCurrentDateTime(now, s.location, pref.Language)

	data := localizedData(r, pageData{
		User: u, Entries: result.Entries, Summary: result.Summary, Trend: trend, CSRF: s.csrf(w, r),
		Today: now.Format("2006-01-02"), From: from.In(s.location).Format("2006-01-02"),
		To: to.In(s.location).AddDate(0, 0, -1).Format("2006-01-02"), Range: rangeName, Timezone: "WIB (Asia/Jakarta)",
		CurrentDateTime: currentDateTime,
		Page: result.Page, Pages: result.Pages, Total: result.Total,
		CurrentURL: r.URL.RequestURI(), ChartRange: rangeName, ChartPeriod: chartPeriod,
		ChartFrom: chartFromStr, ChartTo: chartToStr,
		CalendarMonthTitle: calMonthTitle,
		CalendarMonthLimit: calDaysInMonth,
		CalendarBlanks:     calBlanks,
		CalendarDays:       calDays,
		Catalog: DefaultCatalog, CatalogJSON: CatalogJSON(), CatalogJS: CatalogJS(),
		Greeting: greeting,
	})
	if result.Page > 1 {
		data.PrevURL = pageURL(baseURL, result.Page-1)
	}
	if result.Page < result.Pages {
		data.NextURL = pageURL(baseURL, result.Page+1)
	}
	if u.Role == "superadmin" {
		exportQuery := url.Values{"range": {rangeName}, "from": {data.From}, "to": {data.To}}
		data.ExportURL = "/reports.xlsx?" + exportQuery.Encode()
	}
	s.render(w, "dashboard.html", data)
}

func (s *Server) reportRange(r *http.Request, user auth.User, now time.Time) (time.Time, time.Time, string, error) {
	from, to := businessDayRange(now, s.location)
	rangeName := "today"
	if user.Role != "superadmin" {
		return from, to, rangeName, nil
	}
	switch r.URL.Query().Get("range") {
	case "month":
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, s.location)
		to = from.AddDate(0, 1, 0)
		rangeName = "month"
	case "custom":
		var err error
		from, to, err = customReportRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"), now, s.location)
		if err != nil {
			return time.Time{}, time.Time{}, "", err
		}
		rangeName = "custom"
	}
	return from, to, rangeName, nil
}

func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request) {
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
	filename := "transactions-" + from.In(s.location).Format("20060102") + "-" + to.In(s.location).AddDate(0, 0, -1).Format("20060102") + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Date (WIB)", "Operator", "Type", "Amount", "Category", "Note"})
	for _, entry := range entries {
		_ = writer.Write([]string{entry.OccurredAt.In(s.location).Format("2006-01-02 15:04:05"), entry.OperatorName, entry.Kind, transactionstore.FormatCents(entry.AmountCents), entry.Category, entry.Note})
	}
	writer.Flush()
}

func positiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func pageURL(base string, page int) string {
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	if strings.HasSuffix(base, "?") {
		separator = ""
	}
	return base + separator + "page=" + strconv.Itoa(page)
}

func formatCurrentDateTime(t time.Time, loc *time.Location, lang string) string {
	t = t.In(loc)
	if lang == "id" {
		days := map[time.Weekday]string{
			time.Sunday: "Minggu", time.Monday: "Senin", time.Tuesday: "Selasa",
			time.Wednesday: "Rabu", time.Thursday: "Kamis", time.Friday: "Jumat", time.Saturday: "Sabtu",
		}
		months := map[time.Month]string{
			time.January: "Jan", time.February: "Feb", time.March: "Mar", time.April: "Apr",
			time.May: "Mei", time.June: "Jun", time.July: "Jul", time.August: "Agu",
			time.September: "Sep", time.October: "Okt", time.November: "Nov", time.December: "Des",
		}
		return fmt.Sprintf("%s, %02d %s %d %02d:%02d WIB",
			days[t.Weekday()], t.Day(), months[t.Month()], t.Year(), t.Hour(), t.Minute())
	}
	return t.Format("Monday, 02 Jan 2006 15:04 WIB")
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", 403)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", 400)
		return
	}
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	when, err := time.ParseInLocation("2006-01-02", r.FormValue("date"), s.location)
	if err != nil || when.Format("2006-01-02") != now.Format("2006-01-02") {
		http.Error(w, "operators may enter today's transactions only", http.StatusBadRequest)
		return
	}
	kind := r.FormValue("kind")
	if kind != "income" && kind != "expense" {
		http.Error(w, "invalid transaction kind", 400)
		return
	}

	categories := r.PostForm["category"]
	if len(categories) == 0 {
		categories = r.PostForm["category[]"]
	}
	amounts := r.PostForm["amount"]
	if len(amounts) == 0 {
		amounts = r.PostForm["amount[]"]
	}
	itemNames := r.PostForm["item_name"]
	if len(itemNames) == 0 {
		itemNames = r.PostForm["item_name[]"]
	}
	itemTypes := r.PostForm["item_type"]
	if len(itemTypes) == 0 {
		itemTypes = r.PostForm["item_type[]"]
	}
	barberIDs := r.PostForm["barber_id"]
	if len(barberIDs) == 0 {
		barberIDs = r.PostForm["barber_id[]"]
	}
	discounts := r.PostForm["discount_amount"]
	if len(discounts) == 0 {
		discounts = r.PostForm["discount_amount[]"]
	}
	commissions := r.PostForm["commission_earned"]
	if len(commissions) == 0 {
		commissions = r.PostForm["commission_earned[]"]
	}

	var items []transactionstore.Item
	var totalCents int64
	var distinctCategories []string
	seenCat := make(map[string]bool)

	if len(categories) > 0 {
		for i, cat := range categories {
			cat = strings.TrimSpace(cat)
			if cat == "" {
				continue
			}
			var amtStr string
			if i < len(amounts) {
				amtStr = amounts[i]
			}
			amtStr = strings.TrimSpace(amtStr)
			if amtStr == "" {
				continue
			}
			cents, err := transactionstore.ParseCents(amtStr)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			var itemName string
			if i < len(itemNames) {
				itemName = strings.TrimSpace(itemNames[i])
			}
			var itType string
			if i < len(itemTypes) {
				itType = strings.TrimSpace(itemTypes[i])
			}
			var bID int64
			if i < len(barberIDs) {
				bID, _ = strconv.ParseInt(barberIDs[i], 10, 64)
			}
			var discAmt int64
			if i < len(discounts) {
				discAmt, _ = transactionstore.ParseCents(discounts[i])
			}
			var commEarned int64
			if i < len(commissions) {
				commEarned, _ = transactionstore.ParseCents(commissions[i])
			}
			items = append(items, transactionstore.Item{
				Category:         cat,
				ItemName:         itemName,
				AmountCents:      cents,
				ItemType:         itType,
				BarberID:         bID,
				DiscountAmount:   discAmt,
				CommissionEarned: commEarned,
			})
			totalCents += cents
			if !seenCat[cat] {
				seenCat[cat] = true
				distinctCategories = append(distinctCategories, cat)
			}
		}
	}

	if len(items) == 0 {
		cat := strings.TrimSpace(r.FormValue("category"))
		if cat == "" || len(cat) > 255 {
			http.Error(w, "category is required", 400)
			return
		}
		cents, err := transactionstore.ParseCents(r.FormValue("amount"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		items = append(items, transactionstore.Item{
			Category:    cat,
			AmountCents: cents,
		})
		totalCents = cents
		distinctCategories = append(distinctCategories, cat)
	}

	joinedCategory := strings.Join(distinctCategories, ", ")
	if len(joinedCategory) > 255 {
		joinedCategory = joinedCategory[:255]
	}

	var branchID int64
	if bStr := r.FormValue("branch_id"); bStr != "" {
		branchID, _ = strconv.ParseInt(bStr, 10, 64)
	}
	if branchID <= 0 && u.BranchID > 0 {
		branchID = u.BranchID
	}

	err = s.transactions.Create(r.Context(), transactionstore.Entry{
		OperatorID:  u.ID,
		BranchID:    branchID,
		Kind:        kind,
		AmountCents: totalCents,
		Category:    joinedCategory,
		Note:        strings.TrimSpace(r.FormValue("note")),
		OccurredAt:  s.now().UTC(),
		Items:       items,
	})
	if err != nil {
		http.Error(w, "unable to save transaction", 400)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) reverseTransaction(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	transactionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || transactionID <= 0 {
		http.Error(w, "invalid transaction", http.StatusBadRequest)
		return
	}
	user := userFrom(r.Context())
	if err := s.transactions.Reverse(r.Context(), user.ID, transactionID, r.FormValue("reason"), s.now().UTC()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	redirect := r.FormValue("return_to")
	if redirect == "" || !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (s *Server) operatorsPage(w http.ResponseWriter, r *http.Request) {
	operators, err := s.auth.ListOperators(r.Context())
	if err != nil {
		http.Error(w, "unable to load operators", http.StatusInternalServerError)
		return
	}
	branches, _ := s.backoffice.ListBranches(r.Context())
	savedName, _ := url.QueryUnescape(r.URL.Query().Get("saved"))
	s.render(w, "operators.html", localizedData(r, pageData{User: userFrom(r.Context()), Operators: operators, Branches: branches, CSRF: s.csrf(w, r), CurrentURL: r.URL.RequestURI(), SavedName: savedName}))
}

func (s *Server) createOperator(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	var branchID int64
	if bStr := r.FormValue("branch_id"); bStr != "" {
		branchID, _ = strconv.ParseInt(bStr, 10, 64)
	}
	username := strings.TrimSpace(r.FormValue("username"))
	staffType := strings.TrimSpace(r.FormValue("staff_type"))
	phone := strings.TrimSpace(r.FormValue("phone_number"))
	bankName := strings.TrimSpace(r.FormValue("bank_name"))
	bankAccount := strings.TrimSpace(r.FormValue("bank_account_number"))

	if err := s.auth.CreateOperatorWithBranchAndCredentials(r.Context(), username, r.FormValue("email"), r.FormValue("display_name"), r.FormValue("password"), "operator", staffType, branchID, phone, bankName, bankAccount); err != nil {
		operators, _ := s.auth.ListOperators(r.Context())
		branches, _ := s.backoffice.ListBranches(r.Context())
		s.renderStatus(w, "operators.html", localizedData(r, pageData{User: userFrom(r.Context()), Operators: operators, Branches: branches, CSRF: s.csrf(w, r), Error: err.Error()}), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/operators", http.StatusSeeOther)
}

func (s *Server) updateOperatorCredentials(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	operatorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || operatorID <= 0 {
		http.Error(w, "invalid operator", http.StatusBadRequest)
		return
	}
	var branchID int64
	if bStr := r.FormValue("branch_id"); bStr != "" {
		branchID, _ = strconv.ParseInt(bStr, 10, 64)
	}
	staffType := strings.TrimSpace(r.FormValue("staff_type"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	phone := strings.TrimSpace(r.FormValue("phone_number"))
	bankName := strings.TrimSpace(r.FormValue("bank_name"))
	bankAccount := strings.TrimSpace(r.FormValue("bank_account_number"))

	if err := s.auth.UpdateOperatorCredentials(r.Context(), operatorID, displayName, phone, bankName, bankAccount, branchID, staffType); err != nil {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		operators, _ := s.auth.ListOperators(r.Context())
		branches, _ := s.backoffice.ListBranches(r.Context())
		s.renderStatus(w, "operators.html", localizedData(r, pageData{User: userFrom(r.Context()), Operators: operators, Branches: branches, CSRF: s.csrf(w, r), Error: err.Error()}), http.StatusBadRequest)
		return
	}
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"displayName": displayName,
			"phone":       phone,
			"bankName":    bankName,
			"bankAccount": bankAccount,
			"branchID":    branchID,
			"staffType":   staffType,
		})
		return
	}
	http.Redirect(w, r, "/operators?saved="+url.QueryEscape(displayName), http.StatusSeeOther)
}

func (s *Server) setOperatorActive(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	operatorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || operatorID <= 0 {
		http.Error(w, "invalid operator", http.StatusBadRequest)
		return
	}
	active, err := strconv.ParseBool(r.FormValue("active"))
	if err != nil {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := s.auth.SetOperatorActive(r.Context(), operatorID, active); err != nil {
		http.Error(w, "operator not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/operators", http.StatusSeeOther)
}

func (s *Server) withUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", 303)
			return
		}
		u, err := s.auth.UserForSession(r.Context(), c.Value)
		if err != nil {
			http.Redirect(w, r, "/login", 303)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}

func (s *Server) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userFrom(r.Context()).Role != "superadmin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func userFrom(ctx context.Context) auth.User { return ctx.Value(userKey).(auth.User) }
func (s *Server) csrf(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("csrf"); err == nil && len(c.Value) >= 32 {
		return c.Value
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	v := base64.RawURLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: "csrf", Value: v, Path: "/", Secure: s.secure, SameSite: http.SameSiteStrictMode, MaxAge: 43200})
	return v
}
func (s *Server) validCSRF(r *http.Request) bool {
	c, err := r.Cookie("csrf")
	return err == nil && c.Value != "" && c.Value == r.FormValue("csrf")
}
func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	s.renderStatus(w, name, data, http.StatusOK)
}
func (s *Server) renderStatus(w http.ResponseWriter, name string, data pageData, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = s.templates.ExecuteTemplate(w, name, data)
}
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func businessDayRange(now time.Time, location *time.Location) (time.Time, time.Time) {
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	return start, start.AddDate(0, 0, 1)
}

func customReportRange(fromValue, toValue string, now time.Time, location *time.Location) (time.Time, time.Time, error) {
	from, err := time.ParseInLocation("2006-01-02", fromValue, location)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("enter a valid start date")
	}
	lastDay, err := time.ParseInLocation("2006-01-02", toValue, location)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("enter a valid end date")
	}
	to := lastDay.AddDate(0, 0, 1)
	oldest := time.Date(now.Year()-2, now.Month(), now.Day(), 0, 0, 0, 0, location)
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 1)
	if from.Before(oldest) || !to.After(from) || to.After(tomorrow) || to.Sub(from) > 366*2*24*time.Hour {
		return time.Time{}, time.Time{}, errors.New("report dates must be ordered and within the most recent two years")
	}
	return from, to, nil
}

var _ = strconv.Itoa

type safeFile struct {
	http.File
}

type safeDir string

func (d safeDir) Open(name string) (http.File, error) {
	f, err := http.Dir(d).Open(name)
	if err != nil {
		return nil, err
	}
	return safeFile{f}, nil
}
