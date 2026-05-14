package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// Handler upserts catalog rows from feed deliveries.
type Handler struct {
	pool *storage.Pool
}

// New returns a Handler bound to pool.
func New(pool *storage.Pool) *Handler { return &Handler{pool: pool} }

// Register binds this handler to every catalog message type understood
// by the package. After Register the dispatcher routes:
//
//	catalog.sport / catalog.region / catalog.competition / catalog.market_type
//	fixture / fixture_change
//
// to Handle below. catalog.market_type is currently a no-op (description
// data lives in a later wave) but registering it prevents dead-letter spam.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgCatalogSport, h)
	d.Register(feed.MsgCatalogRegion, h)
	d.Register(feed.MsgCatalogComp, h)
	d.Register(feed.MsgCatalogMarketTyp, h)
	d.Register(feed.MsgFixture, h)
	d.Register(feed.MsgFixtureChange, h)
}

// Handle implements feed.Handler.
func (h *Handler) Handle(ctx context.Context, msgType feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	switch msgType {
	case feed.MsgCatalogSport:
		return h.handleSport(ctx, env)
	case feed.MsgCatalogRegion:
		return h.handleRegion(ctx, env)
	case feed.MsgCatalogComp:
		return h.handleCompetition(ctx, env)
	case feed.MsgCatalogMarketTyp:
		return nil // descriptions wave handles these
	case feed.MsgFixture:
		return h.handleMatch(ctx, env, rawID, false)
	case feed.MsgFixtureChange:
		return h.handleMatch(ctx, env, rawID, true)
	}
	return fmt.Errorf("catalog: unsupported message type %q", msgType)
}

func (h *Handler) handleSport(ctx context.Context, env feed.Envelope) error {
	p, err := parseSport(env.Payload)
	if err != nil {
		return fmt.Errorf("catalog: parse sport: %w", err)
	}
	id, ok := pickID(p.ID, p.ObjectID)
	if !ok {
		return errors.New("catalog: sport without id")
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	_, err = h.pool.Exec(ctx, `
		INSERT INTO sports (id, name, is_active, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (id) DO UPDATE
		   SET name = COALESCE(NULLIF(EXCLUDED.name, ''), sports.name),
		       is_active = EXCLUDED.is_active,
		       updated_at = now()`,
		id, p.Name, active)
	if err != nil {
		return fmt.Errorf("catalog: upsert sport %d: %w", id, err)
	}
	return nil
}

func (h *Handler) handleRegion(ctx context.Context, env feed.Envelope) error {
	p, err := parseRegion(env.Payload)
	if err != nil {
		return fmt.Errorf("catalog: parse region: %w", err)
	}
	id, ok := pickID(p.ID, p.ObjectID)
	if !ok {
		return errors.New("catalog: region without id")
	}
	if p.SportID == nil {
		return fmt.Errorf("catalog: region %d without sportId", id)
	}
	if err := h.ensureSport(ctx, *p.SportID); err != nil {
		return err
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	_, err = h.pool.Exec(ctx, `
		INSERT INTO regions (id, sport_id, name, is_active, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (id) DO UPDATE
		   SET sport_id = EXCLUDED.sport_id,
		       name = COALESCE(NULLIF(EXCLUDED.name, ''), regions.name),
		       is_active = EXCLUDED.is_active,
		       updated_at = now()`,
		id, *p.SportID, p.Name, active)
	if err != nil {
		return fmt.Errorf("catalog: upsert region %d: %w", id, err)
	}
	return nil
}

func (h *Handler) handleCompetition(ctx context.Context, env feed.Envelope) error {
	p, err := parseCompetition(env.Payload)
	if err != nil {
		return fmt.Errorf("catalog: parse competition: %w", err)
	}
	id, ok := pickID(p.ID, p.ObjectID)
	if !ok {
		return errors.New("catalog: competition without id")
	}
	if p.SportID == nil || p.RegionID == nil {
		return fmt.Errorf("catalog: competition %d without sportId/regionId", id)
	}
	if err := h.ensureSport(ctx, *p.SportID); err != nil {
		return err
	}
	if err := h.ensureRegion(ctx, *p.RegionID, *p.SportID); err != nil {
		return err
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	_, err = h.pool.Exec(ctx, `
		INSERT INTO competitions (id, region_id, sport_id, name, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (id) DO UPDATE
		   SET region_id = EXCLUDED.region_id,
		       sport_id = EXCLUDED.sport_id,
		       name = COALESCE(NULLIF(EXCLUDED.name, ''), competitions.name),
		       is_active = EXCLUDED.is_active,
		       updated_at = now()`,
		id, *p.RegionID, *p.SportID, p.Name, active)
	if err != nil {
		return fmt.Errorf("catalog: upsert competition %d: %w", id, err)
	}
	return nil
}

// handleMatch upserts the match row, enforcing the no-regression rule, and
// (when changed=true) appends a fixture_changes history row.
func (h *Handler) handleMatch(ctx context.Context, env feed.Envelope, rawID [16]byte, changed bool) error {
	p, err := parseMatch(env.Payload)
	if err != nil {
		return fmt.Errorf("catalog: parse match: %w", err)
	}
	id, ok := pickID(p.MatchID, p.ID, p.ObjectID, optInt64(env.MatchID))
	if !ok {
		return errors.New("catalog: match without id")
	}
	if p.SportID == nil {
		return fmt.Errorf("catalog: match %d without sportId", id)
	}
	if err := h.ensureSport(ctx, *p.SportID); err != nil {
		return err
	}
	if p.CompetitionID != nil {
		// ensureCompetition needs a region id; competitions row has both
		// already if it was upserted earlier. We skip creating a placeholder
		// here because the FK on matches.competition_id is ON DELETE SET NULL
		// and the column itself allows NULL. The competition_id will be
		// populated later when the catalog.competition delivery arrives.
		var exists bool
		if err := h.pool.QueryRow(ctx,
			`SELECT exists(SELECT 1 FROM competitions WHERE id = $1)`, *p.CompetitionID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("catalog: lookup competition %d: %w", *p.CompetitionID, err)
		}
		if !exists {
			p.CompetitionID = nil
		}
	}

	startAt := pickTime(p.StartAt, p.StartTime)
	status := normaliseStatus(p.Status)
	isLive := false
	if p.IsLive != nil {
		isLive = *p.IsLive
	} else if status == "live" {
		isLive = true
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("catalog: begin match tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		curStatus  string
		curStartAt *string // serialised text to keep diff JSON tidy
	)
	row := tx.QueryRow(ctx,
		`SELECT status, to_char(start_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		   FROM matches WHERE id = $1 FOR UPDATE`, id)
	err = row.Scan(&curStatus, &curStartAt)
	exists := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("catalog: lookup match %d: %w", id, err)
	}

	// Decide on the status to write.
	writeStatus := status
	if exists && !allowsTransition(curStatus, status) {
		writeStatus = curStatus // block regression
		fmt.Printf("status.regress.blocked match=%d from=%s to=%s\n", id, curStatus, status)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO matches (id, sport_id, competition_id, name, home, away, start_at, is_live, status, last_event_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		ON CONFLICT (id) DO UPDATE
		   SET sport_id = EXCLUDED.sport_id,
		       competition_id = COALESCE(EXCLUDED.competition_id, matches.competition_id),
		       name = COALESCE(NULLIF(EXCLUDED.name, ''), matches.name),
		       home = COALESCE(NULLIF(EXCLUDED.home, ''), matches.home),
		       away = COALESCE(NULLIF(EXCLUDED.away, ''), matches.away),
		       start_at = COALESCE(EXCLUDED.start_at, matches.start_at),
		       is_live = EXCLUDED.is_live,
		       status = EXCLUDED.status,
		       last_event_id = EXCLUDED.last_event_id,
		       updated_at = now()`,
		id, *p.SportID, p.CompetitionID, p.Name, p.Home, p.Away,
		startAt, isLive, writeStatus, env.EventKey(),
	)
	if err != nil {
		return fmt.Errorf("catalog: upsert match %d: %w", id, err)
	}

	// fixture_changes history when the status or start_at actually moved.
	statusChanged := exists && curStatus != writeStatus
	startChanged := exists && curStartAt != nil && startAt != nil &&
		(*curStartAt)[:19] != startAt.UTC().Format("2006-01-02T15:04:05")
	if changed || statusChanged || startChanged {
		var newStartAt string
		if startAt != nil {
			newStartAt = startAt.UTC().Format("2006-01-02T15:04:05Z")
		}
		oldJ, _ := json.Marshal(map[string]any{"status": curStatus, "start_at": curStartAt})
		newJ, _ := json.Marshal(map[string]any{"status": writeStatus, "start_at": newStartAt})
		ct := "fixture_change"
		if !exists {
			ct = "fixture_create"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO fixture_changes (match_id, change_type, old, new, raw_message_id, received_at)
			VALUES ($1, $2, $3::jsonb, $4::jsonb, $5, now())`,
			id, ct, oldJ, newJ, nullableRawID(rawID),
		); err != nil {
			return fmt.Errorf("catalog: insert fixture_change %d: %w", id, err)
		}
	}
	return tx.Commit(ctx)
}

// ensureSport inserts a placeholder sport row if missing so the FK on
// regions/competitions/matches holds. The real name + is_active arrive
// later via the catalog.sport delivery.
func (h *Handler) ensureSport(ctx context.Context, id int64) error {
	_, err := h.pool.Exec(ctx, `
		INSERT INTO sports (id, name, updated_at)
		VALUES ($1, '', now())
		ON CONFLICT (id) DO NOTHING`, id)
	if err != nil {
		return fmt.Errorf("catalog: ensure sport %d: %w", id, err)
	}
	return nil
}

func (h *Handler) ensureRegion(ctx context.Context, id, sportID int64) error {
	_, err := h.pool.Exec(ctx, `
		INSERT INTO regions (id, sport_id, name, updated_at)
		VALUES ($1, $2, '', now())
		ON CONFLICT (id) DO NOTHING`, id, sportID)
	if err != nil {
		return fmt.Errorf("catalog: ensure region %d: %w", id, err)
	}
	return nil
}

func optInt64(p *int64) *int64 { return p }

// nullableRawID returns nil when id is the zero UUID so the FK to
// raw_messages stays satisfied for handler calls that have no audit row
// (e.g. recovery-triggered upserts).
func nullableRawID(id [16]byte) any {
	var zero [16]byte
	if id == zero {
		return nil
	}
	return id
}
