-- 4_create_payments.up.sql
-- Tracks all payment events from NowPayments webhooks for KPI calculations

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL DEFAULT 'nowpayments',
    provider_payment_id TEXT NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    plan_id UUID,
    tier TEXT NOT NULL DEFAULT 'free',
    interval TEXT,
    amount_crypto NUMERIC,
    currency_crypto TEXT,
    amount_usd_cents INTEGER,
    status TEXT NOT NULL DEFAULT 'pending',
    raw_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_tier ON payments(tier);
