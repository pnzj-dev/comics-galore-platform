CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '""',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_settings (key, value) VALUES
    ('defaults', '{"default_language":"en","default_content_language":"en","items_per_page":12,"popular_tags_limit":20,"site_name":"Comics Galore","maintenance_mode":false,"registrations_open":true,"max_upload_size_mb":50,"image_serving_mode":"direct","require_email_verify":false,"rate_limit":60,"s3_presigned_ttl_min":15,"cf_presigned_ttl_min":15,"quota_free_gb":1,"quota_bronze_gb":10,"quota_silver_gb":50,"quota_gold_gb":200,"quota_platinum_gb":1000,"boost_5gb_price":5.00,"boost_10gb_price":8.00,"boost_20gb_price":12.00}'::jsonb)
ON CONFLICT (key) DO NOTHING;
