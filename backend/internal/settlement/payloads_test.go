package settlement

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGiven_ResultStrings_When_Normalise_Then_MapsToEnum(t *testing.T) {
	cases := map[string]string{
		"win":       "win",
		"lose":      "lose",
		"LOST":      "lose",
		"void":      "void",
		"refunded":  "void",
		"half_win":  "half_win",
		"half-lose": "half_lose",
	}
	for in, want := range cases {
		got, ok := normaliseResult(in)
		require.True(t, ok, "expected to map %q", in)
		require.Equalf(t, want, got, "input=%q", in)
	}
	_, ok := normaliseResult("weird")
	require.False(t, ok)
}

func TestGiven_FlatSettlement_When_Flatten_Then_OneMarket(t *testing.T) {
	body := []byte(`{"matchId":42,"marketTypeId":1,"specifier":"","outcomes":[
		{"id":1,"result":"win","certainty":1}
	]}`)
	p, err := parseSettlement(body)
	require.NoError(t, err)
	mid, mks, ok := p.flatten()
	require.True(t, ok)
	require.EqualValues(t, 42, mid)
	require.Len(t, mks, 1)
	mt, ok := mks[0].marketTypeID()
	require.True(t, ok)
	require.EqualValues(t, 1, mt)
	require.Len(t, mks[0].outcomes(), 1)
}

func TestGiven_CertaintyDefault_When_Outcome_Then_Returns1(t *testing.T) {
	o := settlementOutcome{}
	require.Equal(t, 1, o.certainty())
	zero := 0
	o.Certainty = &zero
	require.Equal(t, 0, o.certainty())
}

func TestGiven_CancelPayload_When_Parse_Then_ObjectTypeAndVoidAction(t *testing.T) {
	body := []byte(`{"objectType":13,"matchId":42,"marketTypeId":7,"specifier":"","reason":"event_void","voidAction":1}`)
	p, err := parseCancel(body)
	require.NoError(t, err)
	require.Equal(t, 13, p.ObjectType)
	require.NotNil(t, p.MatchID)
	require.EqualValues(t, 42, *p.MatchID)
	require.NotNil(t, p.MarketTypeID)
	require.EqualValues(t, 7, *p.MarketTypeID)
	require.NotNil(t, p.VoidAction)
	require.Equal(t, 1, *p.VoidAction)
}
