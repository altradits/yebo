package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/altradits/yebo/internal/db"
	"github.com/altradits/yebo/internal/middleware"
	"github.com/altradits/yebo/internal/services/lightning"
	"github.com/altradits/yebo/internal/services/mpesa"
	"github.com/altradits/yebo/internal/services/rates"
	"github.com/altradits/yebo/internal/utils"
)

func Deposit(w http.ResponseWriter, r *http.Request) {
	btcKES := rates.Global.GetKES()
	renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{"BtcKES": btcKES})
}

// DepositMpesa initiates an M-Pesa STK Push for the given KES amount.
func DepositMpesa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/deposit", http.StatusSeeOther)
		return
	}
	amountKES, err := strconv.ParseFloat(r.FormValue("amount_kes"), 64)
	if err != nil || amountKES < 10 {
		renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{
			"Error": "Minimum deposit is KES 10",
		})
		return
	}
	phone := utils.NormalisePhone(r.FormValue("phone"))
	if err := utils.ValidatePhone(phone); err != nil {
		renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	walletID := middleware.WalletID(r)

	// Record pending transaction
	var mpesaID int64
	db.DB.QueryRow(`
		INSERT INTO mpesa_transactions (mpesa_receipt, type, phone, amount_kes, status, wallet_id, checkout_request_id)
		VALUES ('PENDING-'||gen_random_uuid(), 'stk_push', $1, $2, 'pending', $3, '') RETURNING id
	`, phone, amountKES, walletID).Scan(&mpesaID)

	stkResp, err := mpesa.STKPush(phone, int64(amountKES), fmt.Sprintf("deposit_%d", mpesaID))
	if err != nil {
		renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{
			"Error": "Could not initiate M-Pesa payment. Please try again.",
		})
		return
	}
	db.DB.Exec(`UPDATE mpesa_transactions SET checkout_request_id=$1 WHERE id=$2`, //nolint:errcheck
		stkResp.CheckoutRequestID, mpesaID)

	renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{
		"Success": fmt.Sprintf("STK Push sent to %s. Enter your M-Pesa PIN to complete.", phone),
	})
}

// DepositLightning creates a Lightning invoice for the given sats amount.
func DepositLightning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/deposit", http.StatusSeeOther)
		return
	}
	amountSats, err := strconv.ParseInt(r.FormValue("amount_sats"), 10, 64)
	if err != nil || amountSats <= 0 {
		renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{
			"Error": "Invalid amount",
		})
		return
	}
	walletID := middleware.WalletID(r)
	userID := middleware.UserID(r)

	inv, err := lightning.CreateInvoice(amountSats, "YeboBank deposit", 3600)
	if err != nil {
		renderTemplate(w, r, "customer/deposit.html", map[string]interface{}{
			"Error": "Could not create invoice. Lightning node may be offline.",
		})
		return
	}
	db.DB.Exec(`
		INSERT INTO ln_invoices (payment_hash, payment_request, amount_sats, wallet_id, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, inv.PaymentHash, inv.PaymentRequest, amountSats, walletID,
		time.Now().Add(time.Hour))

	_ = userID
	renderTemplate(w, r, "customer/receive.html", map[string]interface{}{
		"Invoice":     inv.PaymentRequest,
		"AmountSats":  amountSats,
		"PaymentHash": inv.PaymentHash,
	})
}
