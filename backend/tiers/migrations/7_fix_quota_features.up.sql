-- Quota is now measured in downloads/month, not storage (GB). Fix the plan
-- feature copy and the decorative quota_downloads column to match the values
-- enforced via auth AppSettings (Free 5, Bronze 50, Silver 200, Gold/Platinum
-- unlimited). The quota line is removed from `features` because the UI now
-- derives it from `quota_downloads` (single source of truth).
UPDATE plans SET
    quota_downloads = CASE t.name
        WHEN 'Free' THEN 5
        WHEN 'Bronze' THEN 50
        WHEN 'Silver' THEN 200
        WHEN 'Gold' THEN 999999
        WHEN 'Platinum' THEN 999999
        ELSE plans.quota_downloads
    END,
    features = CASE t.name
        WHEN 'Free' THEN '["Browse comics","Read comments"]'::jsonb
        WHEN 'Bronze' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery"]'::jsonb
        WHEN 'Silver' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery"]'::jsonb
        WHEN 'Gold' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts"]'::jsonb
        WHEN 'Platinum' THEN '["Browse comics","Read comments","Write comments","Download archives","Web reader","Full preview gallery","Premium & exclusive posts"]'::jsonb
        ELSE plans.features
    END
FROM tiers t
WHERE plans.tier_id = t.id;

UPDATE tiers SET description = 'Top tier with full access and unlimited downloads' WHERE name = 'Platinum';
