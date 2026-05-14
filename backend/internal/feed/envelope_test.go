package feed_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

func TestGiven_ObjectType_When_Classify_Then_CanonicalTypeReturned(t *testing.T) {
	cases := []struct {
		name string
		json string
		want feed.MessageType
	}{
		{"sport", `{"objectType":1}`, feed.MsgCatalogSport},
		{"region", `{"objectType":2}`, feed.MsgCatalogRegion},
		{"competition", `{"objectType":3}`, feed.MsgCatalogComp},
		{"match-fixture", `{"objectType":4,"matchId":7}`, feed.MsgFixture},
		{"match-fixture-change", `{"objectType":4,"matchId":7,"statusChange":true}`, feed.MsgFixtureChange},
		{"match-book", `{"objectType":4,"matchId":7,"book":true}`, feed.MsgSubscriptionBook},
		{"market-type", `{"objectType":5}`, feed.MsgCatalogMarketTyp},
		{"market-odds", `{"objectType":13,"matchId":7}`, feed.MsgOddsChange},
		{"market-settled", `{"objectType":13,"matchId":7,"settled":true}`, feed.MsgBetSettlement},
		{"void-cancel", `{"voidAction":1,"objectId":99}`, feed.MsgBetCancel},
		{"void-unvoid", `{"voidAction":2,"objectId":99}`, feed.MsgRollbackCancel},
		{"alive", `{"alive":true}`, feed.MsgAlive},
		{"snapshot", `{"snapshotComplete":true}`, feed.MsgSnapshotComplete},
		{"explicit-type", `{"type":"custom_x"}`, feed.MessageType("custom_x")},
		{"unknown", `{"foo":"bar"}`, feed.MsgUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env, err := feed.DecodeEnvelope([]byte(c.json))
			require.NoError(t, err)
			got := feed.Classify(env, "")
			require.Equal(t, c.want, got)
		})
	}
}

func TestGiven_NonJSON_When_DecodeEnvelope_Then_ErrorAndPayloadPreserved(t *testing.T) {
	body := []byte("not json at all")
	env, err := feed.DecodeEnvelope(body)
	require.Error(t, err)
	require.Equal(t, body, env.Payload, "payload must be preserved for forensic audit")
}

func TestGiven_EnvelopeWithMatchID_When_EventKey_Then_ReturnsMatchID(t *testing.T) {
	env, err := feed.DecodeEnvelope([]byte(`{"objectType":13,"matchId":42}`))
	require.NoError(t, err)
	require.Equal(t, "42", env.EventKey())
}

func TestGiven_EnvelopeWithEventID_When_EventKey_Then_PrefersEventID(t *testing.T) {
	env, err := feed.DecodeEnvelope([]byte(`{"objectType":13,"eventId":"abc","matchId":42}`))
	require.NoError(t, err)
	require.Equal(t, "abc", env.EventKey())
}
