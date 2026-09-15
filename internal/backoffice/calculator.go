package backoffice

// ProfitSharingRule holds the percentage allocation for a single entity.
type ProfitSharingRule struct {
	UserID     int64
	Name       string
	Percentage float64 // 0–100
}

// ProfitSharingConfig holds the complete profit sharing configuration for a branch/month.
type ProfitSharingConfig struct {
	BranchID              int64
	PeriodMonth           string // "YYYY-MM"
	OwnerPercentage       float64
	EmployeeRules         []ProfitSharingRule
	UnallocatedPercentage float64 // computed: 100 - owner - sum(employees)
}

// ProfitSharingResult holds the calculated profit allocation.
type ProfitSharingResult struct {
	NetServiceRevenueCents int64
	OwnerShareCents        int64
	OwnerPercentage        float64
	EmployeeShares         []EmployeeShare
	UnallocatedCents       int64
	UnallocatedPercentage  float64
}

// EmployeeShare holds a single employee's profit sharing allocation.
type EmployeeShare struct {
	UserID         int64
	Name           string
	Percentage     float64
	ShareCents     int64
	CommissionCents int64
	TotalPayCents  int64
}

// ProductCommission holds commission info for product sales by an employee.
type ProductCommission struct {
	UserID           int64
	Name             string
	ProductName      string
	Quantity         int
	CommissionRate   int64 // per-item commission in cents
	TotalCommission  int64
}

// TakeHomePay holds the final salary calculation for an employee.
type TakeHomePay struct {
	UserID              int64
	Name                string
	ServiceShareCents   int64
	ProductCommission   int64
	BonusCents          int64
	DeductionCents      int64
	TotalPayCents       int64
}

// TransactionItem represents a single line item from a transaction for calculation.
type TransactionItem struct {
	ItemType        string // "SERVICE" or "PRODUCT"
	AmountCents     int64
	BarberID        int64
	CommissionRate  int64 // per-item commission for products
	Quantity        int
	DiscountCents   int64
}

// RevenueBreakdown holds the segregated revenue figures.
type RevenueBreakdown struct {
	GrossServiceCents  int64
	GrossProductCents  int64
	ServiceDiscountCents int64
	ProductDiscountCents int64
	NetServiceCents    int64
	NetProductCents    int64
	TotalDiscountCents int64
}

// SegregateRevenue separates transaction items into service vs product pools
// and applies discount allocations. Product revenue is excluded from profit sharing.
func SegregateRevenue(items []TransactionItem) RevenueBreakdown {
	var result RevenueBreakdown
	for _, item := range items {
		switch item.ItemType {
		case "PRODUCT":
			result.GrossProductCents += item.AmountCents
			result.ProductDiscountCents += item.DiscountCents
		default: // SERVICE
			result.GrossServiceCents += item.AmountCents
			result.ServiceDiscountCents += item.DiscountCents
		}
	}
	result.TotalDiscountCents = result.ServiceDiscountCents + result.ProductDiscountCents
	result.NetServiceCents = result.GrossServiceCents - result.ServiceDiscountCents
	result.NetProductCents = result.GrossProductCents - result.ProductDiscountCents
	if result.NetServiceCents < 0 {
		result.NetServiceCents = 0
	}
	if result.NetProductCents < 0 {
		result.NetProductCents = 0
	}
	return result
}

// CalculateProfitSharing computes the profit allocation for a branch/month
// based on net service revenue and the configured percentages.
func CalculateProfitSharing(netServiceCents int64, config ProfitSharingConfig) ProfitSharingResult {
	result := ProfitSharingResult{
		NetServiceRevenueCents: netServiceCents,
		OwnerPercentage:        config.OwnerPercentage,
	}

	// Calculate owner share
	result.OwnerShareCents = int64(float64(netServiceCents) * config.OwnerPercentage / 100.0)

	// Calculate employee shares
	var totalEmployeePercent float64
	for _, rule := range config.EmployeeRules {
		share := EmployeeShare{
			UserID:     rule.UserID,
			Name:       rule.Name,
			Percentage: rule.Percentage,
			ShareCents: int64(float64(netServiceCents) * rule.Percentage / 100.0),
		}
		result.EmployeeShares = append(result.EmployeeShares, share)
		totalEmployeePercent += rule.Percentage
	}

	// Calculate unallocated (reserve) percentage and amount
	result.UnallocatedPercentage = 100.0 - config.OwnerPercentage - totalEmployeePercent
	if result.UnallocatedPercentage < 0 {
		result.UnallocatedPercentage = 0
	}
	result.UnallocatedCents = int64(float64(netServiceCents) * result.UnallocatedPercentage / 100.0)

	return result
}

// CalculateProductCommissions aggregates product commissions per employee.
func CalculateProductCommissions(items []TransactionItem) map[int64]int64 {
	commissions := make(map[int64]int64)
	for _, item := range items {
		if item.ItemType == "PRODUCT" && item.BarberID > 0 && item.CommissionRate > 0 {
			qty := item.Quantity
			if qty < 1 {
				qty = 1
			}
			commissions[item.BarberID] += item.CommissionRate * int64(qty)
		}
	}
	return commissions
}

// CalculateTakeHomePay computes the final take-home pay for an employee.
func CalculateTakeHomePay(serviceShare, productCommission, bonus, deduction int64) int64 {
	total := serviceShare + productCommission + bonus - deduction
	if total < 0 {
		total = 0
	}
	return total
}

// ValidateProfitSharingConfig checks that the total allocation doesn't exceed 100%.
func ValidateProfitSharingConfig(config ProfitSharingConfig) (float64, error) {
	total := config.OwnerPercentage
	for _, rule := range config.EmployeeRules {
		total += rule.Percentage
	}
	remaining := 100.0 - total
	return remaining, nil
}
