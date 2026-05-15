// Package catalog handles FeedConstruct catalog deliveries (sport / region /
// competition / match) and persists the resulting hierarchy. Maps to M03/M04
// in docs/07_frontend_architecture/modules/.
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// MatchStatus mirrors the CHECK constraint in migrations/002_catalog.sql.
type MatchStatus string

const (
	StatusUnknown    MatchStatus = ""
	StatusNotStarted MatchStatus = "not_started"
	StatusLive       MatchStatus = "live"
	StatusPostponed  MatchStatus = "postponed"
	StatusEnded      MatchStatus = "ended"
	StatusCancelled  MatchStatus = "cancelled"
	StatusClosed     MatchStatus = "closed"
)

// statusRank orders the lifecycle from earliest to latest. Anti-regression
// blocks any transition with a lower rank than the persisted one.
var statusRank = map[MatchStatus]int{
	StatusNotStarted: 0,
	StatusPostponed:  1,
	StatusLive:       2,
	StatusEnded:      3,
	StatusCancelled:  3,
	StatusClosed:     4,
}

func rankOf(s MatchStatus) (int, bool) {
	r, ok := statusRank[s]
	return r, ok
}

type Sport struct {
	ID       int32
	Name     string
	IsActive bool
}

type Region struct {
	ID       int32
	SportID  int32
	Name     string
	IsActive bool
}

type Competition struct {
	ID       int32
	RegionID int32
	SportID  int32
	Name     string
	IsActive bool
}

type Match struct {
	ID            int64
	SportID       int32
	CompetitionID *int32
	Name          string
	Home          string
	Away          string
	StartAt       *time.Time
	IsLive        bool
	Status        MatchStatus
	LastEventID   string
}

// FixtureChangeRow is one row of the fixture_changes audit log.
type FixtureChangeRow struct {
	MatchID      int64
	ChangeType   string
	Old          map[string]any
	New          map[string]any
	RawMessageID [16]byte
}

// Repo abstracts persistence. The pgx-backed implementation lives next
// to other storage code; unit tests use an in-memory fake.
type Repo interface {
	UpsertSport(ctx context.Context, s Sport) error
	SoftDeleteSport(ctx context.Context, id int32) error
	GetSport(ctx context.Context, id int32) (Sport, bool, error)

	UpsertRegion(ctx context.Context, r Region) error
	GetRegion(ctx context.Context, id int32) (Region, bool, error)

	UpsertCompetition(ctx context.Context, c Competition) error

	UpsertMatch(ctx context.Context, m Match) error
	GetMatch(ctx context.Context, id int64) (Match, bool, error)

	InsertFixtureChange(ctx context.Context, row FixtureChangeRow) error
}

// AntiRegressionEvent is emitted whenever an incoming status would
// regress the persisted lifecycle.
type AntiRegressionEvent struct {
	MatchID int64
	From    MatchStatus
	To      MatchStatus
}

// Logger receives anti-regression notices. Pass nil to silently drop.
type Logger interface {
	AntiRegressionBlocked(ev AntiRegressionEvent)
}

// LoggerFunc adapts a plain function into a Logger.
type LoggerFunc func(ev AntiRegressionEvent)

func (f LoggerFunc) AntiRegressionBlocked(ev AntiRegressionEvent) { f(ev) }

// MatchObserver is notified whenever HandleMatch successfully applies a
// real status transition (StatusUnknown → X or X → Y where X != Y).
// Anti-regression hits are *not* reported; the observer only sees
// effective transitions persisted to the matches row.
type MatchObserver interface {
	MatchStatusChanged(ctx context.Context, matchID int64, from, to MatchStatus)
}

// MatchObserverFunc adapts a function into a MatchObserver.
type MatchObserverFunc func(ctx context.Context, matchID int64, from, to MatchStatus)

// MatchStatusChanged implements MatchObserver.
func (f MatchObserverFunc) MatchStatusChanged(ctx context.Context, matchID int64, from, to MatchStatus) {
	f(ctx, matchID, from, to)
}

// Handler is the catalog facade. Construct via New, then call Register
// to bind it to the feed dispatcher.
type Handler struct {
	Repo     Repo
	Logger   Logger
	Observer MatchObserver
	Now      func() time.Time

	regressionCount atomic.Int64
}

func New(repo Repo) *Handler { return &Handler{Repo: repo} }

// RegressionCount returns the number of anti-regression events blocked
// since construction.
func (h *Handler) RegressionCount() int64 { return h.regressionCount.Load() }

// Register wires the catalog handler into a feed dispatcher.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgCatalogSport, feed.HandlerFunc(h.HandleSport))
	d.Register(feed.MsgCatalogRegion, feed.HandlerFunc(h.HandleRegion))
	d.Register(feed.MsgCatalogComp, feed.HandlerFunc(h.HandleCompetition))
	d.Register(feed.MsgFixture, feed.HandlerFunc(h.HandleMatch))
	d.Register(feed.MsgFixtureChange, feed.HandlerFunc(h.HandleMatch))
}

type sportPayload struct {
	ID       *int32 `json:"id,omitempty"`
	ObjectID *int32 `json:"objectId,omitempty"`
	Name     string `json:"name,omitempty"`
	Removed  bool   `json:"removed,omitempty"`
	IsActive *bool  `json:"isActive,omitempty"`
}

func (p sportPayload) sportID(env feed.Envelope) int32 {
	if p.ID != nil {
		return *p.ID
	}
	if p.ObjectID != nil {
		return *p.ObjectID
	}
	if env.ObjectID != nil {
		return int32(*env.ObjectID)
	}
	if env.SportID != nil {
		return *env.SportID
	}
	return 0
}

type regionPayload struct {
	ID       *int32 `json:"id,omitempty"`
	ObjectID *int32 `json:"objectId,omitempty"`
	SportID  *int32 `json:"sportId,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (p regionPayload) regionID(env feed.Envelope) int32 {
	if p.ID != nil {
		return *p.ID
	}
	if p.ObjectID != nil {
		return *p.ObjectID
	}
	if env.ObjectID != nil {
		return int32(*env.ObjectID)
	}
	return 0
}

type competitionPayload struct {
	ID       *int32 `json:"id,omitempty"`
	ObjectID *int32 `json:"objectId,omitempty"`
	RegionID *int32 `json:"regionId,omitempty"`
	SportID  *int32 `json:"sportId,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (p competitionPayload) competitionID(env feed.Envelope) int32 {
	if p.ID != nil {
		return *p.ID
	}
	if p.ObjectID != nil {
		return *p.ObjectID
	}
	if env.ObjectID != nil {
		return int32(*env.ObjectID)
	}
	return 0
}

type matchPayload struct {
	ID            *int64     `json:"id,omitempty"`
	MatchID       *int64     `json:"matchId,omitempty"`
	ObjectID      *int64     `json:"objectId,omitempty"`
	SportID       *int32     `json:"sportId,omitempty"`
	RegionID      *int32     `json:"regionId,omitempty"`
	CompetitionID *int32     `json:"competitionId,omitempty"`
	Name          string     `json:"name,omitempty"`
	Home          string     `json:"home,omitempty"`
	Away          string     `json:"away,omitempty"`
	StartAt       *time.Time `json:"startAt,omitempty"`
	Date          *time.Time `json:"date,omitempty"`
	IsLive        *bool      `json:"isLive,omitempty"`
	Status        string     `json:"status,omitempty"`
	EventID       string     `json:"eventId,omitempty"`

	Sport       *sportPayload       `json:"sport,omitempty"`
	Region      *regionPayload      `json:"region,omitempty"`
	Competition *competitionPayload `json:"competition,omitempty"`
}

func (p matchPayload) matchID(env feed.Envelope) int64 {
	if p.ID != nil {
		return *p.ID
	}
	if p.MatchID != nil {
		return *p.MatchID
	}
	if p.ObjectID != nil {
		return *p.ObjectID
	}
	if env.MatchID != nil {
		return *env.MatchID
	}
	if env.ObjectID != nil {
		return *env.ObjectID
	}
	return 0
}

func (p matchPayload) startAt() *time.Time {
	if p.StartAt != nil {
		return p.StartAt
	}
	if p.Date != nil {
		return p.Date
	}
	return nil
}

// HandleSport upserts (or soft-deletes) a sports row.
func (h *Handler) HandleSport(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	var p sportPayload
	if err := unmarshalPayload(env.Payload, &p); err != nil {
		return fmt.Errorf("catalog: parse sport: %w", err)
	}
	id := p.sportID(env)
	if id == 0 {
		return errors.New("catalog: sport missing id")
	}
	if p.Removed || (p.IsActive != nil && !*p.IsActive) {
		return h.Repo.SoftDeleteSport(ctx, id)
	}
	return h.Repo.UpsertSport(ctx, Sport{
		ID:       id,
		Name:     strings.TrimSpace(p.Name),
		IsActive: true,
	})
}

// HandleRegion upserts a regions row, auto-creating a stub sports row
// when the parent has not arrived yet.
func (h *Handler) HandleRegion(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	var p regionPayload
	if err := unmarshalPayload(env.Payload, &p); err != nil {
		return fmt.Errorf("catalog: parse region: %w", err)
	}
	id := p.regionID(env)
	if id == 0 {
		return errors.New("catalog: region missing id")
	}
	sportID := int32(0)
	if p.SportID != nil {
		sportID = *p.SportID
	} else if env.SportID != nil {
		sportID = *env.SportID
	}
	if sportID == 0 {
		return errors.New("catalog: region missing sport_id")
	}
	if err := h.ensureSport(ctx, sportID); err != nil {
		return err
	}
	return h.Repo.UpsertRegion(ctx, Region{
		ID:       id,
		SportID:  sportID,
		Name:     strings.TrimSpace(p.Name),
		IsActive: true,
	})
}

// HandleCompetition upserts a competitions row. If sportId is absent it
// is inherited from the parent region.
func (h *Handler) HandleCompetition(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	var p competitionPayload
	if err := unmarshalPayload(env.Payload, &p); err != nil {
		return fmt.Errorf("catalog: parse competition: %w", err)
	}
	id := p.competitionID(env)
	if id == 0 {
		return errors.New("catalog: competition missing id")
	}
	if p.RegionID == nil {
		return errors.New("catalog: competition missing region_id")
	}
	sportID := int32(0)
	if p.SportID != nil {
		sportID = *p.SportID
	} else if env.SportID != nil {
		sportID = *env.SportID
	}
	if sportID == 0 {
		region, ok, err := h.Repo.GetRegion(ctx, *p.RegionID)
		if err != nil {
			return fmt.Errorf("catalog: load region for competition: %w", err)
		}
		if !ok || region.SportID == 0 {
			return fmt.Errorf("catalog: competition %d cannot derive sport_id from region %d", id, *p.RegionID)
		}
		sportID = region.SportID
	}
	if err := h.ensureSport(ctx, sportID); err != nil {
		return err
	}
	return h.Repo.UpsertCompetition(ctx, Competition{
		ID:       id,
		RegionID: *p.RegionID,
		SportID:  sportID,
		Name:     strings.TrimSpace(p.Name),
		IsActive: true,
	})
}

// HandleMatch upserts a matches row plus a fixture_changes audit row when
// applicable. Anti-regression blocks lifecycle-status regressions; other
// fields continue to flow through.
func (h *Handler) HandleMatch(ctx context.Context, _ feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	var p matchPayload
	if err := unmarshalPayload(env.Payload, &p); err != nil {
		return fmt.Errorf("catalog: parse match: %w", err)
	}
	id := p.matchID(env)
	if id == 0 {
		return errors.New("catalog: match missing id")
	}

	sportID := int32(0)
	switch {
	case p.SportID != nil:
		sportID = *p.SportID
	case p.Sport != nil:
		sportID = p.Sport.sportID(env)
	case env.SportID != nil:
		sportID = *env.SportID
	}
	if sportID == 0 {
		return errors.New("catalog: match missing sport_id")
	}
	if err := h.ensureSport(ctx, sportID); err != nil {
		return err
	}

	var regionID *int32
	switch {
	case p.RegionID != nil:
		regionID = p.RegionID
	case p.Region != nil:
		rid := p.Region.regionID(env)
		if rid != 0 {
			regionID = &rid
		}
	}
	if regionID != nil && *regionID != 0 {
		name := ""
		if p.Region != nil {
			name = p.Region.Name
		}
		if err := h.Repo.UpsertRegion(ctx, Region{
			ID:       *regionID,
			SportID:  sportID,
			Name:     strings.TrimSpace(name),
			IsActive: true,
		}); err != nil {
			return fmt.Errorf("catalog: upsert region from match: %w", err)
		}
	}

	var competitionID *int32
	switch {
	case p.CompetitionID != nil:
		competitionID = p.CompetitionID
	case p.Competition != nil:
		cid := p.Competition.competitionID(env)
		if cid != 0 {
			competitionID = &cid
		}
	}
	if competitionID != nil && *competitionID != 0 {
		if regionID == nil || *regionID == 0 {
			return fmt.Errorf("catalog: competition %d on match %d has no region_id", *competitionID, id)
		}
		name := ""
		if p.Competition != nil {
			name = p.Competition.Name
		}
		if err := h.Repo.UpsertCompetition(ctx, Competition{
			ID:       *competitionID,
			RegionID: *regionID,
			SportID:  sportID,
			Name:     strings.TrimSpace(name),
			IsActive: true,
		}); err != nil {
			return fmt.Errorf("catalog: upsert competition from match: %w", err)
		}
	}

	incomingStatus := parseStatus(p.Status)
	isLive := false
	if p.IsLive != nil {
		isLive = *p.IsLive
	} else if incomingStatus == StatusLive {
		isLive = true
	}

	prev, hadPrev, err := h.Repo.GetMatch(ctx, id)
	if err != nil {
		return fmt.Errorf("catalog: load match: %w", err)
	}

	eventID := strings.TrimSpace(p.EventID)
	if eventID == "" {
		eventID = strings.TrimSpace(env.EventID)
	}
	if hadPrev && eventID != "" && prev.LastEventID == eventID {
		return nil
	}

	effectiveStatus := prev.Status
	switch {
	case !hadPrev:
		if incomingStatus != StatusUnknown {
			effectiveStatus = incomingStatus
		} else {
			effectiveStatus = StatusNotStarted
		}
	case incomingStatus != StatusUnknown && incomingStatus != prev.Status:
		if isRegression(prev.Status, incomingStatus) {
			h.regressionCount.Add(1)
			if h.Logger != nil {
				h.Logger.AntiRegressionBlocked(AntiRegressionEvent{
					MatchID: id, From: prev.Status, To: incomingStatus,
				})
			}
		} else {
			effectiveStatus = incomingStatus
		}
	}

	next := Match{
		ID:            id,
		SportID:       sportID,
		CompetitionID: competitionID,
		Name:          strings.TrimSpace(p.Name),
		Home:          strings.TrimSpace(p.Home),
		Away:          strings.TrimSpace(p.Away),
		StartAt:       p.startAt(),
		IsLive:        isLive,
		Status:        effectiveStatus,
		LastEventID:   eventID,
	}
	if hadPrev {
		if next.Name == "" {
			next.Name = prev.Name
		}
		if next.Home == "" {
			next.Home = prev.Home
		}
		if next.Away == "" {
			next.Away = prev.Away
		}
		if next.StartAt == nil {
			next.StartAt = prev.StartAt
		}
		if next.CompetitionID == nil {
			next.CompetitionID = prev.CompetitionID
		}
		if p.IsLive == nil && incomingStatus == StatusUnknown {
			next.IsLive = prev.IsLive
		}
	}

	if err := h.Repo.UpsertMatch(ctx, next); err != nil {
		return fmt.Errorf("catalog: upsert match: %w", err)
	}

	if h.Observer != nil && next.Status != prev.Status {
		h.Observer.MatchStatusChanged(ctx, id, prev.Status, next.Status)
	}

	if env.StatusChange && hadPrev {
		oldDiff, newDiff := diffMatch(prev, next)
		if len(newDiff) > 0 {
			if err := h.Repo.InsertFixtureChange(ctx, FixtureChangeRow{
				MatchID:      id,
				ChangeType:   fixtureChangeType(newDiff),
				Old:          oldDiff,
				New:          newDiff,
				RawMessageID: rawID,
			}); err != nil {
				return fmt.Errorf("catalog: insert fixture_change: %w", err)
			}
		}
	}

	return nil
}

func (h *Handler) ensureSport(ctx context.Context, id int32) error {
	if id == 0 {
		return nil
	}
	_, ok, err := h.Repo.GetSport(ctx, id)
	if err != nil {
		return fmt.Errorf("catalog: load sport: %w", err)
	}
	if ok {
		return nil
	}
	return h.Repo.UpsertSport(ctx, Sport{ID: id, Name: "", IsActive: true})
}

func parseStatus(s string) MatchStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "not_started", "notstarted":
		return StatusNotStarted
	case "live":
		return StatusLive
	case "postponed":
		return StatusPostponed
	case "ended":
		return StatusEnded
	case "cancelled", "canceled":
		return StatusCancelled
	case "closed":
		return StatusClosed
	default:
		return StatusUnknown
	}
}

func isRegression(from, to MatchStatus) bool {
	fr, ok1 := rankOf(from)
	tr, ok2 := rankOf(to)
	if !ok1 || !ok2 {
		return false
	}
	return tr < fr
}

func diffMatch(prev, next Match) (oldDiff, newDiff map[string]any) {
	oldDiff = map[string]any{}
	newDiff = map[string]any{}
	if !sameTimePtr(prev.StartAt, next.StartAt) {
		oldDiff["start_at"] = timePtrToISO(prev.StartAt)
		newDiff["start_at"] = timePtrToISO(next.StartAt)
	}
	if prev.Status != next.Status {
		oldDiff["status"] = string(prev.Status)
		newDiff["status"] = string(next.Status)
	}
	if prev.Name != next.Name {
		oldDiff["name"] = prev.Name
		newDiff["name"] = next.Name
	}
	if prev.Home != next.Home {
		oldDiff["home"] = prev.Home
		newDiff["home"] = next.Home
	}
	if prev.Away != next.Away {
		oldDiff["away"] = prev.Away
		newDiff["away"] = next.Away
	}
	if prev.IsLive != next.IsLive {
		oldDiff["is_live"] = prev.IsLive
		newDiff["is_live"] = next.IsLive
	}
	if !sameInt32Ptr(prev.CompetitionID, next.CompetitionID) {
		oldDiff["competition_id"] = derefInt32(prev.CompetitionID)
		newDiff["competition_id"] = derefInt32(next.CompetitionID)
	}
	return oldDiff, newDiff
}

func sameTimePtr(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

func sameInt32Ptr(a, b *int32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func derefInt32(p *int32) any {
	if p == nil {
		return nil
	}
	return *p
}

func timePtrToISO(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func fixtureChangeType(diff map[string]any) string {
	if len(diff) == 0 {
		return ""
	}
	keys := make([]string, 0, len(diff))
	for k := range diff {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return strings.Join(keys, ",")
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func unmarshalPayload(b []byte, out any) error {
	if len(b) == 0 {
		return errors.New("empty payload")
	}
	return json.Unmarshal(b, out)
}
