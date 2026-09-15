package backoffice

import (
	"testing"
)

func TestSegregateRevenue(t *testing.T) {
	items := []TransactionItem{
		{ItemType: "SERVICE", AmountCents: 4000000, DiscountCents: 0},
		{ItemType: "SERVICE", AmountCents: 300000, DiscountCents: 0},
		{ItemType: "SERVICE", AmountCents: 6000000, DiscountCents: 200000},
		{ItemType: "PRODUCT", AmountCents: 12000000, DiscountCents: 0},
		{ItemType: "PRODUCT", AmountCents: 5000000, DiscountCents: 100000},
	}
	result := SegregateRevenue(items)
	if result.GrossServiceCents != 10300000 {
		t.Errorf("GrossServiceCents = %d, want 10300000", result.GrossServiceCents)
	}
	if result.GrossProductCents != 17000000 {
		t.Errorf("GrossProductCents = %d, want 17000000", result.GrossProductCents)
	}
	if result.NetServiceCents != 10100000 {
		t.Errorf("NetServiceCents = %d, want 10100000", result.NetServiceCents)
	}
	if result.NetProductCents != 16900000 {
		t.Errorf("NetProductCents = %d, want 16900000", result.NetProductCents)
	}
	if result.TotalDiscountCents != 300000 {
		t.Errorf("TotalDiscountCents = %d, want 300000", result.TotalDiscountCents)
	}
}

func TestSegregateRevenue_ServiceOnly(t *testing.T) {
	items := []TransactionItem{
		{ItemType: "SERVICE", AmountCents: 10000000},
	}
	result := SegregateRevenue(items)
	if result.NetServiceCents != 10000000 {
		t.Errorf("NetServiceCents = %d, want 10000000", result.NetServiceCents)
	}
	if result.NetProductCents != 0 {
		t.Errorf("NetProductCents = %d, want 0", result.NetProductCents)
	}
}

func TestSegregateRevenue_Empty(t *testing.T) {
	result := SegregateRevenue(nil)
	if result.NetServiceCents != 0 || result.NetProductCents != 0 {
		t.Errorf("expected zero for empty items")
	}
}

// TestProfitSharing_Klaseman tests the Klaseman branch scenario from the TODO:
// Omzet Jasa: Rp 10.000.000 (= 1000000000 cents)
// Karyawan 1 (35%): Rp 3.500.000
// Karyawan 2 (35%): Rp 3.500.000
// Owner (20%): Rp 2.000.000
// Unallocated (10%): Rp 1.000.000
func TestProfitSharing_Klaseman(t *testing.T) {
	// Rp 10.000.000 = 1000000000 cents (amount_cents uses Rp * 100)
	netService := int64(1000000000) // Rp 10.000.000,00
	config := ProfitSharingConfig{
		BranchID:        1,
		PeriodMonth:     "2026-01",
		OwnerPercentage: 20,
		EmployeeRules: []ProfitSharingRule{
			{UserID: 1, Name: "Karyawan 1", Percentage: 35},
			{UserID: 2, Name: "Karyawan 2", Percentage: 35},
		},
	}

	result := CalculateProfitSharing(netService, config)

	// Owner: 20% of Rp 10.000.000 = Rp 2.000.000
	expectedOwner := int64(200000000) // Rp 2.000.000,00
	if result.OwnerShareCents != expectedOwner {
		t.Errorf("OwnerShareCents = %d, want %d (Rp 2.000.000)", result.OwnerShareCents, expectedOwner)
	}

	// Karyawan 1: 35% = Rp 3.500.000
	expectedEmp1 := int64(350000000) // Rp 3.500.000,00
	if len(result.EmployeeShares) < 1 || result.EmployeeShares[0].ShareCents != expectedEmp1 {
		actual := int64(0)
		if len(result.EmployeeShares) > 0 {
			actual = result.EmployeeShares[0].ShareCents
		}
		t.Errorf("Employee 1 ShareCents = %d, want %d (Rp 3.500.000)", actual, expectedEmp1)
	}

	// Karyawan 2: 35% = Rp 3.500.000
	expectedEmp2 := int64(350000000) // Rp 3.500.000,00
	if len(result.EmployeeShares) < 2 || result.EmployeeShares[1].ShareCents != expectedEmp2 {
		actual := int64(0)
		if len(result.EmployeeShares) > 1 {
			actual = result.EmployeeShares[1].ShareCents
		}
		t.Errorf("Employee 2 ShareCents = %d, want %d (Rp 3.500.000)", actual, expectedEmp2)
	}

	// Unallocated: 10% = Rp 1.000.000
	expectedUnalloc := int64(100000000) // Rp 1.000.000,00
	if result.UnallocatedCents != expectedUnalloc {
		t.Errorf("UnallocatedCents = %d, want %d (Rp 1.000.000)", result.UnallocatedCents, expectedUnalloc)
	}
	if result.UnallocatedPercentage != 10.0 {
		t.Errorf("UnallocatedPercentage = %f, want 10.0", result.UnallocatedPercentage)
	}
}

// TestProfitSharing_Ledok tests the Ledok branch scenario from the TODO:
// Omzet Jasa: Rp 12.000.000
// Karyawan 1 (40%): Rp 4.800.000
// Karyawan 2 (30%): Rp 3.600.000
// Owner (20%): Rp 2.400.000
// Unallocated (10%): Rp 1.200.000
func TestProfitSharing_Ledok(t *testing.T) {
	netService := int64(1200000000) // Rp 12.000.000,00
	config := ProfitSharingConfig{
		BranchID:        2,
		PeriodMonth:     "2026-01",
		OwnerPercentage: 20,
		EmployeeRules: []ProfitSharingRule{
			{UserID: 3, Name: "Karyawan 1", Percentage: 40},
			{UserID: 4, Name: "Karyawan 2", Percentage: 30},
		},
	}

	result := CalculateProfitSharing(netService, config)

	// Owner: 20% of Rp 12.000.000 = Rp 2.400.000
	if result.OwnerShareCents != 240000000 {
		t.Errorf("OwnerShareCents = %d, want 240000000", result.OwnerShareCents)
	}

	// Karyawan 1: 40% = Rp 4.800.000
	if len(result.EmployeeShares) < 1 || result.EmployeeShares[0].ShareCents != 480000000 {
		t.Errorf("Employee 1 ShareCents = %d, want 480000000", result.EmployeeShares[0].ShareCents)
	}

	// Karyawan 2: 30% = Rp 3.600.000
	if len(result.EmployeeShares) < 2 || result.EmployeeShares[1].ShareCents != 360000000 {
		t.Errorf("Employee 2 ShareCents = %d, want 360000000", result.EmployeeShares[1].ShareCents)
	}

	// Unallocated: 10% = Rp 1.200.000
	if result.UnallocatedCents != 120000000 {
		t.Errorf("UnallocatedCents = %d, want 120000000", result.UnallocatedCents)
	}
}

// TestProductCommissions tests product commission calculation.
// Klaseman scenario:
//   Karyawan 1: 10 pcs Pomade @ Rp 5.000 = Rp 50.000
//   Karyawan 2: 5 pcs Tonic @ Rp 10.000 = Rp 50.000
func TestProductCommissions_Klaseman(t *testing.T) {
	items := []TransactionItem{
		{ItemType: "PRODUCT", BarberID: 1, CommissionRate: 500000, Quantity: 10}, // Rp 5.000 * 10
		{ItemType: "PRODUCT", BarberID: 2, CommissionRate: 1000000, Quantity: 5}, // Rp 10.000 * 5
		{ItemType: "SERVICE", BarberID: 1, AmountCents: 4000000},                // service, no product commission
	}

	commissions := CalculateProductCommissions(items)

	// Karyawan 1: 10 * Rp 5.000 = Rp 50.000 = 5000000 cents
	if commissions[1] != 5000000 {
		t.Errorf("Karyawan 1 commission = %d, want 5000000 (Rp 50.000)", commissions[1])
	}

	// Karyawan 2: 5 * Rp 10.000 = Rp 50.000 = 5000000 cents
	if commissions[2] != 5000000 {
		t.Errorf("Karyawan 2 commission = %d, want 5000000 (Rp 50.000)", commissions[2])
	}
}

// TestProductCommissions_Ledok tests Ledok scenario:
//   Karyawan 1: 8 pcs Hair Clay @ Rp 7.500 = Rp 60.000
func TestProductCommissions_Ledok(t *testing.T) {
	items := []TransactionItem{
		{ItemType: "PRODUCT", BarberID: 3, CommissionRate: 750000, Quantity: 8}, // Rp 7.500 * 8
	}

	commissions := CalculateProductCommissions(items)

	// Karyawan 1: 8 * Rp 7.500 = Rp 60.000 = 6000000 cents
	if commissions[3] != 6000000 {
		t.Errorf("Karyawan 1 Ledok commission = %d, want 6000000 (Rp 60.000)", commissions[3])
	}
}

func TestProductCommissions_Empty(t *testing.T) {
	commissions := CalculateProductCommissions(nil)
	if len(commissions) != 0 {
		t.Errorf("expected empty commissions for nil items, got %d", len(commissions))
	}
}

func TestProductCommissions_ServiceItemsExcluded(t *testing.T) {
	items := []TransactionItem{
		{ItemType: "SERVICE", BarberID: 1, CommissionRate: 500000, Quantity: 10},
	}
	commissions := CalculateProductCommissions(items)
	if len(commissions) != 0 {
		t.Errorf("expected no commission for service items, got %d", len(commissions))
	}
}

// TestTakeHomePay_Klaseman tests the full take-home-pay calculation.
// Karyawan 1: Rp 3.500.000 (service) + Rp 50.000 (product) = Rp 3.550.000
func TestTakeHomePay_Klaseman(t *testing.T) {
	thp := CalculateTakeHomePay(350000000, 5000000, 0, 0) // Rp 3.500.000 + Rp 50.000
	expected := int64(355000000)                           // Rp 3.550.000,00
	if thp != expected {
		t.Errorf("TakeHomePay = %d, want %d (Rp 3.550.000)", thp, expected)
	}
}

// TestTakeHomePay_Ledok tests Ledok karyawan 1:
// Rp 4.800.000 (service) + Rp 60.000 (product) = Rp 4.860.000
func TestTakeHomePay_Ledok(t *testing.T) {
	thp := CalculateTakeHomePay(480000000, 6000000, 0, 0)
	expected := int64(486000000) // Rp 4.860.000,00
	if thp != expected {
		t.Errorf("TakeHomePay = %d, want %d (Rp 4.860.000)", thp, expected)
	}
}

func TestTakeHomePay_WithDeductions(t *testing.T) {
	thp := CalculateTakeHomePay(350000000, 5000000, 1000000, 2000000)
	expected := int64(354000000) // 350000000 + 5000000 + 1000000 - 2000000
	if thp != expected {
		t.Errorf("TakeHomePay = %d, want %d", thp, expected)
	}
}

func TestTakeHomePay_NegativeClampedToZero(t *testing.T) {
	thp := CalculateTakeHomePay(0, 0, 0, 100000000)
	if thp != 0 {
		t.Errorf("TakeHomePay = %d, want 0 (should clamp negative to 0)", thp)
	}
}

func TestValidateProfitSharingConfig(t *testing.T) {
	config := ProfitSharingConfig{
		OwnerPercentage: 20,
		EmployeeRules: []ProfitSharingRule{
			{Percentage: 35},
			{Percentage: 35},
		},
	}
	remaining, err := ValidateProfitSharingConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 10.0 {
		t.Errorf("remaining = %f, want 10.0", remaining)
	}
}

func TestValidateProfitSharingConfig_FullAllocation(t *testing.T) {
	config := ProfitSharingConfig{
		OwnerPercentage: 30,
		EmployeeRules: []ProfitSharingRule{
			{Percentage: 35},
			{Percentage: 35},
		},
	}
	remaining, err := ValidateProfitSharingConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 0.0 {
		t.Errorf("remaining = %f, want 0.0", remaining)
	}
}

func TestProfitSharing_ZeroRevenue(t *testing.T) {
	config := ProfitSharingConfig{
		OwnerPercentage: 20,
		EmployeeRules: []ProfitSharingRule{
			{UserID: 1, Name: "Emp1", Percentage: 35},
		},
	}
	result := CalculateProfitSharing(0, config)
	if result.OwnerShareCents != 0 {
		t.Errorf("OwnerShareCents = %d, want 0", result.OwnerShareCents)
	}
	if len(result.EmployeeShares) != 1 || result.EmployeeShares[0].ShareCents != 0 {
		t.Errorf("employee share should be 0 for zero revenue")
	}
}
