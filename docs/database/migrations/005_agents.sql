-- Agents: cash-in / cash-out operators with commission tracking
CREATE TABLE agents (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL UNIQUE REFERENCES users(id),
    business_name   TEXT        NOT NULL,
    location        TEXT,
    latitude        NUMERIC(9,6),
    longitude       NUMERIC(9,6),
    commission_rate_bps INT     NOT NULL DEFAULT 100,
    float_sats      BIGINT      NOT NULL DEFAULT 0 CHECK (float_sats >= 0),
    status          TEXT        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','active','suspended')),
    approved_by     BIGINT      REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agents_user_id ON agents(user_id);
CREATE INDEX idx_agents_status  ON agents(status);

CREATE TABLE agent_transactions (
    id              BIGSERIAL   PRIMARY KEY,
    agent_id        BIGINT      NOT NULL REFERENCES agents(id),
    customer_wallet BIGINT      NOT NULL REFERENCES wallets(id),
    type            TEXT        NOT NULL CHECK (type IN ('cash_in','cash_out')),
    amount_sats     BIGINT      NOT NULL CHECK (amount_sats > 0),
    commission_sats BIGINT      NOT NULL DEFAULT 0,
    mpesa_ref       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_tx_agent_id   ON agent_transactions(agent_id);
CREATE INDEX idx_agent_tx_created_at ON agent_transactions(created_at);
