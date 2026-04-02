CREATE TYPE job_status AS ENUM ('QUEUED','PROCESSING','COMPLETED','FAILED','DEAD');
CREATE TYPE job_priority AS ENUM ('high','normal','low');

CREATE TABLE jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type          TEXT NOT NULL,
    priority      job_priority NOT NULL DEFAULT 'normal',
    status        job_status NOT NULL DEFAULT 'QUEUED',
    payload       JSONB NOT NULL,
    result        JSONB,
    error_message TEXT,
    attempts      INT NOT NULL DEFAULT 0,
    max_attempts  INT NOT NULL DEFAULT 3,
    worker_id     TEXT,
    callback_url  TEXT,
    namespace     TEXT NOT NULL DEFAULT 'default',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    queued_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_namespace_status ON jobs(namespace, status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
