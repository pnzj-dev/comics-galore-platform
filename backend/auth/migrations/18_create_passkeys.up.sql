-- WebAuthn passkeys. Stores only public credential material (never private
-- keys). `credential_id` is the hex-encoded WebAuthn credential ID (unique
-- lookup key); `credential` holds the full serialized go-webauthn Credential
-- (public key, attestation, transports, AAGUID, sign count) for future
-- assertion verification.
CREATE TABLE IF NOT EXISTS passkeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    credential_id TEXT NOT NULL UNIQUE,
    credential JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_passkeys_user ON passkeys(user_id);
