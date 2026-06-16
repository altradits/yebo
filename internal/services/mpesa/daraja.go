package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	sandboxBase = "https://sandbox.safaricom.co.ke"
	prodBase    = "https://api.safaricom.co.ke"
)

// Client is a Daraja API client with automatic token refresh.
type Client struct {
	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	baseURL     string
	http        *http.Client
}

var Default *Client

// Init creates the default Daraja client from environment variables.
func Init() {
	base := sandboxBase
	if os.Getenv("MPESA_ENV") == "production" {
		base = prodBase
	}
	Default = &Client{
		baseURL: base,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) getToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}
	key := os.Getenv("MPESA_CONSUMER_KEY")
	secret := os.Getenv("MPESA_CONSUMER_SECRET")
	creds := base64.StdEncoding.EncodeToString([]byte(key + ":" + secret))

	req, _ := http.NewRequest("GET", c.baseURL+"/oauth/v1/generate?grant_type=client_credentials", nil)
	req.Header.Set("Authorization", "Basic "+creds)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("daraja: token: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Token     string `json:"access_token"`
		ExpiresIn string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	c.token = result.Token
	c.tokenExpiry = time.Now().Add(55 * time.Minute)
	return c.token, nil
}

func (c *Client) post(path string, body interface{}) ([]byte, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daraja: %s: %w", path, err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
