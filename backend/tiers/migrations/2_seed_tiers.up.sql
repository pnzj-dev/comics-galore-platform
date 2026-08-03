INSERT INTO tiers (name, description, sort_order) VALUES
    ('Free', 'Free tier with limited downloads', 0),
    ('Bronze', 'Entry-level paid tier', 1),
    ('Silver', 'Mid-range paid tier', 2),
    ('Gold', 'Premium tier with maximum benefits', 3);

INSERT INTO plans (tier_id, interval, price_usd_cents, quota_downloads)
    SELECT id, 'monthly', 0, 5 FROM tiers WHERE name = 'Free'
    UNION ALL
    SELECT id, 'lifetime', 0, 5 FROM tiers WHERE name = 'Free'
    UNION ALL
    SELECT id, 'monthly', 500, 50 FROM tiers WHERE name = 'Bronze'
    UNION ALL
    SELECT id, 'lifetime', 5000, 50 FROM tiers WHERE name = 'Bronze'
    UNION ALL
    SELECT id, 'monthly', 1000, 200 FROM tiers WHERE name = 'Silver'
    UNION ALL
    SELECT id, 'lifetime', 10000, 200 FROM tiers WHERE name = 'Silver'
    UNION ALL
    SELECT id, 'monthly', 2000, 999999 FROM tiers WHERE name = 'Gold'
    UNION ALL
    SELECT id, 'lifetime', 20000, 999999 FROM tiers WHERE name = 'Gold';
