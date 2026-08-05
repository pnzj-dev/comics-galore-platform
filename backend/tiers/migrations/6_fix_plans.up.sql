-- Cleanup and re-seed plans with correct pricing, features, and intervals.
-- Run this migration to fix discrepancies from old seed data.

ALTER TABLE plans DROP CONSTRAINT IF EXISTS plans_interval_check;
ALTER TABLE plans ADD CONSTRAINT plans_interval_check CHECK (interval IN ('monthly', 'quarterly', 'semesterly', 'yearly'));

DELETE FROM plans;

INSERT INTO plans (tier_id, name, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, t.name, 'monthly', CASE t.name
    WHEN 'Free' THEN 0 WHEN 'Bronze' THEN 399 WHEN 'Silver' THEN 699 WHEN 'Gold' THEN 999 WHEN 'Platinum' THEN 1999
END, CASE t.name
    WHEN 'Free' THEN 1 WHEN 'Bronze' THEN 10 WHEN 'Silver' THEN 50 WHEN 'Gold' THEN 200 WHEN 'Platinum' THEN 1000000
END, CASE t.name
    WHEN 'Free' THEN '["Browse comics","Read comments","1 GB download quota"]'::jsonb
    WHEN 'Bronze' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","10 GB download quota"]'::jsonb
    WHEN 'Silver' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","50 GB download quota"]'::jsonb
    WHEN 'Gold' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","200 GB download quota"]'::jsonb
    WHEN 'Platinum' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","1 TB download quota"]'::jsonb
END, 30
FROM tiers t;

INSERT INTO plans (tier_id, name, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, t.name, 'quarterly', CASE t.name
    WHEN 'Bronze' THEN 1077 WHEN 'Silver' THEN 1887 WHEN 'Gold' THEN 2697 WHEN 'Platinum' THEN 5397
END, CASE t.name
    WHEN 'Bronze' THEN 10 WHEN 'Silver' THEN 50 WHEN 'Gold' THEN 200 WHEN 'Platinum' THEN 1000000
END, CASE t.name
    WHEN 'Bronze' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","10 GB download quota"]'::jsonb
    WHEN 'Silver' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","50 GB download quota"]'::jsonb
    WHEN 'Gold' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","200 GB download quota"]'::jsonb
    WHEN 'Platinum' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","1 TB download quota"]'::jsonb
END, 90
FROM tiers t
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum');

INSERT INTO plans (tier_id, name, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, t.name, 'semesterly', CASE t.name
    WHEN 'Bronze' THEN 2035 WHEN 'Silver' THEN 3565 WHEN 'Gold' THEN 5095 WHEN 'Platinum' THEN 10195
END, CASE t.name
    WHEN 'Bronze' THEN 10 WHEN 'Silver' THEN 50 WHEN 'Gold' THEN 200 WHEN 'Platinum' THEN 1000000
END, CASE t.name
    WHEN 'Bronze' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","10 GB download quota"]'::jsonb
    WHEN 'Silver' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","50 GB download quota"]'::jsonb
    WHEN 'Gold' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","200 GB download quota"]'::jsonb
    WHEN 'Platinum' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","1 TB download quota"]'::jsonb
END, 180
FROM tiers t
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum');

INSERT INTO plans (tier_id, name, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, t.name, 'yearly', CASE t.name
    WHEN 'Bronze' THEN 3884 WHEN 'Silver' THEN 6806 WHEN 'Gold' THEN 9722 WHEN 'Platinum' THEN 19390
END, CASE t.name
    WHEN 'Bronze' THEN 10 WHEN 'Silver' THEN 50 WHEN 'Gold' THEN 200 WHEN 'Platinum' THEN 1000000
END, CASE t.name
    WHEN 'Bronze' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","10 GB download quota"]'::jsonb
    WHEN 'Silver' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","50 GB download quota"]'::jsonb
    WHEN 'Gold' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","200 GB download quota"]'::jsonb
    WHEN 'Platinum' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts","1 TB download quota"]'::jsonb
END, 365
FROM tiers t
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum');

UPDATE plans SET currency = 'USD' WHERE currency IS NULL;
UPDATE plans SET name = tiers.name FROM tiers WHERE plans.tier_id = tiers.id AND plans.name IS NULL;
