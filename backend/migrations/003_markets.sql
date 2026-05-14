-- 003_markets.sql
-- Markets, outcomes and status history (M05/M06/M07).

CREATE TABLE IF NOT EXISTS markets (
    match_id        bigint NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    market_type_id  integer NOT NULL,
    specifier       text NOT NULL DEFAULT '',
    status          text NOT NULL DEFAULT 'active',
    group_id        integer,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, market_type_id, specifier),
    CONSTRAINT markets_status_chk CHECK (
        status IN ('active','suspended','deactivated','settled','cancelled','handed_over')
    )
);

CREATE INDEX IF NOT EXISTS markets_status_idx ON markets (status);

CREATE TABLE IF NOT EXISTS outcomes (
    match_id        bigint NOT NULL,
    market_type_id  integer NOT NULL,
    specifier       text NOT NULL DEFAULT '',
    outcome_id      integer NOT NULL,
    odds            numeric(12,4),
    is_active       boolean NOT NULL DEFAULT true,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, market_type_id, specifier, outcome_id),
    FOREIGN KEY (match_id, market_type_id, specifier)
        REFERENCES markets(match_id, market_type_id, specifier) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS market_status_history (
    id              bigserial PRIMARY KEY,
    match_id        bigint NOT NULL,
    market_type_id  integer NOT NULL,
    specifier       text NOT NULL DEFAULT '',
    from_status     text,
    to_status       text NOT NULL,
    raw_message_id  uuid REFERENCES raw_messages(id) ON DELETE SET NULL,
    changed_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS market_status_history_market_idx
    ON market_status_history (match_id, market_type_id, specifier);
