package handlers

import (
	"fmt"
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

// LNDInvoiceSettled handles the LND webhook when an invoice is settled.
func LNDInvoiceSettled(w http.ResponseWriter, r *http.Request) {
	// LND streams invoice updates via REST SSE or webhook depending on configuration.
	// This endpoint is called by our own polling loop or LND webhook integration.
	// Full implementation: handlers/webhook.go — LND invoice settlement
	w.WriteHeader(http.StatusOK)
}
