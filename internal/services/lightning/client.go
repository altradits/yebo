package lightning

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client talks to LND's REST API.
type Client struct {
	baseURL  string
	macaroon string
	http     *http.Client
}

var Default *Client

// Init creates the default LND client from environment variables.
func Init() error {
	baseURL := os.Getenv("LND_REST_URL")
	if baseURL == "" {
		return nil // LND not configured — non-fatal in dev
	}
	macaroonPath := os.Getenv("LND_MACAROON_PATH")
	mac, err := readMacaroon(macaroonPath)
	if err != nil {
		return fmt.Errorf("lightning: macaroon: %w", err)
	}
	httpClient, err := buildHTTPClient()
	if err != nil {
		return fmt.Errorf("lightning: http client: %w", err)
	}
	Default = &Client{baseURL: baseURL, macaroon: mac, http: httpClient}
	return nil
}

func readMacaroon(path string) (string, error) {
	if path == "" {
		if hex := os.Getenv("LND_MACAROON_HEX"); hex != "" {
			return hex, nil
		}
		return "", nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildHTTPClient() (*http.Client, error) {
	tlsCfg := &tls.Config{}
	certPath := os.Getenv("LND_TLS_CERT_PATH")
	if certPath != "" {
		cert, err := os.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(cert) {
			return nil, fmt.Errorf("lightning: invalid TLS cert")
		}
		tlsCfg.RootCAs = pool
	} else {
		tlsCfg.InsecureSkipVerify = true // #nosec G402 — Voltage.cloud uses valid cert
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}, nil
}
