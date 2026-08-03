ALTER TABLE users ADD COLUMN IF NOT EXISTS sub_partner_id TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT 'free';
ALTER TABLE users ADD CONSTRAINT users_tier_check CHECK (tier IN ('free', 'bronze', 'silver', 'gold', 'platinum'));
