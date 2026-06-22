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
- Docker + Docker Compose (for PostgreSQL)

---

### Quickstart — Docker Postgres + `go run`

This is the recommended way to develop. Docker provides the database; you run the Go server directly so you get fast reloads.

**1. Start the database**

```bash
git clone https://github.com/altradits/yebo
cd yebo
cp .env.example .env
docker compose up -d postgres
```

PostgreSQL is now available at `localhost:5433`.

**2. Configure `.env`**

Minimum required fields (the rest can stay as defaults for local dev):

```env
DB_URL=postgres://yebobank:change_me@localhost:5433/yebobank?sslmode=disable
PORT=8080
ADMIN_PHONE=+254700000000
ADMIN_PASSWORD=change_this_admin_password
```

**3. Run**

```bash
bash scripts/run_dev.sh
```

Or directly:

```bash
set -a && source .env && set +a
go run cmd/server/main.go
```

The server starts at **http://localhost:8080**.  
Migrations and seed data run automatically on first start.  
LND (Lightning) is optional — the server logs a warning and continues without it.

**Admin login:**
```
URL:      http://localhost:8080/login
Phone:    value of ADMIN_PHONE in .env
Password: value of ADMIN_PASSWORD in .env
```

---

### Option B — Full Docker Compose (no local Go needed)

```bash
cp .env.example .env
# Edit .env with your values
docker compose up --build
```

Server: **http://localhost:8080**  
PostgreSQL data persisted in the `pgdata` Docker volume.  
Nginx (TLS reverse proxy) is production-only: `docker compose --profile production up`

---

### Frontend (Next.js customer app)

A Next.js 16 frontend lives in `frontend/`. It talks to the Go backend via JSON API at `/api/*`.

**Prerequisites:** Node 20+

```bash
cd frontend
cp .env.local.example .env.local   # sets NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev
```

Frontend starts at **http://localhost:3000**.

The app has 5 tabs: Home · Send · Activity · Community · Profile, plus Deposit, Withdraw, Savings, and Chama detail pages. All protected routes redirect to `/login` automatically — unauthenticated users cannot access the dashboard.

> **Auth flow:** Enter phone → receive OTP (printed to server log in dev) → enter code → signed in.

To run both together (backend + frontend) open two terminals:

```bash
# Terminal 1 — backend
set -a && source .env && set +a
go run cmd/server/main.go

# Terminal 2 — frontend
cd frontend && npm run dev
```

Then open **http://localhost:3000**.

---

### JSON API Routes

The Go backend exposes these routes for the frontend:

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/request-otp` | — | Send OTP to phone |
| POST | `/api/auth/verify-otp` | — | Verify OTP, create session |
| POST | `/api/auth/logout` | — | Clear session |
| GET | `/api/user` | ✓ | Profile + balance + rate |
| GET | `/api/user/balance` | ✓ | Balance + BTC/KES rate |
| GET | `/api/user/transactions` | ✓ | Paginated ledger (`?limit=&offset=`) |
| GET | `/api/community/stats` | — | Member count, total savings, interest paid |

---

### Exposing Webhooks (M-Pesa callbacks)

Safaricom requires a public HTTPS URL for STK Push and B2C callbacks.  
Use [ngrok](https://ngrok.com) during development:

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
docker compose --profile production up -d
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

```
yebo/
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
├── docker-compose.yml               ← postgres + server + nginx
├── Dockerfile
├── nginx.conf                       ← production reverse proxy config
├── go.mod                           ← zero external dependencies
├── .env.example
│
├── cmd/
│   └── server/
│       └── main.go                  ← entry point, wires everything together
│
├── internal/
│   ├── pgdrv/                       ← zero-dependency PostgreSQL wire driver
│   │   └── pgdrv.go
│   │
│   ├── db/
│   │   ├── connection.go            ← pool setup, health check
│   │   ├── migrations.go            ← auto-run on startup
│   │   ├── seed.go                  ← default admin + pool settings
│   │   ├── ledger.go                ← CreditSats() DebitSats() — only way to touch balances
│   │   └── migrations/              ← SQL files applied at startup (go run)
│   │       ├── 001_users.sql
│   │       ├── 002_wallets_ledger.sql
│   │       ├── 003_savings.sql
│   │       ├── 004_chamas.sql
│   │       ├── 005_agents.sql
│   │       ├── 006_treasury.sql
│   │       ├── 007_mpesa.sql
│   │       ├── 008_lightning.sql
│   │       ├── 009_rates.sql
│   │       ├── 010_wallets_nullable_user.sql
│   │       └── 011_otp_requests.sql
│   │
│   ├── handlers/
│   │   ├── api.go                   ← JSON API: OTP auth, user profile, community stats
│   │   ├── api_ext.go               ← JSON API: chamas, deposit/withdraw, savings
│   │   ├── auth.go                  ← register, login, logout, session
│   │   ├── customer.go              ← dashboard, history, settings
│   │   ├── deposit.go               ← M-Pesa STK Push + Lightning receive
│   │   ├── withdraw.go              ← M-Pesa B2C + Lightning send
│   │   ├── savings.go               ← lock, unlock, early exit, interest view
│   │   ├── chama.go                 ← group wallet: create, join, contribute, vote
│   │   ├── agent.go                 ← agent dashboard, cash-in, cash-out, commission
│   │   ├── global.go                ← LNURL-pay endpoint, Lightning Address
│   │   ├── admin.go                 ← approvals, customers, settings, distribution
│   │   ├── trader.go                ← treasury assets, profit log
│   │   ├── webhook.go               ← M-Pesa callback, LND invoice settled
│   │   └── helpers.go               ← shared template rendering helpers
│   │
│   ├── services/
│   │   ├── mpesa/
│   │   │   ├── daraja.go            ← HTTP client, token refresh, STK Push
│   │   │   ├── b2c.go               ← B2C withdrawal to phone
│   │   │   ├── callback.go          ← validate + parse Safaricom callbacks
│   │   │   └── idempotency.go       ← receipt dedup before crediting
│   │   │
│   │   ├── lightning/
│   │   │   ├── client.go            ← LND REST client
│   │   │   ├── invoice.go           ← CreateInvoice, InvoiceStatus
│   │   │   ├── payment.go           ← SendPayment, CheckPayment
│   │   │   └── lnurl.go             ← LNURL-pay spec, Lightning Address
│   │   │
│   │   ├── rates/
│   │   │   ├── feed.go              ← fetch BTC/KES rate from CoinGecko
│   │   │   └── cache.go             ← in-memory cache, refresh every 60s
│   │   │
│   │   └── interest/
│   │       ├── engine.go            ← monthly distribution math
│   │       └── scheduler.go         ← goroutine: fires on 1st of month
│   │
│   ├── middleware/
│   │   ├── auth.go                  ← session validate, role guard, JSON 401 for API
│   │   ├── cors.go                  ← CORS headers for Next.js frontend
│   │   └── ratelimit.go             ← token bucket per IP
│   │
│   └── utils/
│       ├── crypto.go                ← PBKDF2 password hash, token gen
│       ├── formatters.go            ← FormatSats, SatsToKES, KESToSats, TimeAgo
│       └── validators.go            ← phone, amount validation
│
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── style.css
│   │   ├── js/
│   │   │   ├── app.js               ← KES↔Sats converter, balance toggle
│   │   │   ├── qr.js                ← QR generation via qrcode.js (CDN)
│   │   │   └── scanner.js           ← camera QR scan for invoices
│   │   └── assets/
│   │       └── logo.svg
│   │
│   └── templates/
│       ├── layout.html              ← base wrapper
│       ├── home.html                ← marketing landing page
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
│   ├── ARCHITECTURE.md              ← system design, data flow, component diagram
│   ├── DATABASE.md                  ← schema, migrations, ledger rules
│   ├── SECURITY.md                  ← password hashing, sessions, rate limiting, TLS
│   ├── MPESA.md                     ← Daraja API setup, STK Push, B2C, callbacks
│   ├── MPESA_SETUP.md               ← Daraja sandbox + production guide
│   ├── LIGHTNING.md                 ← LND node setup, Voltage.cloud, LNURL-pay
│   ├── LIGHTNING_NODE_SETUP.md      ← LND + Voltage.cloud step-by-step
│   ├── INTEREST.md                  ← interest engine math, scheduler, distribution
│   ├── DESIGN.md                    ← colors, fonts, shadows, components, mobile
│   ├── DEPLOYMENT.md                ← Docker Compose, nginx, TLS, production checklist
│   ├── API.md                       ← all endpoints, request/response, error codes
│   ├── BITCOIN_COMPLIANCE.md        ← what the Bitcoin whitepaper requires of us
│   ├── COMPETITORS.md               ← competitor analysis and positioning
│   ├── CONTRIBUTING.md              ← rules for contributors, branch strategy
│   ├── PRODUCT_PROPOSAL.md
│   └── database/
│       └── migrations/              ← SQL files used by Docker build
│           ├── 001_users.sql
│           ├── 002_wallets_ledger.sql
│           ├── 003_savings.sql
│           ├── 004_chamas.sql
│           ├── 005_agents.sql
│           ├── 006_treasury.sql
│           ├── 007_mpesa.sql
│           ├── 008_lightning.sql
│           ├── 009_rates.sql
│           ├── 010_wallets_nullable_user.sql
│           └── 011_otp_requests.sql
│
├── frontend/                        ← Next.js 16 customer-facing app
│   ├── src/
│   │   ├── proxy.ts                 ← auth guard: redirects unauthenticated users to /login
│   │   ├── app/
│   │   │   ├── (auth)/login/        ← phone entry screen
│   │   │   ├── (auth)/verify/       ← 6-digit OTP entry
│   │   │   ├── (app)/home/          ← balance card, quick actions, recent transactions
│   │   │   ├── (app)/send/          ← 3-step send: recipient → amount → confirm
│   │   │   ├── (app)/deposit/       ← M-Pesa STK Push deposit flow
│   │   │   ├── (app)/withdraw/      ← M-Pesa B2C withdrawal flow
│   │   │   ├── (app)/savings/       ← lock savings, view locks, APY display
│   │   │   ├── (app)/activity/      ← full tx history with search + filters
│   │   │   ├── (app)/community/     ← community stats, chama list
│   │   │   ├── (app)/community/chamas/[id]/  ← chama detail: balance, members
│   │   │   └── (app)/profile/       ← account settings, real logout
│   │   ├── components/              ← button, card, input, bottom-nav
│   │   └── lib/                     ← api.ts (typed fetch client), format.ts
│   ├── .env.local                   ← NEXT_PUBLIC_API_URL=http://localhost:8080
│   ├── .env.local.example           ← copy this to .env.local
│   └── package.json
│
└── scripts/
    ├── setup_db.sh                  ← create postgres user + database
    ├── run_dev.sh                   ← load .env and start server
    └── seed_test.sh                 ← populate test data
```

## License

MIT
