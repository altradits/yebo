package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/yebobank/yebobank/internal/db"
	"github.com/yebobank/yebobank/internal/services/mpesa"
	"github.com/yebobank/yebobank/internal/services/rates"
	"github.com/yebobank/yebobank/internal/utils"
)

// MpesaSTKCallback handles the M-Pesa STK Push confirmation callback from Safaricom.
// This must be idempotent — Safaricom may send the same callback more than once.
func MpesaSTKCallback(w http.ResponseWriter, r *http.Request) {
	cb, err := mpesa.ParseSTKCallback(r)
	if err != nil {
		log.Printf("webhook: parse STK callback: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !cb.Succeeded() {
		log.Printf("webhook: STK Push failed: %s", cb.Body.StkCallback.ResultDesc)
		// Acknowledge to Safaricom even on failure
		fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
		return
	}
	meta := cb.Metadata()
	receipt, _ := meta["MpesaReceiptNumber"].(string)
	amountKES, _ := meta["Amount"].(float64)
	phone, _ := meta["PhoneNumber"].(string)

	if receipt == "" {
		log.Printf("webhook: STK callback missing receipt")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if mpesa.IsDuplicate(receipt) {
		fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
		return
	}

	// Find wallet by phone
	phone = utils.NormalisePhone(fmt.Sprint(phone))
	var walletID int64
	err = db.DB.QueryRow(`
		SELECT w.id FROM wallets w JOIN users u ON u.id=w.user_id WHERE u.phone=$1
	`, phone).Scan(&walletID)
	if err != nil {
		log.Printf("webhook: no wallet for phone %s: %v", phone, err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
		return
	}

	btcKES := rates.Global.GetKES()
	amountSats := utils.KESToSats(amountKES, btcKES)
	if amountSats <= 0 {
		log.Printf("webhook: zero sats conversion for KES %.2f at rate %.2f", amountKES, btcKES)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("webhook: begin tx: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() //nolint:errcheck

	if err := mpesa.MarkCompleted(tx, receipt, walletID, amountKES); err != nil {
		log.Printf("webhook: mark completed: %v", err)
		fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
		return
	}
	if err := db.CreditSats(tx, walletID, amountSats, "deposit", receipt,
		fmt.Sprintf("M-Pesa deposit KES %.2f", amountKES), nil); err != nil {
		log.Printf("webhook: credit sats: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("webhook: commit: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("webhook: credited %d sats to wallet %d (M-Pesa %s KES %.2f)", amountSats, walletID, receipt, amountKES)
	fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
}

// LNDInvoiceSettled handles the LND webhook when a Lightning invoice is settled.
// LND calls this with JSON body: {"payment_hash": "<hex>"}
func LNDInvoiceSettled(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PaymentHash string `json:"payment_hash"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil || body.PaymentHash == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var walletID, amountSats int64
	var settled bool
	err := db.DB.QueryRow(`
		SELECT wallet_id, amount_sats, settled FROM ln_invoices WHERE payment_hash=$1
	`, body.PaymentHash).Scan(&walletID, &amountSats, &settled)
	if err != nil {
		log.Printf("webhook: lnd: invoice not found: %s", body.PaymentHash)
		w.WriteHeader(http.StatusOK)
		return
	}
	if settled {
		w.WriteHeader(http.StatusOK)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("webhook: lnd: begin tx: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`UPDATE ln_invoices SET settled=true, settled_at=NOW() WHERE payment_hash=$1`, body.PaymentHash); err != nil {
		log.Printf("webhook: lnd: mark settled: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := db.CreditSats(tx, walletID, amountSats, "deposit", body.PaymentHash,
		"Lightning deposit", nil); err != nil {
		log.Printf("webhook: lnd: credit sats: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("webhook: lnd: commit: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("webhook: lnd: credited %d sats to wallet %d (hash %s)", amountSats, walletID, body.PaymentHash)
	w.WriteHeader(http.StatusOK)
}

// MpesaB2CCallback handles B2C payment result callbacks from Safaricom.
func MpesaB2CCallback(w http.ResponseWriter, r *http.Request) {
	res, err := mpesa.ParseB2CResult(r)
	if err != nil {
		log.Printf("webhook: b2c callback: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if res.Result.ResultCode != 0 {
		log.Printf("webhook: b2c failed: code=%d %s txn=%s",
			res.Result.ResultCode, res.Result.ResultDesc, res.Result.TransactionID)
	} else {
		log.Printf("webhook: b2c success: txn=%s", res.Result.TransactionID)
	}
	fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
}

// MpesaB2CTimeout handles B2C queue timeout callbacks from Safaricom.
func MpesaB2CTimeout(w http.ResponseWriter, r *http.Request) {
	log.Printf("webhook: b2c timeout received")
	fmt.Fprint(w, `{"ResultCode":0,"ResultDesc":"Accepted"}`)
}
