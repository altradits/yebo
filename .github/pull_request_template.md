## Summary

## Type of change
- [ ] Bug fix
- [ ] New feature
- [ ] Refactor
- [ ] Docs / tests only
- [ ] Security fix

## Financial safety checklist
- [ ] No direct `wallets.balance_sats` writes — all changes via `db.CreditSats` / `db.DebitSats`
- [ ] Every new ledger operation produces an audit row
- [ ] M-Pesa callbacks validated before crediting (idempotency key checked)
- [ ] Lightning invoice settlement treated as final
- [ ] No new external Go dependencies (or rationale below)

## External dependencies added
None

## Testing
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes

**Manual test steps:**
1.
2.

## Migration required
- [ ] Yes — migration file added at `docs/database/migrations/NNN_name.sql`
- [ ] No
