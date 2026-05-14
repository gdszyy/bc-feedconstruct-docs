package settlement

import (
	"encoding/json"
	"strings"
	"time"
)

// outcomeSettlement is one entry inside bet_settlement.outcomes /
// .selections. Result and certainty are permissive so we can accept both
// string-coded ("win") and numeric forms while we firm up live samples.
type outcomeSettlement struct {
	ID             *int32   `json:"id,omitempty"`
	OutcomeID      *int32   `json:"outcomeId,omitempty"`
	Result         string   `json:"result,omitempty"`
	ResultCode     *int     `json:"resultCode,omitempty"`
	Certainty      *int     `json:"certainty,omitempty"`
	VoidFactor     *float64 `json:"voidFactor,omitempty"`
	DeadHeatFactor *float64 `json:"deadHeatFactor,omitempty"`
}

func (o outcomeSettlement) outcomeID() (int32, bool) {
	if o.OutcomeID != nil {
		return *o.OutcomeID, true
	}
	if o.ID != nil {
		return *o.ID, true
	}
	return 0, false
}

// marketSettlement supports the markets[] wrapper carrying one
// marketTypeId / specifier and its settled outcomes.
type marketSettlement struct {
	MarketTypeID *int32              `json:"marketTypeId,omitempty"`
	TypeID       *int32              `json:"typeId,omitempty"`
	Specifier    string              `json:"specifier,omitempty"`
	Outcomes     []outcomeSettlement `json:"outcomes,omitempty"`
	Selections   []outcomeSettlement `json:"selections,omitempty"`
	// Certainty may be carried at market level too.
	Certainty *int `json:"certainty,omitempty"`
}

func (m marketSettlement) marketTypeID() (int32, bool) {
	if m.MarketTypeID != nil {
		return *m.MarketTypeID, true
	}
	if m.TypeID != nil {
		return *m.TypeID, true
	}
	return 0, false
}

func (m marketSettlement) selections() []outcomeSettlement {
	if len(m.Outcomes) > 0 {
		return m.Outcomes
	}
	return m.Selections
}

// betSettlementPayload accepts both the markets[] wrapper and the flat
// single-market shape, mirroring the odds package conventions.
type betSettlementPayload struct {
	MatchID  *int64 `json:"matchId,omitempty"`
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`

	Markets []marketSettlement `json:"markets,omitempty"`

	// Flat single-market form fields.
	MarketTypeID *int32              `json:"marketTypeId,omitempty"`
	TypeID       *int32              `json:"typeId,omitempty"`
	Specifier    string              `json:"specifier,omitempty"`
	Outcomes     []outcomeSettlement `json:"outcomes,omitempty"`
	Selections   []outcomeSettlement `json:"selections,omitempty"`
	Certainty    *int                `json:"certainty,omitempty"`
}

func (p betSettlementPayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func (p betSettlementPayload) flatten() []marketSettlement {
	if len(p.Markets) > 0 {
		return p.Markets
	}
	if p.MarketTypeID == nil && p.TypeID == nil {
		return nil
	}
	return []marketSettlement{{
		MarketTypeID: p.MarketTypeID,
		TypeID:       p.TypeID,
		Specifier:    p.Specifier,
		Outcomes:     p.Outcomes,
		Selections:   p.Selections,
		Certainty:    p.Certainty,
	}}
}

func parseBetSettlement(body []byte) (betSettlementPayload, error) {
	var p betSettlementPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// voidNotificationPayload mirrors FC's VoidNotification frame.
// ObjectType=13 targets a single market; ObjectType=4 targets a whole
// match (every market of that match must be cancelled).
type voidNotificationPayload struct {
	MatchID      *int64     `json:"matchId,omitempty"`
	ID           *int64     `json:"id,omitempty"`
	ObjectID     *int64     `json:"objectId,omitempty"`
	ObjectType   *int       `json:"objectType,omitempty"`
	MarketTypeID *int32     `json:"marketTypeId,omitempty"`
	TypeID       *int32     `json:"typeId,omitempty"`
	Specifier    string     `json:"specifier,omitempty"`
	VoidAction   *int       `json:"voidAction,omitempty"`
	VoidReason   *string    `json:"voidReason,omitempty"`
	Reason       *string    `json:"reason,omitempty"`
	FromDate     *time.Time `json:"fromDate,omitempty"`
	ToDate       *time.Time `json:"toDate,omitempty"`
	FromTS       *time.Time `json:"fromTs,omitempty"`
	ToTS         *time.Time `json:"toTs,omitempty"`
	SupercededBy *int64     `json:"supercededBy,omitempty"`
}

func (p voidNotificationPayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func (p voidNotificationPayload) marketTypeID() *int32 {
	if p.MarketTypeID != nil {
		return p.MarketTypeID
	}
	return p.TypeID
}

func (p voidNotificationPayload) reason() *string {
	if p.VoidReason != nil {
		return p.VoidReason
	}
	return p.Reason
}

func (p voidNotificationPayload) fromTime() *time.Time {
	if p.FromDate != nil {
		return p.FromDate
	}
	return p.FromTS
}

func (p voidNotificationPayload) toTime() *time.Time {
	if p.ToDate != nil {
		return p.ToDate
	}
	return p.ToTS
}

// isMatchLevel reports whether the void targets every market of the
// match. FC encodes this as ObjectType=4 (match) or as a payload that
// omits marketTypeId entirely.
func (p voidNotificationPayload) isMatchLevel() bool {
	if p.ObjectType != nil && *p.ObjectType == 4 {
		return true
	}
	return p.marketTypeID() == nil
}

func parseVoidNotification(body []byte) (voidNotificationPayload, error) {
	var p voidNotificationPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// rollbackPayload describes both rolled_back_settlement and
// rolled_back_cancel deliveries. The target is inferred from the message
// type carried in the dispatcher key when present, with the payload's
// 'target' field as a fallback.
type rollbackPayload struct {
	MatchID      *int64 `json:"matchId,omitempty"`
	ID           *int64 `json:"id,omitempty"`
	ObjectID     *int64 `json:"objectId,omitempty"`
	MarketTypeID *int32 `json:"marketTypeId,omitempty"`
	TypeID       *int32 `json:"typeId,omitempty"`
	Specifier    string `json:"specifier,omitempty"`
	OutcomeID    *int32 `json:"outcomeId,omitempty"`
	Target       string `json:"target,omitempty"` // "settlement" | "cancel"
	VoidAction   *int   `json:"voidAction,omitempty"`
}

func (p rollbackPayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func (p rollbackPayload) marketTypeID() *int32 {
	if p.MarketTypeID != nil {
		return p.MarketTypeID
	}
	return p.TypeID
}

func parseRollback(body []byte) (rollbackPayload, error) {
	var p rollbackPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// normaliseResult maps FC result strings (or 0/1/0.5 codes) onto the
// settlements.result CHECK constraint values. Returns "" when no
// reliable mapping is available, so the caller can short-circuit.
func normaliseResult(raw string, code *int) Result {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "win", "won":
		return ResultWin
	case "lose", "lost":
		return ResultLose
	case "void":
		return ResultVoid
	case "half_win", "halfwin", "half-won":
		return ResultHalfWin
	case "half_lose", "halflose", "half-lost":
		return ResultHalfLose
	}
	if code != nil {
		switch *code {
		case 1:
			return ResultWin
		case 0:
			return ResultLose
		}
	}
	return Result("")
}
