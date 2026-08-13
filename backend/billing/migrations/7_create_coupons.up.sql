CREATE TABLE IF NOT EXISTS coupons (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       TEXT UNIQUE NOT NULL,
    percent_off INT NOT NULL DEFAULT 0 CHECK (percent_off >= 0 AND percent_off <= 100),
    tier       TEXT NOT NULL DEFAULT '',
    max_uses   INT NOT NULL DEFAULT 0,
    used       INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
