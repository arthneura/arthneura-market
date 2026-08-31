CREATE TABLE IF NOT EXISTS agents (
    did BYTEA PRIMARY KEY,
    controller TEXT,
    raw_event JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS commitments (
    commitment_id BYTEA PRIMARY KEY,
    provider BYTEA,
    consumer BYTEA,
    merkle_root BYTEA,
    raw_event JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
