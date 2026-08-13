CREATE TABLE IF NOT EXISTS staff_picks (
    comic_id   UUID PRIMARY KEY REFERENCES comics(id) ON DELETE CASCADE,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_staff_picks_position ON staff_picks(position);
