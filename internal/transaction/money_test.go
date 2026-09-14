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
