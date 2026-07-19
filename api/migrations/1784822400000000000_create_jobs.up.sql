BEGIN;

CREATE TABLE jobs (
    id           BIGSERIAL   NOT NULL PRIMARY KEY,
    kind         VARCHAR(64) NOT NULL,
    dedup_key    TEXT        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}'::JSONB,
    status       VARCHAR(16) NOT NULL DEFAULT 'pending',
    attempts     INT         NOT NULL DEFAULT 0,
    max_attempts INT         NOT NULL DEFAULT 5,
    run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at    TIMESTAMPTZ,
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX jobs_pending_dedup_idx ON jobs (dedup_key) WHERE status = 'pending';

CREATE INDEX jobs_claim_idx ON jobs (run_at, id) WHERE status = 'pending';

CREATE INDEX jobs_lease_idx ON jobs (locked_at) WHERE status = 'running';

COMMIT;
