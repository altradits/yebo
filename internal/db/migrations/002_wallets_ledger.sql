-- NEVER update balance_sats directly — use db.CreditSats() / db.DebitSats()
CREATE TABLE wallets (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance_sats    BIGINT      NOT NULL DEFAULT 0 CHECK (balance_sats >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

-- Ledger: append-only source of truth for every satoshi movement
CREATE TABLE ledger_entries (
    id              BIGSERIAL   PRIMARY KEY,
    wallet_id       BIGINT      NOT NULL REFERENCES wallets(id),
    amount_sats     BIGINT      NOT NULL,
    balance_after   BIGINT      NOT NULL,
    type            TEXT        NOT NULL
                                CHECK (type IN (
                                    'deposit','withdrawal','send','receive',
                                    'savings_lock','savings_unlock','savings_interest',
                                    'chama_contribution','chama_payout',
                                    'agent_cashin','agent_cashout','agent_commission',
                                    'admin_adjustment'
                                )),
    ref_id          TEXT,
    note            TEXT,
    actor_id        BIGINT      REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_wallet_id  ON ledger_entries(wallet_id);
CREATE INDEX idx_ledger_type       ON ledger_entries(type);
CREATE INDEX idx_ledger_created_at ON ledger_entries(created_at);
CREATE INDEX idx_ledger_ref_id     ON ledger_entries(ref_id);
