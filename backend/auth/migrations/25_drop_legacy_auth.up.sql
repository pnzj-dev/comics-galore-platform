-- Remove every credential/session mechanism that now lives in Logto.
-- Identity (logto_id) remains on users; role/tier/username/ban/suspend/prefs
-- are app-level and stay on users.
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS auth_accounts;
DROP TABLE IF EXISTS passkeys;
DROP TABLE IF EXISTS webauthn_challenges;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS oauth_exchange_codes;
DROP TABLE IF EXISTS mfa_challenges;

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE users DROP COLUMN IF EXISTS verify_token;
ALTER TABLE users DROP COLUMN IF EXISTS verify_token_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS reset_token;
ALTER TABLE users DROP COLUMN IF EXISTS reset_token_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;
