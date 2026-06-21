-- M-Pesa transactions: idempotency-keyed to prevent double-crediting
CREATE TABLE mpesa_transactions (
    id              BIGSERIAL   PRIMARY KEY,
    mpesa_receipt   TEXT        NOT NULL UNIQUE,
    type            TEXT        NOT NULL CHECK (type IN ('stk_push','b2c')),
    phone           TEXT        NOT NULL,
    amount_kes      NUMERIC(12,2) NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','completed','failed','reversed')),
    wallet_id       BIGINT      REFERENCES wallets(id),
    checkout_request_id TEXT,
    result_code     TEXT,
    result_desc     TEXT,
    initiated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_mpesa_receipt     ON mpesa_transactions(mpesa_receipt);
CREATE INDEX idx_mpesa_phone       ON mpesa_transactions(phone);
CREATE INDEX idx_mpesa_status      ON mpesa_transactions(status);
CREATE INDEX idx_mpesa_initiated   ON mpesa_transactions(initiated_at);
CREATE INDEX idx_mpesa_checkout_id ON mpesa_transactions(checkout_request_id);
