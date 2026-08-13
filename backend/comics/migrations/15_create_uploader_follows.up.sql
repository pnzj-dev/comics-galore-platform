CREATE TABLE IF NOT EXISTS uploader_follows (
    user_id     UUID NOT NULL,
    uploader_id UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, uploader_id)
);

CREATE INDEX IF NOT EXISTS idx_uploader_follows_uploader ON uploader_follows(uploader_id);
CREATE INDEX IF NOT EXISTS idx_uploader_follows_user ON uploader_follows(user_id);
