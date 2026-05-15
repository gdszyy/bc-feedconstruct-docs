package translations_test

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/translations"
)

// fakeRepo is the in-memory Repo used across translations_test files.
type fakeRepo struct {
	mu        sync.Mutex
	languages map[string]translations.Language
	trans     map[string]map[int64]translations.Translation
	refresh   map[string]time.Time

	upsertLanguageCalls int
	upsertTransCalls    int
	getTransCalls       int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		languages: map[string]translations.Language{},
		trans:     map[string]map[int64]translations.Translation{},
		refresh:   map[string]time.Time{},
	}
}

func (r *fakeRepo) UpsertLanguages(_ context.Context, langs []translations.Language) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.upsertLanguageCalls++
	for _, l := range langs {
		cur, ok := r.languages[l.ID]
		if ok && l.Name == "" {
			l.Name = cur.Name
		}
		r.languages[l.ID] = l
	}
	return nil
}

func (r *fakeRepo) ListLanguages(_ context.Context) ([]translations.Language, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]translations.Language, 0, len(r.languages))
	for _, l := range r.languages {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *fakeRepo) GetTranslation(_ context.Context, languageID string, translationID int64) (translations.Translation, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getTransCalls++
	bucket, ok := r.trans[languageID]
	if !ok {
		return translations.Translation{}, false, nil
	}
	t, ok := bucket[translationID]
	return t, ok, nil
}

func (r *fakeRepo) UpsertTranslations(_ context.Context, langID string, items []translations.Translation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.upsertTransCalls++
	bucket, ok := r.trans[langID]
	if !ok {
		bucket = map[int64]translations.Translation{}
		r.trans[langID] = bucket
	}
	for _, t := range items {
		t.LanguageID = langID
		bucket[t.TranslationID] = t
	}
	return nil
}

func (r *fakeRepo) LastFullRefresh(_ context.Context, languageID string) (time.Time, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.refresh[languageID]
	return t, ok, nil
}

func (r *fakeRepo) RecordFullRefresh(_ context.Context, languageID string, at time.Time, _ int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refresh[languageID] = at
	return nil
}

func (r *fakeRepo) snapshotTranslations(lang string) map[int64]translations.Translation {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[int64]translations.Translation{}
	for k, v := range r.trans[lang] {
		out[k] = v
	}
	return out
}

func (r *fakeRepo) seedTranslation(t translations.Translation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bucket, ok := r.trans[t.LanguageID]
	if !ok {
		bucket = map[int64]translations.Translation{}
		r.trans[t.LanguageID] = bucket
	}
	bucket[t.TranslationID] = t
}

// fakeAPI is an API stub that tracks call counts and replays canned
// responses keyed by language. Errors can be injected per method.
type fakeAPI struct {
	mu sync.Mutex

	languagesCalls   int
	byLangCalls      map[string]int
	byIDCalls        map[string]int
	languagesResp    []translations.Language
	byLanguageResp   map[string][]translations.Translation
	byIDResp         map[string]map[int64]translations.Translation
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		byLangCalls:    map[string]int{},
		byIDCalls:      map[string]int{},
		byLanguageResp: map[string][]translations.Translation{},
		byIDResp:       map[string]map[int64]translations.Translation{},
	}
}

func (a *fakeAPI) Languages(_ context.Context) ([]translations.Language, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.languagesCalls++
	out := make([]translations.Language, len(a.languagesResp))
	copy(out, a.languagesResp)
	return out, nil
}

func (a *fakeAPI) ByLanguage(_ context.Context, languageID string) ([]translations.Translation, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.byLangCalls[languageID]++
	items, ok := a.byLanguageResp[languageID]
	if !ok {
		return nil, nil
	}
	out := make([]translations.Translation, len(items))
	copy(out, items)
	return out, nil
}

func (a *fakeAPI) ByID(_ context.Context, languageID string, translationID int64) (translations.Translation, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := byIDKey(languageID, translationID)
	a.byIDCalls[key]++
	bucket, ok := a.byIDResp[languageID]
	if !ok {
		return translations.Translation{}, false, nil
	}
	t, ok := bucket[translationID]
	return t, ok, nil
}

func (a *fakeAPI) setLanguages(langs []translations.Language) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.languagesResp = langs
}

func (a *fakeAPI) setByLanguage(lang string, items []translations.Translation) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.byLanguageResp[lang] = items
}

func (a *fakeAPI) setByID(t translations.Translation) {
	a.mu.Lock()
	defer a.mu.Unlock()
	bucket, ok := a.byIDResp[t.LanguageID]
	if !ok {
		bucket = map[int64]translations.Translation{}
		a.byIDResp[t.LanguageID] = bucket
	}
	bucket[t.TranslationID] = t
}

func (a *fakeAPI) langCallCount(lang string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.byLangCalls[lang]
}

func (a *fakeAPI) idCallCount(lang string, id int64) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.byIDCalls[byIDKey(lang, id)]
}

func byIDKey(lang string, id int64) string {
	return lang + ":" + itoa(id)
}

func itoa(v int64) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%10]
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// captureLogger collects every Logger callback so tests can assert on
// the lifecycle without polluting stdout.
type captureLogger struct {
	mu       sync.Mutex
	langs    int
	lang     []string
	skipped  []skipEntry
	misses   []missEntry
	items    []int
}

type skipEntry struct {
	Language string
	Reason   string
}

type missEntry struct {
	Language      string
	TranslationID int64
}

func (c *captureLogger) LanguagesRefreshed(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.langs = n
}

func (c *captureLogger) LanguageRefreshed(l string, n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lang = append(c.lang, l)
	c.items = append(c.items, n)
}

func (c *captureLogger) LanguageRefreshSkipped(l, r string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.skipped = append(c.skipped, skipEntry{l, r})
}

func (c *captureLogger) LookupMiss(l string, id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.misses = append(c.misses, missEntry{l, id})
}

func (c *captureLogger) skipCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.skipped)
}
