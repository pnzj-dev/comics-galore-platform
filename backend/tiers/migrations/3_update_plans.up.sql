UPDATE plans SET interval = 'monthly' WHERE interval = 'lifetime';

ALTER TABLE plans DROP CONSTRAINT IF EXISTS plans_interval_check;
ALTER TABLE plans ADD CONSTRAINT plans_interval_check CHECK (interval IN ('monthly', 'quarterly', 'yearly'));

ALTER TABLE plans ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '[]';
ALTER TABLE plans ADD COLUMN IF NOT EXISTS provider_plan_id TEXT;
ALTER TABLE plans ADD COLUMN IF NOT EXISTS provider_interval_days INT DEFAULT 0;
