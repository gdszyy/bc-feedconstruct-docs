package catalog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGiven_StatusStrings_When_Normalise_Then_MapToEnum(t *testing.T) {
	cases := map[string]string{
		"":           "not_started",
		"Scheduled":  "not_started",
		"PRE_MATCH":  "not_started",
		"live":       "live",
		"in_play":    "live",
		"Finished":   "ended",
		"settled":    "closed",
		"cancelled":  "cancelled",
		"canceled":   "cancelled",
		"postponed":  "postponed",
		"weird":      "not_started",
	}
	for in, want := range cases {
		require.Equalf(t, want, normaliseStatus(in), "input=%q", in)
	}
}

func TestGiven_StatusTransitions_When_Check_Then_NoRegression(t *testing.T) {
	// allowed
	require.True(t, allowsTransition("", "live"))
	require.True(t, allowsTransition("not_started", "live"))
	require.True(t, allowsTransition("live", "ended"))
	require.True(t, allowsTransition("ended", "closed"))
	require.True(t, allowsTransition("ended", "cancelled"))
	require.True(t, allowsTransition("live", "live"))
	// blocked
	require.False(t, allowsTransition("ended", "live"))
	require.False(t, allowsTransition("closed", "live"))
	require.False(t, allowsTransition("cancelled", "live"))
	require.False(t, allowsTransition("ended", "not_started"))
	require.False(t, allowsTransition("live", "not_started"))
}

func TestGiven_MatchPayload_When_Parse_Then_AcceptsIDAndMatchID(t *testing.T) {
	p, err := parseMatch([]byte(`{"matchId":42,"sportId":1,"status":"live","home":"A","away":"B"}`))
	require.NoError(t, err)
	id, ok := pickID(p.MatchID, p.ID, p.ObjectID)
	require.True(t, ok)
	require.EqualValues(t, 42, id)
	require.Equal(t, "live", p.Status)

	p, err = parseMatch([]byte(`{"id":7,"sportId":2}`))
	require.NoError(t, err)
	id, ok = pickID(p.MatchID, p.ID, p.ObjectID)
	require.True(t, ok)
	require.EqualValues(t, 7, id)
}
