-- Purchased download-quota boosts. A boost permanently adds `downloads` to the
-- user's monthly download allowance. Granted by the billing service once the
-- associated NowPayments deposit is confirmed.
CREATE TABLE quota_boosts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    downloads  INT  NOT NULL CHECK (downloads > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quota_boosts_user ON quota_boosts(user_id);
