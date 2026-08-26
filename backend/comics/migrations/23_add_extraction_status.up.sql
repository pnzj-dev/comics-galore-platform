-- Server-side archive extraction state for comics whose archive can't be
-- extracted client-side (CBR/RAR, PDF). Set to 'processing' at creation and
-- 'done'/'failed' by the extraction worker.
ALTER TABLE comics ADD COLUMN extraction_status TEXT NOT NULL DEFAULT 'none'
    CHECK (extraction_status IN ('none', 'processing', 'done', 'failed'));
