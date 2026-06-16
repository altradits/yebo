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
		return urlToDSN(url)
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

// urlToDSN converts a postgres://user:pass@host:port/dbname?sslmode=x URL
// to the space-separated DSN format that pgdrv understands.
func urlToDSN(rawURL string) string {
	// Strip scheme
	s := rawURL
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if len(s) > len(prefix) && s[:len(prefix)] == prefix {
			s = s[len(prefix):]
			break
		}
	}
	// Split query string
	sslmode := "disable"
	if qi := indexOf(s, '?'); qi >= 0 {
		for _, param := range splitOn(s[qi+1:], '&') {
			kv := splitN(param, '=', 2)
			if len(kv) == 2 && kv[0] == "sslmode" {
				sslmode = kv[1]
			}
		}
		s = s[:qi]
	}
	// Split user:pass@host:port/dbname
	user, pass, host, port, dbname := "", "", "localhost", "5432", ""
	if at := lastIndexOf(s, '@'); at >= 0 {
		userInfo := s[:at]
		s = s[at+1:]
		if ci := indexOf(userInfo, ':'); ci >= 0 {
			user = userInfo[:ci]
			pass = userInfo[ci+1:]
		} else {
			user = userInfo
		}
	}
	if slash := indexOf(s, '/'); slash >= 0 {
		dbname = s[slash+1:]
		s = s[:slash]
	}
	if ci := lastIndexOf(s, ':'); ci >= 0 {
		host = s[:ci]
		port = s[ci+1:]
	} else if s != "" {
		host = s
	}
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, pass, sslmode)
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func lastIndexOf(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func splitOn(s string, sep byte) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func splitN(s string, sep byte, n int) []string {
	var out []string
	for i := 0; i < len(s) && len(out) < n-1; i++ {
		if s[i] == sep {
			out = append(out, s[:i])
			s = s[i+1:]
			i = -1
		}
	}
	return append(out, s)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
