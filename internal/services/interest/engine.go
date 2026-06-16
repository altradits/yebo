package interest

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/yebobank/yebobank/internal/db"
)

// Distribute runs one interest distribution cycle.
// It credits interest to every active savings lock proportionally.
// Interest comes ONLY from treasury_assets — never from deposits.
func Distribute(runByID *int64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// Advisory lock prevents concurrent runs (e.g. two pods in Docker Compose).
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(1234567890)`); err != nil {
		return fmt.Errorf("interest: advisory lock: %w", err)
	}

	rateBPS, err := getRate(tx)
	if err != nil {
		return err
	}

	type lock struct {
		id       int64
		walletID int64
		sats     int64
	}
	rows, err := tx.Query(`
		SELECT id, wallet_id, amount_sats
		FROM savings_locks
		WHERE status = 'active'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var locks []lock
	var totalLocked int64
	for rows.Next() {
		var l lock
		if err := rows.Scan(&l.id, &l.walletID, &l.sats); err != nil {
			return err
		}
		locks = append(locks, l)
		totalLocked += l.sats
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Annual rate ÷ 12 for monthly distribution.
	// monthlyRateBPS = rateBPS / 12, applied to each lock's principal.
	var totalInterest int64
	for _, l := range locks {
		interest := monthlyInterest(l.sats, rateBPS)
		if interest <= 0 {
			continue
		}
		if err := db.CreditSats(tx, l.walletID, interest,
			"savings_interest", "", fmt.Sprintf("Monthly interest on lock %d", l.id), runByID); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			UPDATE savings_locks SET interest_earned_sats = interest_earned_sats + $1 WHERE id = $2
		`, interest, l.id); err != nil {
			return err
		}
		totalInterest += interest
	}

	_, err = tx.Exec(`
		INSERT INTO interest_distributions
			(total_locked_sats, total_interest_sats, treasury_profit_sats, accounts_credited, rate_bps, run_by)
		VALUES ($1, $2, 0, $3, $4, $5)
	`, totalLocked, totalInterest, len(locks), rateBPS, runByID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// monthlyInterest computes one month's interest for a principal at an annual rate in BPS.
func monthlyInterest(principal int64, annualBPS int) int64 {
	// monthly = principal * (annualBPS / 10000) / 12
	return int64(float64(principal) * float64(annualBPS) / 10000.0 / 12.0)
}

func getRate(tx *sql.Tx) (int, error) {
	var bps int
	err := tx.QueryRow(`SELECT interest_rate_bps FROM pool_settings WHERE id = 1`).Scan(&bps)
	if err != nil {
		return 0, fmt.Errorf("interest: get rate: %w", err)
	}
	return bps, nil
}

// UnlockMatured releases savings locks whose unlock date has passed.
func UnlockMatured() error {
	rows, err := db.DB.Query(`
		SELECT id, wallet_id, amount_sats
		FROM savings_locks
		WHERE status = 'active' AND unlocks_at <= $1
	`, time.Now())
	if err != nil {
		return err
	}
	defer rows.Close()

	type lock struct{ id, walletID, sats int64 }
	var locks []lock
	for rows.Next() {
		var l lock
		if err := rows.Scan(&l.id, &l.walletID, &l.sats); err != nil {
			return err
		}
		locks = append(locks, l)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, l := range locks {
		tx, err := db.DB.Begin()
		if err != nil {
			return err
		}
		if err := db.CreditSats(tx, l.walletID, l.sats, "savings_unlock",
			"", fmt.Sprintf("Matured lock %d", l.id), nil); err != nil {
			tx.Rollback() //nolint:errcheck
			return err
		}
		if _, err := tx.Exec(`
			UPDATE savings_locks SET status='unlocked', unlocked_at=NOW() WHERE id=$1
		`, l.id); err != nil {
			tx.Rollback() //nolint:errcheck
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
