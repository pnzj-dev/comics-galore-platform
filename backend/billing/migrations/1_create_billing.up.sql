CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    provider TEXT NOT NULL DEFAULT 'nowpayments',
    provider_subscription_id TEXT,
    provider_invoice_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'expired', 'cancelled', 'failed')),
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    provider TEXT NOT NULL DEFAULT 'nowpayments',
    provider_invoice_id TEXT,
    amount_crypto TEXT,
    currency_crypto TEXT NOT NULL DEFAULT 'BTC',
    amount_usd_cents INT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'expired', 'failed')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL DEFAULT 'nowpayments',
    event_type TEXT NOT NULL,
    external_id TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_payment_attempts_user ON payment_attempts(user_id);
CREATE INDEX idx_webhook_events_provider ON webhook_events(provider);
CREATE INDEX idx_webhook_events_external ON webhook_events(external_id);
