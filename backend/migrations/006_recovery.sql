-- 006_recovery.sql
-- Recovery jobs and producer health (M10/M15).

CREATE TABLE IF NOT EXISTS recovery_jobs (
    id              bigserial PRIMARY KEY,
    scope           text NOT NULL,
    product         text,
    match_id        bigint,
    requested_at    timestamptz NOT NULL DEFAULT now(),
    started_at      timestamptz,
    finished_at     timestamptz,
    status          text NOT NULL DEFAULT 'queued',
    attempt         smallint NOT NULL DEFAULT 0,
    next_retry_at   timestamptz,
    detail          jsonb,
    CONSTRAINT recovery_jobs_scope_chk CHECK (
        scope IN ('startup','product','event','stateful','fixture_change')
    ),
    CONSTRAINT recovery_jobs_status_chk CHECK (
        status IN ('queued','running','success','failed','rate_limited')
    )
);

CREATE INDEX IF NOT EXISTS recovery_jobs_status_idx
    ON recovery_jobs (status, next_retry_at);

CREATE TABLE IF NOT EXISTS producer_health (
    product             text PRIMARY KEY,
    last_alive_at       timestamptz,
    last_message_at     timestamptz,
    is_down             boolean NOT NULL DEFAULT false,
    detail              jsonb,
    CONSTRAINT producer_health_product_chk CHECK (product IN ('live','prematch'))
);

INSERT INTO producer_health (product) VALUES ('live'), ('prematch')
    ON CONFLICT (product) DO NOTHING;
