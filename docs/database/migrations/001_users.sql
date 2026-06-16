-- Users: phone-first registration, optional KYC, role-based access
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    phone           TEXT        NOT NULL UNIQUE,
    full_name       TEXT,
    email           TEXT,
    password_hash   TEXT        NOT NULL,
    pin_hash        TEXT,
    role            TEXT        NOT NULL DEFAULT 'customer'
                                CHECK (role IN ('customer','agent','trader','admin')),
    kyc_status      TEXT        NOT NULL DEFAULT 'none'
                                CHECK (kyc_status IN ('none','pending','verified','rejected')),
    language        TEXT        NOT NULL DEFAULT 'en'
                                CHECK (language IN ('en','sw')),
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role  ON users(role);

-- Sessions: server-side only, no JWT
CREATE TABLE sessions (
    token       TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '72 hours'
);

CREATE INDEX idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Audit log: immutable record of every admin action
CREATE TABLE audit_log (
    id          BIGSERIAL   PRIMARY KEY,
    actor_id    BIGINT      REFERENCES users(id),
    action      TEXT        NOT NULL,
    target_type TEXT,
    target_id   BIGINT,
    detail      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
