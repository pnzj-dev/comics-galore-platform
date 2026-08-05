CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    email_new_from_following BOOLEAN DEFAULT true,
    email_support_replies BOOLEAN DEFAULT true,
    email_marketing BOOLEAN DEFAULT false,
    in_app_enabled BOOLEAN DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
