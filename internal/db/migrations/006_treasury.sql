-- Treasury: yield-generating assets that fund interest payments
-- Interest paid to users must come ONLY from treasury yield, never deposits
CREATE TABLE treasury_assets (
    id              BIGSERIAL   PRIMARY KEY,
    name            TEXT        NOT NULL,
    asset_type      TEXT        NOT NULL CHECK (asset_type IN ('lightning_routing','btc_yield','other')),
    balance_sats    BIGINT      NOT NULL DEFAULT 0,
    apy_bps         INT         NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Interest distribution log: one row per monthly run
CREATE TABLE interest_distributions (
    id                  BIGSERIAL   PRIMARY KEY,
    run_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_locked_sats   BIGINT      NOT NULL,
    total_interest_sats BIGINT      NOT NULL,
    treasury_profit_sats BIGINT     NOT NULL,
    accounts_credited   INT         NOT NULL,
    rate_bps            INT         NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'complete'
                                    CHECK (status IN ('running','complete','failed')),
    run_by              BIGINT      REFERENCES users(id)
);

INSERT INTO treasury_assets (name, asset_type, apy_bps, notes)
VALUES ('Lightning Routing Fees', 'lightning_routing', 0, 'Accumulated routing fees from LND node');
