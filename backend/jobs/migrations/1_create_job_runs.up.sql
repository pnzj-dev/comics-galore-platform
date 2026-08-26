CREATE TABLE job_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name TEXT NOT NULL,
    ref TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('running', 'success', 'failed')),
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_job_runs_job_name ON job_runs(job_name);
CREATE INDEX idx_job_runs_status ON job_runs(status);
CREATE INDEX idx_job_runs_started_at ON job_runs(started_at DESC);
