package mpesa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// STKCallback is the structure Safaricom POSTs to our callback URL.
type STKCallback struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  *struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// ParseSTKCallback reads and validates an STK Push callback from the request body.
func ParseSTKCallback(r *http.Request) (*STKCallback, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		return nil, fmt.Errorf("callback: read: %w", err)
	}
	var cb STKCallback
	if err := json.Unmarshal(body, &cb); err != nil {
		return nil, fmt.Errorf("callback: decode: %w", err)
	}
	return &cb, nil
}

// Metadata extracts key/value pairs from the callback metadata.
func (cb *STKCallback) Metadata() map[string]interface{} {
	m := make(map[string]interface{})
	if cb.Body.StkCallback.CallbackMetadata == nil {
		return m
	}
	for _, item := range cb.Body.StkCallback.CallbackMetadata.Item {
		m[item.Name] = item.Value
	}
	return m
}

// Succeeded returns true if the STK Push was successful.
func (cb *STKCallback) Succeeded() bool {
	return cb.Body.StkCallback.ResultCode == 0
}

// B2CResult is the structure Safaricom POSTs for B2C results.
type B2CResult struct {
	Result struct {
		ResultType               int    `json:"ResultType"`
		ResultCode               int    `json:"ResultCode"`
		ResultDesc               string `json:"ResultDesc"`
		OriginatorConversationID string `json:"OriginatorConversationID"`
		ConversationID           string `json:"ConversationID"`
		TransactionID            string `json:"TransactionID"`
	} `json:"Result"`
}

func ParseB2CResult(r *http.Request) (*B2CResult, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	var res B2CResult
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
