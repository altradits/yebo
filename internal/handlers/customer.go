package handlers

import (
	"net/http"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	walletID := middleware.WalletID(r)
	userID := middleware.UserID(r)

	bal, err := db.BalanceSats(walletID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	btcKES := rates.Global.GetKES()

	var fullName, phone string
	db.DB.QueryRow(`SELECT full_name, phone FROM users WHERE id=$1`, userID).Scan(&fullName, &phone) //nolint:errcheck

	// Recent transactions
	rows, _ := db.DB.Query(`
		SELECT amount_sats, type, note, created_at
		FROM ledger_entries WHERE wallet_id=$1
		ORDER BY created_at DESC LIMIT 10
	`, walletID)
	defer rows.Close()
	type entry struct {
		AmountSats int64
		Type       string
		Note       string
		CreatedAt  string
		IsCredit   bool
	}
	var entries []entry
	for rows.Next() {
		var e entry
		var t string
		rows.Scan(&e.AmountSats, &e.Type, &e.Note, &t) //nolint:errcheck
		e.IsCredit = e.AmountSats > 0
		if e.AmountSats < 0 {
			e.AmountSats = -e.AmountSats
		}
		entries = append(entries, e)
	}

	renderTemplate(w, r, "customer/dashboard.html", map[string]interface{}{
		"FullName":   fullName,
		"Phone":      phone,
		"BalanceSats": bal,
		"BalanceKES": utils.SatsToKES(bal, btcKES),
		"BtcKES":    btcKES,
		"Entries":   entries,
	})
}

func History(w http.ResponseWriter, r *http.Request) {
	walletID := middleware.WalletID(r)
	rows, err := db.DB.Query(`
		SELECT amount_sats, type, ref_id, note, created_at
		FROM ledger_entries WHERE wallet_id=$1
		ORDER BY created_at DESC LIMIT 100
	`, walletID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type entry struct {
		AmountSats int64
		Type, RefID, Note, CreatedAt string
		IsCredit bool
	}
	var entries []entry
	for rows.Next() {
		var e entry
		rows.Scan(&e.AmountSats, &e.Type, &e.RefID, &e.Note, &e.CreatedAt) //nolint:errcheck
		e.IsCredit = e.AmountSats > 0
		if e.AmountSats < 0 {
			e.AmountSats = -e.AmountSats
		}
		entries = append(entries, e)
	}
	renderTemplate(w, r, "customer/history.html", map[string]interface{}{"Entries": entries})
}

func Settings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	if r.Method == http.MethodGet {
		var fullName, email, phone, language string
		db.DB.QueryRow(`SELECT full_name, email, phone, language FROM users WHERE id=$1`, userID).
			Scan(&fullName, &email, &phone, &language) //nolint:errcheck
		renderTemplate(w, r, "customer/settings.html", map[string]interface{}{
			"FullName": fullName, "Email": email, "Phone": phone, "Language": language,
		})
		return
	}
	fullName := r.FormValue("full_name")
	email := r.FormValue("email")
	language := r.FormValue("language")
	db.DB.Exec(`UPDATE users SET full_name=$1, email=$2, language=$3, updated_at=NOW() WHERE id=$4`,
		fullName, email, language, userID) //nolint:errcheck
	http.Redirect(w, r, "/settings?saved=1", http.StatusSeeOther)
}
