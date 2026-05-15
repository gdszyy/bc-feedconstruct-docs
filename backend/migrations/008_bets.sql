-- 008_bets.sql
-- Bet aggregate (M13/M14). The state column on bets carries the LIVE
-- value for fast reads; bet_transitions is the append-only history that
-- M14 surfaces as the timeline. bet_transitions never UPDATEs — every
-- state change inserts a new row, including rollbacks.

CREATE TABLE IF NOT EXISTS bets (
    id                text          PRIMARY KEY,
    user_id           text          NOT NULL,
    placed_at         timestamptz   NOT NULL,
    stake             numeric(20,4) NOT NULL,
    currency          text          NOT NULL,
    bet_type          text          NOT NULL,
    state             text          NOT NULL,
    idempotency_key   text          NOT NULL,
    payout_gross      numeric(20,4),
    payout_currency   text,
    void_factor       numeric(5,4),
    dead_heat_factor  numeric(5,4),
    created_at        timestamptz   NOT NULL DEFAULT now(),
    updated_at        timestamptz   NOT NULL DEFAULT now()
);

-- Idempotency-Key dedupes per user, not globally — two users can both
-- ship a UUIDv7 collision (vanishingly rare, but not impossible).
CREATE UNIQUE INDEX IF NOT EXISTS bets_idempotency_key_user_idx
    ON bets (user_id, idempotency_key);

CREATE INDEX IF NOT EXISTS bets_user_status_idx
    ON bets (user_id, state, placed_at DESC);

CREATE TABLE IF NOT EXISTS bet_selections (
    bet_id        text          NOT NULL REFERENCES bets(id) ON DELETE CASCADE,
    position      int           NOT NULL,
    match_id      text          NOT NULL,
    market_id     text          NOT NULL,
    outcome_id    text          NOT NULL,
    locked_odds   numeric(10,4) NOT NULL,
    PRIMARY KEY (bet_id, position)
);

CREATE INDEX IF NOT EXISTS bet_selections_outcome_idx
    ON bet_selections (match_id, market_id, outcome_id);

CREATE TABLE IF NOT EXISTS bet_transitions (
    id              bigserial   PRIMARY KEY,
    bet_id          text        NOT NULL REFERENCES bets(id) ON DELETE CASCADE,
    at              timestamptz NOT NULL DEFAULT now(),
    from_state      text        NOT NULL DEFAULT '',
    to_state        text        NOT NULL,
    reason          text        NOT NULL DEFAULT '',
    event_id        text        NOT NULL DEFAULT '',
    correlation_id  text        NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS bet_transitions_bet_at_idx
    ON bet_transitions (bet_id, at);

-- Replayed events must NOT double-write transitions. Synthesized
-- transitions (e.g. the initial Pending row created at place() time)
-- carry an empty event_id and are excluded from the uniqueness rule.
CREATE UNIQUE INDEX IF NOT EXISTS bet_transitions_event_unique_idx
    ON bet_transitions (bet_id, event_id) WHERE event_id <> '';
