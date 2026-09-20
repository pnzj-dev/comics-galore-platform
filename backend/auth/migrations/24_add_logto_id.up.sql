ALTER TABLE users ADD COLUMN IF NOT EXISTS logto_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_logto_id ON users(logto_id);
