// Package catalog handles sport / region / competition / match / fixture_change
// updates. Maps to upload-guideline 业务域 "赛事主数据" (M03/M04).
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RegionRef / CompetitionRef are the minimum tuples needed to upsert a
// row in the regions / competitions tables.
type RegionRef struct {
	ID      int32
	SportID int32
	Name    string
}

type CompetitionRef struct {
	ID       int32
	RegionID int32
	SportID  int32
	Name     string
}

// MatchRecord is the in-process representation of a row in the matches
// table. Pointer fields are nullable on disk.
type MatchRecord struct {
	ID            int64
	SportID       int32
	CompetitionID *int32
	Name          string
	Home          string
	Away          string
	StartAt       time.Time
	IsLive        bool
	Status        string
}

// FixtureChange is appended to the fixture_changes table whenever a
// fixture_change delivery materially mutates the match row.
type FixtureChange struct {
	MatchID      int64
	ChangeType   string
	Old          map[string]any
	New          map[string]any
	RawMessageID [16]byte
	ReceivedAt   time.Time
}

// Repo is the persistence contract the handler depends on.
type Repo interface {
	UpsertSport(ctx context.Context, id int32, name string) error
	UpsertRegion(ctx context.Context, r RegionRef) error
	UpsertCompetition(ctx context.Context, c CompetitionRef) error
	LoadMatch(ctx context.Context, id int64) (MatchRecord, bool, error)
	UpsertMatch(ctx context.Context, m MatchRecord) error
	AppendFixtureChange(ctx context.Context, fc FixtureChange) error
}

// Logger is a tiny WARN-only interface; structured fields stay free-form
// so the BFF can adapt slog, zap or a no-op in tests.
type Logger interface {
	Warn(event string, fields map[string]any)
}

type noopLogger struct{}

func (noopLogger) Warn(string, map[string]any) {}

// Options configures Handler behaviour.
type Options struct {
	Logger Logger
	Now    func() time.Time
}

// Handler turns a raw Match payload into upserts on the catalog tables.
type Handler struct {
	repo Repo
	log  Logger
	now  func() time.Time
}

// NewHandler constructs a Handler. A nil Logger is replaced with a noop.
func NewHandler(repo Repo, opts Options) *Handler {
	if opts.Logger == nil {
		opts.Logger = noopLogger{}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &Handler{repo: repo, log: opts.Logger, now: opts.Now}
}

// HandleMatch processes a Match / fixture delivery: it upserts the
// hierarchy and (when no terminal-status regression would result) the
// match row itself.
func (h *Handler) HandleMatch(ctx context.Context, payload []byte, rawID [16]byte) error {
	return h.handle(ctx, payload, rawID, false)
}

// HandleFixtureChange is the same flow as HandleMatch but also appends a
// fixture_changes row whenever the resulting MatchRecord differs from
// the previously stored one.
func (h *Handler) HandleFixtureChange(ctx context.Context, payload []byte, rawID [16]byte) error {
	return h.handle(ctx, payload, rawID, true)
}

func (h *Handler) handle(ctx context.Context, payload []byte, rawID [16]byte, recordFixtureChange bool) error {
	wire, err := decodeMatch(payload)
	if err != nil {
		return fmt.Errorf("catalog: decode match: %w", err)
	}
	if wire.ID == 0 {
		return errors.New("catalog: match payload missing Id")
	}

	if err := h.repo.UpsertSport(ctx, wire.SportID, ""); err != nil {
		return fmt.Errorf("catalog: upsert sport: %w", err)
	}
	if wire.RegionID != 0 {
		if err := h.repo.UpsertRegion(ctx, RegionRef{ID: wire.RegionID, SportID: wire.SportID}); err != nil {
			return fmt.Errorf("catalog: upsert region: %w", err)
		}
	}
	if wire.CompetitionID != 0 && wire.RegionID != 0 {
		if err := h.repo.UpsertCompetition(ctx, CompetitionRef{
			ID:       wire.CompetitionID,
			RegionID: wire.RegionID,
			SportID:  wire.SportID,
		}); err != nil {
			return fmt.Errorf("catalog: upsert competition: %w", err)
		}
	}

	prev, prevOK, err := h.repo.LoadMatch(ctx, wire.ID)
	if err != nil {
		return fmt.Errorf("catalog: load match: %w", err)
	}

	next := wireToRecord(wire)
	if prevOK {
		applyRegressionGuard(&next, prev, h.log)
	}

	if err := h.repo.UpsertMatch(ctx, next); err != nil {
		return fmt.Errorf("catalog: upsert match: %w", err)
	}

	if recordFixtureChange && prevOK {
		if diff := diffMatch(prev, next); diff != nil {
			fc := FixtureChange{
				MatchID:      next.ID,
				ChangeType:   "fixture_change",
				Old:          diff.old,
				New:          diff.new,
				RawMessageID: rawID,
				ReceivedAt:   h.now(),
			}
			if err := h.repo.AppendFixtureChange(ctx, fc); err != nil {
				return fmt.Errorf("catalog: append fixture_change: %w", err)
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Wire ↔ domain
// ---------------------------------------------------------------------------

type matchMember struct {
	Type int    `json:"Type"`
	Name string `json:"Name"`
}

type wireMatch struct {
	ID            int64         `json:"Id"`
	SportID       int32         `json:"SportId"`
	RegionID      int32         `json:"RegionId"`
	CompetitionID int32         `json:"CompetitionId"`
	Date          string        `json:"Date"`
	IsLive        bool          `json:"IsLive"`
	MatchStatus   int           `json:"MatchStatus"`
	MatchMembers  []matchMember `json:"MatchMembers"`
}

func decodeMatch(payload []byte) (wireMatch, error) {
	var w wireMatch
	if err := json.Unmarshal(payload, &w); err != nil {
		return w, err
	}
	return w, nil
}

func wireToRecord(w wireMatch) MatchRecord {
	rec := MatchRecord{
		ID:      w.ID,
		SportID: w.SportID,
		IsLive:  w.IsLive,
		Status:  statusFromWire(w.MatchStatus, w.IsLive),
	}
	if w.CompetitionID != 0 {
		c := w.CompetitionID
		rec.CompetitionID = &c
	}
	if t, err := time.Parse(time.RFC3339, w.Date); err == nil {
		rec.StartAt = t.UTC()
	}
	for _, m := range w.MatchMembers {
		switch m.Type {
		case 1:
			rec.Home = m.Name
		case 2:
			rec.Away = m.Name
		}
	}
	return rec
}

// statusFromWire converts FeedConstruct's numeric MatchStatus into the
// internal status enum. IsLive lifts NotStarted/Started -> live when set,
// so the BFF reflects the live producer's authoritative flag.
func statusFromWire(matchStatus int, isLive bool) string {
	switch matchStatus {
	case 2:
		return "ended"
	case 3:
		return "cancelled"
	}
	if isLive {
		return "live"
	}
	if matchStatus == 1 {
		return "live"
	}
	return "not_started"
}

// ---------------------------------------------------------------------------
// Regression guard
// ---------------------------------------------------------------------------

func terminalStatus(s string) bool {
	switch s {
	case "ended", "closed", "cancelled":
		return true
	}
	return false
}

// applyRegressionGuard mutates next in place to prevent a forbidden
// regression and emits a structured warn log when it does so.
func applyRegressionGuard(next *MatchRecord, prev MatchRecord, log Logger) {
	if !terminalStatus(prev.Status) {
		return
	}
	// Terminal -> terminal transitions are still allowed (real corrections
	// like ended -> cancelled). Only terminal -> non-terminal is blocked.
	if terminalStatus(next.Status) {
		return
	}
	log.Warn("status.regress.blocked", map[string]any{
		"match_id":   next.ID,
		"from":       prev.Status,
		"attempted":  next.Status,
		"is_live_in": next.IsLive,
	})
	next.Status = prev.Status
	next.IsLive = false
}

// ---------------------------------------------------------------------------
// Diff
// ---------------------------------------------------------------------------

type matchDiff struct {
	old map[string]any
	new map[string]any
}

func diffMatch(prev, next MatchRecord) *matchDiff {
	d := &matchDiff{old: map[string]any{}, new: map[string]any{}}
	if prev.Status != next.Status {
		d.old["status"] = prev.Status
		d.new["status"] = next.Status
	}
	if !prev.StartAt.Equal(next.StartAt) {
		d.old["start_at"] = formatTime(prev.StartAt)
		d.new["start_at"] = formatTime(next.StartAt)
	}
	if prev.IsLive != next.IsLive {
		d.old["is_live"] = prev.IsLive
		d.new["is_live"] = next.IsLive
	}
	if len(d.old) == 0 {
		return nil
	}
	return d
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
