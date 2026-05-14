package subscription

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGiven_BookingPayload_When_Parse_Then_ProductDetected(t *testing.T) {
	cases := map[string]string{
		`{"matchId":42,"product":"live"}`:           "live",
		`{"id":42,"queueType":"prematch"}`:          "prematch",
		`{"objectId":42,"queue":"pre-match"}`:       "prematch",
		`{"matchId":42}`:                            "",
		`{"matchId":42,"product":"weird"}`:          "",
	}
	for body, want := range cases {
		p, err := parseBooking([]byte(body))
		require.NoError(t, err)
		require.Equalf(t, want, p.product(), "input=%q", body)
	}
}

func TestGiven_BookingPayload_When_MatchID_Then_FallsBackAcrossFields(t *testing.T) {
	cases := []string{
		`{"matchId":42}`,
		`{"id":42}`,
		`{"objectId":42}`,
	}
	for _, body := range cases {
		p, err := parseBooking([]byte(body))
		require.NoError(t, err)
		id, ok := p.matchID()
		require.True(t, ok)
		require.EqualValues(t, 42, id)
	}
	p, err := parseBooking([]byte(`{"product":"live"}`))
	require.NoError(t, err)
	_, ok := p.matchID()
	require.False(t, ok)
}
