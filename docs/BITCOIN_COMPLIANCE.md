# YeboBank - Bitcoin Whitepaper Compliance

## What the Bitcoin Whitepaper Requires of YeboBank

The Bitcoin whitepaper (Nakamoto, 2008) is a technical specification for a peer-to-peer
electronic cash system. It is not a legal document. It contains no licensing requirements.

YeboBank is a custodial service layer built ON TOP of Bitcoin and Lightning.
The whitepaper's rules apply to us as design values, not as a regulatory checklist.

## Five Obligations from the Whitepapers

### 1. Transaction Immutability (Bitcoin Whitepaper, Section 2)
"We define an electronic coin as a chain of digital signatures."

Once a Lightning payment confirms, it is final.
Your code must treat confirmed Lightning payments as immutable.
No chargebacks. No reversals without explicit admin action with full audit trail.

Implementation:
- ln_payments table: status only goes from pending -> succeeded (or failed)
- Succeeded payments are never deleted or modified
- Any correction requires a new reverse ledger entry, not an update

### 2. Proof of Work / Append-Only Ledger (Bitcoin Whitepaper, Section 3-4)
"The proof-of-work involves scanning for a value that when hashed..."

Bitcoin's blockchain is append-only. Your internal ledger must be the same.
ledger_entries is INSERT-ONLY. No UPDATE. No DELETE. Ever.
If a correction is needed, add a reverse entry with type="admin_correction".

### 3. No Double-Spend (Bitcoin Whitepaper, Section 2)
The entire Bitcoin design exists to prevent double-spending.

Your implementation:
- mpesa_reference UNIQUE index: prevents crediting the same M-Pesa receipt twice
- payment_hash UNIQUE index: prevents crediting the same Lightning payment twice
- FOR UPDATE lock in CreditSats/DebitSats: prevents race conditions on same wallet
- All balance changes in SERIALIZABLE transactions

### 4. Watchtower Requirement (Lightning Network Whitepaper, Section 3.3.5)
"It should be expected that... a third party can be delegated by only giving the
Breach Remedy transaction to this third party."

If a channel partner broadcasts a stale commitment transaction, you have a timelock
window to respond. If you miss the window, they can steal funds.

Implementation:
- Enable LND built-in watchtower: [watchtower] watchtower.active=1 in lnd.conf
- OR use a third-party watchtower service
- Monitor channel health daily
- Set up alerts for unexpected channel closures

### 5. Incentive Alignment (Bitcoin Whitepaper, Section 6)
"The incentive can help encourage nodes to stay honest."

Nakamoto designed miners to earn only when they act honestly.
YeboBank earns only when customers earn: 2% of pool profit, never of principal.
If the pool earns nothing, YeboBank earns nothing.
This is not charity — it is correct incentive design per the whitepaper.

Implementation:
- admin_fee_sats = pool_profit * 0.02 only
- Principal (savings_locks.principal_sats) is never touched for fees
- If pool_profit = 0, admin_fee = 0

## What YeboBank Does NOT Need to Do

- Run a Bitcoin full node (we use LND which handles this)
- Mine Bitcoin (not applicable for a service layer)
- Implement HTLC scripts (LND handles this)
- Manage private keys for each user (we are custodial - one node key for all)
- Implement the Bitcoin wire protocol (LND handles this)

## Legal Context (Kenya)

The Bitcoin whitepaper does not supersede Kenyan law.
The following Kenyan laws apply to YeboBank:

- National Payments System Act (NPSA) - covers payment facilitation
- Capital Markets Authority Act - covers investment products (savings with yield)
- Central Bank of Kenya Act - covers money transmission

Recommended path: Apply for CBK Regulatory Sandbox
URL: https://www.centralbank.go.ke/fintech/
This gives 12-24 months to operate legally while proving the product.

The human rights case (financial access for low-income Kenyans) is strengthened,
not weakened, by operating within the regulatory framework. A licensed YeboBank
cannot be shut down on a regulator's whim.
