package utils

import (
	"fmt"
	"math"
	"time"
)

// FormatSats formats satoshis with thousands separator: "1,234,567 sats"
func FormatSats(sats int64) string {
	return fmt.Sprintf("%s sats", commaSep(sats))
}

// SatsToKES converts satoshis to KES using the current BTC/KES rate.
func SatsToKES(sats int64, btcKES float64) float64 {
	if btcKES <= 0 {
		return 0
	}
	return float64(sats) / 1e8 * btcKES
}

// KESToSats converts KES to satoshis using the current BTC/KES rate.
func KESToSats(kes, btcKES float64) int64 {
	if btcKES <= 0 {
		return 0
	}
	return int64(math.Round(kes / btcKES * 1e8))
}

// FormatKES formats a KES amount: "KES 4,320.50"
func FormatKES(kes float64) string {
	return fmt.Sprintf("KES %s", commaSepFloat(kes))
}

// TimeAgo returns a human-readable relative time string.
func TimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2 Jan 2006")
	}
}

func commaSep(n int64) string {
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		pos := len(s) - i
		if i > 0 && pos%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func commaSepFloat(f float64) string {
	whole := int64(f)
	frac := int(math.Round((f-float64(whole))*100)) % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%s.%02d", commaSep(whole), frac)
}
