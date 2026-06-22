package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

func Savings(w http.ResponseWriter, r *http.Request) {
	walletID := middleware.WalletID(r)
	btcKES := rates.Global.GetKES()
	bal, _ := db.BalanceSats(walletID)

	rows, _ := db.DB.Query(`
		SELECT id, amount_sats, lock_days, interest_rate_bps, locked_at, unlocks_at, status, interest_earned_sats
		FROM savings_locks WHERE wallet_id=$1 ORDER BY locked_at DESC
	`, walletID)
	defer rows.Close()
	type lock struct {
		ID, AmountSats, InterestEarned int64
		LockDays, RateBPS              int
		LockedAt, UnlocksAt            string
		Status                         string
	}
	var locks []lock
	for rows.Next() {
		var l lock
		rows.Scan(&l.ID, &l.AmountSats, &l.LockDays, &l.RateBPS, &l.LockedAt, &l.UnlocksAt, &l.Status, &l.InterestEarned) //nolint:errcheck
		locks = append(locks, l)
	}
	renderTemplate(w, r, "customer/savings.html", map[string]interface{}{
		"BalanceSats": bal,
		"BalanceKES":  utils.SatsToKES(bal, btcKES),
		"BtcKES":      btcKES,
		"Locks":       locks,
	})
}

func SavingsLock(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var rateBPS, lockDays int
		var minSats, maxSats int64
		db.DB.QueryRow(`SELECT interest_rate_bps, lock_days, min_savings_sats, max_savings_sats FROM pool_settings WHERE id=1`).
			Scan(&rateBPS, &lockDays, &minSats, &maxSats) //nolint:errcheck
		renderTemplate(w, r, "customer/savings_lock.html", map[string]interface{}{
			"RateBPS": rateBPS, "LockDays": lockDays, "MinSats": minSats, "MaxSats": maxSats,
		})
		return
	}
	walletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)
	amountSats, err := strconv.ParseInt(r.FormValue("amount_sats"), 10, 64)
	if err != nil {
		renderTemplate(w, r, "customer/savings_lock.html", map[string]interface{}{"Error": "Invalid amount"})
		return
	}
	var rateBPS, lockDays int
	var minSats, maxSats int64
	db.DB.QueryRow(`SELECT interest_rate_bps, lock_days, min_savings_sats, max_savings_sats FROM pool_settings WHERE id=1`).
		Scan(&rateBPS, &lockDays, &minSats, &maxSats) //nolint:errcheck

	if err := utils.ValidateSatsAmount(amountSats, minSats, maxSats); err != nil {
		renderTemplate(w, r, "customer/savings_lock.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, walletID, amountSats, "savings_lock",
		"", fmt.Sprintf("Savings lock %d days", lockDays), &actorID); err != nil {
		renderTemplate(w, r, "customer/savings_lock.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	unlocksAt := time.Now().AddDate(0, 0, lockDays)
	tx.Exec(`
		INSERT INTO savings_locks (wallet_id, amount_sats, lock_days, interest_rate_bps, unlocks_at, early_exit_penalty_bps)
		VALUES ($1, $2, $3, $4, $5, 2000)
	`, walletID, amountSats, lockDays, rateBPS, unlocksAt) //nolint:errcheck
	tx.Commit() //nolint:errcheck
	http.Redirect(w, r, "/savings", http.StatusSeeOther)
}
