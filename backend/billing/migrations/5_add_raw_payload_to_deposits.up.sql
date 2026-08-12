-- 5_add_raw_payload_to_deposits.up.sql
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS raw_payload JSONB;
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
