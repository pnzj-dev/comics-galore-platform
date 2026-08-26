-- Record explicit consent to the Terms of Service / Privacy Policy. Set on
-- password registration and on OAuth sign-ups that went through the gated
-- register flow (client enforces the checkbox). Existing-user logins do not
-- update it.
ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ;

COMMENT ON COLUMN users.terms_accepted_at IS 'When the user explicitly accepted the Terms of Service (password or gated OAuth sign-up).';
