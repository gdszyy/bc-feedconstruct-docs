-- 007_translations.sql
-- Translation cache (M12). Maps to FeedConstruct Translation Web API
-- (/api/Translation/Languages, /api/Translation/ByLanguage/{id},
-- /api/Translation/ById/{id}?id=...). Cache is replaced on refresh, not
-- merged, so a removed translation upstream eventually disappears here.

CREATE TABLE IF NOT EXISTS translation_languages (
    language_id     text PRIMARY KEY,
    name            text NOT NULL DEFAULT '',
    is_default      boolean NOT NULL DEFAULT false,
    fetched_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS translations (
    language_id     text   NOT NULL REFERENCES translation_languages(language_id) ON DELETE CASCADE,
    translation_id  bigint NOT NULL,
    text            text   NOT NULL DEFAULT '',
    fetched_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (language_id, translation_id)
);

CREATE INDEX IF NOT EXISTS translations_language_idx
    ON translations (language_id);

-- ByLanguage refreshes are rate-limited upstream (1/hour). The manager
-- stores the last successful refresh per language so subsequent calls
-- can short-circuit without hitting the network.
CREATE TABLE IF NOT EXISTS translation_refresh_log (
    language_id     text PRIMARY KEY REFERENCES translation_languages(language_id) ON DELETE CASCADE,
    last_full_refresh_at timestamptz NOT NULL DEFAULT now(),
    item_count      integer NOT NULL DEFAULT 0
);
