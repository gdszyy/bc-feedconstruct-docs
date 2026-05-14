package settlement

import (
	"encoding/json"
	"strings"
	"time"
)

// settlementOutcome is one outcome inside a bet_settlement delivery.
type settlementOutcome struct {
	ID              *int64   `json:"id,omitempty"`
	OutcomeID       *int64   `json:"outcomeId,omitempty"`
	Result          string   `json:"result,omitempty"`
	Certainty       *int     `json:"certainty,omitempty"`
	VoidFactor      *float64 `json:"voidFactor,omitempty"`
	DeadHeatFactor  *float64 `json:"deadHeatFactor,omitempty"`
}

// settlementMarket is one market inside a bet_settlement delivery.
type settlementMarket struct {
	MarketTypeID *int64              `json:"marketTypeId,omitempty"`
	TypeID       *int64              `json:"typeId,omitempty"`
	Specifier    string              `json:"specifier,omitempty"`
	Outcomes     []settlementOutcome `json:"outcomes,omitempty"`
	Selections   []settlementOutcome `json:"selections,omitempty"`
}

// settlementPayload supports both the flat single-market shape and the
// multi-market markets[] shape.
type settlementPayload struct {
	MatchID  *int64             `json:"matchId,omitempty"`
	ID       *int64             `json:"id,omitempty"`
	ObjectID *int64             `json:"objectId,omitempty"`
	Markets  []settlementMarket `json:"markets,omitempty"`

	// flat single-market form
	MarketTypeID *int64              `json:"marketTypeId,omitempty"`
	TypeID       *int64              `json:"typeId,omitempty"`
	Specifier    string              `json:"specifier,omitempty"`
	Outcomes     []settlementOutcome `json:"outcomes,omitempty"`
	Selections   []settlementOutcome `json:"selections,omitempty"`

	// optional explicit timestamp for the settlement.
	SettledAt *time.Time `json:"settledAt,omitempty"`
}

func parseSettlement(body []byte) (settlementPayload, error) {
	var p settlementPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func (p settlementPayload) flatten() (matchID int64, markets []settlementMarket, ok bool) {
	id, k := pickID(p.MatchID, p.ID, p.ObjectID)
	if !k {
		return 0, nil, false
	}
	if len(p.Markets) > 0 {
		return id, p.Markets, true
	}
	mt := p.MarketTypeID
	if mt == nil {
		mt = p.TypeID
	}
	if mt == nil && len(p.Outcomes) == 0 && len(p.Selections) == 0 {
		return id, nil, true
	}
	return id, []settlementMarket{{
		MarketTypeID: mt,
		Specifier:    p.Specifier,
		Outcomes:     p.Outcomes,
		Selections:   p.Selections,
	}}, true
}

func (m settlementMarket) outcomes() []settlementOutcome {
	if len(m.Outcomes) > 0 {
		return m.Outcomes
	}
	return m.Selections
}

func (m settlementMarket) marketTypeID() (int64, bool) {
	return pickID(m.MarketTypeID, m.TypeID)
}

func (o settlementOutcome) outcomeID() (int64, bool) {
	return pickID(o.OutcomeID, o.ID)
}

func (o settlementOutcome) certainty() int {
	if o.Certainty != nil {
		return *o.Certainty
	}
	return 1 // FC's default is "certain" when the field is omitted
}

// cancelPayload mirrors FeedConstruct's VoidNotification.
// ObjectType: 4 = match, 13 = market, 16 = selection.
type cancelPayload struct {
	ObjectType   int        `json:"objectType,omitempty"`
	ObjectID     *int64     `json:"objectId,omitempty"`
	MatchID      *int64     `json:"matchId,omitempty"`
	MarketTypeID *int64     `json:"marketTypeId,omitempty"`
	TypeID       *int64     `json:"typeId,omitempty"`
	Specifier    string     `json:"specifier,omitempty"`
	VoidAction   *int       `json:"voidAction,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	FromDate     *time.Time `json:"fromDate,omitempty"`
	ToDate       *time.Time `json:"toDate,omitempty"`
	SupercededBy *int64     `json:"supercededBy,omitempty"`
}

func parseCancel(body []byte) (cancelPayload, error) {
	var p cancelPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// rollbackPayload describes rollback_bet_settlement or rollback_cancel deliveries.
type rollbackPayload struct {
	MatchID      *int64 `json:"matchId,omitempty"`
	MarketTypeID *int64 `json:"marketTypeId,omitempty"`
	TypeID       *int64 `json:"typeId,omitempty"`
	Specifier    string `json:"specifier,omitempty"`
	OutcomeID    *int64 `json:"outcomeId,omitempty"`
	Target       string `json:"target,omitempty"` // "settlement" / "cancel"
}

func parseRollback(body []byte) (rollbackPayload, error) {
	var p rollbackPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// pickID returns the first non-nil pointer value.
func pickID(c ...*int64) (int64, bool) {
	for _, v := range c {
		if v != nil {
			return *v, true
		}
	}
	return 0, false
}

// normaliseResult maps incoming FC result strings to the settlements.result
// CHECK constraint enum.
func normaliseResult(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "win":
		return "win", true
	case "lose", "lost":
		return "lose", true
	case "void", "refunded":
		return "void", true
	case "half_win", "halfwin", "half-win":
		return "half_win", true
	case "half_lose", "halflose", "half-lose":
		return "half_lose", true
	}
	return "", false
}
