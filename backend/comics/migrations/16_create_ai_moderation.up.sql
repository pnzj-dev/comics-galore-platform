CREATE TABLE IF NOT EXISTS ai_decisions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('comic', 'comment')),
    target_id   UUID NOT NULL,
    decision    TEXT NOT NULL CHECK (decision IN ('approved', 'rejected', 'uncertain')),
    confidence  NUMERIC,
    reason      TEXT,
    model       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_decisions_target ON ai_decisions(target_type, target_id);

CREATE TABLE IF NOT EXISTS ai_review_queue (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type  TEXT NOT NULL CHECK (target_type IN ('comic', 'comment')),
    target_id    UUID NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'resolved')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_by  UUID,
    resolved_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_review_queue_status ON ai_review_queue(status);
