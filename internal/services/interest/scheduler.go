package interest

import (
	"log"
	"time"
)

// StartScheduler launches a goroutine that runs Distribute on the 1st of each month.
func StartScheduler() {
	go func() {
		for {
			next := nextFirstOfMonth()
			log.Printf("interest: next distribution at %s", next.Format(time.RFC3339))
			time.Sleep(time.Until(next))
			if err := Distribute(nil); err != nil {
				log.Printf("interest: distribution error: %v", err)
			} else {
				log.Printf("interest: distribution complete")
			}
			// Sleep 25 hours to avoid firing twice on DST transitions.
			time.Sleep(25 * time.Hour)
		}
	}()
}

// StartUnlockChecker runs every 10 minutes to release matured savings locks.
func StartUnlockChecker() {
	go func() {
		for {
			if err := UnlockMatured(); err != nil {
				log.Printf("interest: unlock check error: %v", err)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}

func nextFirstOfMonth() time.Time {
	now := time.Now().UTC()
	// First of next month, midnight UTC.
	first := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	if first.Before(now) {
		first = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	}
	return first
}
