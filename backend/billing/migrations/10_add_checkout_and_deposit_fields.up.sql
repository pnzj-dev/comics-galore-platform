-- Link deposits/subscriptions to a Temporal "checkout" workflow so webhooks
-- can map a payment event back to the running workflow. Also store the deposit
-- fields the frontend needs to render the payment screen (extra id, network,
-- expiry) so the state endpoint can return them without recomputation.
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS checkout_id TEXT;
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS payin_extra_id TEXT;
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS network TEXT;
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS checkout_id TEXT;

CREATE INDEX IF NOT EXISTS idx_deposits_checkout ON deposits(checkout_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_checkout ON subscriptions(checkout_id);
