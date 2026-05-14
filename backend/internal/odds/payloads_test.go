package odds

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGiven_MarketStatusStrings_When_Normalise_Then_MapToEnum(t *testing.T) {
	cases := map[string]MarketStatus{
		"":             StatusUnknown,
		"active":       StatusActive,
		"open":         StatusActive,
		"suspended":    StatusSuspended,
		"stopped":      StatusSuspended,
		"deactivated":  StatusDeactivated,
		"inactive":     StatusDeactivated,
		"settled":      StatusSettled,
		"resulted":     StatusSettled,
		"cancelled":    StatusCancelled,
		"canceled":     StatusCancelled,
		"handed_over":  StatusHandedOver,
		"weird":        StatusUnknown,
	}
	for in, want := range cases {
		require.Equalf(t, want, normaliseMarketStatus(in), "input=%q", in)
	}
}

func TestGiven_MarketTransitions_When_Check_Then_NoRegression(t *testing.T) {
	// from "" (unknown) always passes
	require.True(t, allowsTransition("", StatusActive))
	// active -> suspended OK
	require.True(t, allowsTransition(StatusActive, StatusSuspended))
	// suspended -> active is forbidden by rank (suspended=2, active=1)
	require.False(t, allowsTransition(StatusSuspended, StatusActive))
	// active -> settled OK (1 -> 10)
	require.True(t, allowsTransition(StatusActive, StatusSettled))
	// settled -> cancelled OK (10 -> 20)
	require.True(t, allowsTransition(StatusSettled, StatusCancelled))
	// settled -> active rejected (acceptance #12 market-level)
	require.False(t, allowsTransition(StatusSettled, StatusActive))
	// cancelled -> anything below rank 20 rejected
	require.False(t, allowsTransition(StatusCancelled, StatusActive))
	require.False(t, allowsTransition(StatusCancelled, StatusSettled))
	require.False(t, allowsTransition(StatusHandedOver, StatusActive))
	// to "" never passes
	require.False(t, allowsTransition(StatusActive, ""))
}

func TestGiven_FlatOddsChange_When_Flatten_Then_OneMarket(t *testing.T) {
	body := []byte(`{"matchId":42,"marketTypeId":1,"specifier":"",
		"outcomes":[{"id":1,"odds":1.85}]}`)
	p, err := parseOddsChange(body)
	require.NoError(t, err)
	id, ok := p.matchID()
	require.True(t, ok)
	require.EqualValues(t, 42, id)
	markets := p.flatten()
	require.Len(t, markets, 1)
	mt, ok := markets[0].marketTypeID()
	require.True(t, ok)
	require.EqualValues(t, 1, mt)
	require.Len(t, markets[0].outcomes(), 1)
}

func TestGiven_MultiMarketOddsChange_When_Flatten_Then_TakesMarketsArray(t *testing.T) {
	body := []byte(`{"matchId":42,"markets":[
		{"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]},
		{"marketTypeId":2,"specifier":"hcap=-1","outcomes":[{"id":1,"odds":2.10}]}
	]}`)
	p, err := parseOddsChange(body)
	require.NoError(t, err)
	markets := p.flatten()
	require.Len(t, markets, 2)
	require.Equal(t, "hcap=-1", markets[1].Specifier)
}

func TestGiven_OutcomeFields_When_Active_Then_RespectsBothFlags(t *testing.T) {
	tr, fa := true, false
	require.True(t, (outcomePayload{}).active(), "default true")
	require.True(t, (outcomePayload{Active: &tr}).active())
	require.False(t, (outcomePayload{Active: &fa}).active())
	require.False(t, (outcomePayload{IsActive: &fa}).active())
	require.True(t, (outcomePayload{IsActive: &tr, Active: &fa}).active(),
		"isActive wins over active when both are present")
}

func TestGiven_NoMarketTypeID_When_Flatten_Then_EmptyMarkets(t *testing.T) {
	p, err := parseOddsChange([]byte(`{"matchId":42,"outcomes":[]}`))
	require.NoError(t, err)
	require.Empty(t, p.flatten(), "no marketTypeId means no actionable market")
}
