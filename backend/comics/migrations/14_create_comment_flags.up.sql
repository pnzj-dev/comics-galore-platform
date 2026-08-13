CREATE TABLE IF NOT EXISTS comment_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id  UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open'
                CHECK (status IN ('open', 'resolved')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by UUID
);

CREATE INDEX IF NOT EXISTS idx_comment_flags_comment ON comment_flags(comment_id);
CREATE INDEX IF NOT EXISTS idx_comment_flags_status ON comment_flags(status);

-- One flag per user per comment (idempotent flagging).
CREATE UNIQUE INDEX IF NOT EXISTS idx_comment_flags_user_comment
ON comment_flags(comment_id, user_id);
