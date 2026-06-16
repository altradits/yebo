# YeboBank — Database Schema

## The Ledger Rule (Read This First)

The ledger_entries table is the SOURCE OF TRUTH.
wallets.balance_sats is a CACHE of the ledger.
Every write to wallets.balance_sats MUST have a ledger_entries row in the SAME transaction.
The only way to touch a balance is through db.CreditSats() or db.DebitSats().
These are defined in internal/db/ledger.go.
Never bypass them. Never.

## Migration Files

Run in order. Auto-applied on server startup via internal/db/migrations.go.

- 001_users.sql
- 002_wallets_ledger.sql
- 003_savings.sql
- 004_chamas.sql
- 005_agents.sql
- 006_treasury.sql
- 007_mpesa.sql
- 008_lightning.sql
- 009_rates.sql

## Schema Reference

### users
| Column         | Type         | Notes                                         |
|----------------|--------------|-----------------------------------------------|
| id             | UUID PK      | gen_random_uuid()                             |
| phone          | VARCHAR(20)  | UNIQUE, +2547XXXXXXXX format                  |
| email          | VARCHAR(255) | UNIQUE, optional                              |
| password_hash  | VARCHAR(128) | PBKDF2-SHA256 output, hex-encoded             |
| password_salt  | VARCHAR(64)  | 32 random bytes, hex-encoded                  |
| full_name      | VARCHAR(255) | Optional                                      |
| role           | VARCHAR(20)  | customer / agent / trader / admin             |
| is_blocked     | BOOLEAN      | Default FALSE                                 |
| is_verified    | BOOLEAN      | Default FALSE (optional KYC tier)             |
| pin_hash       | VARCHAR(128) | 6-digit transaction PIN, same PBKDF2 treatment|
| pin_salt       | VARCHAR(64)  |                                               |
| language       | VARCHAR(5)   | 'en' or 'sw' (Swahili)                        |
| created_at     | TIMESTAMPTZ  |                                               |
| last_seen      | TIMESTAMPTZ  | Updated on every authenticated request        |
| last_ip        | INET         |                                               |

### sessions
| Column       | Type         | Notes                                           |
|--------------|--------------|-------------------------------------------------|
| id           | UUID PK      |                                                 |
| user_id      | UUID FK      | -> users(id) CASCADE DELETE                     |
| token        | VARCHAR(64)  | UNIQUE, 32 random bytes hex-encoded             |
| expires_at   | TIMESTAMPTZ  | Absolute expiry (72 hours from creation)        |
| last_active  | TIMESTAMPTZ  | Updated on each request, checked for idle (2hr) |
| ip_address   | INET         |                                                 |
| user_agent   | TEXT         |                                                 |

### wallets
| Column        | Type         | Notes                                          |
|---------------|--------------|------------------------------------------------|
| id            | UUID PK      |                                                |
| user_id       | UUID FK      | UNIQUE -> users(id)                            |
| balance_sats  | BIGINT       | CHECK >= 0. CACHE ONLY. Use ledger to change.  |
| lightning_addr| VARCHAR(255) | UNIQUE, user@yebobank.com                      |
| updated_at    | TIMESTAMPTZ  | Updated every time balance changes             |

### ledger_entries (APPEND ONLY — NO UPDATE, NO DELETE)
| Column         | Type         | Notes                                         |
|----------------|--------------|-----------------------------------------------|
| id             | UUID PK      |                                               |
| wallet_id      | UUID FK      | -> wallets(id)                                |
| type           | VARCHAR(40)  | See credit/debit type list below              |
| direction      | VARCHAR(6)   | CHECK IN ('credit', 'debit')                  |
| amount_sats    | BIGINT       | CHECK > 0, always positive                    |
| balance_after  | BIGINT       | Wallet balance after this entry               |
| reference_id   | UUID         | FK to mpesa_transactions, ln_invoices, etc.   |
| reference_type | VARCHAR(40)  | 'mpesa_deposit', 'ln_invoice', etc.           |
| actor_id       | UUID FK      | Who initiated (user or admin)                 |
| note           | TEXT         |                                               |
| created_at     | TIMESTAMPTZ  |                                               |

Credit types: deposit_mpesa, deposit_lightning, interest_payment,
              chama_distribution, agent_commission, refund, admin_credit

Debit types: withdrawal_mpesa, withdrawal_lightning, savings_lock,
             chama_contribution, agent_cash_out, fee, admin_debit

### savings_locks
| Column                | Type         | Notes                                    |
|-----------------------|--------------|------------------------------------------|
| id                    | UUID PK      |                                          |
| wallet_id             | UUID FK      | -> wallets(id)                           |
| user_id               | UUID FK      | -> users(id)                             |
| principal_sats        | BIGINT       | CHECK >= 1000 (minimum lock: 1000 sats)  |
| accrued_sats          | BIGINT       | Interest accumulated, starts at 0        |
| lock_years            | INT          | CHECK >= 5                               |
| status                | VARCHAR(20)  | active / matured / withdrawn / early_exit|
| locked_at             | TIMESTAMPTZ  |                                          |
| matures_at            | TIMESTAMPTZ  |                                          |
| withdrawn_at          | TIMESTAMPTZ  |                                          |
| early_exit_penalty_pct| INT          | 0-100, set at time of early exit         |

Early exit penalty schedule:
- Under 1 year: 100% of accrued interest forfeited
- 1-2 years: 75%
- 2-3 years: 50%
- 3-4 years: 25%
- 4-5 years: 10%
- At or after maturity: 0%

### interest_distributions
| Column           | Type        | Notes                                      |
|------------------|-------------|---------------------------------------------|
| id               | UUID PK     |                                             |
| period_month     | DATE        | UNIQUE. First day of month distributed.     |
| pool_profit_sats | BIGINT      | Total profit logged by trader this month    |
| admin_fee_sats   | BIGINT      | 2% of pool_profit                           |
| distributed_sats | BIGINT      | pool_profit - admin_fee (less rounding rem) |
| total_locked_sats| BIGINT      | Sum of all active principals at run time    |
| lock_count       | INT         | Number of active locks                      |
| run_by           | UUID FK     | Admin user who triggered distribution       |
| run_at           | TIMESTAMPTZ |                                             |

### chamas
| Column             | Type         | Notes                                      |
|--------------------|-------------|---------------------------------------------|
| id                 | UUID PK     |                                             |
| name               | VARCHAR(255)|                                             |
| description        | TEXT        |                                             |
| created_by         | UUID FK     | -> users(id)                                |
| wallet_id          | UUID FK     | UNIQUE -> wallets(id), chama has own wallet |
| contribution_sats  | BIGINT      | Expected monthly contribution per member    |
| max_members        | INT         | Default 50                                  |
| status             | VARCHAR(20) | active / suspended / dissolved              |
| created_at         | TIMESTAMPTZ |                                             |

### agents
| Column             | Type           | Notes                                    |
|--------------------|----------------|------------------------------------------|
| id                 | UUID PK        |                                          |
| user_id            | UUID FK        | UNIQUE -> users(id)                      |
| status             | VARCHAR(20)    | pending / active / suspended             |
| location_name      | VARCHAR(255)   | "Gikomba Market, Nairobi"                |
| location_lat       | NUMERIC(10,7)  |                                          |
| location_lng       | NUMERIC(10,7)  |                                          |
| float_limit_sats   | BIGINT         | Max they can facilitate, default 10M     |
| commission_rate    | NUMERIC(5,4)   | Default 0.005 (0.5%)                     |
| total_earned_sats  | BIGINT         | Lifetime commission                      |
| approved_by        | UUID FK        |                                          |
| approved_at        | TIMESTAMPTZ    |                                          |

### mpesa_transactions
| Column              | Type            | Notes                                  |
|---------------------|-----------------|----------------------------------------|
| id                  | UUID PK         |                                        |
| user_id             | UUID FK         |                                        |
| type                | VARCHAR(20)     | stk_push / b2c                         |
| direction           | VARCHAR(6)      | in / out                               |
| kes_amount          | NUMERIC(14,2)   |                                        |
| sats_amount         | BIGINT          | NULL until rate confirmed              |
| kes_rate            | NUMERIC(14,8)   | Rate used (sats per KES)               |
| mpesa_reference     | VARCHAR(50)     | Safaricom confirmation code. IDEMPOTENCY KEY |
| checkout_request_id | VARCHAR(100)    | UNIQUE. From STK Push response.        |
| result_code         | INT             | 0 = success                            |
| result_desc         | TEXT            |                                        |
| phone               | VARCHAR(20)     |                                        |
| status              | VARCHAR(20)     | pending / confirmed / failed / expired |
| initiated_at        | TIMESTAMPTZ     |                                        |
| confirmed_at        | TIMESTAMPTZ     |                                        |
| ledger_entry_id     | UUID FK         | -> ledger_entries(id)                  |

### ln_invoices
| Column          | Type         | Notes                                         |
|-----------------|--------------|-----------------------------------------------|
| id              | UUID PK      |                                               |
| user_id         | UUID FK      |                                               |
| payment_hash    | VARCHAR(64)  | UNIQUE. 32 bytes hex. IDEMPOTENCY KEY.        |
| payment_request | TEXT         | BOLT11 invoice string                         |
| amount_sats     | BIGINT       |                                               |
| description     | TEXT         |                                               |
| status          | VARCHAR(20)  | pending / settled / expired / cancelled       |
| expires_at      | TIMESTAMPTZ  |                                               |
| settled_at      | TIMESTAMPTZ  |                                               |
| ledger_entry_id | UUID FK      |                                               |

### rate_snapshots
| Column      | Type           | Notes                                         |
|-------------|----------------|-----------------------------------------------|
| id          | UUID PK        |                                               |
| btc_usd     | NUMERIC(18,2)  |                                               |
| btc_kes     | NUMERIC(18,2)  |                                               |
| sats_per_kes| NUMERIC(14,8)  | How many sats 1 KES buys                      |
| kes_per_sat | NUMERIC(14,8)  | How much 1 sat costs in KES                   |
| source      | VARCHAR(50)    | 'coingecko'                                   |
| fetched_at  | TIMESTAMPTZ    |                                               |

Get current rate: SELECT * FROM rate_snapshots ORDER BY fetched_at DESC LIMIT 1

## Indexes (Critical for Performance)

All foreign keys are indexed.
All status columns are indexed.
All created_at columns on high-volume tables have DESC indexes.
payment_hash and mpesa_reference have UNIQUE indexes (idempotency).
