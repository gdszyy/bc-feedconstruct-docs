package odds

import (
	"encoding/json"
	"strings"
)

// outcomePayload mirrors FeedConstruct Selection fields used for live odds.
type outcomePayload struct {
	ID       *int64   `json:"id,omitempty"`
	Odds     *float64 `json:"odds,omitempty"`
	IsActive *bool    `json:"isActive,omitempty"`
	Active   *bool    `json:"active,omitempty"`
}

// marketPayload is one logical market inside a delivery. odds_change for
// a single market arrives flat at the top level; the multi-market form
// carries a "markets" array.
type marketPayload struct {
	MarketTypeID *int64           `json:"marketTypeId,omitempty"`
	TypeID       *int64           `json:"typeId,omitempty"`
	Specifier    string           `json:"specifier,omitempty"`
	GroupID      *int64           `json:"groupId,omitempty"`
	Status       string           `json:"status,omitempty"`       // optional per-market status
	MarketStatus string           `json:"marketStatus,omitempty"` // FC variant name
	Outcomes     []outcomePayload `json:"outcomes,omitempty"`
	Selections   []outcomePayload `json:"selections,omitempty"`
}

// oddsChangePayload is the permissive top-level view.
type oddsChangePayload struct {
	MatchID      *int64           `json:"matchId,omitempty"`
	ID           *int64           `json:"id,omitempty"`
	ObjectID     *int64           `json:"objectId,omitempty"`
	SportID      *int64           `json:"sportId,omitempty"`
	Status       string           `json:"status,omitempty"`
	MarketStatus string           `json:"marketStatus,omitempty"`
	Markets      []marketPayload  `json:"markets,omitempty"`

	// Flat form (single market):
	MarketTypeID *int64           `json:"marketTypeId,omitempty"`
	TypeID       *int64           `json:"typeId,omitempty"`
	Specifier    string           `json:"specifier,omitempty"`
	GroupID      *int64           `json:"groupId,omitempty"`
	Outcomes     []outcomePayload `json:"outcomes,omitempty"`
	Selections   []outcomePayload `json:"selections,omitempty"`
}

// flatten converts the flat form into a markets slice of length 1 (or 0
// when no market data is present). Returns the canonical match id too.
func (p oddsChangePayload) flatten() (matchID int64, markets []marketPayload, ok bool) {
	id, k := pickID(p.MatchID, p.ID, p.ObjectID)
	if !k {
		return 0, nil, false
	}
	if len(p.Markets) > 0 {
		return id, p.Markets, true
	}
	if p.MarketTypeID == nil && p.TypeID == nil {
		return id, nil, true
	}
	mt := p.MarketTypeID
	if mt == nil {
		mt = p.TypeID
	}
	m := marketPayload{
		MarketTypeID: mt,
		Specifier:    p.Specifier,
		GroupID:      p.GroupID,
		Status:       firstNonEmpty(p.Status, p.MarketStatus),
		Outcomes:     p.Outcomes,
		Selections:   p.Selections,
	}
	return id, []marketPayload{m}, true
}

// betStopPayload is the bet_stop body. FC sends either a top-level
// marketStatus + optional groupId / marketTypeId targeting, or a flat
// per-market entry carried via the same odds_change envelope.
type betStopPayload struct {
	MatchID      *int64 `json:"matchId,omitempty"`
	ID           *int64 `json:"id,omitempty"`
	ObjectID     *int64 `json:"objectId,omitempty"`
	GroupID      *int64 `json:"groupId,omitempty"`
	MarketTypeID *int64 `json:"marketTypeId,omitempty"`
	TypeID       *int64 `json:"typeId,omitempty"`
	Specifier    string `json:"specifier,omitempty"`
	Status       string `json:"status,omitempty"`
	MarketStatus string `json:"marketStatus,omitempty"`
}

func parseOddsChange(body []byte) (oddsChangePayload, error) {
	var p oddsChangePayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func parseBetStop(body []byte) (betStopPayload, error) {
	var p betStopPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// outcomes returns the merged outcomes list (outcomes wins over selections).
func (m marketPayload) outcomes() []outcomePayload {
	if len(m.Outcomes) > 0 {
		return m.Outcomes
	}
	return m.Selections
}

func (m marketPayload) marketTypeID() (int64, bool) {
	return pickID(m.MarketTypeID, m.TypeID)
}

func (m marketPayload) statusString() string {
	return firstNonEmpty(m.Status, m.MarketStatus)
}

// pickID returns the first non-nil id.
func pickID(c ...*int64) (int64, bool) {
	for _, v := range c {
		if v != nil {
			return *v, true
		}
	}
	return 0, false
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// normaliseMarketStatus maps incoming strings to the markets.status enum.
// Empty / unknown strings yield "" so callers can decide whether to leave
// the current status untouched.
func normaliseMarketStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "active", "open":
		return "active"
	case "suspended", "stopped":
		return "suspended"
	case "deactivated", "inactive":
		return "deactivated"
	case "settled", "resulted":
		return "settled"
	case "cancelled", "canceled":
		return "cancelled"
	case "handed_over", "handover":
		return "handed_over"
	}
	return ""
}

// statusRank encodes market status precedence (acceptance #12 market-level).
// Higher rank wins. Terminal states (settled / cancelled / handed_over) sit
// above active/suspended/deactivated; cancelled tops them all so a late
// cancel can override a prior settlement.
func statusRank(status string) int {
	switch status {
	case "active":
		return 1
	case "suspended":
		return 2
	case "deactivated":
		return 3
	case "settled":
		return 10
	case "handed_over":
		return 11
	case "cancelled":
		return 20
	}
	return 0
}

// allowsTransition reports whether `to` may overwrite `from`.
func allowsTransition(from, to string) bool {
	if from == "" {
		return true
	}
	if to == "" {
		return false
	}
	return statusRank(to) >= statusRank(from)
}

// outcomeActive resolves the active flag respecting both `isActive` and
// `active` fields. Defaults to true when neither is set.
func (o outcomePayload) active() bool {
	if o.IsActive != nil {
		return *o.IsActive
	}
	if o.Active != nil {
		return *o.Active
	}
	return true
}
