CREATE TABLE workers (
    id              TEXT PRIMARY KEY,
    status          TEXT NOT NULL DEFAULT 'alive',
    jobs_processed  INT NOT NULL DEFAULT 0,
    last_heartbeat  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
