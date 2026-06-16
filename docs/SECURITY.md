# YeboBank — Security Specification

## Password Storage

Algorithm: PBKDF2-HMAC-SHA256
Iterations: 310,000 (NIST SP 800-132 recommendation, 2023)
Salt length: 32 bytes (256 bits), randomly generated per user
Output length: 32 bytes (256 bits)
Encoding: hex string stored in password_hash column

This is implemented WITHOUT external libraries using only Go stdlib (crypto/sha256).
The full implementation is in internal/utils/crypto.go.

Never use MD5, SHA-1, or plain SHA-256 for passwords.
Never use a shared/global salt.
Never store plaintext passwords anywhere, including logs.

## Transaction PIN

Separate 6-digit PIN required to confirm any debit:
- Withdrawals
- Lightning sends
- Savings lock creation
- Chama contributions
- Agent cash-out

Same PBKDF2 treatment as the main password.
PIN is stored in users.pin_hash and users.pin_salt.
Never transmitted without TLS.
Never logged.

## Session Security

Token generation: crypto/rand, 32 bytes, hex-encoded = 64 char string
Token storage: users see it as a cookie only. DB stores the token.
Idle timeout: 2 hours (SESSION_IDLE_TIMEOUT_HOURS env var)
Absolute expiry: 72 hours from creation
On logout: DELETE FROM sessions WHERE token = $1
On password change: DELETE FROM sessions WHERE user_id = $1 (all sessions)
Concurrent sessions: allowed (user on phone + desktop at same time)

Session cookie settings:
- HttpOnly: true (JavaScript cannot read it)
- SameSite: Lax
- Secure: true (HTTPS only in production)
- Path: /

## Rate Limiting

Auth endpoints (/login, /register): 10 requests per minute per IP
Payment endpoints: 20 transactions per hour per user
Admin endpoints: 100 requests per minute (no user rate limit)

Implementation: token bucket in memory (internal/middleware/ratelimit.go)
For multi-instance deployments: move counters to Redis.

## M-Pesa Callback Authentication

Step 1: Allowlist Safaricom IP ranges in nginx BEFORE request hits Go.
  Sandbox IPs: 196.201.214.0/24
  Production IPs: published at https://developer.safaricom.co.ke/docs

nginx config:
  location /api/mpesa/callback {
      allow 196.201.214.0/24;
      deny all;
      proxy_pass http://localhost:8080;
  }

Step 2: In Go, reject any callback without a valid CheckoutRequestID matching
a pending mpesa_transactions row. Do not process unknown callbacks.

Step 3: After processing, always respond HTTP 200 with:
  {"ResultCode": 0, "ResultDesc": "Accepted"}
  Safaricom retries on non-200 responses.

## LND Macaroon Security

Use readonly.macaroon for: balance queries, invoice lookups
Use admin.macaroon for: CreateInvoice, SendPayment

Store macaroon path in LND_MACAROON_PATH environment variable.
Never commit macaroon files to git.
Never log macaroon hex.
Set file permissions: chmod 600 on macaroon files.

## TLS

All production traffic must be HTTPS.
Go server does NOT handle TLS directly.
nginx handles TLS termination with Let's Encrypt certificates.
Certificate renewal: certbot --nginx, runs via cron.
HTTP to HTTPS redirect: enforced in nginx.

## Actions Requiring Audit Log

Every admin action must write to ledger_entries or a dedicated audit table:
- Deposit approved
- Deposit rejected
- Withdrawal approved
- Withdrawal rejected
- User blocked
- User unblocked
- Agent approved
- Agent suspended
- Interest distribution run
- Pool settings changed
- Admin manual credit
- Admin manual debit

## SQL Injection Prevention

All database queries use parameterized $1, $2, $3 placeholders.
No string concatenation in SQL queries. Ever.
The custom pgdrv interpolates values safely (see internal/pgdrv/pgdrv.go).

## What Not To Store In Logs

- Passwords (any form)
- PINs (any form)
- Session tokens
- Macaroon hex
- Full payment hashes (truncate to first 8 chars in logs)
- Full M-Pesa receipt numbers in high-volume logs (use last 4 chars)
- Phone numbers in high-volume logs (use last 4 digits)
