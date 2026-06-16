# YeboBank — Contributing Guidelines
## For 1,000 Open Source Developers

## The Five Rules (Read Before Writing Any Code)

1. Never write directly to wallets.balance_sats.
   Use db.CreditSats() or db.DebitSats() in internal/db/ledger.go.
   Both functions write the ledger entry in the same DB transaction.
   If either fails, both fail. No exceptions.

2. Never delete from ledger_entries.
   The ledger is append-only. If a correction is needed, add a reverse entry.

3. Every external callback (M-Pesa, Lightning) must check idempotency before crediting.
   If the same receipt/payment_hash appears twice, credit once and log the duplicate.

4. Every migration is forward-only.
   No down migrations. Migrations are numbered and applied in sequence at startup.
   If you need to undo something, add a new migration.

5. All SQL uses parameterized queries.
   No string concatenation in SQL. Ever. $1, $2, $3 placeholders only.

## Branch Strategy

main         Protected. Production-ready only. Requires 2 approvals + passing CI.
develop      Integration branch. CI must pass before merge.
feature/*    Feature branches (feature/mpesa-stk-push)
fix/*        Bug fixes (fix/interest-rounding-error)
docs/*       Documentation only (docs/lightning-setup)

## Issue Labels

good-first-issue     New contributors start here
core                 Touches ledger or auth. Requires senior review.
frontend             Templates and CSS only
backend              Go code only
mpesa                M-Pesa integration
lightning            Lightning/LND integration
database             Schema migrations
security             Security-sensitive. Mandatory review.
docs                 Documentation only
chama                Chama/group wallet feature
agent                Agent network feature

## Work Assignment

### 10 Senior Developers (Core Team)
These are the only developers who touch the following files:
- internal/db/ledger.go
- internal/middleware/auth.go
- internal/services/mpesa/ (all files)
- internal/services/lightning/ (all files)
- internal/services/interest/ (all files)
- internal/handlers/webhook.go
- docs/database/migrations/ (all SQL files)

### 100 Mid-Level Developers
- All customer handlers (deposit, withdraw, send, receive, savings, chama, agent)
- All admin handlers
- Rate feed service
- QR code generation (web/static/js/qr.js)
- Balance toggle (web/static/js/app.js)
- Chama voting system

### 200 Frontend Developers
- All 25 HTML templates following DESIGN.md exactly
- Mobile bottom navigation
- Responsive CSS in web/static/css/style.css
- Home page (standalone, does not use layout.html)

### 100 Test Developers
- Unit tests for interest engine (all edge cases listed in INTEREST.md)
- Unit tests for PBKDF2 implementation
- Unit tests for sats conversion math
- Integration tests for M-Pesa callback handler
- Integration tests for LND invoice settlement
- Load tests for concurrent deposit handling

### 50 Documentation Developers
- docs/API.md (every endpoint documented with curl examples)
- docs/LIGHTNING_NODE_SETUP.md (step-by-step Voltage.cloud guide)
- docs/MPESA_SETUP.md (Daraja sandbox registration walkthrough)
- docs/DEPLOYMENT.md (Docker + nginx + TLS)
- In-code comments on all exported functions

### 40 DevOps Developers
- Dockerfile (multi-stage, final image under 20MB)
- docker-compose.yml (postgres + server + nginx)
- .github/workflows/ci.yml (go build, go test, go vet, gosec)
- .github/workflows/docker.yml (build + push to GHCR on main merge)
- nginx.conf (TLS termination, IP allowlisting for M-Pesa callbacks)
- scripts/setup_db.sh
- scripts/run_dev.sh

### 10 Localization Developers
- All UI text in English AND Swahili
- users.language field drives language selection
- Template function t(key, lang) for translations
- translations/en.json and translations/sw.json

### 500 Community Developers
- Test the product end-to-end
- File bug reports with reproduction steps
- Write the GitHub wiki
- Create YouTube tutorials
- Translate documentation
- Start with good-first-issue label

## Pull Request Requirements

All PRs must include:
1. Description of what changed and why
2. Tests for any new logic
3. Documentation update if adding a new endpoint or feature
4. go build ./... passes
5. go vet ./... passes
6. go test ./... passes

PRs touching ledger.go or auth.go require 2 senior developer approvals.
PRs touching migrations require 1 senior developer approval + test on staging DB.
All other PRs require 1 approval.

## Testing Requirements

Every package must have a _test.go file.
Target: 80% coverage on internal/services/* packages.
Target: 100% coverage on internal/db/ledger.go.
Integration tests must use a real PostgreSQL instance (not mocks).

Run tests:
  go test ./...
  go test -race ./...  (detect race conditions)
  go test -cover ./...

## Code Style

gofmt -w . before every commit.
No unused imports.
No unused variables.
Error returns checked everywhere.
Log errors with context: log.Printf("[mpesa callback] error: %v context: %s", err, ctx)
No panic() in handlers.

## Environment Setup for Contributors

1. Install Go 1.22+
2. Install PostgreSQL 14+
3. cp .env.example .env
4. Edit .env with your credentials
5. ./scripts/setup_db.sh
6. go run cmd/server/main.go
7. Visit http://localhost:8080

For M-Pesa testing:
- Register free at https://developer.safaricom.co.ke
- Use sandbox credentials (no real money involved)
- Use ngrok to expose local callback URL for Safaricom webhooks:
  ngrok http 8080
  Set MPESA_CALLBACK_URL=https://YOUR-NGROK-URL.ngrok.io

For Lightning testing:
- Create free account at https://voltage.cloud
- Or run Polar (https://lightningpolar.com) locally for development
