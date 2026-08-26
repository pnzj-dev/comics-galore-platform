-- OAuth login state (CSRF protection) and one-time exchange codes.
--
-- oauth_states: short-lived `state` generated when starting a provider flow,
-- carrying PKCE code_verifier and any link-intent user. Single-use.
--
-- oauth_exchange_codes: after a provider callback authenticates a user, the
-- backend issues a short-lived single-use code handed to the browser; the
-- SvelteKit server exchanges it for a session token without ever putting the
-- session id in a URL.
CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    code_verifier TEXT NOT NULL,
    link_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oauth_states_expires ON oauth_states(expires_at);

CREATE TABLE IF NOT EXISTS oauth_exchange_codes (
    code TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oauth_exchange_expires ON oauth_exchange_codes(expires_at);
