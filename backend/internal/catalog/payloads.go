package catalog

import (
	"encoding/json"
	"strings"
	"time"
)

// sportPayload is the permissive view of an ObjectType=1 delivery.
type sportPayload struct {
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`
	Name     string `json:"name,omitempty"`
	IsActive *bool  `json:"isActive,omitempty"`
}

type regionPayload struct {
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`
	SportID  *int64 `json:"sportId,omitempty"`
	Name     string `json:"name,omitempty"`
	IsActive *bool  `json:"isActive,omitempty"`
}

type competitionPayload struct {
	ID       *int64 `json:"id,omitempty"`
	ObjectID *int64 `json:"objectId,omitempty"`
	SportID  *int64 `json:"sportId,omitempty"`
	RegionID *int64 `json:"regionId,omitempty"`
	Name     string `json:"name,omitempty"`
	IsActive *bool  `json:"isActive,omitempty"`
}

type matchPayload struct {
	ID            *int64     `json:"id,omitempty"`
	MatchID       *int64     `json:"matchId,omitempty"`
	ObjectID      *int64     `json:"objectId,omitempty"`
	SportID       *int64     `json:"sportId,omitempty"`
	CompetitionID *int64     `json:"competitionId,omitempty"`
	Name          string     `json:"name,omitempty"`
	Home          string     `json:"home,omitempty"`
	Away          string     `json:"away,omitempty"`
	StartAt       *time.Time `json:"startAt,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	IsLive        *bool      `json:"isLive,omitempty"`
	Status        string     `json:"status,omitempty"`
}

func parseSport(body []byte) (sportPayload, error) {
	var p sportPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func parseRegion(body []byte) (regionPayload, error) {
	var p regionPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func parseCompetition(body []byte) (competitionPayload, error) {
	var p competitionPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func parseMatch(body []byte) (matchPayload, error) {
	var p matchPayload
	err := json.Unmarshal(body, &p)
	return p, err
}

func pickID(candidates ...*int64) (int64, bool) {
	for _, c := range candidates {
		if c != nil {
			return *c, true
		}
	}
	return 0, false
}

func pickTime(candidates ...*time.Time) *time.Time {
	for _, c := range candidates {
		if c != nil {
			return c
		}
	}
	return nil
}

// normaliseStatus maps various FC strings to one of the matches.status enum
// values. Unknown values fall back to "not_started" so the CHECK constraint
// holds; the caller logs the raw value for forensics.
func normaliseStatus(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "", "scheduled", "not_started", "pre", "pre_match":
		return "not_started"
	case "live", "started", "in_play", "inplay":
		return "live"
	case "ended", "finished":
		return "ended"
	case "closed", "settled":
		return "closed"
	case "cancelled", "canceled":
		return "cancelled"
	case "postponed", "delayed":
		return "postponed"
	}
	return "not_started"
}

// statusRank encodes the no-regression precedence (acceptance #12).
// Higher rank wins. cancelled tops everything because a cancel always
// supersedes earlier states.
func statusRank(status string) int {
	switch status {
	case "not_started":
		return 0
	case "postponed":
		return 1
	case "live":
		return 2
	case "ended":
		return 3
	case "closed":
		return 4
	case "cancelled":
		return 5
	}
	return -1
}

// allowsTransition reports whether a transition from `from` to `to` should
// be applied. Same-or-higher rank wins; lower rank is rejected.
func allowsTransition(from, to string) bool {
	if from == "" {
		return true
	}
	fr := statusRank(from)
	tr := statusRank(to)
	if tr < 0 {
		return false
	}
	return tr >= fr
}
