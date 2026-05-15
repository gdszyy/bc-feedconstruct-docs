package translations

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// RefreshLanguages fetches /api/Translation/Languages, upserts the
// language list and ensures translation_languages has a row for every
// language seen upstream. It must be called once at boot before
// RefreshLanguage to satisfy the FK from translations.
func (m *Manager) RefreshLanguages(ctx context.Context) error {
	if m.API == nil {
		return errors.New("translations: API not configured")
	}
	if m.Repo == nil {
		return errors.New("translations: Repo not configured")
	}
	now := m.now()
	upstream, err := m.API.Languages(ctx)
	if err != nil {
		return fmt.Errorf("translations: languages: %w", err)
	}
	for i := range upstream {
		upstream[i].FetchedAt = now
	}
	if err := m.Repo.UpsertLanguages(ctx, upstream); err != nil {
		return fmt.Errorf("translations: upsert languages: %w", err)
	}
	if m.Logger != nil {
		m.Logger.LanguagesRefreshed(len(upstream))
	}
	return nil
}

// RefreshLanguage performs a full ByLanguage fetch and replaces the
// cached translations for the given language. Honours the per-language
// cooldown both in memory and via the persisted refresh log; returns
// SkippedRateLimit=true (and no error) when the cooldown has not
// elapsed.
//
// The language row in translation_languages must already exist (either
// from RefreshLanguages or from a manual UpsertLanguages); if not, the
// FK on translations forces us to upsert a stub here.
func (m *Manager) RefreshLanguage(ctx context.Context, languageID string) (RefreshSummary, error) {
	languageID = strings.TrimSpace(languageID)
	if languageID == "" {
		return RefreshSummary{}, errors.New("translations: empty languageID")
	}
	if m.API == nil {
		return RefreshSummary{}, errors.New("translations: API not configured")
	}
	if m.Repo == nil {
		return RefreshSummary{}, errors.New("translations: Repo not configured")
	}

	now := m.now()
	summary := RefreshSummary{LanguageID: languageID}

	m.mu.Lock()
	lastMem, hasMem := m.lastByLang[languageID]
	m.mu.Unlock()

	if hasMem && now.Sub(lastMem) < m.cooldown() {
		summary.SkippedRateLimit = true
		summary.LastRefreshAt = lastMem
		m.skipTotal.Add(1)
		if m.Logger != nil {
			m.Logger.LanguageRefreshSkipped(languageID, ReasonRateLimit)
		}
		return summary, nil
	}

	lastDB, hasDB, err := m.Repo.LastFullRefresh(ctx, languageID)
	if err != nil {
		return summary, fmt.Errorf("translations: read refresh log: %w", err)
	}
	if hasDB && now.Sub(lastDB) < m.cooldown() {
		summary.SkippedRateLimit = true
		summary.LastRefreshAt = lastDB
		m.mu.Lock()
		m.lastByLang[languageID] = lastDB
		m.mu.Unlock()
		m.skipTotal.Add(1)
		if m.Logger != nil {
			m.Logger.LanguageRefreshSkipped(languageID, ReasonRateLimit)
		}
		return summary, nil
	}

	// Ensure the language row exists so the FK from translations does
	// not fail when the caller didn't run RefreshLanguages first.
	if err := m.Repo.UpsertLanguages(ctx, []Language{{
		ID:        languageID,
		FetchedAt: now,
	}}); err != nil {
		return summary, fmt.Errorf("translations: ensure language: %w", err)
	}

	items, err := m.API.ByLanguage(ctx, languageID)
	if err != nil {
		return summary, fmt.Errorf("translations: by language: %w", err)
	}
	for i := range items {
		items[i].LanguageID = languageID
		items[i].FetchedAt = now
	}
	if err := m.Repo.UpsertTranslations(ctx, languageID, items); err != nil {
		return summary, fmt.Errorf("translations: upsert: %w", err)
	}
	if err := m.Repo.RecordFullRefresh(ctx, languageID, now, len(items)); err != nil {
		return summary, fmt.Errorf("translations: record refresh: %w", err)
	}

	m.mu.Lock()
	m.lastByLang[languageID] = now
	m.mu.Unlock()

	m.refreshTotal.Add(1)
	summary.ItemCount = len(items)
	summary.LastRefreshAt = now
	if m.Logger != nil {
		m.Logger.LanguageRefreshed(languageID, len(items))
	}
	return summary, nil
}

// Lookup returns the cached translation text for the given pair, or
// fetches it via /api/Translation/ById on cache miss and persists the
// result. The cache-miss path is not rate limited because ById is a
// per-translation endpoint, not the bulk ByLanguage one.
func (m *Manager) Lookup(ctx context.Context, languageID string, translationID int64) (Translation, bool, error) {
	languageID = strings.TrimSpace(languageID)
	if languageID == "" {
		return Translation{}, false, errors.New("translations: empty languageID")
	}
	if m.Repo == nil {
		return Translation{}, false, errors.New("translations: Repo not configured")
	}

	hit, ok, err := m.Repo.GetTranslation(ctx, languageID, translationID)
	if err != nil {
		return Translation{}, false, fmt.Errorf("translations: get: %w", err)
	}
	if ok {
		return hit, true, nil
	}

	if m.API == nil {
		// Repo-only mode (e.g. offline tests). Treat the miss as a
		// definite negative result, not an error.
		m.missTotal.Add(1)
		if m.Logger != nil {
			m.Logger.LookupMiss(languageID, translationID)
		}
		return Translation{}, false, nil
	}

	fetched, ok, err := m.API.ByID(ctx, languageID, translationID)
	if err != nil {
		return Translation{}, false, fmt.Errorf("translations: by id: %w", err)
	}
	m.missTotal.Add(1)
	if m.Logger != nil {
		m.Logger.LookupMiss(languageID, translationID)
	}
	if !ok {
		return Translation{}, false, nil
	}
	now := m.now()
	fetched.LanguageID = languageID
	fetched.TranslationID = translationID
	fetched.FetchedAt = now

	// Ensure the parent language row exists so the FK on translations
	// does not fail under a cold cache.
	if err := m.Repo.UpsertLanguages(ctx, []Language{{ID: languageID, FetchedAt: now}}); err != nil {
		return Translation{}, false, fmt.Errorf("translations: ensure language: %w", err)
	}
	if err := m.Repo.UpsertTranslations(ctx, languageID, []Translation{fetched}); err != nil {
		return Translation{}, false, fmt.Errorf("translations: cache by id: %w", err)
	}
	return fetched, true, nil
}

// RunRefreshLoop drives RefreshLanguage for every language listed by
// the repo on the given interval. Errors are logged but not fatal so
// the loop survives intermittent upstream outages.
//
// interval defaults to the cooldown when <= 0 so back-to-back ticks
// naturally collapse via the rate-limit guard.
func (m *Manager) RunRefreshLoop(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = m.cooldown()
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			langs, err := m.Repo.ListLanguages(ctx)
			if err != nil {
				if m.Logger != nil {
					m.Logger.LanguageRefreshSkipped("", fmt.Sprintf("list_failed: %v", err))
				}
				continue
			}
			for _, l := range langs {
				if _, err := m.RefreshLanguage(ctx, l.ID); err != nil && m.Logger != nil {
					m.Logger.LanguageRefreshSkipped(l.ID, fmt.Sprintf("refresh_failed: %v", err))
				}
			}
		}
	}
}
