CREATE TABLE IF NOT EXISTS series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    uploader_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS series_follows (
    user_id UUID NOT NULL,
    series_id UUID NOT NULL REFERENCES series(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, series_id)
);

ALTER TABLE comics ADD COLUMN IF NOT EXISTS series_id UUID REFERENCES series(id);
ALTER TABLE comics ADD COLUMN IF NOT EXISTS series_order INT DEFAULT 1;

CREATE INDEX idx_series_slug ON series(slug);
CREATE INDEX idx_series_uploader ON series(uploader_id);
CREATE INDEX idx_comics_series ON comics(series_id);
CREATE INDEX idx_series_follows_user ON series_follows(user_id);
