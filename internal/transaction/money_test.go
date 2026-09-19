package transaction

import "testing"

func TestParseCents(t *testing.T) {
	tests := []struct {
		in   string
		want int64
		ok   bool
	}{{"12", 1200, true}, {"12.3", 1230, true}, {"12.34", 1234, true}, {"0.01", 1, true}, {"0", 0, false}, {"1.234", 0, false}, {"-1", 0, false}, {"1e2", 0, false}}
	for _, tt := range tests {
		got, err := ParseCents(tt.in)
		if (err == nil) != tt.ok || got != tt.want {
			t.Errorf("ParseCents(%q)=(%d,%v), want (%d,ok=%v)", tt.in, got, err, tt.want, tt.ok)
		}
	}
}

func TestEntryDiscountCalculations(t *testing.T) {
	entry := Entry{
		AmountCents: 3600000,
		Items: []Item{
			{
				Category:       "Haircut",
				ItemName:       "Regular",
				AmountCents:    4000000,
				DiscountAmount: 400000,
			},
		},
	}

	if entry.TotalDiscountCents() != 400000 {
		t.Fatalf("expected TotalDiscountCents=400000, got %d", entry.TotalDiscountCents())
	}
	if entry.GrossSubtotalCents() != 4000000 {
		t.Fatalf("expected GrossSubtotalCents=4000000, got %d", entry.GrossSubtotalCents())
	}

	noDiscEntry := Entry{
		AmountCents: 5000000,
		Items: []Item{
			{
				Category:       "Shampoo",
				AmountCents:    5000000,
				DiscountAmount: 0,
			},
		},
	}
	if noDiscEntry.TotalDiscountCents() != 0 {
		t.Fatalf("expected TotalDiscountCents=0, got %d", noDiscEntry.TotalDiscountCents())
	}
	if noDiscEntry.GrossSubtotalCents() != 5000000 {
		t.Fatalf("expected GrossSubtotalCents=5000000, got %d", noDiscEntry.GrossSubtotalCents())
	}
}

