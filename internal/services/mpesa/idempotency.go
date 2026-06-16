package mpesa

import (
	"database/sql"
	"fmt"

	"github.com/yebobank/yebobank/internal/db"
)

// MarkCompleted records a completed M-Pesa transaction.
// Returns an error if the receipt has already been processed (duplicate callback).
func MarkCompleted(tx *sql.Tx, receipt string, walletID int64, amountKES float64) error {
	var existing string
	err := tx.QueryRow(
		`SELECT mpesa_receipt FROM mpesa_transactions WHERE mpesa_receipt = $1`, receipt,
	).Scan(&existing)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("idempotency: check: %w", err)
	}
	if err == nil {
		return fmt.Errorf("idempotency: receipt %s already processed", receipt)
	}
	_, err = tx.Exec(`
		INSERT INTO mpesa_transactions (mpesa_receipt, type, phone, amount_kes, status, wallet_id, completed_at)
		VALUES ($1, 'stk_push', '', $2, 'completed', $3, NOW())
	`, receipt, amountKES, walletID)
	return err
}

// IsDuplicate returns true if a receipt has already been credited.
func IsDuplicate(receipt string) bool {
	var count int
	db.DB.QueryRow( //nolint:errcheck
		`SELECT COUNT(*) FROM mpesa_transactions WHERE mpesa_receipt=$1 AND status='completed'`,
		receipt,
	).Scan(&count)
	return count > 0
}
