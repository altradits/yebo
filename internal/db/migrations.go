package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrate applies all pending SQL migration files in lexicographic order.
// Tracks applied migrations in the schema_migrations table.
func Migrate(dir string) error {
	if err := ensureMigrationsTable(); err != nil {
		return err
	}
	applied, err := appliedMigrations()
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("db: glob %s: %w", dir, err)
	}
	sort.Strings(files)
	for _, f := range files {
		name := filepath.Base(f)
		if applied[name] {
			continue
		}
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("db: read %s: %w", name, err)
		}
		if err := applyMigration(name, string(sql)); err != nil {
			return fmt.Errorf("db: migration %s: %w", name, err)
		}
	}
	return nil
}

func ensureMigrationsTable() error {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name       TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	return err
}

func appliedMigrations() (map[string]bool, error) {
	rows, err := DB.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		m[name] = true
	}
	return m, rows.Err()
}

func applyMigration(name, sql string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	for _, stmt := range splitSQL(sql) {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(name) VALUES($1)`, name); err != nil {
		return err
	}
	return tx.Commit()
}

func splitSQL(sql string) []string {
	var out []string
	for _, s := range strings.Split(sql, ";") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
