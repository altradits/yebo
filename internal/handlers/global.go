package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/services/lightning"
)

// Home serves the marketing landing page.
func Home(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "home.html", nil)
}

// LNURLPay serves the first step of the LNURL-pay flow for a username.
// Route: GET /.well-known/lnurlp/{username}
func LNURLPay(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/.well-known/lnurlp/")
	if username == "" {
		http.NotFound(w, r)
		return
	}
	var userID int64
	err := db.DB.QueryRow(`SELECT id FROM users WHERE phone=$1 OR email=$1`, username).Scan(&userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	lightning.HandleLNURLPay(w, r, username)
}

// LNURLPayCallback serves the invoice for the second step of LNURL-pay.
// Route: GET /lnurl/pay/{username}/callback?amount={msats}
func LNURLPayCallback(w http.ResponseWriter, r *http.Request) {
	// Extract username from path
	path := strings.TrimPrefix(r.URL.Path, "/lnurl/pay/")
	username := strings.TrimSuffix(path, "/callback")

	amountMsats := r.URL.Query().Get("amount")
	if amountMsats == "" {
		http.Error(w, `{"status":"ERROR","reason":"amount required"}`, http.StatusBadRequest)
		return
	}

	var amountMsat int64
	_, err := fmt.Sscanf(amountMsats, "%d", &amountMsat)
	if err != nil || amountMsat < 1000 {
		http.Error(w, `{"status":"ERROR","reason":"invalid amount"}`, http.StatusBadRequest)
		return
	}
	amountSats := amountMsat / 1000

	var walletID int64
	err = db.DB.QueryRow(`SELECT w.id FROM wallets w JOIN users u ON u.id=w.user_id WHERE u.phone=$1 OR u.email=$1`, username).
		Scan(&walletID)
	if err != nil {
		http.Error(w, `{"status":"ERROR","reason":"user not found"}`, http.StatusNotFound)
		return
	}

	inv, err := lightning.CreateInvoice(amountSats, "YeboBank payment", 300)
	if err != nil {
		http.Error(w, `{"status":"ERROR","reason":"could not create invoice"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"pr":%q,"routes":[]}`, inv.PaymentRequest)
}
