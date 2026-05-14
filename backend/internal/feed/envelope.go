package feed

import (
	"encoding/json"
	"strings"
	"time"
)

// MessageType is the canonical internal message-type name used in
// raw_messages.message_type and as the prefix of the internal RabbitMQ
// routing key. Values match docs/08_backend_railway/02_postgres_schema.md.
type MessageType string

const (
	MsgUnknown          MessageType = "unknown"
	MsgCatalogSport     MessageType = "catalog.sport"
	MsgCatalogRegion    MessageType = "catalog.region"
	MsgCatalogComp      MessageType = "catalog.competition"
	MsgCatalogMarketTyp MessageType = "catalog.market_type"
	MsgFixture          MessageType = "fixture"
	MsgFixtureChange    MessageType = "fixture_change"
	MsgOddsChange       MessageType = "odds_change"
	MsgBetStop          MessageType = "bet_stop"
	MsgBetSettlement    MessageType = "bet_settlement"
	MsgBetCancel        MessageType = "bet_cancel"
	MsgRollback         MessageType = "rollback"
	MsgRollbackCancel   MessageType = "rollback_cancel"
	MsgAlive            MessageType = "alive"
	MsgSnapshotComplete MessageType = "snapshot_complete"
	MsgSubscriptionBook MessageType = "subscription.book"
	MsgSubscriptionUnbk MessageType = "subscription.unbook"
)

// Envelope is the minimal subset of FeedConstruct delivery fields the
// ingest layer cares about. Everything else stays in Payload as raw JSON
// so business handlers receive the original content.
type Envelope struct {
	// Type, when present, overrides classification by ObjectType.
	Type string `json:"type,omitempty"`

	// ObjectType is the FeedConstruct ObjectType ID
	// (1=Sport, 2=Region, 3=Competition, 4=Match, 5=MarketType, 13=Market, 16=Selection).
	ObjectType int `json:"objectType,omitempty"`
	ObjectID   *int64 `json:"objectId,omitempty"`

	// EventID is preferred when present; otherwise MatchID is used.
	EventID string `json:"eventId,omitempty"`
	MatchID *int64 `json:"matchId,omitempty"`

	SportID   *int32 `json:"sportId,omitempty"`
	ProductID *int16 `json:"productId,omitempty"`

	// Timestamp is the producer-side time. Accepts RFC3339 strings.
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// VoidAction is populated for VoidNotification messages.
	// 1 = void, 2 = unvoid (rollback of cancel).
	VoidAction *int `json:"voidAction,omitempty"`

	// Settled is true for bet_settlement payloads (FC may carry "settled"
	// or a dedicated SettlementNotification). Permissive on purpose.
	Settled bool `json:"settled,omitempty"`

	// Snapshot is true on the "GetDataSnapshot complete" signal.
	SnapshotComplete bool `json:"snapshotComplete,omitempty"`

	// Status change hint for ObjectType=4 (match) deliveries.
	StatusChange bool `json:"statusChange,omitempty"`

	// Alive is true for keep-alive frames.
	Alive bool `json:"alive,omitempty"`

	// Book/Unbook hints for ObjectType=4 deliveries that wrap a Book result.
	Book   bool `json:"book,omitempty"`
	Unbook bool `json:"unbook,omitempty"`

	// Payload always holds the original JSON body unchanged.
	Payload []byte `json:"-"`
}

// DecodeEnvelope parses jsonBytes into an Envelope. The Payload field is
// always populated even on parse error so the caller can persist the raw
// body for forensic review.
func DecodeEnvelope(jsonBytes []byte) (Envelope, error) {
	env := Envelope{Payload: jsonBytes}
	dec := json.NewDecoder(strings.NewReader(string(jsonBytes)))
	dec.UseNumber() // tolerate numeric fields without losing precision
	// Decode into an intermediate map then re-marshal into Envelope so we
	// don't fail on unrelated extra fields the broker may add.
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return env, err
	}
	// Reuse json.Unmarshal for known fields.
	if err := json.Unmarshal(jsonBytes, &env); err != nil {
		// Preserve Payload, return parse error so caller can flag process_error.
		env.Payload = jsonBytes
		return env, err
	}
	env.Payload = jsonBytes
	return env, nil
}

// Classify maps an envelope (and originating queue) to the canonical
// internal MessageType used for storage and routing.
func Classify(env Envelope, queue string) MessageType {
	if t := strings.TrimSpace(env.Type); t != "" {
		return MessageType(t)
	}
	switch {
	case env.SnapshotComplete:
		return MsgSnapshotComplete
	case env.Alive:
		return MsgAlive
	}

	switch env.ObjectType {
	case 1:
		return MsgCatalogSport
	case 2:
		return MsgCatalogRegion
	case 3:
		return MsgCatalogComp
	case 4:
		switch {
		case env.Book:
			return MsgSubscriptionBook
		case env.Unbook:
			return MsgSubscriptionUnbk
		case env.StatusChange:
			return MsgFixtureChange
		}
		return MsgFixture
	case 5:
		return MsgCatalogMarketTyp
	case 13, 16:
		if env.Settled {
			return MsgBetSettlement
		}
		// FeedConstruct doesn't reliably mark "bet_stop" on the envelope;
		// downstream odds handler decides between odds_change and bet_stop
		// after inspecting the full payload. The audit type stays odds_change
		// for now, which is consistent with the routing-key scheme.
		return MsgOddsChange
	}

	if env.VoidAction != nil {
		switch *env.VoidAction {
		case 1:
			return MsgBetCancel
		case 2:
			return MsgRollbackCancel
		}
	}

	// Unknown — caller routes to dead-letter.
	return MsgUnknown
}

// EventKey returns the preferred event_id string for the row, choosing
// EventID, then MatchID, then ObjectID.
func (e Envelope) EventKey() string {
	if e.EventID != "" {
		return e.EventID
	}
	if e.MatchID != nil {
		return formatInt64(*e.MatchID)
	}
	if e.ObjectID != nil {
		return formatInt64(*e.ObjectID)
	}
	return ""
}

func formatInt64(v int64) string {
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
