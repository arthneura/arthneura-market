ALTER TABLE commitments
  ADD COLUMN IF NOT EXISTS deliver_url TEXT;
