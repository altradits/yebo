-- Lightning invoices: created by our LND node for deposits
CREATE TABLE ln_invoices (
    id              BIGSERIAL   PRIMARY KEY,
    payment_hash    TEXT        NOT NULL UNIQUE,
    payment_request TEXT        NOT NULL,
    amount_sats     BIGINT      NOT NULL CHECK (amount_sats > 0),
    memo            TEXT,
    wallet_id       BIGINT      NOT NULL REFERENCES wallets(id),
    status          TEXT        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','settled','expired','cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    settled_at      TIMESTAMPTZ
);

CREATE INDEX idx_ln_invoices_payment_hash ON ln_invoices(payment_hash);
CREATE INDEX idx_ln_invoices_wallet_id    ON ln_invoices(wallet_id);
CREATE INDEX idx_ln_invoices_status       ON ln_invoices(status);

-- Lightning payments: outbound payments sent via LND
CREATE TABLE ln_payments (
    id              BIGSERIAL   PRIMARY KEY,
    payment_hash    TEXT        NOT NULL UNIQUE,
    payment_request TEXT        NOT NULL,
    amount_sats     BIGINT      NOT NULL,
    fee_sats        BIGINT      NOT NULL DEFAULT 0,
    wallet_id       BIGINT      NOT NULL REFERENCES wallets(id),
    status          TEXT        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','succeeded','failed')),
    failure_reason  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at      TIMESTAMPTZ
);

CREATE INDEX idx_ln_payments_payment_hash ON ln_payments(payment_hash);
CREATE INDEX idx_ln_payments_wallet_id    ON ln_payments(wallet_id);
CREATE INDEX idx_ln_payments_status       ON ln_payments(status);
