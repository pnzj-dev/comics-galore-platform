-- Optional volume (collected edition / chapter) and issue number for
-- nicer download filenames and richer metadata. Both are NULL for comics
-- that don't carry them (e.g. a one-shot with no numbering).
ALTER TABLE comics
    ADD COLUMN volume TEXT,
    ADD COLUMN issue_number TEXT;

COMMENT ON COLUMN comics.volume IS 'Free-form volume or chapter label (e.g. "Vol. 2", "Ch. 5").';
COMMENT ON COLUMN comics.issue_number IS 'Issue number for single-issue comics (e.g. "12", "#3").';
