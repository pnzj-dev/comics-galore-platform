-- Linked external identities. A single user may have many authentication
-- methods (password, passkeys, Google, Facebook, Twitter/X, Apple).
-- Linking is by verified provider identity, never by email alone.
CREATE TABLE IF NOT EXISTS auth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_account_id TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_accounts_provider
    ON auth_accounts(provider, provider_account_id);
CREATE INDEX IF NOT EXISTS idx_auth_accounts_user ON auth_accounts(user_id);
