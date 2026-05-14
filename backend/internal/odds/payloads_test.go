package odds

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGiven_MarketStatusInputs_When_Normalise_Then_MapToEnum(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"active":      "active",
		"open":        "active",
		"suspended":   "suspended",
		"stopped":     "suspended",
		"deactivated": "deactivated",
		"inactive":    "deactivated",
		"settled":     "settled",
		"resulted":    "settled",
		"cancelled":   "cancelled",
		"canceled":    "cancelled",
		"handed_over": "handed_over",
		"weird":       "",
	}
	for in, want := range cases {
		require.Equalf(t, want, normaliseMarketStatus(in), "input=%q", in)
	}
}

func TestGiven_MarketTransitions_When_Check_Then_NoRegression(t *testing.T) {
	require.True(t, allowsTransition("", "active"))
	require.True(t, allowsTransition("active", "suspended"))
	require.True(t, allowsTransition("suspended", "active") == false)
	require.True(t, allowsTransition("active", "settled"))
	require.True(t, allowsTransition("settled", "cancelled"))
	require.False(t, allowsTransition("settled", "active"))
	require.False(t, allowsTransition("cancelled", "active"))
	require.False(t, allowsTransition("handed_over", "active"))
}

func TestGiven_FlatOddsChange_When_Flatten_Then_OneMarket(t *testing.T) {
	body := []byte(`{"matchId":42,"marketTypeId":1,"specifier":"","outcomes":[{"id":1,"odds":1.85}]}`)
	p, err := parseOddsChange(body)
	require.NoError(t, err)
	mid, markets, ok := p.flatten()
	require.True(t, ok)
	require.EqualValues(t, 42, mid)
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
	mid, markets, ok := p.flatten()
	require.True(t, ok)
	require.EqualValues(t, 42, mid)
	require.Len(t, markets, 2)
	require.Equal(t, "hcap=-1", markets[1].Specifier)
}

func TestGiven_OutcomeFields_When_Active_Then_RespectsBothFlags(t *testing.T) {
	tr := true
	fa := false
	require.True(t, (outcomePayload{}).active(), "default true")
	require.True(t, (outcomePayload{Active: &tr}).active())
	require.False(t, (outcomePayload{Active: &fa}).active())
	require.False(t, (outcomePayload{IsActive: &fa}).active())
}
