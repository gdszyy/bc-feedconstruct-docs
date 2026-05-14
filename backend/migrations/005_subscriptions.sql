-- 005_subscriptions.sql
-- Subscription lifecycle (M11). Map FeedConstruct Book/Unbook objects.

CREATE TABLE IF NOT EXISTS subscriptions (
    match_id        bigint PRIMARY KEY,
    product         text NOT NULL,
    status          text NOT NULL,
    requested_at    timestamptz,
    subscribed_at   timestamptz,
    released_at     timestamptz,
    last_event_id   text,
    reason          text,
    CONSTRAINT subscriptions_product_chk CHECK (product IN ('live','prematch')),
    CONSTRAINT subscriptions_status_chk CHECK (
        status IN ('requested','subscribed','unsubscribed','expired','failed')
    )
);

CREATE TABLE IF NOT EXISTS subscription_events (
    id              bigserial PRIMARY KEY,
    match_id        bigint NOT NULL,
    from_status     text,
    to_status       text NOT NULL,
    reason          text,
    occurred_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS subscription_events_match_idx
    ON subscription_events (match_id);
