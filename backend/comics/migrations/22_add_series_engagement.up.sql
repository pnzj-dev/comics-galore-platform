-- Series engagement aggregates + schedule for the home page.
ALTER TABLE series
    ADD COLUMN views_count  BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN hearts_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN schedule_day TEXT;

COMMENT ON COLUMN series.views_count  IS 'Aggregate views across the series comics (BIGINT — can reach millions).';
COMMENT ON COLUMN series.hearts_count IS 'Aggregate favorites across the series comics (BIGINT).';
COMMENT ON COLUMN series.schedule_day IS 'Day of week new episodes drop (mon..sun) or ''completed''; NULL = unscheduled.';

ALTER TABLE series ADD CONSTRAINT series_schedule_day_check
    CHECK (schedule_day IS NULL OR schedule_day IN ('mon','tue','wed','thu','fri','sat','sun','completed'));

CREATE INDEX idx_series_views    ON series(views_count DESC);
CREATE INDEX idx_series_hearts   ON series(hearts_count DESC);
CREATE INDEX idx_series_schedule ON series(schedule_day);

-- Backfill existing aggregates from comics.
UPDATE series s SET
    views_count  = COALESCE((SELECT SUM(c.view_count) FROM comics c WHERE c.series_id = s.id), 0),
    hearts_count = COALESCE((SELECT SUM(c.fav_count)  FROM comics c WHERE c.series_id = s.id), 0);
