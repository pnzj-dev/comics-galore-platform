CREATE TABLE IF NOT EXISTS saved_views (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id   UUID NOT NULL,
    resource   TEXT NOT NULL,
    name       TEXT NOT NULL,
    filters    JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_saved_views_admin ON saved_views(admin_id);
