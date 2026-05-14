package subscription

import (
	"encoding/json"
	"strings"
)

// bookingPayload is the permissive view for both book and unbook deliveries.
// FC carries the match id under matchId / id / objectId; the product (live
// vs. prematch) under product / queueType / queue. We accept any of them.
type bookingPayload struct {
	MatchID   *int64 `json:"matchId,omitempty"`
	ID        *int64 `json:"id,omitempty"`
	ObjectID  *int64 `json:"objectId,omitempty"`
	Product   string `json:"product,omitempty"`
	QueueType string `json:"queueType,omitempty"`
	Queue     string `json:"queue,omitempty"`
	Reason    string `json:"reason,omitempty"`
	// Some deliveries flag success/failure inline.
	Success *bool `json:"success,omitempty"`
}

func parseBooking(body []byte) (bookingPayload, error) {
	var p bookingPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func (p bookingPayload) matchID() (int64, bool) {
	if p.MatchID != nil {
		return *p.MatchID, true
	}
	if p.ID != nil {
		return *p.ID, true
	}
	if p.ObjectID != nil {
		return *p.ObjectID, true
	}
	return 0, false
}

func (p bookingPayload) product() string {
	for _, s := range []string{p.Product, p.QueueType, p.Queue} {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "live":
			return "live"
		case "prematch", "pre_match", "pre-match":
			return "prematch"
		}
	}
	return ""
}
