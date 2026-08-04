ALTER TABLE comics DROP CONSTRAINT IF EXISTS comics_status_check;
ALTER TABLE comics ADD CONSTRAINT comics_status_check CHECK (status IN ('pending_review', 'published', 'rejected', 'scheduled'));

UPDATE comics SET status = 'pending_review' WHERE status = 'draft';
