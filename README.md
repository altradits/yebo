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

## How to Run

### Prerequisites

- Go 1.22+
- PostgreSQL 16+ **or** Docker + Docker Compose

---

### Option 1 — Docker Compose (recommended)

```bash
git clone https://github.com/altradits/yebo
cd yebo
cp .env.example .env
# Edit .env — minimum required fields are marked below
docker compose up --build
```

The server starts at **http://localhost:8080** (nginx proxies port 80/443).  
PostgreSQL data is persisted in the `pgdata` Docker volume.

---

### Option 2 — Local (Go + Postgres)

**1. Database**

```bash
createuser -P yebobank        # pick a password
createdb -O yebobank yebobank
```

**2. Environment**

```bash
cp .env.example .env
```

Edit `.env` — required fields to start:

```env
# Database
DB_URL=postgres://yebobank:<password>@localhost:5432/yebobank?sslmode=disable

# Server
PORT=8080

# M-Pesa (use Safaricom Sandbox for development)
MPESA_ENV=sandbox
MPESA_CONSUMER_KEY=<from sandbox.safaricom.co.ke>
MPESA_CONSUMER_SECRET=<from sandbox.safaricom.co.ke>
MPESA_SHORTCODE=174379
MPESA_PASSKEY=<sandbox passkey>
MPESA_CALLBACK_URL=https://<your-ngrok-or-domain>

# Lightning (optional — server starts without LND)
LND_HOST=localhost:10009
LND_MACAROON_HEX=<admin.macaroon hex>
LND_TLS_CERT_PATH=/path/to/tls.cert
```

**3. Run**

```bash
go run cmd/server/main.go
```

Migrations and seed data run automatically on startup. The default admin account is created by the seed:

```
Phone: +254700000000
Password: admin1234
```

Change the password immediately after first login at `/settings/password`.

---

### Exposing Webhooks (M-Pesa callbacks)

Safaricom needs a public HTTPS URL to deliver STK Push and B2C callbacks.  
Use [ngrok](https://ngrok.com) or [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) during development:

```bash
ngrok http 8080
# Copy the https URL → set MPESA_CALLBACK_URL in .env
```

---

### TLS in Production

The included `nginx.conf` expects Let's Encrypt certificates at:
```
/etc/letsencrypt/live/yebobank.com/
```

Obtain them with [certbot](https://certbot.eff.org/) before starting nginx, then:

```bash
docker compose up -d
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
