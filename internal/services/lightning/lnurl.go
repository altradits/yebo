package lightning

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// LNURLPayResponse is the first response in the LNURL-pay flow.
type LNURLPayResponse struct {
	Callback    string   `json:"callback"`
	MinSendable int64    `json:"minSendable"` // millisats
	MaxSendable int64    `json:"maxSendable"` // millisats
	Metadata    string   `json:"metadata"`
	Tag         string   `json:"tag"`
}

// LNURLPayCallbackResponse is the second response containing the invoice.
type LNURLPayCallbackResponse struct {
	PR     string `json:"pr"`
	Routes []interface{} `json:"routes"`
}

// HandleLNURLPay serves the first step of LNURL-pay for a given user.
func HandleLNURLPay(w http.ResponseWriter, r *http.Request, username string) {
	domain := os.Getenv("DOMAIN")
	resp := LNURLPayResponse{
		Tag:         "payRequest",
		Callback:    fmt.Sprintf("https://%s/lnurl/pay/%s/callback", domain, username),
		MinSendable: 1000,            // 1 sat minimum
		MaxSendable: 10_000_000_000,  // 100,000 sats maximum
		Metadata:    fmt.Sprintf(`[["text/plain","Pay %s via YeboBank"]]`, username),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// LightningAddress resolves a Lightning Address (user@domain) to a LNURL-pay endpoint.
// Lightning Address format: https://domain/.well-known/lnurlp/username
func LightningAddress(address string) (*LNURLPayResponse, error) {
	parts := strings.SplitN(address, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("lnurl: invalid lightning address: %s", address)
	}
	url := fmt.Sprintf("https://%s/.well-known/lnurlp/%s", parts[1], parts[0])
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil, fmt.Errorf("lnurl: resolve %s: %w", address, err)
	}
	defer resp.Body.Close()
	var pay LNURLPayResponse
	if err := json.NewDecoder(resp.Body).Decode(&pay); err != nil {
		return nil, err
	}
	return &pay, nil
}
