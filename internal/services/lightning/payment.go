package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PaymentResult struct {
	PaymentHash     string `json:"payment_hash"`
	PaymentPreimage string `json:"payment_preimage"`
	FeeMsat         string `json:"fee_msat"`
	Status          string `json:"status"`
	FailureReason   string `json:"failure_reason"`
}

// SendPayment pays a Lightning invoice. Amount in sats (only used for zero-amount invoices).
func SendPayment(paymentRequest string, feeLimitSats int64) (*PaymentResult, error) {
	if Default == nil {
		return nil, fmt.Errorf("lightning: client not initialised")
	}
	body := map[string]interface{}{
		"payment_request": paymentRequest,
		"fee_limit":       map[string]int64{"fixed": feeLimitSats},
		"timeout_seconds": 60,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", Default.baseURL+"/v2/router/send", bytes.NewReader(b))
	req.Header.Set("Grpc-Metadata-macaroon", Default.macaroon)
	req.Header.Set("Content-Type", "application/json")

	resp, err := Default.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lightning: send payment: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lightning: send payment: HTTP %d: %s", resp.StatusCode, body)
	}
	var result PaymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Status != "SUCCEEDED" {
		return nil, fmt.Errorf("lightning: payment failed: %s", result.FailureReason)
	}
	return &result, nil
}

// DecodeInvoice decodes a BOLT11 payment request and returns the amount in sats.
func DecodeInvoice(paymentRequest string) (amountSats int64, err error) {
	if Default == nil {
		return 0, fmt.Errorf("lightning: client not initialised")
	}
	req, _ := http.NewRequest("GET",
		Default.baseURL+"/v1/payreq/"+paymentRequest, nil)
	req.Header.Set("Grpc-Metadata-macaroon", Default.macaroon)
	resp, err := Default.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		NumSatoshis string `json:"num_satoshis"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	fmt.Sscanf(result.NumSatoshis, "%d", &amountSats)
	return amountSats, nil
}
