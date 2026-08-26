-- Download quota now lives on the tier (single source of truth), instead of
-- the auth AppSettings blob and a denormalized plans column. Enforcement, the
-- plan grid, and the admin panel all read/write `tiers.quota_downloads`.
ALTER TABLE tiers ADD COLUMN quota_downloads INT NOT NULL DEFAULT 0;

UPDATE tiers SET quota_downloads = CASE name
    WHEN 'Free' THEN 5
    WHEN 'Bronze' THEN 50
    WHEN 'Silver' THEN 200
    WHEN 'Gold' THEN 999999
    WHEN 'Platinum' THEN 999999
    ELSE 0
END;

ALTER TABLE plans DROP COLUMN IF EXISTS quota_downloads;
