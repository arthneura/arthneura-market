CREATE TABLE IF NOT EXISTS offers (
  id            BIGSERIAL PRIMARY KEY,
  listing_id    BIGINT NOT NULL,
  from_did      BYTEA NOT NULL,
  to_did        BYTEA NOT NULL,
  price         BIGINT NOT NULL,
  expires_at    TIMESTAMPTZ NOT NULL,
  status        TEXT NOT NULL DEFAULT 'open',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS offers_listing ON offers (listing_id);
