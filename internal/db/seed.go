package db

import (
	"fmt"
	"os"
	"strconv"

	"github.com/yebobank/yebobank/internal/utils"
)

// Seed ensures the default admin and pool settings exist. Idempotent.
func Seed() error {
	if err := seedAdmin(); err != nil {
		return fmt.Errorf("db: seed admin: %w", err)
	}
	return seedPoolSettings()
}

func seedAdmin() error {
	phone := os.Getenv("ADMIN_PHONE")
	password := os.Getenv("ADMIN_PASSWORD")
	if phone == "" || password == "" {
		return nil
	}
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	var userID int64
	err = DB.QueryRow(`
		INSERT INTO users (phone, password_hash, role, kyc_status, full_name)
		VALUES ($1, $2, 'admin', 'verified', 'System Admin') RETURNING id
	`, phone, hash).Scan(&userID)
	if err != nil {
		return err
	}
	_, err = DB.Exec(`INSERT INTO wallets (user_id) VALUES ($1)`, userID)
	return err
}

func seedPoolSettings() error {
	bps := envInt("INTEREST_RATE_ANNUAL_BPS", 500)
	_, err := DB.Exec(`
		INSERT INTO pool_settings (id, interest_rate_bps, min_savings_sats, max_savings_sats, lock_days)
		VALUES (1, $1, 50000, 100000000000, 30)
		ON CONFLICT (id) DO NOTHING
	`, bps)
	return err
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
