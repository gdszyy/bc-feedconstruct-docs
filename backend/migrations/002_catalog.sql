-- 002_catalog.sql
-- Sports catalog (M03/M04). Maps to FeedConstruct ObjectType IDs 1/2/3/4.

CREATE TABLE IF NOT EXISTS sports (
    id          integer PRIMARY KEY,
    name        text NOT NULL,
    is_active   boolean NOT NULL DEFAULT true,
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS regions (
    id          integer PRIMARY KEY,
    sport_id    integer NOT NULL REFERENCES sports(id) ON DELETE RESTRICT,
    name        text NOT NULL,
    is_active   boolean NOT NULL DEFAULT true,
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS regions_sport_idx ON regions (sport_id);

CREATE TABLE IF NOT EXISTS competitions (
    id          integer PRIMARY KEY,
    region_id   integer NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    sport_id    integer NOT NULL REFERENCES sports(id) ON DELETE RESTRICT,
    name        text NOT NULL,
    is_active   boolean NOT NULL DEFAULT true,
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS competitions_region_idx ON competitions (region_id);

CREATE TABLE IF NOT EXISTS matches (
    id              bigint PRIMARY KEY,
    sport_id        integer NOT NULL REFERENCES sports(id) ON DELETE RESTRICT,
    competition_id  integer REFERENCES competitions(id) ON DELETE SET NULL,
    name            text,
    home            text,
    away            text,
    start_at        timestamptz,
    is_live         boolean NOT NULL DEFAULT false,
    status          text NOT NULL DEFAULT 'not_started',
    last_event_id   text,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT matches_status_chk CHECK (
        status IN ('not_started','live','ended','closed','cancelled','postponed')
    )
);

CREATE INDEX IF NOT EXISTS matches_competition_idx ON matches (competition_id);
CREATE INDEX IF NOT EXISTS matches_live_idx ON matches (is_live) WHERE is_live;

CREATE TABLE IF NOT EXISTS fixture_changes (
    id              bigserial PRIMARY KEY,
    match_id        bigint NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    change_type     text NOT NULL,
    old             jsonb,
    new             jsonb,
    raw_message_id  uuid REFERENCES raw_messages(id) ON DELETE SET NULL,
    received_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS fixture_changes_match_idx ON fixture_changes (match_id);
