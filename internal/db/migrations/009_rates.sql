-- BTC/KES rate snapshots: cached every 60 seconds from CoinGecko
CREATE TABLE rate_snapshots (
    id              BIGSERIAL   PRIMARY KEY,
    btc_kes         NUMERIC(16,4) NOT NULL,
    btc_usd         NUMERIC(16,4),
    source          TEXT        NOT NULL DEFAULT 'coingecko',
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rates_fetched_at ON rate_snapshots(fetched_at DESC);

-- Latest rate view for quick access
CREATE VIEW latest_rate AS
    SELECT btc_kes, btc_usd, fetched_at
    FROM rate_snapshots
    ORDER BY fetched_at DESC
    LIMIT 1;
