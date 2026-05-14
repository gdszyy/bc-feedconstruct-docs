-- 001_init.sql
-- Foundation: extensions + raw_messages (M01 message audit log).
-- See docs/08_backend_railway/02_postgres_schema.md for the authoritative spec.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS raw_messages (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    received_at     timestamptz NOT NULL DEFAULT now(),
    source          text NOT NULL,
    routing_key     text,
    queue           text,
    message_type    text NOT NULL,
    event_id        text,
    product_id      smallint,
    sport_id        integer,
    ts_provider     timestamptz,
    payload         jsonb NOT NULL,
    raw_blob        bytea,
    processed_at    timestamptz,
    process_error   text
);

-- Idempotency (acceptance #11): identical deliveries collapse.
-- NULL event_id is normalised to empty string via the constraint expression.
CREATE UNIQUE INDEX IF NOT EXISTS raw_messages_uniq
    ON raw_messages (
        source,
        message_type,
        COALESCE(event_id, ''),
        COALESCE(ts_provider, 'epoch'::timestamptz)
    );

CREATE INDEX IF NOT EXISTS raw_messages_received_at_idx
    ON raw_messages (received_at);

CREATE INDEX IF NOT EXISTS raw_messages_event_idx
    ON raw_messages (event_id) WHERE event_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS raw_messages_type_idx
    ON raw_messages (message_type);

-- Internal metrics_counters used across modules (acceptance #15/#16).
CREATE TABLE IF NOT EXISTS metrics_counters (
    name        text PRIMARY KEY,
    value       bigint NOT NULL DEFAULT 0,
    updated_at  timestamptz NOT NULL DEFAULT now()
);
