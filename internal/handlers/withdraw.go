package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/middleware"
	"github.com/yebobank/yebobank/internal/services/lightning"
	"github.com/yebobank/yebobank/internal/services/mpesa"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

func Withdraw(w http.ResponseWriter, r *http.Request) {
	walletID := middleware.WalletID(r)
	bal, _ := db.BalanceSats(walletID)
	btcKES := rates.Global.GetKES()
	renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{
		"BalanceSats": bal,
		"BalanceKES": utils.SatsToKES(bal, btcKES),
		"BtcKES":    btcKES,
	})
}

// WithdrawMpesa sends KES to the customer's M-Pesa phone via B2C.
func WithdrawMpesa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/withdraw", http.StatusSeeOther)
		return
	}
	walletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)

	amountKES, err := strconv.ParseFloat(r.FormValue("amount_kes"), 10)
	if err != nil || amountKES < 10 {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{"Error": "Minimum withdrawal is KES 10"})
		return
	}
	phone := utils.NormalisePhone(r.FormValue("phone"))

	btcKES := rates.Global.GetKES()
	amountSats := utils.KESToSats(amountKES, btcKES)

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, walletID, amountSats, "withdrawal",
		"", fmt.Sprintf("M-Pesa withdrawal to %s", phone), &actorID); err != nil {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	_, err2 := mpesa.SendB2C(phone, int64(amountKES), fmt.Sprintf("wallet_%d", walletID))
	if err2 != nil {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{
			"Error": "M-Pesa transfer failed: " + err2.Error(),
		})
		return
	}
	tx.Commit() //nolint:errcheck
	renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{
		"Success": fmt.Sprintf("KES %.0f sent to %s", amountKES, phone),
	})
}

// WithdrawLightning pays a Lightning invoice from the customer's wallet.
func WithdrawLightning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/withdraw", http.StatusSeeOther)
		return
	}
	walletID := middleware.WalletID(r)
	actorID := middleware.UserID(r)
	payReq := r.FormValue("payment_request")

	amountSats, err := lightning.DecodeInvoice(payReq)
	if err != nil {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{"Error": "Invalid invoice: " + err.Error()})
		return
	}

	tx, _ := db.DB.Begin()
	defer tx.Rollback() //nolint:errcheck

	if err := db.DebitSats(tx, walletID, amountSats, "send",
		"", "Lightning payment", &actorID); err != nil {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	result, err := lightning.SendPayment(payReq, 10)
	if err != nil {
		renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	tx.Commit() //nolint:errcheck
	renderTemplate(w, r, "customer/withdraw.html", map[string]interface{}{
		"Success": fmt.Sprintf("Payment sent! Hash: %s", result.PaymentHash),
	})
}
