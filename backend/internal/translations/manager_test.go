package translations_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/translations"
)

// 验收 14 — 静态描述与 i18n（M12）— Translation cache 行为

// Given the translation manager has never fetched ByLanguage(en)
// When manager.RefreshLanguage("en") is called
// Then the upstream WebAPI is hit exactly once and every Translation
//      from the response is upserted into the translations table.
func TestGiven_ColdManager_When_RefreshLanguage_Then_UpstreamFetchAndCachePopulated(t *testing.T) {
	api := newFakeAPI()
	api.setByLanguage("en", []translations.Translation{
		{TranslationID: 1, Text: "Match Result"},
		{TranslationID: 2, Text: "Total Goals"},
	})
	repo := newFakeRepo()
	logger := &captureLogger{}
	mgr := translations.New(api, repo)
	mgr.Logger = logger
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) }

	summary, err := mgr.RefreshLanguage(context.Background(), "en")
	require.NoError(t, err)
	require.False(t, summary.SkippedRateLimit)
	require.Equal(t, 2, summary.ItemCount)

	require.Equal(t, 1, api.langCallCount("en"))

	cached := repo.snapshotTranslations("en")
	require.Len(t, cached, 2)
	require.Equal(t, "Match Result", cached[1].Text)
	require.Equal(t, "Total Goals", cached[2].Text)
	require.Equal(t, "en", cached[1].LanguageID)

	require.Equal(t, int64(1), mgr.RefreshCount())
	require.Equal(t, []string{"en"}, logger.lang)
	require.Equal(t, []int{2}, logger.items)
}

// Given a successful ByLanguage(en) refresh that completed less than
//       one hour ago
// When manager.RefreshLanguage("en") is called again
// Then no HTTP request is sent and the manager reports a rate-limit
//      skip via Logger.
func TestGiven_RecentRefresh_When_CalledAgain_Then_SkippedPerRateLimit(t *testing.T) {
	api := newFakeAPI()
	api.setByLanguage("en", []translations.Translation{{TranslationID: 1, Text: "Match Result"}})
	repo := newFakeRepo()
	logger := &captureLogger{}
	mgr := translations.New(api, repo)
	mgr.Logger = logger

	clock := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	mgr.Now = func() time.Time { return clock }

	_, err := mgr.RefreshLanguage(context.Background(), "en")
	require.NoError(t, err)
	require.Equal(t, 1, api.langCallCount("en"))

	// Advance the clock 30 minutes — still within the cooldown.
	clock = clock.Add(30 * time.Minute)

	summary, err := mgr.RefreshLanguage(context.Background(), "en")
	require.NoError(t, err)
	require.True(t, summary.SkippedRateLimit)
	require.Equal(t, 1, api.langCallCount("en"),
		"second call must not hit upstream")

	require.Equal(t, 1, logger.skipCount())
	require.Equal(t, "en", logger.skipped[0].Language)
	require.Equal(t, translations.ReasonRateLimit, logger.skipped[0].Reason)
	require.Equal(t, int64(1), mgr.SkipCount())
}

// Given the in-memory cooldown was cleared (process restart simulation)
//       but the persisted refresh log still shows a recent refresh
// When manager.RefreshLanguage("en") is called
// Then the rate-limit is still honoured (the persisted log is the
//      source of truth across restarts).
func TestGiven_PersistedCooldownAfterRestart_When_RefreshLanguage_Then_StillSkipped(t *testing.T) {
	api := newFakeAPI()
	api.setByLanguage("en", []translations.Translation{{TranslationID: 1, Text: "X"}})
	repo := newFakeRepo()
	require.NoError(t, repo.RecordFullRefresh(context.Background(), "en",
		time.Date(2026, 5, 14, 11, 30, 0, 0, time.UTC), 100))

	mgr := translations.New(api, repo)
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) }

	summary, err := mgr.RefreshLanguage(context.Background(), "en")
	require.NoError(t, err)
	require.True(t, summary.SkippedRateLimit, "persisted refresh log must veto upstream call")
	require.Equal(t, 0, api.langCallCount("en"))
}

// Given a cached translation row exists for (lang="en", id=42)
// When manager.Lookup("en", 42) is called
// Then the cached Text is returned and no HTTP request is sent.
func TestGiven_TranslationCached_When_LookupHit_Then_NoHTTP(t *testing.T) {
	api := newFakeAPI()
	repo := newFakeRepo()
	repo.seedTranslation(translations.Translation{
		LanguageID: "en", TranslationID: 42, Text: "Both Teams To Score",
	})
	mgr := translations.New(api, repo)

	got, ok, err := mgr.Lookup(context.Background(), "en", 42)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Both Teams To Score", got.Text)
	require.Equal(t, 0, api.idCallCount("en", 42))
	require.Equal(t, int64(0), mgr.MissCount())
}

// Given the cache has no entry for (lang="en", id=42)
// When manager.Lookup("en", 42) is called
// Then /api/Translation/ById/en?id=42 is fetched, the returned Text is
//      upserted into the translations table, and the value is returned
//      to the caller.
func TestGiven_TranslationCacheMiss_When_Lookup_Then_FetchByIdAndCache(t *testing.T) {
	api := newFakeAPI()
	api.setByID(translations.Translation{
		LanguageID: "en", TranslationID: 42, Text: "Both Teams To Score",
	})
	repo := newFakeRepo()
	logger := &captureLogger{}
	mgr := translations.New(api, repo)
	mgr.Logger = logger
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) }

	got, ok, err := mgr.Lookup(context.Background(), "en", 42)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Both Teams To Score", got.Text)
	require.Equal(t, 1, api.idCallCount("en", 42))

	cached := repo.snapshotTranslations("en")
	require.Len(t, cached, 1)
	require.Equal(t, "Both Teams To Score", cached[42].Text)

	require.Equal(t, int64(1), mgr.MissCount())
	require.Len(t, logger.misses, 1)

	// Second lookup must hit the cache.
	_, ok, err = mgr.Lookup(context.Background(), "en", 42)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1, api.idCallCount("en", 42),
		"second lookup must not hit upstream again")
}

// Given the cache miss falls through to ById and upstream knows nothing
// When manager.Lookup is invoked
// Then (Translation{}, false, nil) is returned and nothing is persisted
func TestGiven_UnknownTranslation_When_Lookup_Then_NegativeWithNoUpsert(t *testing.T) {
	api := newFakeAPI()
	repo := newFakeRepo()
	mgr := translations.New(api, repo)

	_, ok, err := mgr.Lookup(context.Background(), "en", 9999)
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, repo.snapshotTranslations("en"))
}
