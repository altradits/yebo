package interest

import "testing"

func TestMonthlyInterest(t *testing.T) {
	cases := []struct {
		principal int64
		rateBPS   int
		want      int64
	}{
		{1_000_000, 1200, 10000}, // 1M sats @ 12% pa = 10k sats/month
		{100_000, 600, 500},      // 100k sats @ 6% pa = 500 sats/month
		{0, 1200, 0},
		{1_000_000, 0, 0},
	}
	for _, c := range cases {
		got := monthlyInterest(c.principal, c.rateBPS)
		if got != c.want {
			t.Errorf("monthlyInterest(%d, %d) = %d, want %d", c.principal, c.rateBPS, got, c.want)
		}
	}
}
