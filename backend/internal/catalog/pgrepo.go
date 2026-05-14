package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo persists the M03/M04 catalog hierarchy via pgx. Satisfies Repo.
type PgRepo struct{ pool *pgxpool.Pool }

func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *PgRepo) UpsertSport(ctx context.Context, s Sport) error {
	const q = `
		INSERT INTO sports (id, name, is_active, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (id) DO UPDATE
		   SET name       = CASE
		                       WHEN EXCLUDED.name = '' THEN sports.name
		                       ELSE EXCLUDED.name
		                    END,
		       is_active  = EXCLUDED.is_active,
		       updated_at = now()`
	if _, err := r.pool.Exec(ctx, q, s.ID, s.Name, s.IsActive); err != nil {
		return fmt.Errorf("storage: upsert sport: %w", err)
	}
	return nil
}

func (r *PgRepo) SoftDeleteSport(ctx context.Context, id int32) error {
	const q = `
		INSERT INTO sports (id, name, is_active, updated_at)
		VALUES ($1, '', false, now())
		ON CONFLICT (id) DO UPDATE
		   SET is_active = false, updated_at = now()`
	if _, err := r.pool.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("storage: soft-delete sport: %w", err)
	}
	return nil
}

func (r *PgRepo) GetSport(ctx context.Context, id int32) (Sport, bool, error) {
	const q = `SELECT id, name, is_active FROM sports WHERE id = $1`
	var s Sport
	err := r.pool.QueryRow(ctx, q, id).Scan(&s.ID, &s.Name, &s.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Sport{}, false, nil
		}
		return Sport{}, false, fmt.Errorf("storage: get sport: %w", err)
	}
	return s, true, nil
}

func (r *PgRepo) UpsertRegion(ctx context.Context, rg Region) error {
	const q = `
		INSERT INTO regions (id, sport_id, name, is_active, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (id) DO UPDATE
		   SET sport_id   = EXCLUDED.sport_id,
		       name       = CASE
		                       WHEN EXCLUDED.name = '' THEN regions.name
		                       ELSE EXCLUDED.name
		                    END,
		       is_active  = EXCLUDED.is_active,
		       updated_at = now()`
	if _, err := r.pool.Exec(ctx, q, rg.ID, rg.SportID, rg.Name, rg.IsActive); err != nil {
		return fmt.Errorf("storage: upsert region: %w", err)
	}
	return nil
}

func (r *PgRepo) GetRegion(ctx context.Context, id int32) (Region, bool, error) {
	const q = `SELECT id, sport_id, name, is_active FROM regions WHERE id = $1`
	var rg Region
	err := r.pool.QueryRow(ctx, q, id).Scan(&rg.ID, &rg.SportID, &rg.Name, &rg.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Region{}, false, nil
		}
		return Region{}, false, fmt.Errorf("storage: get region: %w", err)
	}
	return rg, true, nil
}

func (r *PgRepo) UpsertCompetition(ctx context.Context, c Competition) error {
	const q = `
		INSERT INTO competitions (id, region_id, sport_id, name, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (id) DO UPDATE
		   SET region_id  = EXCLUDED.region_id,
		       sport_id   = EXCLUDED.sport_id,
		       name       = CASE
		                       WHEN EXCLUDED.name = '' THEN competitions.name
		                       ELSE EXCLUDED.name
		                    END,
		       is_active  = EXCLUDED.is_active,
		       updated_at = now()`
	if _, err := r.pool.Exec(ctx, q, c.ID, c.RegionID, c.SportID, c.Name, c.IsActive); err != nil {
		return fmt.Errorf("storage: upsert competition: %w", err)
	}
	return nil
}

func (r *PgRepo) UpsertMatch(ctx context.Context, m Match) error {
	const q = `
		INSERT INTO matches
		    (id, sport_id, competition_id, name, home, away, start_at,
		     is_live, status, last_event_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		ON CONFLICT (id) DO UPDATE
		   SET sport_id       = EXCLUDED.sport_id,
		       competition_id = EXCLUDED.competition_id,
		       name           = EXCLUDED.name,
		       home           = EXCLUDED.home,
		       away           = EXCLUDED.away,
		       start_at       = EXCLUDED.start_at,
		       is_live        = EXCLUDED.is_live,
		       status         = EXCLUDED.status,
		       last_event_id  = EXCLUDED.last_event_id,
		       updated_at     = now()`
	_, err := r.pool.Exec(ctx, q,
		m.ID, m.SportID, m.CompetitionID,
		nullableStr(m.Name), nullableStr(m.Home), nullableStr(m.Away),
		m.StartAt, m.IsLive, string(m.Status), nullableStr(m.LastEventID),
	)
	if err != nil {
		return fmt.Errorf("storage: upsert match: %w", err)
	}
	return nil
}

func (r *PgRepo) GetMatch(ctx context.Context, id int64) (Match, bool, error) {
	const q = `
		SELECT id, sport_id, competition_id, COALESCE(name, ''),
		       COALESCE(home, ''), COALESCE(away, ''), start_at,
		       is_live, status, COALESCE(last_event_id, '')
		  FROM matches
		 WHERE id = $1`
	var m Match
	var status string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.SportID, &m.CompetitionID,
		&m.Name, &m.Home, &m.Away, &m.StartAt,
		&m.IsLive, &status, &m.LastEventID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Match{}, false, nil
		}
		return Match{}, false, fmt.Errorf("storage: get match: %w", err)
	}
	m.Status = MatchStatus(status)
	return m, true, nil
}

func (r *PgRepo) InsertFixtureChange(ctx context.Context, row FixtureChangeRow) error {
	oldJSON, err := json.Marshal(row.Old)
	if err != nil {
		return fmt.Errorf("storage: marshal fixture_change.old: %w", err)
	}
	newJSON, err := json.Marshal(row.New)
	if err != nil {
		return fmt.Errorf("storage: marshal fixture_change.new: %w", err)
	}
	const q = `
		INSERT INTO fixture_changes (match_id, change_type, old, new, raw_message_id)
		VALUES ($1, $2, $3::jsonb, $4::jsonb, $5)`
	if _, err := r.pool.Exec(ctx, q, row.MatchID, row.ChangeType, string(oldJSON), string(newJSON), row.RawMessageID); err != nil {
		return fmt.Errorf("storage: insert fixture_change: %w", err)
	}
	return nil
}
