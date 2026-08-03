CREATE TABLE reading_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    comic_id UUID NOT NULL,
    current_page INT NOT NULL DEFAULT 0,
    total_pages INT NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, comic_id)
);

CREATE TABLE download_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    comic_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reading_progress_user ON reading_progress(user_id);
CREATE INDEX idx_reading_progress_comic ON reading_progress(comic_id);
CREATE INDEX idx_download_logs_user ON download_logs(user_id);
CREATE INDEX idx_download_logs_date ON download_logs(created_at);
