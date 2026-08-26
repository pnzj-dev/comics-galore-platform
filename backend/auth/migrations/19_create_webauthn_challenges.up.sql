-- Single-use, short-lived WebAuthn challenges, bound to an intended ceremony.
-- `purpose` is 'register' or 'login'. `user_id` is nullable: login ceremonies
-- are discoverable (usernameless) and resolve the user only at finish time.
-- `session_data` holds the go-webauthn SessionData JSON needed to complete
-- the ceremony. Rows are deleted on first use (single-use).
CREATE TABLE IF NOT EXISTS webauthn_challenges (
    challenge TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    purpose TEXT NOT NULL,
    session_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webauthn_challenges_expires ON webauthn_challenges(expires_at);
