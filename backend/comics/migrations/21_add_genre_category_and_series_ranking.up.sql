-- Category and genre on comics (primary labels; complement the free-form tags array).
ALTER TABLE comics
    ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS genre    TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN comics.category IS 'Primary category (e.g. Manga, Webcomic).';
COMMENT ON COLUMN comics.genre    IS 'Primary genre label (Action, Romance, Drama…); complements tags.';

-- Series reader + ranking display fields. Rank/rank_change are computed from
-- engagement sums (view_count / like_count) and not stored.
ALTER TABLE series
    ADD COLUMN cover_key     TEXT NOT NULL DEFAULT '',
    ADD COLUMN genre         TEXT NOT NULL DEFAULT '',
    ADD COLUMN category      TEXT NOT NULL DEFAULT '',
    ADD COLUMN overlay_title TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN series.cover_key     IS 'Cover image key (resolved like comic covers).';
COMMENT ON COLUMN series.genre         IS 'Primary genre label.';
COMMENT ON COLUMN series.category      IS 'Primary category.';
COMMENT ON COLUMN series.overlay_title  IS 'Stylized title overlaid on the cover; defaults to title when empty.';
