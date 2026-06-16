package rates

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yebobank/yebobank/internal/db"
)

const refreshInterval = 60 * time.Second

// StartFeed begins the background goroutine that fetches BTC/KES every 60s.
func StartFeed() {
	go func() {
		for {
			if err := fetch(); err != nil {
				log.Printf("rates: fetch error: %v", err)
			}
			time.Sleep(refreshInterval)
		}
	}()
}

func fetch() error {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=kes,usd"
	apiKey := os.Getenv("COINGECKO_API_KEY")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", apiKey)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("coingecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Bitcoin struct {
			KES float64 `json:"kes"`
			USD float64 `json:"usd"`
		} `json:"bitcoin"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("coingecko: decode: %w", err)
	}
	if payload.Bitcoin.KES <= 0 {
		return fmt.Errorf("coingecko: got zero KES rate")
	}

	Global.Set(payload.Bitcoin.KES, payload.Bitcoin.USD)

	if db.DB != nil {
		db.DB.Exec(`INSERT INTO rate_snapshots (btc_kes, btc_usd) VALUES ($1, $2)`,
			payload.Bitcoin.KES, payload.Bitcoin.USD)
	}
	return nil
}
