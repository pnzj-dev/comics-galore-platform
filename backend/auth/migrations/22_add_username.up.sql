-- Unique public username/handle. Nullable so pre-existing and OAuth-created
-- accounts without a chosen handle remain valid.
ALTER TABLE users ADD COLUMN username TEXT;

CREATE UNIQUE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
