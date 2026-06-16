# Contributing to YeboBank

YeboBank is the world's first open-source Bitcoin community bank built for Africa's low-income earners. Every line of code must reflect that mission.

## Hard Rules — No Exceptions

1. **Never write directly to `wallets.balance_sats`** — use `db.CreditSats()` or `db.DebitSats()`. These are the only safe paths to touch a balance.
2. **Never skip an audit log entry.** Every debit, credit, approval, and admin action must produce a row in `ledger_entries`.
3. **Every new feature needs a test in `*_test.go`.**
4. **Every migration is forward-only.** No down migrations. Add a new numbered file.
5. **All SQL is raw.** No ORM. `database/sql` only.
6. **Zero external Go dependencies.** Every package added to `go.mod` must have a written rationale in the PR. Stdlib is the only allowed dependency.

## Getting Started

```bash
cp .env.example .env
# fill in your credentials
bash scripts/setup_db.sh
bash scripts/run_dev.sh
```

## Branch Strategy

```
main          ← production-ready only. Protected. Requires 2 approvals.
develop       ← integration branch. CI must pass.
feature/*     ← individual features  (feature/mpesa-integration)
fix/*         ← bug fixes            (fix/interest-rounding)
```

## Issue Labels

| Label | Meaning |
|---|---|
| `good-first-issue` | Safe for new contributors |
| `core` | Touches ledger or auth — requires senior review |
| `frontend` | Templates + CSS only |
| `backend` | Go only |
| `mpesa` | M-Pesa integration |
| `lightning` | Lightning/LND integration |
| `database` | Schema migrations |
| `security` | Security-sensitive — additional review required |
| `docs` | Documentation only |

## PR Requirements

- `go build ./...` passes
- `go test ./...` passes
- `gosec -severity high ./...` — no new high findings
- Financial safety checklist in PR template completed
- Migration file added if schema changed

## Work Breakdown (for new contributors)

Read the issue assigned to you. Read the section of `docs/` covering that feature. Read Bitcoin whitepaper Section 2 (3 pages). Then write the code.
