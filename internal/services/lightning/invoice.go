package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Invoice struct {
	PaymentRequest string `json:"payment_request"`
	PaymentHash    string `json:"r_hash"`
	AddIndex       string `json:"add_index"`
}

// CreateInvoice generates a new Lightning invoice on the LND node.
func CreateInvoice(amountSats int64, memo string, expirySeconds int64) (*Invoice, error) {
	if Default == nil {
		return nil, fmt.Errorf("lightning: client not initialised")
	}
	body := map[string]interface{}{
		"value":  amountSats,
		"memo":   memo,
		"expiry": expirySeconds,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", Default.baseURL+"/v1/invoices", bytes.NewReader(b))
	req.Header.Set("Grpc-Metadata-macaroon", Default.macaroon)
	req.Header.Set("Content-Type", "application/json")

	resp, err := Default.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lightning: create invoice: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lightning: create invoice: HTTP %d: %s", resp.StatusCode, body)
	}
	var inv Invoice
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// InvoiceStatus checks whether an invoice has been settled.
func InvoiceStatus(paymentHash string) (settled bool, err error) {
	if Default == nil {
		return false, nil
	}
	req, _ := http.NewRequest("GET", Default.baseURL+"/v1/invoice/"+paymentHash, nil)
	req.Header.Set("Grpc-Metadata-macaroon", Default.macaroon)
	resp, err := Default.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var result struct {
		Settled   bool   `json:"settled"`
		SettleDate string `json:"settle_date"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Settled, nil
}

// WatchInvoice polls until the invoice is settled or the timeout is reached.
func WatchInvoice(paymentHash string, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		settled, err := InvoiceStatus(paymentHash)
		if err != nil {
			return false, err
		}
		if settled {
			return true, nil
		}
		time.Sleep(2 * time.Second)
	}
	return false, nil
}
