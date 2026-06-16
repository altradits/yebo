# YeboBank — System Architecture

## Technology Stack

| Layer       | Choice                         | Reason                                          |
|-------------|--------------------------------|--------------------------------------------------|
| Language    | Go 1.22, stdlib only           | No module proxy needed, compiles anywhere        |
| Database    | PostgreSQL 14+                 | ACID transactions for financial data             |
| DB driver   | internal/pgdrv (custom)        | Zero external dependencies                       |
| Templates   | html/template                  | No build step, works immediately                 |
| Passwords   | PBKDF2-SHA256, 310k iterations | Secure without external libraries                |
| Sessions    | DB-backed + 2hr idle expiry    | Protects shared-device users                     |
| Ledger unit | Satoshis (int64)               | No floating-point arithmetic on money            |
| M-Pesa      | Daraja API (net/http only)     | STK Push deposits, B2C withdrawals               |
| Lightning   | LND REST API (net/http only)   | Invoice, payment, LNURL-pay                      |
| Rate feed   | CoinGecko free API             | BTC/KES every 60 seconds, cached in DB           |
| Interest    | Background goroutine           | Runs 2nd of each month, single DB transaction    |

## Repository Structure

```
yebobank/
├── cmd/server/main.go                  Entry point, wire everything
├── go.mod                              Zero external dependencies
├── .env.example
├── Dockerfile
├── docker-compose.yml
│
├── internal/
│   ├── pgdrv/pgdrv.go                  Custom PostgreSQL wire driver
│   ├── db/
│   │   ├── connection.go               Pool setup, health check
│   │   ├── migrations.go               Auto-run on startup
│   │   ├── seed.go                     Default admin + pool settings
│   │   └── ledger.go                   CreditSats() DebitSats() ONLY way to touch balances
│   ├── handlers/
│   │   ├── auth.go                     register, login, logout, session
│   │   ├── customer.go                 dashboard, history, settings
│   │   ├── deposit.go                  M-Pesa STK Push + Lightning receive
│   │   ├── withdraw.go                 M-Pesa B2C + Lightning send
│   │   ├── savings.go                  lock, unlock, early exit, interest view
│   │   ├── chama.go                    group wallet: create, join, contribute, vote
│   │   ├── agent.go                    agent dashboard, cash-in, cash-out, commission
│   │   ├── global.go                   LNURL-pay endpoint, payment links
│   │   ├── admin.go                    approvals, customers, settings, distribution
│   │   ├── trader.go                   treasury assets, profit log
│   │   └── webhook.go                  M-Pesa callback, LND invoice settled
│   ├── services/
│   │   ├── mpesa/
│   │   │   ├── daraja.go               HTTP client, token refresh, STK Push
│   │   │   ├── b2c.go                  B2C withdrawal to phone
│   │   │   ├── callback.go             Validate + parse Safaricom callbacks
│   │   │   └── idempotency.go          Receipt dedup before crediting
│   │   ├── lightning/
│   │   │   ├── client.go               LND REST connection
│   │   │   ├── invoice.go              CreateInvoice, DecodeInvoice, CheckInvoice
│   │   │   ├── payment.go              SendPayment, CheckPayment
│   │   │   └── lnurl.go                LNURL-pay spec, Lightning Address
│   │   ├── rates/
│   │   │   ├── feed.go                 Fetch BTC/KES from CoinGecko
│   │   │   └── cache.go                In-memory cache, refresh every 60s
│   │   └── interest/
│   │       ├── engine.go               Monthly distribution math
│   │       └── scheduler.go            Goroutine: sleep until 2nd of month
│   ├── middleware/
│   │   ├── auth.go                     Session validate, role guard, idle timeout
│   │   └── ratelimit.go                Token bucket per IP
│   └── utils/
│       ├── crypto.go                   PBKDF2 password hash, token gen, salt gen
│       ├── formatters.go               FormatSats, SatsToKES, KESToSats, TimeAgo
│       └── validators.go               Phone, email, amount validation
│
├── web/
│   ├── static/
│   │   ├── css/style.css               Single file, all styles
│   │   ├── js/
│   │   │   ├── app.js                  KES/Sats/BTC toggle, live conversion
│   │   │   ├── qr.js                   QR code generation
│   │   │   └── scanner.js              Camera QR scan for Lightning invoices
│   │   └── assets/icons/               SVG icons (Tabler Icons, subset)
│   └── templates/
│       ├── layout.html                 Base wrapper
│       ├── home.html                   Marketing page (standalone)
│       ├── login.html
│       ├── register.html
│       ├── customer/                   (10 templates)
│       ├── agent/                      (3 templates)
│       ├── trader/                     (3 templates)
│       └── admin/                      (8 templates)
│
└── docs/database/migrations/           9 SQL migration files
```

## Data Flow

### Deposit via M-Pesa
```
User submits amount
  → Handler calls mpesa.InitiateSTKPush()
  → Safaricom sends PIN prompt to user's phone
  → User approves on phone
  → Safaricom POST to /api/mpesa/callback
  → webhook.HandleMpesaCallback()
  → Check idempotency (mpesa_reference unique)
  → db.Begin()
  → db.CreditSats(tx, walletID, sats, "deposit_mpesa", ...)
    → UPDATE wallets SET balance_sats = balance_sats + sats
    → INSERT INTO ledger_entries (...)
  → db.Commit()
  → User sees updated balance
```

### Lightning Payment Receive
```
User clicks "Receive"
  → Handler calls lightning.CreateInvoice(sats, memo)
  → LND returns BOLT11 invoice + payment_hash
  → Store in ln_invoices table (status=pending)
  → Display QR code to user
  → Payer scans and pays
  → LND webhook fires to /api/lightning/invoice-settled
  → Check payment_hash not already processed
  → db.CreditSats(tx, walletID, sats, "deposit_lightning", ...)
  → Update ln_invoices status=settled
```

### Monthly Interest Distribution
```
Scheduler goroutine wakes on 2nd of month 00:01 UTC
  → interest.Calculate(db, month, adminFeePct)
    → SUM profit_logs for that month
    → GET all active savings_locks
    → Compute proportional amounts (integer math, no floats)
  → interest.Execute(db, distribution)
    → db.Begin()
    → INSERT interest_distributions (header row)
    → FOR EACH lock:
        → db.CreditSats(tx, walletID, amount, "interest_payment", ...)
        → UPDATE savings_locks SET accrued_sats = accrued_sats + amount
        → INSERT interest_distribution_lines
    → db.CreditSats(tx, adminWalletID, adminFee, "admin_fee", ...)
    → db.Commit()
```

## Architecture Diagram

```
                    +------------------------+
                    |   YeboBank Frontend    |
                    |  (Go html/template)    |
                    | Marketing | Customer   |
                    | Agent | Admin | Trader |
                    +----------+-------------+
                               |
                    +----------v-------------+
                    |    Go HTTP Server      |
                    |    net/http mux        |
                    | Handlers + Middleware  |
                    +----+--------+----------+
                         |        |
          +--------------+        +------------------+
          |                                          |
+---------v---------+                   +-----------v----------+
|    PostgreSQL      |                   |   External Services  |
|                    |                   |                      |
| wallets            |                   | M-Pesa Daraja API    |
| ledger_entries     |                   | LND Node (Voltage)   |
| savings_locks      |                   | CoinGecko Rate Feed  |
| chamas             |                   +-----------+----------+
| agents             |                               |
| mpesa_transactions |                   +-----------v----------+
| ln_invoices        |                   |   Webhook Handler    |
| rate_snapshots     |                   |   (Idempotent)       |
| ...                |<------------------+ -> db.CreditSats()  |
+--------------------+                   +---------------------+
          |
+---------v---------+
|  Background Jobs  |
| Rate feed (60s)   |
| Interest (1/mo)   |
+-------------------+
```
