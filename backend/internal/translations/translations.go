// Package translations is the FeedConstruct Translation Web API client
// plus a persisted cache of (language_id, translation_id) → text. Maps
// to upload-guideline 业务域 "静态描述" and to the frontend module M12
// (docs/07_frontend_architecture/modules/M12_descriptions_i18n.md).
//
// Upstream rate limit: /api/Translation/ByLanguage and
// /api/Translation/Languages must not be called more than once per hour
// per language. The Manager enforces this in two layers: an in-memory
// per-language cooldown and a Postgres-backed refresh log so a fresh
// process restart immediately after a refresh does not violate the
// upstream contract. See docs/02_translations/translations-rmq-web-api/.
package translations

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Language is one row of translation_languages.
type Language struct {
	ID        string
	Name      string
	IsDefault bool
	FetchedAt time.Time
}

// Translation is one row of translations.
type Translation struct {
	LanguageID    string
	TranslationID int64
	Text          string
	FetchedAt     time.Time
}

// RefreshSummary captures the per-language outcome of RefreshLanguage,
// useful for tests and metrics.
type RefreshSummary struct {
	LanguageID       string
	ItemCount        int
	LastRefreshAt    time.Time
	SkippedRateLimit bool
}

// Repo abstracts persistence. PgRepo implements it for production; unit
// tests use the in-memory fake in helpers_test.go.
type Repo interface {
	UpsertLanguages(ctx context.Context, langs []Language) error
	ListLanguages(ctx context.Context) ([]Language, error)

	GetTranslation(ctx context.Context, languageID string, translationID int64) (Translation, bool, error)
	UpsertTranslations(ctx context.Context, langID string, items []Translation) error

	// LastFullRefresh returns the wall-clock time of the most recent
	// successful ByLanguage refresh for langID. The boolean is false
	// when no refresh has been recorded yet.
	LastFullRefresh(ctx context.Context, languageID string) (time.Time, bool, error)
	RecordFullRefresh(ctx context.Context, languageID string, at time.Time, itemCount int) error
}

// API is the slice of Client the Manager calls. Defined as an interface
// so the Manager can be tested against a stub without spinning up
// httptest.Server (which is still covered by api_test.go).
type API interface {
	Languages(ctx context.Context) ([]Language, error)
	ByLanguage(ctx context.Context, languageID string) ([]Translation, error)
	ByID(ctx context.Context, languageID string, translationID int64) (Translation, bool, error)
}

// Logger observes refresh lifecycle. Pass nil to silently drop.
type Logger interface {
	LanguagesRefreshed(count int)
	LanguageRefreshed(languageID string, items int)
	LanguageRefreshSkipped(languageID string, reason string)
	LookupMiss(languageID string, translationID int64)
}

// LoggerFunc adapts a struct of plain functions into a Logger; nil
// fields short-circuit to no-op.
type LoggerFunc struct {
	OnLanguages    func(count int)
	OnLanguage     func(languageID string, items int)
	OnLanguageSkip func(languageID string, reason string)
	OnMiss         func(languageID string, translationID int64)
}

// LanguagesRefreshed implements Logger.
func (f LoggerFunc) LanguagesRefreshed(c int) {
	if f.OnLanguages != nil {
		f.OnLanguages(c)
	}
}

// LanguageRefreshed implements Logger.
func (f LoggerFunc) LanguageRefreshed(l string, n int) {
	if f.OnLanguage != nil {
		f.OnLanguage(l, n)
	}
}

// LanguageRefreshSkipped implements Logger.
func (f LoggerFunc) LanguageRefreshSkipped(l, r string) {
	if f.OnLanguageSkip != nil {
		f.OnLanguageSkip(l, r)
	}
}

// LookupMiss implements Logger.
func (f LoggerFunc) LookupMiss(l string, id int64) {
	if f.OnMiss != nil {
		f.OnMiss(l, id)
	}
}

// DefaultRefreshCooldown matches the FeedConstruct doc: getByLanguage
// and Languages methods should not be called more frequently than once
// per hour for each language.
const DefaultRefreshCooldown = time.Hour

// Reason strings reported via Logger.LanguageRefreshSkipped.
const (
	ReasonRateLimit = "rate_limited"
)

// Manager orchestrates the Translation cache. Construct via New, set
// the optional Logger / Now, then call RefreshLanguages once at boot
// and RefreshLanguage on demand (or on a periodic ticker).
type Manager struct {
	API    API
	Repo   Repo
	Logger Logger
	Now    func() time.Time

	// RefreshCooldown is the minimum elapsed time between two
	// successful ByLanguage refreshes for the same language. Defaults
	// to DefaultRefreshCooldown.
	RefreshCooldown time.Duration

	mu         sync.Mutex
	lastByLang map[string]time.Time

	refreshTotal atomic.Int64
	skipTotal    atomic.Int64
	missTotal    atomic.Int64
}

// New returns a Manager bound to api and repo with safe defaults.
func New(api API, repo Repo) *Manager {
	return &Manager{
		API:             api,
		Repo:            repo,
		RefreshCooldown: DefaultRefreshCooldown,
		lastByLang:      map[string]time.Time{},
	}
}

// RefreshCount returns the number of successful ByLanguage refreshes
// since construction.
func (m *Manager) RefreshCount() int64 { return m.refreshTotal.Load() }

// SkipCount returns the number of ByLanguage refreshes skipped because
// the cooldown had not elapsed.
func (m *Manager) SkipCount() int64 { return m.skipTotal.Load() }

// MissCount returns the number of Manager.Lookup calls that fell
// through to the upstream ById endpoint.
func (m *Manager) MissCount() int64 { return m.missTotal.Load() }

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now().UTC()
}

func (m *Manager) cooldown() time.Duration {
	if m.RefreshCooldown > 0 {
		return m.RefreshCooldown
	}
	return DefaultRefreshCooldown
}
