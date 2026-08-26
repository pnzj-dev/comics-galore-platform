-- Authentication-method neutrality: a user may authenticate via password,
-- passkey, or OAuth. Neither email nor password is required for every user.
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
