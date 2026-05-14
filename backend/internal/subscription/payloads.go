package subscription

import (
	"encoding/json"
	"strings"
)

// bookPayload accepts FC Book / Unbook objects. Both shapes carry the
// match identifier and (optionally) a product hint. The same payload
// type is reused for unbook since fields overlap.
type bookPayload struct {
	MatchID  *int64 `json:"matchId,omitempty"`
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`
	Product  string `json:"product,omitempty"`
}

func (p bookPayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

// product resolves the wire product string onto the Product enum. FC
// defaults to live when omitted.
func (p bookPayload) product() Product {
	switch strings.ToLower(strings.TrimSpace(p.Product)) {
	case "prematch", "pre_match":
		return ProductPrematch
	}
	return ProductLive
}

func parseBook(body []byte) (bookPayload, error) {
	var p bookPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// fixturePayload is a thin slice of the match payload that the
// subscription manager needs to react to terminal status changes. The
// catalog handler owns the full upsert; this parser only looks at the
// matchId + status fields.
type fixturePayload struct {
	MatchID  *int64 `json:"matchId,omitempty"`
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`
	Status   string `json:"status,omitempty"`
}

func (p fixturePayload) matchID() (int64, bool) {
	for _, c := range []*int64{p.MatchID, p.ID, p.ObjectID} {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func parseFixture(body []byte) (fixturePayload, error) {
	var p fixturePayload
	err := json.Unmarshal(body, &p)
	return p, err
}

// isTerminal reports whether the (normalised) match status means the
// match is over and any active subscription should be released. Mirrors
// the catalog MatchStatus enum without taking a hard dependency.
func isTerminal(rawStatus string) bool {
	switch strings.ToLower(strings.TrimSpace(rawStatus)) {
	case "ended", "closed", "cancelled", "canceled", "abandoned":
		return true
	}
	return false
}
