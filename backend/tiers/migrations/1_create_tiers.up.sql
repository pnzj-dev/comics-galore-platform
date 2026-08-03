CREATE TABLE tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id UUID NOT NULL REFERENCES tiers(id),
    interval TEXT NOT NULL DEFAULT 'monthly'
        CHECK (interval IN ('monthly', 'quarterly', 'yearly', 'lifetime')),
    price_usd_cents INT NOT NULL DEFAULT 0,
    quota_downloads INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_plans_tier_id ON plans(tier_id);
