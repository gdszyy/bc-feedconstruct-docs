package translations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo is the production pgxpool-backed Repo implementation.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

// UpsertLanguages writes every language in langs. An empty Name field
// preserves the existing one so a stub upsert (used by RefreshLanguage
// before RefreshLanguages has run) does not blank a real name.
func (r *PgRepo) UpsertLanguages(ctx context.Context, langs []Language) error {
	if len(langs) == 0 {
		return nil
	}
	const q = `
		INSERT INTO translation_languages (language_id, name, is_default, fetched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (language_id) DO UPDATE SET
			name       = CASE WHEN EXCLUDED.name = '' THEN translation_languages.name ELSE EXCLUDED.name END,
			is_default = EXCLUDED.is_default,
			fetched_at = EXCLUDED.fetched_at`
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("storage: begin upsert languages: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, l := range langs {
		fetchedAt := l.FetchedAt
		if fetchedAt.IsZero() {
			fetchedAt = time.Now().UTC()
		}
		if _, err := tx.Exec(ctx, q, l.ID, l.Name, l.IsDefault, fetchedAt); err != nil {
			return fmt.Errorf("storage: upsert language %q: %w", l.ID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("storage: commit upsert languages: %w", err)
	}
	return nil
}

// ListLanguages returns every row in translation_languages ordered by
// language_id (deterministic for the refresh loop).
func (r *PgRepo) ListLanguages(ctx context.Context) ([]Language, error) {
	const q = `
		SELECT language_id, COALESCE(name, ''), is_default, fetched_at
		  FROM translation_languages
		 ORDER BY language_id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("storage: list languages: %w", err)
	}
	defer rows.Close()
	var out []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.ID, &l.Name, &l.IsDefault, &l.FetchedAt); err != nil {
			return nil, fmt.Errorf("storage: scan language: %w", err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetTranslation returns the cached row, if any.
func (r *PgRepo) GetTranslation(ctx context.Context, languageID string, translationID int64) (Translation, bool, error) {
	const q = `
		SELECT language_id, translation_id, text, fetched_at
		  FROM translations
		 WHERE language_id = $1 AND translation_id = $2`
	var t Translation
	err := r.pool.QueryRow(ctx, q, languageID, translationID).Scan(
		&t.LanguageID, &t.TranslationID, &t.Text, &t.FetchedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Translation{}, false, nil
		}
		return Translation{}, false, fmt.Errorf("storage: get translation: %w", err)
	}
	return t, true, nil
}

// UpsertTranslations writes every translation in items for langID.
// Empty Text overrides any cached value (the upstream is authoritative).
func (r *PgRepo) UpsertTranslations(ctx context.Context, langID string, items []Translation) error {
	if len(items) == 0 {
		return nil
	}
	const q = `
		INSERT INTO translations (language_id, translation_id, text, fetched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (language_id, translation_id) DO UPDATE SET
			text       = EXCLUDED.text,
			fetched_at = EXCLUDED.fetched_at`
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("storage: begin upsert translations: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, t := range items {
		fetchedAt := t.FetchedAt
		if fetchedAt.IsZero() {
			fetchedAt = time.Now().UTC()
		}
		if _, err := tx.Exec(ctx, q, langID, t.TranslationID, t.Text, fetchedAt); err != nil {
			return fmt.Errorf("storage: upsert translation %s/%d: %w", langID, t.TranslationID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("storage: commit upsert translations: %w", err)
	}
	return nil
}

// LastFullRefresh reads the most recent successful ByLanguage refresh
// for langID from translation_refresh_log.
func (r *PgRepo) LastFullRefresh(ctx context.Context, languageID string) (time.Time, bool, error) {
	const q = `SELECT last_full_refresh_at FROM translation_refresh_log WHERE language_id = $1`
	var t time.Time
	if err := r.pool.QueryRow(ctx, q, languageID).Scan(&t); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, fmt.Errorf("storage: last refresh: %w", err)
	}
	return t, true, nil
}

// RecordFullRefresh upserts the timestamp of the most recent successful
// ByLanguage refresh. The languages row must already exist (FK).
func (r *PgRepo) RecordFullRefresh(ctx context.Context, languageID string, at time.Time, itemCount int) error {
	const q = `
		INSERT INTO translation_refresh_log (language_id, last_full_refresh_at, item_count)
		VALUES ($1, $2, $3)
		ON CONFLICT (language_id) DO UPDATE SET
			last_full_refresh_at = EXCLUDED.last_full_refresh_at,
			item_count           = EXCLUDED.item_count`
	if _, err := r.pool.Exec(ctx, q, languageID, at, itemCount); err != nil {
		return fmt.Errorf("storage: record refresh: %w", err)
	}
	return nil
}
