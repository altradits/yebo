# YeboBank — Lightning Network Integration

## Overview

Lightning Network is used for:
1. Receiving sats from external Lightning wallets
2. Sending sats to external Lightning wallets
3. Lightning Addresses (user@yebobank.com) so anyone can pay a YeboBank user
4. Internal transfers between YeboBank users (on-us, zero routing fee)

## LND Node Options (Choose One)

### Option A: Voltage.cloud (Recommended for Launch)
URL: https://voltage.cloud
Cost: $5/month for a managed LND node
Setup time: 15 minutes
What you get: LND node with REST API, managed TLS, dashboard, backups

Steps:
1. Sign up at voltage.cloud
2. Create node > Select LND > Pick region (EU West or South Africa closest to Kenya)
3. Download: admin.macaroon, readonly.macaroon, tls.cert
4. Note your node endpoint: your-node.voltageapp.io:8080 (REST)
5. Fund node: send on-chain BTC to displayed address
6. Open channels: connect to ACINQ, Wallet of Satoshi, or other well-connected nodes

### Option B: Self-hosted LND on VPS
Cost: ~$10-20/month VPS
Setup time: 2-3 hours
Guide: https://github.com/lightningnetwork/lnd/blob/master/docs/INSTALL.md

Minimum VPS specs: 2 CPU, 4GB RAM, 80GB SSD
Recommended VPS: Hetzner (Finland/Germany) or DigitalOcean (Frankfurt)

LND config (lnd.conf):
  [Application Options]
  alias=YeboBank
  maxpendingchannels=10
  
  [Bitcoin]
  bitcoin.active=1
  bitcoin.mainnet=1
  bitcoin.node=neutrino
  
  [Neutrino]
  neutrino.addpeer=btcd-mainnet.lightning.computer
  neutrino.addpeer=faucet.lightning.community

## Environment Variables

LND_REST_URL=https://your-node.voltageapp.io:8080
LND_MACAROON_PATH=/path/to/admin.macaroon
# For self-hosted only:
LND_TLS_CERT_PATH=/path/to/tls.cert

## REST API Endpoints Used

All requests require header: Grpc-Metadata-macaroon: <hex-encoded macaroon>

### Get node info
GET /v1/getinfo

### Create invoice (receive payment)
POST /v1/invoices
Body: {"value": 1000, "memo": "YeboBank deposit", "expiry": 3600}
Response: {"payment_request": "lnbc...", "r_hash": "..."}

### Check invoice status
GET /v1/invoice/:r_hash_str
Response: {"settled": true/false, "state": "SETTLED/OPEN/CANCELLED", "amt_paid_sat": "1000"}

### Send payment
POST /v2/router/send
Body: {"payment_request": "lnbc...", "timeout_seconds": 60, "fee_limit_sat": 100}
Response: {"status": "SUCCEEDED/FAILED", "payment_hash": "...", "failure_reason": ""}

### Decode invoice
GET /v1/payreq/:pay_req
Response: {"num_satoshis": "1000", "description": "...", "destination": "..."}

## Lightning Address (LNURL-pay)

Lightning Address format: user@yebobank.com
This lets any Lightning wallet pay a YeboBank user without an invoice.

How it works:
1. Payer's wallet does: GET https://yebobank.com/.well-known/lnurlp/username
2. Your server responds with LNURL metadata JSON
3. Payer's wallet does: GET https://yebobank.com/lnurlp/username?amount=50000 (millisats)
4. Your server creates a LND invoice and returns it
5. Payer's wallet pays the invoice
6. Your webhook handler credits the wallet

Required routes:
GET /.well-known/lnurlp/:username -> LNURL metadata
GET /lnurlp/:username?amount=N    -> Returns BOLT11 invoice

LNURL metadata response (Step 2):
{
  "callback": "https://yebobank.com/lnurlp/username",
  "maxSendable": 100000000,
  "minSendable": 1000,
  "metadata": "[["text/plain","Pay username@yebobank.com"]]",
  "tag": "payRequest",
  "commentAllowed": 255
}

Invoice response (Step 4):
{
  "pr": "lnbc50n1p...",
  "routes": []
}

Spec reference: https://github.com/lnurl/luds/blob/legacy/lnurl-pay.md

## Invoice Webhook (LND Settled Callback)

LND can POST to your server when an invoice is settled.
Configure in lnd.conf:
  [Application Options]
  invoicemacaroonpath=/path/to/invoice.macaroon
  invoiceexpirydelta=40

Or poll: every 10 seconds, check all pending ln_invoices WHERE status='pending'
AND expires_at > NOW() by calling GET /v1/invoice/:r_hash_str

When settled:
1. Check payment_hash not already processed (unique index on payment_hash)
2. db.CreditSats(tx, walletID, sats, "deposit_lightning", invoiceID, "ln_invoice", ...)
3. UPDATE ln_invoices SET status='settled', settled_at=NOW()

## Internal Transfers (On-Us)

When user A sends to user B's Lightning address (also on YeboBank):
- Detect: destination is @yebobank.com
- Do NOT route through LND (wastes routing fees and liquidity)
- Instead: db.Begin(), db.DebitSats(A), db.CreditSats(B), db.Commit()
- Mark as internal transfer in ledger_entries

This is a significant advantage: unlimited zero-fee instant transfers between YeboBank users.

## Watchtower

The Lightning whitepaper requires watchtower functionality.
If a channel partner broadcasts a stale commitment transaction, you have a timelock
window to respond with a breach remedy transaction.

For LND: Enable built-in watchtower.
  Add to lnd.conf:
  [watchtower]
  watchtower.active=1

Or use a third-party watchtower service (LND supports this natively).
This is required per the Lightning Network whitepaper, Section 3.3.

## Channel Management

Open channels to:
- ACINQ: 03864ef025fde8fb587d989186ce6a4a186895ee44a926bfc370e2c366597a3f8
- Wallet of Satoshi: 035e4ff418fc8b5554c5d9eea66396c227bd429a3251c8cbc711002ba215bfc226
- Voltage: node address in their dashboard

Minimum channel size: 1,000,000 sats (0.01 BTC)
Target outbound liquidity: 10,000,000 sats (for user sends)
Target inbound liquidity: 10,000,000 sats (for user receives)
