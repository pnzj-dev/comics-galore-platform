CREATE TABLE comics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uploader_id UUID NOT NULL,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    content_language TEXT NOT NULL DEFAULT 'en',
    status TEXT NOT NULL DEFAULT 'pending_review'
        CHECK (status IN ('pending_review', 'published', 'rejected', 'draft')),
    cover_key TEXT NOT NULL,
    file_key TEXT NOT NULL,
    page_keys JSONB DEFAULT '[]',
    file_size_bytes BIGINT NOT NULL DEFAULT 0,
    min_tier_id UUID,
    age_rating TEXT NOT NULL DEFAULT 'all_ages'
        CHECK (age_rating IN ('all_ages', 'teen', 'mature', 'explicit')),
    tags JSONB DEFAULT '[]',
    rejection_reason TEXT,
    published_at TIMESTAMPTZ,
    view_count BIGINT NOT NULL DEFAULT 0,
    download_count BIGINT NOT NULL DEFAULT 0,
    like_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_comics_status ON comics(status);
CREATE INDEX idx_comics_uploader ON comics(uploader_id);
CREATE INDEX idx_comics_published_at ON comics(published_at DESC) WHERE status = 'published';
CREATE INDEX idx_comics_slug ON comics(slug);
