# YeboBank — Development Guidelines
## The World's First Open-Source Bitcoin Community Bank for Africa

> Built on the Bitcoin and Lightning Network whitepapers.
> Designed for low-income earners in Kenya and Africa.
> Open source. One phase. No guessing.

---

## What This Is

YeboBank is a custodial, open-source Bitcoin community bank. It is not a wallet, not a
payment processor, and not an exchange. It is a **bank** — with savings accounts, interest
income, group wallets (chamas), agent cash networks, and a global payment receive endpoint.
Everything denominated in Bitcoin (satoshis), displayed in Kenyan Shillings.

## What Makes It Different

| Feature                | YeboBank | Strike | Galoy/Blink | Machankura | BTCPay |
|------------------------|----------|--------|-------------|------------|--------|
| Open source            | YES      | NO     | YES         | NO         | YES    |
| M-Pesa integration     | YES      | YES    | NO          | YES        | NO     |
| Savings with interest  | YES      | NO     | NO          | NO         | NO     |
| Chama group wallets    | YES      | NO     | NO          | NO         | NO     |
| Agent cash network     | YES      | NO     | NO          | NO         | NO     |
| Lightning addresses    | YES      | YES    | YES         | NO         | YES    |
| Global payments in     | YES      | YES    | NO          | NO         | YES    |
| Zero external Go deps  | YES      | NO     | NO          | NO         | NO     |

## Competitors to Study

- Galoy: https://galoy.io | https://github.com/GaloyMoney/galoy
- Blink: https://blink.sv
- Machankura (USSD Bitcoin - partner, do not compete): https://8333.mobi
- Strike: https://strike.me | https://strike.me/blog/announcing-strike-africa/
- Bitnob: https://bitnob.com
- BTCPay Server: https://btcpayserver.org | https://github.com/btcpayserver/btcpayserver
- LNbits: https://lnbits.com
- Yellow Card: https://yellowcard.io

## Quick Start

```bash
git clone https://github.com/yebobank/yebobank
cd yebobank
cp .env.example .env
# Edit .env with M-Pesa, LND, and database credentials
./scripts/setup_db.sh
go run cmd/server/main.go
```

## THE RULE EVERY DEVELOPER MUST KNOW

Never write directly to wallets.balance_sats.
Use db.CreditSats() or db.DebitSats() only.
Both functions write the ledger entry in the same DB transaction.
If either fails, both fail. This is non-negotiable.

## Documentation Index

- docs/ARCHITECTURE.md     System design, data flow, component diagram
- docs/DATABASE.md         Complete schema, all migrations, ledger rules
- docs/SECURITY.md         Password hashing, sessions, rate limiting, TLS
- docs/MPESA.md            Daraja API setup, STK Push, B2C, callbacks
- docs/LIGHTNING.md        LND node setup, Voltage.cloud, LNURL-pay
- docs/INTEREST.md         Interest engine math, scheduler, distribution
- docs/DESIGN.md           Colors, fonts, shadows, components, mobile
- docs/CONTRIBUTING.md     Rules for 1,000 developers, branch strategy
- docs/DEPLOYMENT.md       Docker, nginx, TLS, production checklist
- docs/API.md              All endpoints, request/response, error codes
- docs/BITCOIN_COMPLIANCE.md  What the Bitcoin whitepaper requires of us
- docs/COMPETITORS.md      Competitor analysis and positioning

## Structure
yebobank/
│
├── README.md
├── LICENSE                          ← MIT License
├── CONTRIBUTING.md
├── SECURITY.md                      ← responsible disclosure policy
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   ├── workflows/
│   │   ├── ci.yml                   ← go build + go test on every PR
│   │   ├── security.yml             ← gosec scanner
│   │   └── docker.yml               ← build + push to GHCR on main
│   └── pull_request_template.md
│
├── cmd/
│   └── server/
│       └── main.go                  ← entry point, wire everything together
│
├── internal/
│   ├── pgdrv/                       ← zero-dependency PostgreSQL wire driver
│   │   └── pgdrv.go
│   │
│   ├── db/
│   │   ├── connection.go            ← pool setup, health check
│   │   ├── migrations.go            ← auto-run on startup
│   │   ├── seed.go                  ← default admin + pool settings
│   │   └── ledger.go                ← CreditSats() DebitSats() — only way to touch balances
│   │
│   ├── handlers/
│   │   ├── auth.go                  ← register, login, logout, session
│   │   ├── customer.go              ← dashboard, history, settings
│   │   ├── deposit.go               ← M-Pesa STK Push + Lightning receive
│   │   ├── withdraw.go              ← M-Pesa B2C + Lightning send
│   │   ├── savings.go               ← lock, unlock, early exit, interest view
│   │   ├── chama.go                 ← group wallet: create, join, contribute, vote
│   │   ├── agent.go                 ← agent dashboard, cash-in, cash-out, commission
│   │   ├── global.go                ← LNURL-pay endpoint, payment links
│   │   ├── admin.go                 ← approvals, customers, settings, distribution
│   │   ├── trader.go                ← treasury assets, profit log
│   │   └── webhook.go               ← M-Pesa callback, LND invoice settled
│   │
│   ├── services/
│   │   ├── mpesa/
│   │   │   ├── daraja.go            ← HTTP client, token refresh, STK Push
│   │   │   ├── b2c.go               ← B2C withdrawal to phone
│   │   │   ├── callback.go          ← validate + parse Safaricom callbacks
│   │   │   └── idempotency.go       ← receipt → dedup before crediting
│   │   │
│   │   ├── lightning/
│   │   │   ├── client.go            ← LND gRPC connection (or REST fallback)
│   │   │   ├── invoice.go           ← CreateInvoice, DecodeInvoice, WatchInvoice
│   │   │   ├── payment.go           ← SendPayment, CheckPayment
│   │   │   └── lnurl.go             ← LNURL-pay spec, Lightning Address
│   │   │
│   │   ├── rates/
│   │   │   ├── feed.go              ← Fetch BTC/KES rate from CoinGecko or LNURL
│   │   │   └── cache.go             ← In-memory cache, refresh every 60s
│   │   │
│   │   └── interest/
│   │       ├── engine.go            ← Monthly distribution math
│   │       └── scheduler.go         ← Goroutine: sleep until 1st of month
│   │
│   ├── middleware/
│   │   ├── auth.go                  ← Session validate, role guard, idle timeout
│   │   └── ratelimit.go             ← Simple token bucket per IP
│   │
│   └── utils/
│       ├── crypto.go                ← PBKDF2 password hash, token gen, salt gen
│       ├── formatters.go            ← FormatSats, SatsToKES, KESToSats, TimeAgo
│       └── validators.go            ← Phone, email, amount validation
│
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── style.css            ← Single file, final, documented below
│   │   ├── js/
│   │   │   ├── app.js               ← KES↔Sats converter, balance toggle
│   │   │   ├── qr.js                ← QR generation via qrcode.js (CDN)
│   │   │   └── scanner.js           ← Camera-based QR scan for invoices
│   │   └── assets/
│   │       ├── logo.svg             ← YeboBank mark — documented in §3
│   │       └── icons/               ← SVG icon set — documented in §3
│   │
│   └── templates/
│       ├── layout.html              ← Base wrapper
│       ├── home.html                ← Marketing/landing page (standalone)
│       ├── login.html
│       ├── register.html
│       ├── customer/
│       │   ├── dashboard.html
│       │   ├── deposit.html
│       │   ├── withdraw.html
│       │   ├── send.html
│       │   ├── receive.html
│       │   ├── history.html
│       │   ├── savings.html
│       │   ├── savings_lock.html
│       │   ├── chama.html
│       │   ├── chama_create.html
│       │   └── settings.html
│       ├── agent/
│       │   ├── dashboard.html
│       │   ├── cash_in.html
│       │   └── cash_out.html
│       ├── trader/
│       │   ├── dashboard.html
│       │   ├── assets.html
│       │   └── profit.html
│       └── admin/
│           ├── dashboard.html
│           ├── deposits.html
│           ├── withdrawals.html
│           ├── customers.html
│           ├── chamas.html
│           ├── agents.html
│           ├── distribution.html
│           └── settings.html
│
├── docs/
│   ├── database/
│   │   └── migrations/
│   │       ├── 001_users.sql
│   │       ├── 002_wallets_ledger.sql
│   │       ├── 003_savings.sql
│   │       ├── 004_chamas.sql
│   │       ├── 005_agents.sql
│   │       ├── 006_treasury.sql
│   │       ├── 007_mpesa.sql
│   │       ├── 008_lightning.sql
│   │       └── 009_rates.sql
│   ├── API.md                       ← All endpoints documented
│   ├── LIGHTNING_NODE_SETUP.md      ← LND + Voltage.cloud instructions
│   ├── MPESA_SETUP.md               ← Daraja sandbox + production guide
│   ├── DEPLOYMENT.md                ← Docker Compose + bare metal
│   └── PRODUCT_PROPOSAL.md          ← This document
│
├── scripts/
│   ├── setup_db.sh                  ← Create postgres user + database
│   ├── run_dev.sh                   ← Start server with env vars
│   └── seed_test.sh                 ← Populate test data
│
├── docker-compose.yml               ← postgres + server + nginx
├── Dockerfile
├── go.mod                           ← Zero external dependencies
└── .env.example

## License

MIT
