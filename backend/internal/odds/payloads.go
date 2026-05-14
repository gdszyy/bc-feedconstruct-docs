package odds

import (
	"encoding/json"
	"strings"
)

// outcomePayload is one entry inside markets[].outcomes / .selections.
type outcomePayload struct {
	ID       *int32   `json:"id,omitempty"`
	Odds     *float64 `json:"odds,omitempty"`
	Active   *bool    `json:"active,omitempty"`
	IsActive *bool    `json:"isActive,omitempty"`
}

// active resolves the active flag respecting both isActive and active
// fields. Defaults to true when neither is set.
func (o outcomePayload) active() bool {
	if o.IsActive != nil {
		return *o.IsActive
	}
	if o.Active != nil {
		return *o.Active
	}
	return true
}

// marketPayload supports both flat and nested forms.
type marketPayload struct {
	MarketTypeID *int32           `json:"marketTypeId,omitempty"`
	TypeID       *int32           `json:"typeId,omitempty"`
	Specifier    string           `json:"specifier,omitempty"`
	GroupID      *int32           `json:"groupId,omitempty"`
	Status       string           `json:"status,omitempty"`
	MarketStatus string           `json:"marketStatus,omitempty"`
	Outcomes     []outcomePayload `json:"outcomes,omitempty"`
	Selections   []outcomePayload `json:"selections,omitempty"`
}

func (m marketPayload) marketTypeID() (int32, bool) {
	if m.MarketTypeID != nil {
		return *m.MarketTypeID, true
	}
	if m.TypeID != nil {
		return *m.TypeID, true
	}
	return 0, false
}

func (m marketPayload) statusString() string {
	return firstNonEmpty(m.Status, m.MarketStatus)
}

func (m marketPayload) outcomes() []outcomePayload {
	if len(m.Outcomes) > 0 {
		return m.Outcomes
	}
	return m.Selections
}

// oddsChangePayload supports both the flat single-market shape and the
// markets[] multi-market shape.
type oddsChangePayload struct {
	MatchID  *int64 `json:"matchId,omitempty"`
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`

	Markets []marketPayload `json:"markets,omitempty"`

	// Flat single-market form fields:
	MarketTypeID *int32           `json:"marketTypeId,omitempty"`
	TypeID       *int32           `json:"typeId,omitempty"`
	Specifier    string           `json:"specifier,omitempty"`
	GroupID      *int32           `json:"groupId,omitempty"`
	Status       string           `json:"status,omitempty"`
	MarketStatus string           `json:"marketStatus,omitempty"`
	Outcomes     []outcomePayload `json:"outcomes,omitempty"`
	Selections   []outcomePayload `json:"selections,omitempty"`
}

func (p oddsChangePayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

// flatten returns the markets[] slice the handler should iterate. The flat
// shape is normalised into a single-element slice; when no market data is
// present (no markets[] AND no flat marketTypeId) the result is an empty
// slice so the handler can short-circuit.
func (p oddsChangePayload) flatten() []marketPayload {
	if len(p.Markets) > 0 {
		return p.Markets
	}
	if p.MarketTypeID == nil && p.TypeID == nil {
		return nil
	}
	return []marketPayload{{
		MarketTypeID: p.MarketTypeID,
		TypeID:       p.TypeID,
		Specifier:    p.Specifier,
		GroupID:      p.GroupID,
		Status:       p.Status,
		MarketStatus: p.MarketStatus,
		Outcomes:     p.Outcomes,
		Selections:   p.Selections,
	}}
}

func parseOddsChange(body []byte) (oddsChangePayload, error) {
	var p oddsChangePayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// betStopPayload mirrors the FC bet_stop body. It targets either a single
// market (marketTypeId + optional specifier), a group (groupId), or the
// whole match when nothing is set.
type betStopPayload struct {
	MatchID      *int64 `json:"matchId,omitempty"`
	ID           *int64 `json:"id,omitempty"`
	ObjectID     *int64 `json:"objectId,omitempty"`
	GroupID      *int32 `json:"groupId,omitempty"`
	MarketTypeID *int32 `json:"marketTypeId,omitempty"`
	TypeID       *int32 `json:"typeId,omitempty"`
	Specifier    string `json:"specifier,omitempty"`
	Status       string `json:"status,omitempty"`
	MarketStatus string `json:"marketStatus,omitempty"`
}

func (p betStopPayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func (p betStopPayload) marketTypeID() *int32 {
	if p.MarketTypeID != nil {
		return p.MarketTypeID
	}
	return p.TypeID
}

func (p betStopPayload) statusString() string {
	return firstNonEmpty(p.Status, p.MarketStatus)
}

func parseBetStop(body []byte) (betStopPayload, error) {
	var p betStopPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// normaliseMarketStatus maps FC strings to the markets.status enum.
// Returns StatusUnknown for empty / unrecognised inputs so callers can
// decide whether to leave the persisted status untouched or fall back
// to a default (active on odds_change, suspended on bet_stop).
func normaliseMarketStatus(raw string) MarketStatus {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "active", "open":
		return StatusActive
	case "suspended", "stopped":
		return StatusSuspended
	case "deactivated", "inactive":
		return StatusDeactivated
	case "settled", "resulted":
		return StatusSettled
	case "cancelled", "canceled":
		return StatusCancelled
	case "handed_over", "handover":
		return StatusHandedOver
	}
	return StatusUnknown
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
