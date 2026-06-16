# YeboBank Product Proposal

## The Gap

YeboBank is the world's first open-source Bitcoin community bank designed specifically for Africa's low-income earners.

Not a wallet. Not a payment processor. A **bank** — with savings accounts, interest income, agent cash networks, group savings (chamas), and a global payment rail. Denominated entirely in Bitcoin (satoshis), displayed in local currency.

## Why This Doesn't Exist Yet

Every existing player chose one part of this problem. See `docs/` for competitive analysis.

## Core Features

1. **Savings accounts** — Lock sats for 30+ days, earn interest from transparent treasury yield
2. **M-Pesa integration** — Deposit and withdraw via the 51M-user M-Pesa network
3. **Agent cash network** — Physical cash-in/out points for users without smartphones
4. **Chama group wallets** — Community savings groups with on-chain voting
5. **Lightning payments** — Send to anyone globally, instantly
6. **LNURL / Lightning Address** — Receive payments at user@yebobank.com
7. **Interest engine** — Monthly distribution from treasury yield, never from deposits

## Bitcoin Whitepaper Compliance

1. We hold keys — custodial by design for low-income accessibility
2. Every satoshi movement is append-only, timestamped, auditable
3. Confirmed Lightning payments are final — no silent reversals
4. Interest comes only from treasury yield — no fractional reserve
5. LND watchtower active — HTLC protection

## Architecture

See `docs/API.md` for routes. Zero external Go dependencies. PostgreSQL. LND REST.

## Phase 1 Scope

Web app (works on phone browser). M-Pesa on/off ramp. Lightning send/receive. Savings. Chamas. Agent network.

## Phase 2 (future)

PWA manifest (install to homescreen). Native mobile app. USSD interface (partner with Machankura).
