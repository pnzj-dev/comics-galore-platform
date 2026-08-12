-- 6_prevent_free_subscriptions.up.sql
-- The subscriptions table must never contain free-tier rows.
-- Free tier is represented by the absence of a paid subscription.

DELETE FROM subscriptions WHERE tier = 'free';

ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_tier_paid CHECK (tier <> 'free');
