INSERT INTO tiers (name, description, sort_order) VALUES
    ('Platinum', 'Top tier with full access and 1 TB quota', 4)
ON CONFLICT DO NOTHING;

DELETE FROM plans WHERE interval = 'lifetime';

INSERT INTO plans (tier_id, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, 'monthly', CASE t.name
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
FROM tiers t
WHERE NOT EXISTS (SELECT 1 FROM plans WHERE plans.tier_id = t.id AND plans.interval = 'monthly');

INSERT INTO plans (tier_id, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, 'quarterly', CASE t.name
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
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum')
AND NOT EXISTS (SELECT 1 FROM plans WHERE plans.tier_id = t.id AND plans.interval = 'quarterly');

INSERT INTO plans (tier_id, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, 'semesterly', CASE t.name
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
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum')
AND NOT EXISTS (SELECT 1 FROM plans WHERE plans.tier_id = t.id AND plans.interval = 'semesterly');

INSERT INTO plans (tier_id, interval, price_usd_cents, quota_downloads, features, provider_interval_days)
SELECT t.id, 'yearly', CASE t.name
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
WHERE t.name IN ('Bronze', 'Silver', 'Gold', 'Platinum')
AND NOT EXISTS (SELECT 1 FROM plans WHERE plans.tier_id = t.id AND plans.interval = 'yearly');
