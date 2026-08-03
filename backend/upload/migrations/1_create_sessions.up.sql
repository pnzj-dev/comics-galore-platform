CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    mode TEXT NOT NULL DEFAULT 'manual'
        CHECK (mode IN ('manual', 'archive')),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'finalising', 'completed', 'failed', 'expired')),
    s3_prefix TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    parts JSONB DEFAULT '[]',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_upload_sessions_user ON upload_sessions(user_id);
CREATE INDEX idx_upload_sessions_status ON upload_sessions(status);
