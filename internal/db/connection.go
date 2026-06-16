package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/yebobank/yebobank/internal/pgdrv"
)

var DB *sql.DB

func Open() error {
	db, err := sql.Open("pgdrv", buildDSN())
	if err != nil {
		return fmt.Errorf("db: open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("db: ping: %w", err)
	}
	DB = db
	return nil
}

func buildDSN() string {
	if url := os.Getenv("DB_URL"); url != "" {
		return url
	}
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_NAME", "yebobank"),
		env("DB_USER", "yebobank"),
		os.Getenv("DB_PASSWORD"),
		env("DB_SSLMODE", "disable"),
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
