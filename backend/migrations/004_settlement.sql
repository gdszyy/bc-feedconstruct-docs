-- 004_settlement.sql
-- Settlements, cancels, rollbacks (M08/M09).

CREATE TABLE IF NOT EXISTS settlements (
    id                  bigserial PRIMARY KEY,
    match_id            bigint NOT NULL,
    market_type_id      integer NOT NULL,
    specifier           text NOT NULL DEFAULT '',
    outcome_id          integer NOT NULL,
    result              text NOT NULL,
    certainty           smallint NOT NULL DEFAULT 1,
    void_factor         numeric(5,4),
    dead_heat_factor    numeric(5,4),
    raw_message_id      uuid REFERENCES raw_messages(id) ON DELETE SET NULL,
    settled_at          timestamptz NOT NULL DEFAULT now(),
    rolled_back_at      timestamptz,
    CONSTRAINT settlements_result_chk CHECK (
        result IN ('win','lose','void','half_win','half_lose')
    ),
    CONSTRAINT settlements_certainty_chk CHECK (certainty IN (0,1))
);

-- Idempotency for settlements (acceptance #11): same outcome + settled_at
-- collapses; certainty=1 supersedes 0 logically (handled in code).
CREATE UNIQUE INDEX IF NOT EXISTS settlements_uniq
    ON settlements (match_id, market_type_id, specifier, outcome_id, settled_at);

CREATE INDEX IF NOT EXISTS settlements_match_idx ON settlements (match_id);

CREATE TABLE IF NOT EXISTS cancels (
    id              bigserial PRIMARY KEY,
    match_id        bigint NOT NULL,
    market_type_id  integer,
    specifier       text NOT NULL DEFAULT '',
    void_reason     text,
    void_action     smallint NOT NULL DEFAULT 1,
    superceded_by   bigint REFERENCES cancels(id) ON DELETE SET NULL,
    from_ts         timestamptz,
    to_ts           timestamptz,
    raw_message_id  uuid REFERENCES raw_messages(id) ON DELETE SET NULL,
    cancelled_at    timestamptz NOT NULL DEFAULT now(),
    rolled_back_at  timestamptz,
    CONSTRAINT cancels_action_chk CHECK (void_action IN (1,2))
);

CREATE INDEX IF NOT EXISTS cancels_match_idx ON cancels (match_id);

CREATE TABLE IF NOT EXISTS rollbacks (
    id              bigserial PRIMARY KEY,
    target          text NOT NULL,
    target_id       bigint NOT NULL,
    raw_message_id  uuid REFERENCES raw_messages(id) ON DELETE SET NULL,
    applied_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT rollbacks_target_chk CHECK (target IN ('settlement','cancel'))
);

CREATE UNIQUE INDEX IF NOT EXISTS rollbacks_uniq
    ON rollbacks (target, target_id, raw_message_id);
