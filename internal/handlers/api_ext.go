package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/mpesa"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

// ── Chamas ────────────────────────────────────────────────────────────────────

// APIGetChamas returns all chamas the authenticated user belongs to.
func APIGetChamas(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	btcKES, _ := rates.GetBTCKES()

	rows, err := db.DB.Query(`
		SELECT c.id, c.name, COALESCE(c.description,''), COALESCE(w.balance_sats,0),
		       c.status, cm.role,
		       (SELECT COUNT(*) FROM chama_members WHERE chama_id=c.id) AS member_count,
		       c.contribution_sats, c.cycle_days
		FROM chamas c
		JOIN chama_members cm ON cm.chama_id = c.id AND cm.user_id = $1
		JOIN wallets w ON w.id = c.wallet_id
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	defer rows.Close()

	type chama struct {
		ID               int64   `json:"id"`
		Name             string  `json:"name"`
		Description      string  `json:"description"`
		BalanceSats      int64   `json:"balance_sats"`
		BalanceKES       float64 `json:"balance_kes"`
		Status           string  `json:"status"`
		Role             string  `json:"role"`
		MemberCount      int     `json:"member_count"`
		ContributionSats int64   `json:"contribution_sats"`
		CycleDays        int     `json:"cycle_days"`
		BtcKES           float64 `json:"btc_kes"`
	}
	var chamas []chama
	for rows.Next() {
		var c chama
		rows.Scan(&c.ID, &c.Name, &c.Description, &c.BalanceSats, &c.Status, &c.Role, &c.MemberCount, &c.ContributionSats, &c.CycleDays) //nolint:errcheck
		c.BtcKES = btcKES
		if btcKES > 0 {
			c.BalanceKES = float64(c.BalanceSats) / 1e8 * btcKES
		}
		chamas = append(chamas, c)
	}
	if chamas == nil {
		chamas = []chama{}
	}
	jsonOK(w, chamas)
}

// APIGetChama returns a single chama with recent transactions.
func APIGetChama(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	btcKES, _ := rates.GetBTCKES()

	// Extract id from path: /api/chamas/{id}
	idStr := r.URL.Path[len("/api/chamas/"):]
	chamaID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid chama id")
		return
	}

	// Check membership
	var role string
	err = db.DB.QueryRow(`SELECT role FROM chama_members WHERE chama_id=$1 AND user_id=$2`, chamaID, userID).Scan(&role)
	if err != nil {
		jsonErr(w, http.StatusNotFound, "chama not found")
		return
	}

	var id int64
	var name, description, status string
	var balanceSats, contributionSats int64
	var memberCount int
	var cycleDays int
	var createdAt time.Time
	db.DB.QueryRow(`
		SELECT c.id, c.name, COALESCE(c.description,''), COALESCE(w.balance_sats,0),
		       c.status, c.contribution_sats, c.cycle_days, c.created_at,
		       (SELECT COUNT(*) FROM chama_members WHERE chama_id=c.id)
		FROM chamas c
		JOIN wallets w ON w.id = c.wallet_id
		WHERE c.id = $1
	`, chamaID).Scan(&id, &name, &description, &balanceSats, &status, &contributionSats, &cycleDays, &createdAt, &memberCount) //nolint:errcheck

	// Members
	memberRows, _ := db.DB.Query(`
		SELECT u.id, COALESCE(u.full_name, u.phone), cm.role, cm.joined_at
		FROM chama_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.chama_id = $1
		ORDER BY cm.role DESC, cm.joined_at
	`, chamaID)
	defer memberRows.Close()
	type member struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		JoinedAt string `json:"joined_at"`
		IsMe     bool   `json:"is_me"`
	}
	var members []member
	for memberRows.Next() {
		var m member
		var joinedAt time.Time
		memberRows.Scan(&m.ID, &m.Name, &m.Role, &joinedAt) //nolint:errcheck
		m.JoinedAt = joinedAt.Format(time.RFC3339)
		m.IsMe = m.ID == userID
		members = append(members, m)
	}
	if members == nil {
		members = []member{}
	}

	balanceKES := 0.0
	if btcKES > 0 {
		balanceKES = float64(balanceSats) / 1e8 * btcKES
	}

	jsonOK(w, map[string]interface{}{
		"id":                chamaID,
		"name":              name,
		"description":       description,
		"balance_sats":      balanceSats,
		"balance_kes":       balanceKES,
		"btc_kes":           btcKES,
		"status":            status,
		"role":              role,
		"member_count":      memberCount,
		"contribution_sats": contributionSats,
		"cycle_days":        cycleDays,
		"created_at":        createdAt.Format(time.RFC3339),
		"members":           members,
	})
}

// ── Deposit ───────────────────────────────────────────────────────────────────

// APIDepositMpesa initiates an M-Pesa STK Push and returns the checkout_request_id.
func APIDepositMpesa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		AmountKES float64 `json:"amount_kes"`
		Phone     string  `json:"phone"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.AmountKES < 10 {
		jsonErr(w, http.StatusBadRequest, "minimum deposit is KES 10")
		return
	}
	phone := utils.NormalisePhone(req.Phone)
	if err := utils.ValidatePhone(phone); err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}
	walletID := middleware.WalletID(r)

	var mpesaID int64
	db.DB.QueryRow(`
		INSERT INTO mpesa_transactions (mpesa_receipt, type, phone, amount_kes, status, wallet_id, checkout_request_id)
		VALUES ('PENDING-'||extract(epoch from now())::bigint, 'stk_push', $1, $2, 'pending', $3, '') RETURNING id
	`, phone, req.AmountKES, walletID).Scan(&mpesaID) //nolint:errcheck

	stkResp, err := mpesa.STKPush(phone, int64(req.AmountKES), fmt.Sprintf("deposit_%d", mpesaID))
	if err != nil {
		jsonErr(w, http.StatusServiceUnavailable, "could not initiate M-Pesa payment")
		return
	}
	db.DB.Exec(`UPDATE mpesa_transactions SET checkout_request_id=$1 WHERE id=$2`, stkResp.CheckoutRequestID, mpesaID) //nolint:errcheck

	jsonOK(w, map[string]interface{}{
		"message":             fmt.Sprintf("STK Push sent to %s — enter your M-Pesa PIN", phone),
		"checkout_request_id": stkResp.CheckoutRequestID,
		"mpesa_transaction_id": mpesaID,
	})
}

// ── Withdraw ──────────────────────────────────────────────────────────────────

// APIWithdrawMpesa sends KES to the user's phone via M-Pesa B2C.
func APIWithdrawMpesa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		AmountKES float64 `json:"amount_kes"`
		Phone     string  `json:"phone"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.AmountKES < 10 {
		jsonErr(w, http.StatusBadRequest, "minimum withdrawal is KES 10")
		return
	}
	phone := utils.NormalisePhone(req.Phone)
	if err := utils.ValidatePhone(phone); err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}

	walletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)
	btcKES := rates.Global.GetKES()
	if btcKES <= 0 {
		jsonErr(w, http.StatusServiceUnavailable, "exchange rate unavailable")
		return
	}
	amountSats := utils.KESToSats(req.AmountKES, btcKES)

	tx, err := db.DB.Begin()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, walletID, amountSats, "withdrawal",
		"", fmt.Sprintf("M-Pesa withdrawal to %s", phone), &actorID); err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := mpesa.SendB2C(phone, int64(req.AmountKES), fmt.Sprintf("wallet_%d", walletID)); err != nil {
		jsonErr(w, http.StatusServiceUnavailable, "M-Pesa transfer failed: "+err.Error())
		return
	}
	tx.Commit() //nolint:errcheck

	jsonOK(w, map[string]interface{}{
		"message":     fmt.Sprintf("KES %.0f sent to %s", req.AmountKES, phone),
		"amount_sats": amountSats,
	})
}

// ── Savings ───────────────────────────────────────────────────────────────────

// APIGetSavings returns the user's savings locks and pool settings.
func APIGetSavings(w http.ResponseWriter, r *http.Request) {
	walletID := middleware.WalletID(r)
	btcKES, _ := rates.GetBTCKES()

	var rateBPS, lockDays int
	var minSats, maxSats int64
	db.DB.QueryRow(`SELECT interest_rate_bps, lock_days, min_savings_sats, max_savings_sats FROM pool_settings WHERE id=1`).
		Scan(&rateBPS, &lockDays, &minSats, &maxSats) //nolint:errcheck

	rows, _ := db.DB.Query(`
		SELECT id, amount_sats, lock_days, interest_rate_bps, locked_at, unlocks_at, status, interest_earned_sats
		FROM savings_locks WHERE wallet_id=$1 ORDER BY locked_at DESC
	`, walletID)
	defer rows.Close()

	type lock struct {
		ID             int64   `json:"id"`
		AmountSats     int64   `json:"amount_sats"`
		AmountKES      float64 `json:"amount_kes"`
		LockDays       int     `json:"lock_days"`
		RateBPS        int     `json:"rate_bps"`
		LockedAt       string  `json:"locked_at"`
		UnlocksAt      string  `json:"unlocks_at"`
		Status         string  `json:"status"`
		InterestEarned int64   `json:"interest_earned_sats"`
	}
	var locks []lock
	for rows.Next() {
		var l lock
		var lockedAt, unlocksAt time.Time
		rows.Scan(&l.ID, &l.AmountSats, &l.LockDays, &l.RateBPS, &lockedAt, &unlocksAt, &l.Status, &l.InterestEarned) //nolint:errcheck
		l.LockedAt = lockedAt.Format(time.RFC3339)
		l.UnlocksAt = unlocksAt.Format(time.RFC3339)
		if btcKES > 0 {
			l.AmountKES = float64(l.AmountSats) / 1e8 * btcKES
		}
		locks = append(locks, l)
	}
	if locks == nil {
		locks = []lock{}
	}

	jsonOK(w, map[string]interface{}{
		"locks":        locks,
		"pool_rate_bps": rateBPS,
		"pool_lock_days": lockDays,
		"min_sats":     minSats,
		"max_sats":     maxSats,
		"btc_kes":      btcKES,
	})
}

// APILockSavings creates a new savings lock.
func APILockSavings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		AmountSats int64 `json:"amount_sats"`
	}
	if err := decodeJSON(r, &req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.AmountSats <= 0 {
		jsonErr(w, http.StatusBadRequest, "invalid amount")
		return
	}

	walletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)

	var rateBPS, lockDays int
	var minSats, maxSats int64
	db.DB.QueryRow(`SELECT interest_rate_bps, lock_days, min_savings_sats, max_savings_sats FROM pool_settings WHERE id=1`).
		Scan(&rateBPS, &lockDays, &minSats, &maxSats) //nolint:errcheck

	if req.AmountSats < minSats {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("minimum lock is %d sats", minSats))
		return
	}
	if maxSats > 0 && req.AmountSats > maxSats {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("maximum lock is %d sats", maxSats))
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	unlocksAt := time.Now().AddDate(0, 0, lockDays)
	if err := db.DebitSats(tx, walletID, req.AmountSats, "savings_lock",
		"", fmt.Sprintf("Savings lock — %d days", lockDays), &actorID); err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}
	var lockID int64
	tx.QueryRow(`
		INSERT INTO savings_locks (wallet_id, amount_sats, lock_days, interest_rate_bps, unlocks_at, status)
		VALUES ($1, $2, $3, $4, $5, 'locked') RETURNING id
	`, walletID, req.AmountSats, lockDays, rateBPS, unlocksAt).Scan(&lockID) //nolint:errcheck

	if err := tx.Commit(); err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}

	jsonOK(w, map[string]interface{}{
		"message":    "Savings locked successfully",
		"lock_id":    lockID,
		"unlocks_at": unlocksAt.Format(time.RFC3339),
		"rate_bps":   rateBPS,
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
