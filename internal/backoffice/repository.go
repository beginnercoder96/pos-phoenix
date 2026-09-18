package backoffice

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Branch represents a barbershop branch.
type Branch struct {
	ID        int64
	Code      string
	Name      string
	Address   string
	IsActive  bool
	CreatedAt time.Time
}

// CatalogItem represents a service or product in the catalog.
type CatalogItem struct {
	ID               int64
	Name             string
	Category         string
	ItemType         string // "SERVICE" or "PRODUCT"
	PriceCents       int64
	CommissionAmount int64 // per-item commission for products (in cents)
	IsActive         bool
}

// DiscountBundle represents a discount or bundle configuration.
type DiscountBundle struct {
	ID                     int64
	Code                   string
	Name                   string
	Type                   string // "PERCENTAGE", "FIXED_AMOUNT", "BUNDLE"
	Value                  int64
	ServiceAllocationRatio float64
	ProductAllocationRatio float64
	IsActive               bool
}

// BranchProfitRule holds a profit sharing rule record from DB.
type BranchProfitRule struct {
	ID                    int64
	BranchID              int64
	PeriodMonth           string
	OwnerPercentage       float64
	UnallocatedPercentage float64
	UpdatedBy             sql.NullInt64
	UpdatedAt             time.Time
}

// EmployeeProfitRule holds an employee profit sharing percentage.
type EmployeeProfitRule struct {
	ID         int64
	RuleID     int64
	UserID     int64
	Percentage float64
}

// MonthlyAnalytics holds aggregated monthly analytics data.
type MonthlyAnalytics struct {
	Month              string // "YYYY-MM"
	GrossRevenueCents  int64
	ServiceRevenueCents int64
	ProductRevenueCents int64
	DiscountCents       int64
	ReserveCents        int64
}

// EmployeePayrollSummary holds a summary for payroll display.
type EmployeePayrollSummary struct {
	UserID              int64
	DisplayName         string
	BranchName          string
	ServiceSharePercent float64
	ServiceShareCents   int64
	ProductCommission   int64
	BonusCents          int64
	DeductionCents      int64
	TotalPayCents       int64
}

// ProductSaleRecord holds an itemized product sale record for Sheet 4 of financial report.
type ProductSaleRecord struct {
	BranchName      string
	PeriodMonth     string
	ProductName     string
	BarberName      string
	Quantity        int
	PriceCents      int64
	CommissionRate  int64
	TotalRevenue    int64
	TotalCommission int64
}

// Repository provides data access for backoffice features.
type Repository struct {
	DB *sql.DB
}

// ListBranches returns all active branches.
func (r *Repository) ListBranches(ctx context.Context) ([]Branch, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, code, name, address, is_active, created_at FROM branches WHERE is_active=1 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var branches []Branch
	for rows.Next() {
		var b Branch
		if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Address, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

// GetBranchByCode returns a branch by its code.
func (r *Repository) GetBranchByCode(ctx context.Context, code string) (Branch, error) {
	var b Branch
	err := r.DB.QueryRowContext(ctx, `SELECT id, code, name, address, is_active, created_at FROM branches WHERE code=?`, code).Scan(&b.ID, &b.Code, &b.Name, &b.Address, &b.IsActive, &b.CreatedAt)
	return b, err
}

// SaveProfitSharingConfig saves or updates profit sharing rules for a branch/month.
func (r *Repository) SaveProfitSharingConfig(ctx context.Context, branchID int64, periodMonth string, ownerPct float64, employeeRules []EmployeeProfitRule, updatedBy int64) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Calculate unallocated
	totalEmpPct := 0.0
	for _, rule := range employeeRules {
		totalEmpPct += rule.Percentage
	}
	unallocatedPct := 100.0 - ownerPct - totalEmpPct
	if unallocatedPct < 0 {
		unallocatedPct = 0
	}

	// Upsert branch rule
	_, err = tx.ExecContext(ctx,
		`INSERT INTO branch_profit_sharing_rules (branch_id, period_month, owner_percentage, unallocated_percentage, updated_by, updated_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(branch_id, period_month) DO UPDATE SET
		   owner_percentage=excluded.owner_percentage,
		   unallocated_percentage=excluded.unallocated_percentage,
		   updated_by=excluded.updated_by,
		   updated_at=CURRENT_TIMESTAMP`,
		branchID, periodMonth, ownerPct, unallocatedPct, updatedBy)
	if err != nil {
		return fmt.Errorf("upsert branch rule: %w", err)
	}

	// Get the rule ID
	var ruleID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM branch_profit_sharing_rules WHERE branch_id=? AND period_month=?`, branchID, periodMonth).Scan(&ruleID)
	if err != nil {
		return fmt.Errorf("get rule id: %w", err)
	}

	// Delete existing employee rules for this rule
	if _, err = tx.ExecContext(ctx, `DELETE FROM employee_profit_sharing_rules WHERE rule_id=?`, ruleID); err != nil {
		return fmt.Errorf("delete old employee rules: %w", err)
	}

	// Insert new employee rules
	for _, rule := range employeeRules {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO employee_profit_sharing_rules (rule_id, user_id, percentage) VALUES (?, ?, ?)`,
			ruleID, rule.UserID, rule.Percentage); err != nil {
			return fmt.Errorf("insert employee rule: %w", err)
		}
	}

	return tx.Commit()
}

// GetProfitSharingConfig retrieves the profit sharing configuration for a branch/month.
func (r *Repository) GetProfitSharingConfig(ctx context.Context, branchID int64, periodMonth string) (*BranchProfitRule, []EmployeeProfitRule, error) {
	var rule BranchProfitRule
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, branch_id, period_month, owner_percentage, unallocated_percentage, updated_by, updated_at
		 FROM branch_profit_sharing_rules WHERE branch_id=? AND period_month=?`,
		branchID, periodMonth).Scan(&rule.ID, &rule.BranchID, &rule.PeriodMonth, &rule.OwnerPercentage, &rule.UnallocatedPercentage, &rule.UpdatedBy, &rule.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.DB.QueryContext(ctx,
		`SELECT epr.id, epr.rule_id, epr.user_id, epr.percentage
		 FROM employee_profit_sharing_rules epr WHERE epr.rule_id=?`, rule.ID)
	if err != nil {
		return &rule, nil, err
	}
	defer rows.Close()
	var empRules []EmployeeProfitRule
	for rows.Next() {
		var er EmployeeProfitRule
		if err := rows.Scan(&er.ID, &er.RuleID, &er.UserID, &er.Percentage); err != nil {
			return &rule, nil, err
		}
		empRules = append(empRules, er)
	}
	return &rule, empRules, rows.Err()
}

// GetMonthlyServiceRevenue returns the total net service revenue for a branch in a given month.
func (r *Repository) GetMonthlyServiceRevenue(ctx context.Context, branchID int64, periodMonth string) (int64, error) {
	// periodMonth is "YYYY-MM", we need to match transactions in that month
	startDate := periodMonth + "-01"
	// Parse to get end of month
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, err
	}
	end := start.AddDate(0, 1, 0)

	var total sql.NullInt64
	query := `SELECT COALESCE(SUM(ti.amount_cents - COALESCE(ti.discount_amount, 0)), 0)
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		WHERE t.branch_id = ?
		AND t.occurred_at >= ? AND t.occurred_at < ?
		AND COALESCE(ti.item_type, 'SERVICE') = 'SERVICE'
		AND t.kind = 'income'
		AND t.reversal_of_id IS NULL`
	err = r.DB.QueryRowContext(ctx, query, branchID, start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

// GetMonthlyProductRevenue returns the total product revenue for a branch in a given month.
func (r *Repository) GetMonthlyProductRevenue(ctx context.Context, branchID int64, periodMonth string) (int64, error) {
	startDate := periodMonth + "-01"
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, err
	}
	end := start.AddDate(0, 1, 0)

	var total sql.NullInt64
	query := `SELECT COALESCE(SUM(ti.amount_cents), 0)
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		WHERE t.branch_id = ?
		AND t.occurred_at >= ? AND t.occurred_at < ?
		AND ti.item_type = 'PRODUCT'
		AND t.kind = 'income'
		AND t.reversal_of_id IS NULL`
	err = r.DB.QueryRowContext(ctx, query, branchID, start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

// GetMonthlyAnalytics24 returns 24 continuous months of analytics data for a branch or all branches.
func (r *Repository) GetMonthlyAnalytics24(ctx context.Context, branchFilter string, now time.Time) ([]MonthlyAnalytics, error) {
	start := now.AddDate(0, -23, 0)
	startMonth := start.Format("2006-01")
	endMonth := now.Format("2006-01")

	branchCondition := ""
	var args []any
	args = append(args, startMonth+"-01 00:00:00", endMonth+"-31 23:59:59")

	if branchFilter != "" && strings.ToLower(branchFilter) != "all" {
		if _, err := strconv.ParseInt(branchFilter, 10, 64); err == nil {
			branchCondition = " AND t.branch_id = ?"
			args = append(args, branchFilter)
		} else {
			branchCondition = " AND t.branch_id = (SELECT id FROM branches WHERE code = ? COLLATE NOCASE)"
			args = append(args, strings.ToUpper(branchFilter))
		}
	}

	query := fmt.Sprintf(`SELECT
		strftime('%%Y-%%m', t.occurred_at) as month,
		COALESCE(SUM(t.amount_cents), 0) as gross,
		COALESCE(SUM(CASE WHEN ti_type.svc_total IS NOT NULL THEN ti_type.svc_total ELSE t.amount_cents END), 0) as service,
		COALESCE(SUM(CASE WHEN ti_type.prd_total IS NOT NULL THEN ti_type.prd_total ELSE 0 END), 0) as product,
		COALESCE(SUM(CASE WHEN ti_type.disc_total IS NOT NULL THEN ti_type.disc_total ELSE 0 END), 0) as discount
	FROM transactions t
	LEFT JOIN (
		SELECT transaction_id,
			SUM(CASE WHEN COALESCE(item_type, 'SERVICE') = 'SERVICE' THEN amount_cents ELSE 0 END) as svc_total,
			SUM(CASE WHEN item_type = 'PRODUCT' THEN amount_cents ELSE 0 END) as prd_total,
			SUM(COALESCE(discount_amount, 0)) as disc_total
		FROM transaction_items GROUP BY transaction_id
	) ti_type ON ti_type.transaction_id = t.id
	WHERE t.kind = 'income'
	AND t.reversal_of_id IS NULL
	AND t.occurred_at >= ? AND t.occurred_at <= ?
	%s
	GROUP BY strftime('%%Y-%%m', t.occurred_at)
	ORDER BY month`, branchCondition)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queriedMap := make(map[string]MonthlyAnalytics)
	for rows.Next() {
		var a MonthlyAnalytics
		if err := rows.Scan(&a.Month, &a.GrossRevenueCents, &a.ServiceRevenueCents, &a.ProductRevenueCents, &a.DiscountCents); err != nil {
			return nil, err
		}
		queriedMap[a.Month] = a
	}

	// Fetch unallocated rules for computing reserve
	rulesQuery := `SELECT period_month, AVG(unallocated_percentage) FROM branch_profit_sharing_rules WHERE period_month >= ? AND period_month <= ?`
	var rulesArgs []any
	rulesArgs = append(rulesArgs, startMonth, endMonth)
	if branchFilter != "" && strings.ToLower(branchFilter) != "all" {
		if _, err := strconv.ParseInt(branchFilter, 10, 64); err == nil {
			rulesQuery += " AND branch_id = ?"
			rulesArgs = append(rulesArgs, branchFilter)
		} else {
			rulesQuery += " AND branch_id = (SELECT id FROM branches WHERE code = ? COLLATE NOCASE)"
			rulesArgs = append(rulesArgs, strings.ToUpper(branchFilter))
		}
	}
	rulesQuery += " GROUP BY period_month"
	rRows, err := r.DB.QueryContext(ctx, rulesQuery, rulesArgs...)
	unallocMap := make(map[string]float64)
	if err == nil {
		defer rRows.Close()
		for rRows.Next() {
			var m string
			var unalloc float64
			if err := rRows.Scan(&m, &unalloc); err == nil {
				unallocMap[m] = unalloc
			}
		}
	}

	// Build continuous 24 months slice
	var analytics []MonthlyAnalytics
	cur := start
	for i := 0; i < 24; i++ {
		m := cur.Format("2006-01")
		item, exists := queriedMap[m]
		if !exists {
			item = MonthlyAnalytics{Month: m}
		}
		unallocPct, hasRule := unallocMap[m]
		if !hasRule && exists && item.ServiceRevenueCents > 0 {
			unallocPct = 10.0 // Default 10% reserve from business rules
		}
		item.ReserveCents = int64(float64(item.ServiceRevenueCents) * unallocPct / 100.0)
		analytics = append(analytics, item)
		cur = cur.AddDate(0, 1, 0)
	}

	return analytics, nil
}

// GetProductSalesSummary retrieves itemized product sales and commissions for Sheet 4.
func (r *Repository) GetProductSalesSummary(ctx context.Context, now time.Time) ([]ProductSaleRecord, error) {
	start := now.AddDate(0, -23, 0)
	startMonth := start.Format("2006-01")
	endMonth := now.Format("2006-01")

	query := `SELECT
		COALESCE(b.name, 'Pardis Barbershop'),
		strftime('%Y-%m', t.occurred_at) as month,
		ti.item_name,
		COALESCE(u.display_name, 'Barberman'),
		COUNT(ti.id) as qty,
		COALESCE(ci.price_cents, ti.amount_cents) as price,
		COALESCE(ti.commission_earned, COALESCE(ci.commission_amount, 0)) as comm_rate,
		SUM(ti.amount_cents) as total_rev,
		SUM(COALESCE(ti.commission_earned, 0)) as total_comm
	FROM transaction_items ti
	JOIN transactions t ON t.id = ti.transaction_id
	LEFT JOIN branches b ON b.id = t.branch_id
	LEFT JOIN users u ON u.id = ti.barber_id
	LEFT JOIN catalog_items ci ON ci.name = ti.item_name AND ci.item_type = 'PRODUCT'
	WHERE ti.item_type = 'PRODUCT'
	AND t.kind = 'income'
	AND t.reversal_of_id IS NULL
	AND t.occurred_at >= ? AND t.occurred_at <= ?
	GROUP BY b.name, strftime('%Y-%m', t.occurred_at), ti.item_name, u.display_name
	ORDER BY month DESC, b.name, ti.item_name`

	rows, err := r.DB.QueryContext(ctx, query, startMonth+"-01 00:00:00", endMonth+"-31 23:59:59")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ProductSaleRecord
	for rows.Next() {
		var rec ProductSaleRecord
		if err := rows.Scan(&rec.BranchName, &rec.PeriodMonth, &rec.ProductName, &rec.BarberName, &rec.Quantity, &rec.PriceCents, &rec.CommissionRate, &rec.TotalRevenue, &rec.TotalCommission); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// ListCatalogItems returns all catalog items, optionally filtered by type.
func (r *Repository) ListCatalogItems(ctx context.Context, itemType string) ([]CatalogItem, error) {
	query := `SELECT id, name, category, item_type, price_cents, commission_amount, is_active FROM catalog_items`
	var args []any
	if itemType != "" {
		query += ` WHERE item_type = ?`
		args = append(args, itemType)
	}
	query += ` ORDER BY item_type, category, name`
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CatalogItem
	for rows.Next() {
		var item CatalogItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ItemType, &item.PriceCents, &item.CommissionAmount, &item.IsActive); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// SaveCatalogItem creates or updates a catalog item.
func (r *Repository) SaveCatalogItem(ctx context.Context, item CatalogItem) error {
	if item.ID > 0 {
		_, err := r.DB.ExecContext(ctx,
			`UPDATE catalog_items SET name=?, category=?, item_type=?, price_cents=?, commission_amount=?, is_active=? WHERE id=?`,
			item.Name, item.Category, item.ItemType, item.PriceCents, item.CommissionAmount, item.IsActive, item.ID)
		return err
	}
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO catalog_items (name, category, item_type, price_cents, commission_amount, is_active) VALUES (?, ?, ?, ?, ?, ?)`,
		item.Name, item.Category, item.ItemType, item.PriceCents, item.CommissionAmount, item.IsActive)
	return err
}

// ListDiscounts returns all discounts and bundles.
func (r *Repository) ListDiscounts(ctx context.Context) ([]DiscountBundle, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, code, name, type, value, service_allocation_ratio, product_allocation_ratio, is_active FROM discounts_and_bundles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var discounts []DiscountBundle
	for rows.Next() {
		var d DiscountBundle
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Type, &d.Value, &d.ServiceAllocationRatio, &d.ProductAllocationRatio, &d.IsActive); err != nil {
			return nil, err
		}
		discounts = append(discounts, d)
	}
	return discounts, rows.Err()
}

// SaveDiscount creates or updates a discount/bundle.
func (r *Repository) SaveDiscount(ctx context.Context, d DiscountBundle) error {
	if d.ID > 0 {
		_, err := r.DB.ExecContext(ctx,
			`UPDATE discounts_and_bundles SET code=?, name=?, type=?, value=?, service_allocation_ratio=?, product_allocation_ratio=?, is_active=? WHERE id=?`,
			d.Code, d.Name, d.Type, d.Value, d.ServiceAllocationRatio, d.ProductAllocationRatio, d.IsActive, d.ID)
		return err
	}
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO discounts_and_bundles (code, name, type, value, service_allocation_ratio, product_allocation_ratio, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.Code, d.Name, d.Type, d.Value, d.ServiceAllocationRatio, d.ProductAllocationRatio, d.IsActive)
	return err
}

// DeleteDiscount removes a discount/bundle by ID.
func (r *Repository) DeleteDiscount(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM discounts_and_bundles WHERE id=?`, id)
	return err
}

// ToggleDiscountStatus toggles the is_active status of a discount/bundle and returns the new status.
func (r *Repository) ToggleDiscountStatus(ctx context.Context, id int64) (bool, error) {
	var current bool
	err := r.DB.QueryRowContext(ctx, `SELECT is_active FROM discounts_and_bundles WHERE id=?`, id).Scan(&current)
	if err != nil {
		return false, err
	}
	next := !current
	_, err = r.DB.ExecContext(ctx, `UPDATE discounts_and_bundles SET is_active=? WHERE id=?`, next, id)
	return next, err
}

// GetEmployeeProductCommissions returns product commissions for employees in a branch for a month.
func (r *Repository) GetEmployeeProductCommissions(ctx context.Context, branchID int64, periodMonth string) (map[int64]int64, error) {
	startDate := periodMonth + "-01"
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end := start.AddDate(0, 1, 0)

	query := `SELECT ti.barber_id, COALESCE(SUM(ti.commission_earned), 0)
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		WHERE t.branch_id = ?
		AND t.occurred_at >= ? AND t.occurred_at < ?
		AND ti.item_type = 'PRODUCT'
		AND ti.barber_id IS NOT NULL
		AND t.kind = 'income'
		AND t.reversal_of_id IS NULL
		GROUP BY ti.barber_id`
	rows, err := r.DB.QueryContext(ctx, query, branchID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commissions := make(map[int64]int64)
	for rows.Next() {
		var userID, commission int64
		if err := rows.Scan(&userID, &commission); err != nil {
			return nil, err
		}
		commissions[userID] = commission
	}
	return commissions, rows.Err()
}

// EmployeeProductSaleItem represents an itemized product sale record for payroll slip.
type EmployeeProductSaleItem struct {
	ProductName     string
	Quantity        int
	PriceCents      int64
	CommissionRate  int64
	TotalCommission int64
}

// GetEmployeeItemizedProductSales returns itemized product sales for a specific barber in a month.
func (r *Repository) GetEmployeeItemizedProductSales(ctx context.Context, branchID, barberID int64, periodMonth string) ([]EmployeeProductSaleItem, error) {
	startDate := periodMonth + "-01"
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end := start.AddDate(0, 1, 0)

	query := `SELECT
		ti.item_name,
		COUNT(ti.id) as qty,
		COALESCE(ci.price_cents, ti.amount_cents) as price,
		COALESCE(ti.commission_earned, COALESCE(ci.commission_amount, 0)) as comm_rate,
		SUM(COALESCE(ti.commission_earned, 0)) as total_comm
	FROM transaction_items ti
	JOIN transactions t ON t.id = ti.transaction_id
	LEFT JOIN catalog_items ci ON ci.name = ti.item_name AND ci.item_type = 'PRODUCT'
	WHERE t.branch_id = ?
	AND ti.barber_id = ?
	AND t.occurred_at >= ? AND t.occurred_at < ?
	AND ti.item_type = 'PRODUCT'
	AND t.kind = 'income'
	AND t.reversal_of_id IS NULL
	GROUP BY ti.item_name
	ORDER BY ti.item_name`

	rows, err := r.DB.QueryContext(ctx, query, branchID, barberID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EmployeeProductSaleItem
	for rows.Next() {
		var item EmployeeProductSaleItem
		if err := rows.Scan(&item.ProductName, &item.Quantity, &item.PriceCents, &item.CommissionRate, &item.TotalCommission); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListEmployeesByBranch returns employees (barbermen) assigned to a branch.
func (r *Repository) ListEmployeesByBranch(ctx context.Context, branchID int64) ([]struct {
	ID          int64
	DisplayName string
	StaffType   string
}, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, display_name, COALESCE(staff_type, 'barberman')
		FROM users WHERE branch_id = ? AND active = 1
		AND COALESCE(staff_type, 'barberman') IN ('barberman', 'cashier')
		ORDER BY display_name`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var employees []struct {
		ID          int64
		DisplayName string
		StaffType   string
	}
	for rows.Next() {
		var emp struct {
			ID          int64
			DisplayName string
			StaffType   string
		}
		if err := rows.Scan(&emp.ID, &emp.DisplayName, &emp.StaffType); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, rows.Err()
}
