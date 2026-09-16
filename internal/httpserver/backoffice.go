package httpserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mekari/pos-phoenix/internal/auth"
	"github.com/mekari/pos-phoenix/internal/backoffice"
)

// backofficeData holds template data for backoffice pages.
type backofficeData struct {
	User                   auth.User
	CSRF                   string
	Error                  string
	Success                string
	Language               string
	Theme                  string
	CurrentURL             string
	CurrentDateTime        string
	Branches               []backoffice.Branch
	SelectedBranch         string
	SelectedPeriod         string
	OwnerPercentage        float64
	EmployeeRules          []employeeRuleView
	UnallocatedPercentage  float64
	NetServiceRevenue      int64
	OwnerShareAmount       int64
	UnallocatedAmount      int64
	CatalogItems           []backoffice.CatalogItem
	Discounts              []backoffice.DiscountBundle
	Analytics              []backoffice.MonthlyAnalytics
	AnalyticsJSON          string
	TotalGrossRevenue      int64
	TotalServiceRevenue    int64
	TotalProductRevenue    int64
	TotalReserve           int64
	TotalDiscount          int64
	PayrollSummaries       []backoffice.EmployeePayrollSummary
	Employees              []auth.User
	Greeting               string
}

type employeeRuleView struct {
	UserID      int64
	DisplayName string
	Percentage  float64
	ShareAmount int64
}

func (s *Server) backofficeDashboard(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	pref := readPreferences(r)

	branchFilter := r.URL.Query().Get("branch")
	if branchFilter == "" {
		branchFilter = "all"
	}

	branches, _ := s.backoffice.ListBranches(r.Context())
	analytics, _ := s.backoffice.GetMonthlyAnalytics24(r.Context(), branchFilter, now)

	var totalGross, totalService, totalProduct, totalReserve, totalDiscount int64
	for _, a := range analytics {
		totalGross += a.GrossRevenueCents
		totalService += a.ServiceRevenueCents
		totalProduct += a.ProductRevenueCents
		totalReserve += a.ReserveCents
		totalDiscount += a.DiscountCents
	}

	analyticsJSON, _ := json.Marshal(analytics)

	data := backofficeData{
		User:                u,
		CSRF:                s.csrf(w, r),
		Language:            pref.Language,
		Theme:               pref.Theme,
		CurrentURL:          r.URL.RequestURI(),
		CurrentDateTime:     formatCurrentDateTime(now, s.location, pref.Language),
		Branches:            branches,
		SelectedBranch:      branchFilter,
		Analytics:           analytics,
		AnalyticsJSON:       string(analyticsJSON),
		TotalGrossRevenue:   totalGross,
		TotalServiceRevenue: totalService,
		TotalProductRevenue: totalProduct,
		TotalReserve:        totalReserve,
		TotalDiscount:       totalDiscount,
	}

	s.renderBackoffice(w, "backoffice_dashboard.html", data)
}

func (s *Server) backofficeProfitSharing(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	pref := readPreferences(r)

	branchCode := r.URL.Query().Get("branch")
	if branchCode == "" {
		branchCode = "KLASEMAN"
	}
	periodMonth := r.URL.Query().Get("period")
	if periodMonth == "" {
		periodMonth = now.Format("2006-01")
	}

	branches, _ := s.backoffice.ListBranches(r.Context())
	branch, err := s.backoffice.GetBranchByCode(r.Context(), branchCode)
	if err != nil {
		http.Error(w, "branch not found", http.StatusNotFound)
		return
	}

	employees, _ := s.auth.ListEmployees(r.Context(), branch.ID)
	netServiceRevenue, _ := s.backoffice.GetMonthlyServiceRevenue(r.Context(), branch.ID, periodMonth)

	var ownerPct float64
	var empRuleViews []employeeRuleView
	var unallocPct float64

	rule, empRules, err := s.backoffice.GetProfitSharingConfig(r.Context(), branch.ID, periodMonth)
	if err == nil && rule != nil {
		ownerPct = rule.OwnerPercentage
		unallocPct = rule.UnallocatedPercentage
		for _, er := range empRules {
			name := ""
			for _, emp := range employees {
				if emp.ID == er.UserID {
					name = emp.DisplayName
					break
				}
			}
			empRuleViews = append(empRuleViews, employeeRuleView{
				UserID:      er.UserID,
				DisplayName: name,
				Percentage:  er.Percentage,
				ShareAmount: int64(float64(netServiceRevenue) * er.Percentage / 100.0),
			})
		}
	} else {
		// Default: show all employees with 0%
		for _, emp := range employees {
			empRuleViews = append(empRuleViews, employeeRuleView{
				UserID:      emp.ID,
				DisplayName: emp.DisplayName,
				Percentage:  0,
			})
		}
		unallocPct = 100
	}

	data := backofficeData{
		User:                  u,
		CSRF:                  s.csrf(w, r),
		Language:              pref.Language,
		Theme:                 pref.Theme,
		CurrentURL:            r.URL.RequestURI(),
		CurrentDateTime:       formatCurrentDateTime(now, s.location, pref.Language),
		Branches:              branches,
		SelectedBranch:        branchCode,
		SelectedPeriod:        periodMonth,
		OwnerPercentage:       ownerPct,
		EmployeeRules:         empRuleViews,
		UnallocatedPercentage: unallocPct,
		NetServiceRevenue:     netServiceRevenue,
		OwnerShareAmount:      int64(float64(netServiceRevenue) * ownerPct / 100.0),
		UnallocatedAmount:     int64(float64(netServiceRevenue) * unallocPct / 100.0),
		Employees:             employees,
	}

	s.renderBackoffice(w, "backoffice_profit_sharing.html", data)
}

func (s *Server) backofficeSaveProfitSharing(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}

	u := userFrom(r.Context())
	if u.Role != "superadmin" {
		http.Error(w, "forbidden: only Ipang (Owner) can modify profit sharing", http.StatusForbidden)
		return
	}

	branchCode := r.FormValue("branch")
	periodMonth := r.FormValue("period")
	branch, err := s.backoffice.GetBranchByCode(r.Context(), branchCode)
	if err != nil {
		http.Error(w, "branch not found", http.StatusNotFound)
		return
	}

	ownerPct, err := strconv.ParseFloat(r.FormValue("owner_percentage"), 64)
	if err != nil || ownerPct < 0 || ownerPct > 100 {
		http.Error(w, "invalid owner percentage", http.StatusBadRequest)
		return
	}

	// Parse employee percentages
	empIDs := r.PostForm["employee_id"]
	empPcts := r.PostForm["employee_percentage"]
	var empRules []backoffice.EmployeeProfitRule
	for i, idStr := range empIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		pct := 0.0
		if i < len(empPcts) {
			pct, _ = strconv.ParseFloat(empPcts[i], 64)
		}
		if pct < 0 {
			pct = 0
		}
		empRules = append(empRules, backoffice.EmployeeProfitRule{
			UserID:     id,
			Percentage: pct,
		})
	}

	// Validate total doesn't exceed 100%
	totalPct := ownerPct
	for _, rule := range empRules {
		totalPct += rule.Percentage
	}
	if totalPct > 100 {
		http.Error(w, fmt.Sprintf("total percentage %.1f%% exceeds 100%%", totalPct), http.StatusBadRequest)
		return
	}

	if err := s.backoffice.SaveProfitSharingConfig(r.Context(), branch.ID, periodMonth, ownerPct, empRules, u.ID); err != nil {
		http.Error(w, "failed to save: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/backoffice/profit-sharing?branch=%s&period=%s", branchCode, periodMonth), http.StatusSeeOther)
}

func (s *Server) backofficeDiscounts(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	pref := readPreferences(r)
	now := s.now().In(s.location)
	discounts, _ := s.backoffice.ListDiscounts(r.Context())

	data := backofficeData{
		User:            u,
		CSRF:            s.csrf(w, r),
		Language:        pref.Language,
		Theme:           pref.Theme,
		CurrentURL:      r.URL.RequestURI(),
		CurrentDateTime: formatCurrentDateTime(now, s.location, pref.Language),
		Discounts:       discounts,
	}
	s.renderBackoffice(w, "backoffice_discounts.html", data)
}

func (s *Server) backofficeSaveDiscount(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}

	d := backoffice.DiscountBundle{
		Code: strings.TrimSpace(r.FormValue("code")),
		Name: strings.TrimSpace(r.FormValue("name")),
		Type: r.FormValue("type"),
	}
	if d.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if d.Type != "PERCENTAGE" && d.Type != "FIXED_AMOUNT" && d.Type != "BUNDLE" {
		http.Error(w, "invalid discount type", http.StatusBadRequest)
		return
	}

	val, _ := strconv.ParseInt(r.FormValue("value"), 10, 64)
	d.Value = val
	d.ServiceAllocationRatio, _ = strconv.ParseFloat(r.FormValue("service_ratio"), 64)
	d.ProductAllocationRatio, _ = strconv.ParseFloat(r.FormValue("product_ratio"), 64)
	d.IsActive = r.FormValue("is_active") == "1" || r.FormValue("is_active") == "on"

	if idStr := r.FormValue("id"); idStr != "" {
		d.ID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	if err := s.backoffice.SaveDiscount(r.Context(), d); err != nil {
		http.Error(w, "failed to save discount: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/backoffice/discounts", http.StatusSeeOther)
}

func (s *Server) backofficeDeleteDiscount(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid discount ID", http.StatusBadRequest)
		return
	}
	if err := s.backoffice.DeleteDiscount(r.Context(), id); err != nil {
		http.Error(w, "failed to delete discount", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/backoffice/discounts", http.StatusSeeOther)
}

func (s *Server) backofficeProducts(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	pref := readPreferences(r)
	now := s.now().In(s.location)
	items, _ := s.backoffice.ListCatalogItems(r.Context(), "")

	data := backofficeData{
		User:            u,
		CSRF:            s.csrf(w, r),
		Language:        pref.Language,
		Theme:           pref.Theme,
		CurrentURL:      r.URL.RequestURI(),
		CurrentDateTime: formatCurrentDateTime(now, s.location, pref.Language),
		CatalogItems:    items,
	}
	s.renderBackoffice(w, "backoffice_products.html", data)
}

func (s *Server) backofficeSaveProduct(w http.ResponseWriter, r *http.Request) {
	if !s.validCSRF(r) {
		http.Error(w, "invalid request", http.StatusForbidden)
		return
	}

	item := backoffice.CatalogItem{
		Name:     strings.TrimSpace(r.FormValue("name")),
		Category: strings.TrimSpace(r.FormValue("category")),
		ItemType: r.FormValue("item_type"),
		IsActive: true,
	}
	if item.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if item.ItemType != "SERVICE" && item.ItemType != "PRODUCT" {
		item.ItemType = "SERVICE"
	}

	item.PriceCents, _ = strconv.ParseInt(r.FormValue("price_cents"), 10, 64)
	item.CommissionAmount, _ = strconv.ParseInt(r.FormValue("commission_amount"), 10, 64)

	if idStr := r.FormValue("id"); idStr != "" {
		item.ID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	if err := s.backoffice.SaveCatalogItem(r.Context(), item); err != nil {
		http.Error(w, "failed to save product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/backoffice/products", http.StatusSeeOther)
}

func (s *Server) backofficePayroll(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	now := s.now().In(s.location)
	pref := readPreferences(r)

	branchCode := r.URL.Query().Get("branch")
	if branchCode == "" {
		branchCode = "KLASEMAN"
	}
	periodMonth := r.URL.Query().Get("period")
	if periodMonth == "" {
		periodMonth = now.Format("2006-01")
	}

	branches, _ := s.backoffice.ListBranches(r.Context())
	branch, err := s.backoffice.GetBranchByCode(r.Context(), branchCode)
	if err != nil {
		http.Error(w, "branch not found", http.StatusNotFound)
		return
	}

	employees, _ := s.auth.ListEmployees(r.Context(), branch.ID)
	netServiceRevenue, _ := s.backoffice.GetMonthlyServiceRevenue(r.Context(), branch.ID, periodMonth)
	productCommissions, _ := s.backoffice.GetEmployeeProductCommissions(r.Context(), branch.ID, periodMonth)

	// Get profit sharing config
	var summaries []backoffice.EmployeePayrollSummary
	rule, empRules, configErr := s.backoffice.GetProfitSharingConfig(r.Context(), branch.ID, periodMonth)

	for _, emp := range employees {
		summary := backoffice.EmployeePayrollSummary{
			UserID:      emp.ID,
			DisplayName: emp.DisplayName,
			BranchName:  branch.Name,
		}

		if configErr == nil && rule != nil {
			for _, er := range empRules {
				if er.UserID == emp.ID {
					summary.ServiceSharePercent = er.Percentage
					summary.ServiceShareCents = int64(float64(netServiceRevenue) * er.Percentage / 100.0)
					break
				}
			}
		}

		if comm, ok := productCommissions[emp.ID]; ok {
			summary.ProductCommission = comm
		}

		summary.TotalPayCents = backoffice.CalculateTakeHomePay(
			summary.ServiceShareCents,
			summary.ProductCommission,
			summary.BonusCents,
			summary.DeductionCents,
		)
		summaries = append(summaries, summary)
	}

	data := backofficeData{
		User:             u,
		CSRF:             s.csrf(w, r),
		Language:         pref.Language,
		Theme:            pref.Theme,
		CurrentURL:       r.URL.RequestURI(),
		CurrentDateTime:  formatCurrentDateTime(now, s.location, pref.Language),
		Branches:         branches,
		SelectedBranch:   branchCode,
		SelectedPeriod:   periodMonth,
		PayrollSummaries: summaries,
		Employees:        employees,
	}

	s.renderBackoffice(w, "backoffice_payroll.html", data)
}

func (s *Server) backofficeAnalyticsAPI(w http.ResponseWriter, r *http.Request) {
	branchFilter := r.URL.Query().Get("branch_id")
	if branchFilter == "" {
		branchFilter = "all"
	}
	now := s.now().In(s.location)
	analytics, err := s.backoffice.GetMonthlyAnalytics24(r.Context(), branchFilter, now)
	if err != nil {
		http.Error(w, "unable to load analytics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (s *Server) backofficePayrollSlip(w http.ResponseWriter, r *http.Request) {
	branchCode := r.URL.Query().Get("branch")
	periodMonth := r.URL.Query().Get("period")
	employeeIDStr := r.URL.Query().Get("employee_id")

	if branchCode == "" || periodMonth == "" || employeeIDStr == "" {
		http.Error(w, "branch, period, and employee_id are required", http.StatusBadRequest)
		return
	}

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid employee_id", http.StatusBadRequest)
		return
	}

	branch, err := s.backoffice.GetBranchByCode(r.Context(), branchCode)
	if err != nil {
		http.Error(w, "branch not found", http.StatusNotFound)
		return
	}

	netServiceRevenue, _ := s.backoffice.GetMonthlyServiceRevenue(r.Context(), branch.ID, periodMonth)
	productCommissions, _ := s.backoffice.GetEmployeeProductCommissions(r.Context(), branch.ID, periodMonth)

	rule, empRules, _ := s.backoffice.GetProfitSharingConfig(r.Context(), branch.ID, periodMonth)

	var empName string
	var empStaffType string
	var empPhone string
	var empBankName string
	var empBankAccount string
	var empPct float64
	var empServiceShare int64
	employees, _ := s.auth.ListEmployees(r.Context(), branch.ID)
	for _, emp := range employees {
		if emp.ID == employeeID {
			empName = emp.DisplayName
			empStaffType = emp.StaffType
			empPhone = emp.PhoneNumber
			empBankName = emp.BankName
			empBankAccount = emp.BankAccountNumber
			break
		}
	}
	if rule != nil {
		for _, er := range empRules {
			if er.UserID == employeeID {
				empPct = er.Percentage
				empServiceShare = int64(float64(netServiceRevenue) * empPct / 100.0)
				break
			}
		}
	}
	prodComm := productCommissions[employeeID]
	thp := backoffice.CalculateTakeHomePay(empServiceShare, prodComm, 0, 0)

	prodSales, _ := s.backoffice.GetEmployeeItemizedProductSales(r.Context(), branch.ID, employeeID, periodMonth)
	now := s.now().In(s.location)

	slipData := payrollSlipData{
		BranchName:        branch.Name,
		EmployeeID:        employeeID,
		EmployeeName:      empName,
		StaffType:         empStaffType,
		PhoneNumber:       empPhone,
		BankName:          empBankName,
		BankAccountNumber: empBankAccount,
		Period:            formatPeriodIndo(periodMonth),
		PrintDate:         now.Format("02-01-2006"),
		NetServiceRev:     netServiceRevenue,
		SharePercentage:   empPct,
		ServiceShare:      empServiceShare,
		ProductComm:       prodComm,
		TakeHomePay:       thp,
		ProductSales:      prodSales,
	}

	if r.URL.Query().Get("format") == "excel" {
		excelBytes, err := generatePayrollSlipExcel(slipData)
		if err != nil {
			http.Error(w, "unable to generate payroll slip", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("slip-gaji-%s-%s.xlsx", empName, periodMonth)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Header().Set("Content-Length", strconv.Itoa(len(excelBytes)))
		_, _ = w.Write(excelBytes)
		return
	}

	viewData := payrollSlipViewData{
		Slips:          []payrollSlipData{slipData},
		IsBulk:         false,
		SelectedBranch: branchCode,
		SelectedPeriod: periodMonth,
		ExcelURL:       fmt.Sprintf("/backoffice/payroll/slip?branch=%s&period=%s&employee_id=%d&format=excel", branchCode, periodMonth, employeeID),
	}
	s.renderTemplate(w, "payroll_slip.html", viewData)
}

func (s *Server) backofficePayrollSlipAll(w http.ResponseWriter, r *http.Request) {
	branchCode := r.URL.Query().Get("branch")
	periodMonth := r.URL.Query().Get("period")

	if branchCode == "" || periodMonth == "" {
		http.Error(w, "branch and period are required", http.StatusBadRequest)
		return
	}

	branch, err := s.backoffice.GetBranchByCode(r.Context(), branchCode)
	if err != nil {
		http.Error(w, "branch not found", http.StatusNotFound)
		return
	}

	employees, _ := s.auth.ListEmployees(r.Context(), branch.ID)
	netServiceRevenue, _ := s.backoffice.GetMonthlyServiceRevenue(r.Context(), branch.ID, periodMonth)
	productCommissions, _ := s.backoffice.GetEmployeeProductCommissions(r.Context(), branch.ID, periodMonth)
	rule, empRules, _ := s.backoffice.GetProfitSharingConfig(r.Context(), branch.ID, periodMonth)

	var slips []payrollSlipData
	now := s.now().In(s.location)
	for _, emp := range employees {
		var empPct float64
		var empServiceShare int64
		if rule != nil {
			for _, er := range empRules {
				if er.UserID == emp.ID {
					empPct = er.Percentage
					empServiceShare = int64(float64(netServiceRevenue) * empPct / 100.0)
					break
				}
			}
		}
		prodComm := productCommissions[emp.ID]
		thp := backoffice.CalculateTakeHomePay(empServiceShare, prodComm, 0, 0)
		empProdSales, _ := s.backoffice.GetEmployeeItemizedProductSales(r.Context(), branch.ID, emp.ID, periodMonth)

		slips = append(slips, payrollSlipData{
			BranchName:        branch.Name,
			EmployeeID:        emp.ID,
			EmployeeName:      emp.DisplayName,
			StaffType:         emp.StaffType,
			PhoneNumber:       emp.PhoneNumber,
			BankName:          emp.BankName,
			BankAccountNumber: emp.BankAccountNumber,
			Period:            formatPeriodIndo(periodMonth),
			PrintDate:         now.Format("02-01-2006"),
			NetServiceRev:     netServiceRevenue,
			SharePercentage:   empPct,
			ServiceShare:      empServiceShare,
			ProductComm:       prodComm,
			TakeHomePay:       thp,
			ProductSales:      empProdSales,
		})
	}

	if r.URL.Query().Get("format") == "excel" {
		excelBytes, err := generatePayrollSlipAllExcel(slips)
		if err != nil {
			http.Error(w, "unable to generate payroll slips", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("slip-gaji-semua-%s-%s.xlsx", branchCode, periodMonth)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(excelBytes)))
		w.Write(excelBytes)
		return
	}

	viewData := payrollSlipViewData{
		Slips:          slips,
		IsBulk:         true,
		SelectedBranch: branchCode,
		SelectedPeriod: periodMonth,
		ExcelURL:       fmt.Sprintf("/backoffice/payroll/slip-all?branch=%s&period=%s&format=excel", branchCode, periodMonth),
	}
	s.renderTemplate(w, "payroll_slip.html", viewData)
}

func (s *Server) backofficeFinancialReport(w http.ResponseWriter, r *http.Request) {
	now := s.now().In(s.location)
	branches, _ := s.backoffice.ListBranches(r.Context())

	var branchData []branchReportData
	for _, b := range branches {
		analytics, _ := s.backoffice.GetMonthlyAnalytics24(r.Context(), b.Code, now)
		branchData = append(branchData, branchReportData{
			Branch:    b,
			Analytics: analytics,
		})
	}

	allAnalytics, _ := s.backoffice.GetMonthlyAnalytics24(r.Context(), "all", now)
	productSales, _ := s.backoffice.GetProductSalesSummary(r.Context(), now)

	excelBytes, err := generateFinancialReportExcel(allAnalytics, branchData, productSales)
	if err != nil {
		http.Error(w, "unable to generate report", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("laporan-keuangan-%s.xlsx", now.Format("200601"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(excelBytes)))
	w.Write(excelBytes)
}

type payrollSlipViewData struct {
	Slips          []payrollSlipData
	IsBulk         bool
	SelectedBranch string
	SelectedPeriod string
	ExcelURL       string
}

func formatPeriodIndo(p string) string {
	t, err := time.Parse("2006-01", p)
	if err != nil {
		return p
	}
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if int(t.Month()) < len(months) {
		return fmt.Sprintf("%s %d", months[t.Month()], t.Year())
	}
	return p
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderBackoffice(w http.ResponseWriter, name string, data backofficeData) {
	s.renderTemplate(w, name, data)
}

// Ensure unused import is consumed.
var _ = sql.ErrNoRows
