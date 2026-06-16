package db

import (
	"database/sql"
	"fmt"
)

// CreditSats adds sats to a wallet and records the ledger entry.
// This is the ONLY safe way to increase a balance. Never update wallets.balance_sats directly.
func CreditSats(tx *sql.Tx, walletID, amount int64, entryType, refID, note string, actorID *int64) error {
	if amount <= 0 {
		return fmt.Errorf("ledger: credit amount must be positive, got %d", amount)
	}
	return record(tx, walletID, amount, entryType, refID, note, actorID)
}

// DebitSats removes sats from a wallet and records the ledger entry.
// This is the ONLY safe way to decrease a balance. Never update wallets.balance_sats directly.
func DebitSats(tx *sql.Tx, walletID, amount int64, entryType, refID, note string, actorID *int64) error {
	if amount <= 0 {
		return fmt.Errorf("ledger: debit amount must be positive, got %d", amount)
	}
	return record(tx, walletID, -amount, entryType, refID, note, actorID)
}

func record(tx *sql.Tx, walletID, delta int64, entryType, refID, note string, actorID *int64) error {
	var balanceAfter int64
	err := tx.QueryRow(`
		UPDATE wallets
		SET balance_sats = balance_sats + $1, updated_at = NOW()
		WHERE id = $2
		RETURNING balance_sats
	`, delta, walletID).Scan(&balanceAfter)
	if err != nil {
		return fmt.Errorf("ledger: update wallet %d: %w", walletID, err)
	}
	if balanceAfter < 0 {
		return fmt.Errorf("ledger: insufficient balance in wallet %d", walletID)
	}
	_, err = tx.Exec(`
		INSERT INTO ledger_entries
			(wallet_id, amount_sats, balance_after, type, ref_id, note, actor_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, walletID, delta, balanceAfter, entryType, strOrNil(refID), strOrNil(note), actorID)
	return err
}

// WalletByUserID returns the wallet ID for a user.
func WalletByUserID(userID int64) (int64, error) {
	var id int64
	err := DB.QueryRow(`SELECT id FROM wallets WHERE user_id = $1`, userID).Scan(&id)
	return id, err
}

// BalanceSats returns the current balance for a wallet.
func BalanceSats(walletID int64) (int64, error) {
	var b int64
	err := DB.QueryRow(`SELECT balance_sats FROM wallets WHERE id = $1`, walletID).Scan(&b)
	return b, err
}

func strOrNil(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
