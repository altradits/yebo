-- Chamas: group savings wallets
CREATE TABLE chamas (
    id              BIGSERIAL   PRIMARY KEY,
    name            TEXT        NOT NULL,
    description     TEXT,
    wallet_id       BIGINT      NOT NULL UNIQUE REFERENCES wallets(id),
    created_by      BIGINT      NOT NULL REFERENCES users(id),
    contribution_sats BIGINT    NOT NULL DEFAULT 0,
    cycle_days      INT         NOT NULL DEFAULT 30,
    max_members     INT         NOT NULL DEFAULT 20,
    status          TEXT        NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active','paused','closed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chamas_created_by ON chamas(created_by);
CREATE INDEX idx_chamas_status     ON chamas(status);

CREATE TABLE chama_members (
    chama_id        BIGINT      NOT NULL REFERENCES chamas(id) ON DELETE CASCADE,
    user_id         BIGINT      NOT NULL REFERENCES users(id),
    role            TEXT        NOT NULL DEFAULT 'member'
                                CHECK (role IN ('admin','member')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chama_id, user_id)
);

CREATE TABLE chama_votes (
    id              BIGSERIAL   PRIMARY KEY,
    chama_id        BIGINT      NOT NULL REFERENCES chamas(id),
    proposal        TEXT        NOT NULL,
    proposed_by     BIGINT      NOT NULL REFERENCES users(id),
    status          TEXT        NOT NULL DEFAULT 'open'
                                CHECK (status IN ('open','passed','rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closes_at       TIMESTAMPTZ NOT NULL
);

CREATE TABLE chama_vote_responses (
    vote_id         BIGINT      NOT NULL REFERENCES chama_votes(id),
    user_id         BIGINT      NOT NULL REFERENCES users(id),
    choice          TEXT        NOT NULL CHECK (choice IN ('yes','no','abstain')),
    voted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (vote_id, user_id)
);
