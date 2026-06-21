-- Savings locks: fixed-term deposits with interest
CREATE TABLE savings_locks (
    id              BIGSERIAL   PRIMARY KEY,
    wallet_id       BIGINT      NOT NULL REFERENCES wallets(id),
    amount_sats     BIGINT      NOT NULL CHECK (amount_sats > 0),
    lock_days       INT         NOT NULL DEFAULT 30,
    interest_rate_bps INT       NOT NULL,
    locked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unlocks_at      TIMESTAMPTZ NOT NULL,
    unlocked_at     TIMESTAMPTZ,
    status          TEXT        NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active','unlocked','early_exit')),
    interest_earned_sats BIGINT NOT NULL DEFAULT 0,
    early_exit_penalty_bps INT  NOT NULL DEFAULT 2000
);

CREATE INDEX idx_savings_wallet_id  ON savings_locks(wallet_id);
CREATE INDEX idx_savings_status     ON savings_locks(status);
CREATE INDEX idx_savings_unlocks_at ON savings_locks(unlocks_at);

-- Pool settings: admin-controlled global parameters
CREATE TABLE pool_settings (
    id                  INT  PRIMARY KEY DEFAULT 1,
    interest_rate_bps   INT  NOT NULL DEFAULT 500,
    min_savings_sats    BIGINT NOT NULL DEFAULT 50000,
    max_savings_sats    BIGINT NOT NULL DEFAULT 100000000000,
    lock_days           INT  NOT NULL DEFAULT 30,
    early_exit_penalty_bps INT NOT NULL DEFAULT 2000,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
