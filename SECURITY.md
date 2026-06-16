# Security Policy

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Email `security@yebobank.org` with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested mitigations

Response within 48 hours, full resolution within 7 days.

## Non-Negotiable Security Requirements

### Password & PIN Storage
- PBKDF2-SHA256, 310,000 iterations, 32-byte random salt
- Transaction PIN (6-digit) stored with identical treatment
- Never transmitted in plaintext

### Sessions
- 32 random bytes encoded as 64-char hex — stored server-side in DB
- Idle timeout: 2 hours (configurable via `SESSION_IDLE_TIMEOUT_HOURS`)
- Absolute expiry: 72 hours
- On logout: DELETE the session row
- On password change: DELETE ALL session rows for that user

### TLS
- nginx handles TLS termination (Let's Encrypt)
- Go server never exposes port 443 directly

### Rate Limiting
- Auth endpoints: 10 requests/minute per IP
- Payment endpoints: 20 transactions/hour per user
- Token bucket in memory (middleware/ratelimit.go)

### M-Pesa Callbacks
- Safaricom IP allowlist enforced at nginx level
- Sandbox: `196.201.214.0/24`
- Production: see Daraja documentation

### LND Macaroons
- `readonly.macaroon` for balance queries only
- `admin.macaroon` for invoice creation and payments only
- Never logged, never committed to git

### Mandatory Audit Log Entries
Every deposit approval, withdrawal approval, user block/unblock, interest distribution run, pool settings change, and agent approval must produce an audit row.

## Disclosure Policy
We credit researchers who report valid issues (unless they prefer anonymity). We do not pursue legal action against good-faith researchers.
