package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

// ── Auth ──────────────────────────────────────────────────────────────────────

// APIRequestOTP sends a one-time code to the given phone number.
// For development, it prints the OTP to the server log instead of SMS.
func APIRequestOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	phone := utils.NormalisePhone(req.Phone)
	if err := utils.ValidatePhone(phone); err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}

	// Generate a 6-digit OTP and store it in a temp table.
	otp, err := utils.GenerateOTP()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	expires := time.Now().Add(10 * time.Minute)
	db.DB.Exec(`
		INSERT INTO otp_requests (phone, otp, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE SET otp=$2, expires_at=$3, created_at=NOW()
	`, phone, otp, expires) //nolint:errcheck

	// TODO: send via Africa's Talking or Twilio
	// For now, log to server — in prod remove this line
	println("OTP for", phone, ":", otp)

	jsonOK(w, map[string]string{"message": "Code sent"})
}

// APIVerifyOTP validates the OTP and returns a session token.
func APIVerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	phone := utils.NormalisePhone(req.Phone)

	// Verify OTP
	var storedOTP string
	var expiresAt time.Time
	err := db.DB.QueryRow(`
		SELECT otp, expires_at FROM otp_requests WHERE phone=$1
	`, phone).Scan(&storedOTP, &expiresAt)
	if err == sql.ErrNoRows || storedOTP != req.OTP || time.Now().After(expiresAt) {
		jsonErr(w, http.StatusUnauthorized, "Invalid or expired code")
		return
	}
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	db.DB.Exec(`DELETE FROM otp_requests WHERE phone=$1`, phone) //nolint:errcheck

	// Get or create user
	var userID int64
	var fullName string
	var role string
	err = db.DB.QueryRow(`SELECT id, COALESCE(full_name,''), role FROM users WHERE phone=$1`, phone).
		Scan(&userID, &fullName, &role)
	if err == sql.ErrNoRows {
		tx, _ := db.DB.Begin()
		defer tx.Rollback() //nolint:errcheck
		if err := tx.QueryRow(`
			INSERT INTO users (phone, password_hash, role) VALUES ($1, '', 'customer') RETURNING id
		`, phone).Scan(&userID); err != nil {
			jsonErr(w, http.StatusInternalServerError, "server error")
			return
		}
		tx.Exec(`INSERT INTO wallets (user_id) VALUES ($1)`, userID) //nolint:errcheck
		if err := tx.Commit(); err != nil {
			jsonErr(w, http.StatusInternalServerError, "server error")
			return
		}
		role = "customer"
	} else if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}

	// Create session
	token, err := utils.GenerateToken()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	db.DB.Exec(`INSERT INTO sessions (token, user_id) VALUES ($1, $2)`, token, userID) //nolint:errcheck
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(72 * time.Hour),
	})

	jsonOK(w, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":        userID,
			"phone":     phone,
			"full_name": fullName,
			"role":      role,
		},
	})
}

// APILogout invalidates the session.
func APILogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cookie, err := r.Cookie("session")
	if err == nil {
		db.DB.Exec(`DELETE FROM sessions WHERE token=$1`, cookie.Value) //nolint:errcheck
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
	jsonOK(w, map[string]string{"message": "logged out"})
}

// ── User ──────────────────────────────────────────────────────────────────────

// APIGetUser returns the authenticated user's profile + balance.
func APIGetUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	btcKES, _ := rates.GetBTCKES()

	var phone, fullName, role string
	var balanceSats, totalInterest int64
	var createdAt time.Time

	db.DB.QueryRow(`
		SELECT u.phone, COALESCE(u.full_name,''), u.role, COALESCE(w.balance_sats,0), u.created_at
		FROM users u LEFT JOIN wallets w ON w.user_id=u.id
		WHERE u.id=$1
	`, userID).Scan(&phone, &fullName, &role, &balanceSats, &createdAt) //nolint:errcheck

	db.DB.QueryRow(`
		SELECT COALESCE(SUM(interest_earned_sats),0) FROM savings_locks WHERE wallet_id=(SELECT id FROM wallets WHERE user_id=$1)
	`, userID).Scan(&totalInterest) //nolint:errcheck

	jsonOK(w, map[string]interface{}{
		"id":                    userID,
		"phone":                 phone,
		"full_name":             fullName,
		"role":                  role,
		"balance_sats":          balanceSats,
		"btc_kes":               btcKES,
		"total_interest_earned": totalInterest,
		"created_at":            createdAt,
	})
}

// APIGetBalance returns just the balance + current rate.
func APIGetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	btcKES, _ := rates.GetBTCKES()
	var balanceSats int64
	db.DB.QueryRow(`SELECT COALESCE(balance_sats,0) FROM wallets WHERE user_id=$1`, userID).
		Scan(&balanceSats) //nolint:errcheck
	jsonOK(w, map[string]interface{}{
		"balance_sats": balanceSats,
		"btc_kes":      btcKES,
	})
}

// APIGetTransactions returns paginated ledger history.
func APIGetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := db.DB.Query(`
		SELECT l.id, l.type, l.amount_sats, COALESCE(l.note,''), COALESCE(l.ref_id,''), l.created_at
		FROM ledger_entries l
		JOIN wallets w ON w.id = l.wallet_id
		WHERE w.user_id = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "server error")
		return
	}
	defer rows.Close()

	type tx struct {
		ID         int64     `json:"id"`
		Type       string    `json:"type"`
		AmountSats int64     `json:"amount_sats"`
		Note       string    `json:"note"`
		RefID      string    `json:"ref_id"`
		CreatedAt  time.Time `json:"created_at"`
		IsCredit   bool      `json:"is_credit"`
	}
	creditTypes := map[string]bool{
		"deposit": true, "receive": true, "savings_interest": true,
		"chama_payout": true, "agent_commission": true, "admin_adjustment": true,
		"savings_unlock": true,
	}
	var txs []tx
	for rows.Next() {
		var t tx
		rows.Scan(&t.ID, &t.Type, &t.AmountSats, &t.Note, &t.RefID, &t.CreatedAt) //nolint:errcheck
		t.IsCredit = creditTypes[t.Type]
		txs = append(txs, t)
	}
	if txs == nil {
		txs = []tx{}
	}
	jsonOK(w, txs)
}

// ── Community ─────────────────────────────────────────────────────────────────

// APICommunityStats returns aggregate platform stats.
func APICommunityStats(w http.ResponseWriter, r *http.Request) {
	btcKES, _ := rates.GetBTCKES()
	var memberCount int
	var totalSavings, totalInterest int64
	db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role='customer' AND is_active`).Scan(&memberCount)          //nolint:errcheck
	db.DB.QueryRow(`SELECT COALESCE(SUM(balance_sats),0) FROM wallets WHERE user_id IS NOT NULL`).Scan(&totalSavings) //nolint:errcheck
	db.DB.QueryRow(`SELECT COALESCE(SUM(total_interest_sats),0) FROM interest_distributions`).Scan(&totalInterest) //nolint:errcheck
	jsonOK(w, map[string]interface{}{
		"member_count":              memberCount,
		"total_savings_sats":        totalSavings,
		"total_interest_paid_sats":  totalInterest,
		"btc_kes":                   btcKES,
	})
}
