-- TOTP (authenticator-app) two-factor authentication. NULL secret = disabled.
ALTER TABLE users ADD COLUMN totp_secret TEXT;

-- Short-lived, single-use challenge bridging the password step and the TOTP
-- step of login when 2FA is enabled.
CREATE TABLE mfa_challenges (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_mfa_challenges_user ON mfa_challenges(user_id);
