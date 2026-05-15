package subscription

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// bookPayload accepts the FeedConstruct BookObject (RMQ) and the wider
// PartnerBooking (TCP) shapes simultaneously. Field names follow the
// official documentation:
//
//   - BookObject:    ObjectId, ObjectTypeId, IsLive
//   - PartnerBooking: ObjectId, ObjectTypeId, SportId, RegionId,
//                     CompetitionId, IsLive, IsSubscribed
//
// We accept both camelCase and PascalCase to remain tolerant of
// upstream variations.
type bookPayload struct {
	// IDs — at least one of these must resolve to the match id.
	ObjectID    *int64 `json:"objectId,omitempty"`
	ObjectIDPC  *int64 `json:"ObjectId,omitempty"`
	MatchID     *int64 `json:"matchId,omitempty"`
	ID          *int64 `json:"id,omitempty"`

	// ObjectTypeId — accepted but ignored for routing; the dispatcher
	// already classified this delivery as Book/Unbook via the envelope.
	ObjectTypeID   *int  `json:"objectTypeId,omitempty"`
	ObjectTypeIDPC *int  `json:"ObjectTypeId,omitempty"`

	IsLive       *bool `json:"isLive,omitempty"`
	IsLivePC     *bool `json:"IsLive,omitempty"`

	IsSubscribed   *bool `json:"isSubscribed,omitempty"`
	IsSubscribedPC *bool `json:"IsSubscribed,omitempty"`

	// EventID is preferred when present on the payload itself rather
	// than the envelope.
	EventID string `json:"eventId,omitempty"`
}

// matchID resolves the match id from either the payload or the
// envelope. Returns (0, false) when no id is present.
func (p bookPayload) matchID(env feed.Envelope) (int64, bool) {
	switch {
	case p.MatchID != nil && *p.MatchID != 0:
		return *p.MatchID, true
	case p.ID != nil && *p.ID != 0:
		return *p.ID, true
	case p.ObjectID != nil && *p.ObjectID != 0:
		return *p.ObjectID, true
	case p.ObjectIDPC != nil && *p.ObjectIDPC != 0:
		return *p.ObjectIDPC, true
	case env.MatchID != nil && *env.MatchID != 0:
		return *env.MatchID, true
	case env.ObjectID != nil && *env.ObjectID != 0:
		return *env.ObjectID, true
	}
	return 0, false
}

// product resolves live vs prematch. Defaults to ProductLive when no
// hint is present (BookObject without IsLive is ambiguous; the
// FeedConstruct partner queue defaults to live).
func (p bookPayload) product() Product {
	live := true
	switch {
	case p.IsLive != nil:
		live = *p.IsLive
	case p.IsLivePC != nil:
		live = *p.IsLivePC
	}
	if live {
		return ProductLive
	}
	return ProductPrematch
}

// isSubscribed reports the IsSubscribed flag. The boolean is true when
// the payload explicitly sets the flag, so callers can distinguish an
// implicit Book (no flag, presumed subscribed) from an explicit Unbook.
func (p bookPayload) isSubscribed() (value, present bool) {
	switch {
	case p.IsSubscribed != nil:
		return *p.IsSubscribed, true
	case p.IsSubscribedPC != nil:
		return *p.IsSubscribedPC, true
	}
	return false, false
}

func (p bookPayload) eventID(env feed.Envelope) string {
	if s := strings.TrimSpace(p.EventID); s != "" {
		return s
	}
	return strings.TrimSpace(env.EventID)
}

func parseBookPayload(b []byte) (bookPayload, error) {
	var p bookPayload
	if len(b) == 0 {
		return p, errors.New("empty payload")
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return p, err
	}
	return p, nil
}
