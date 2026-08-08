ALTER TABLE comics ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_comics_deleted ON comics(deleted_at) WHERE deleted_at IS NOT NULL;
