package settlement_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/settlement"
)

// 验收 8 — 取消（VoidNotification, VoidAction=1）
//
// Given a VoidNotification with VoidAction=1, ObjectType=13 (market),
//       FromDate, ToDate, Reason
// When CancelHandler processes it
// Then a cancels row is inserted carrying void_reason / from_ts / to_ts
//      AND the targeted markets row transitions to status=cancelled
func TestGiven_VoidNotificationVoid_When_Handled_Then_CancelRowAndMarketCancelled(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)

	body := `{"objectType":13,"matchId":42,"marketTypeId":1,"specifier":"",
		"voidAction":1,"voidReason":"venue issue",
		"fromDate":"2026-05-14T10:00:00Z","toDate":"2026-05-14T12:00:00Z"}`
	require.NoError(t, h.HandleBetCancel(context.Background(), feed.MsgBetCancel,
		envWith(body), [16]byte{0x08}))

	_, cancels, _, markets := repo.snapshot()
	require.Len(t, cancels, 1)
	c := cancels[0]
	require.EqualValues(t, 42, c.MatchID)
	require.NotNil(t, c.MarketTypeID)
	require.EqualValues(t, 1, *c.MarketTypeID)
	require.NotNil(t, c.VoidReason)
	require.Equal(t, "venue issue", *c.VoidReason)
	require.NotNil(t, c.FromTS)
	require.Equal(t, time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC), c.FromTS.UTC())
	require.NotNil(t, c.ToTS)
	require.Equal(t, time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC), c.ToTS.UTC())
	require.EqualValues(t, settlement.VoidActionVoid, c.VoidAction)

	require.Equal(t, settlement.StatusCancelled, markets[marketKey{42, 1, ""}].current,
		"acceptance 8: targeted market must transition to cancelled")
	require.EqualValues(t, 1, h.CancelCount())
}

// Given a cancel referencing a superceded_by id
// When processed
// Then the superceded_by column links to the original cancel; chain queryable
func TestGiven_CancelWithSupercededBy_When_Handled_Then_LinkPreserved(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)

	// First cancel — establishes a row to be referenced.
	first := `{"objectType":13,"matchId":42,"marketTypeId":1,"voidAction":1,"voidReason":"initial"}`
	require.NoError(t, h.HandleBetCancel(context.Background(), feed.MsgBetCancel, envWith(first), [16]byte{0x01}))

	// Second cancel references the first via supercededBy.
	second := `{"objectType":13,"matchId":42,"marketTypeId":1,"voidAction":1,
		"voidReason":"superseding","supercededBy":1}`
	require.NoError(t, h.HandleBetCancel(context.Background(), feed.MsgBetCancel, envWith(second), [16]byte{0x02}))

	_, cancels, _, _ := repo.snapshot()
	require.Len(t, cancels, 2)
	require.Nil(t, cancels[0].SupercededBy, "the original cancel has no predecessor")
	require.NotNil(t, cancels[1].SupercededBy)
	require.EqualValues(t, 1, *cancels[1].SupercededBy,
		"acceptance 8: supercededBy must link to the original cancel id")
}

// Given a cancel for ObjectType=4 (match) covering all markets of a match
// When processed
// Then every market of that match transitions to status=cancelled
func TestGiven_MatchLevelCancel_When_Handled_Then_AllMarketsCancelled(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	repo.seedMarket(42, 2, "hcap=-1", settlement.StatusActive, settlement.StatusUnknown)
	repo.seedMarket(42, 3, "", settlement.StatusSuspended, settlement.StatusActive)
	// Foreign match, must not be touched.
	repo.seedMatch(99)
	repo.seedMarket(99, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)

	body := `{"objectType":4,"matchId":42,"voidAction":1,"voidReason":"match abandoned"}`
	require.NoError(t, h.HandleBetCancel(context.Background(), feed.MsgBetCancel,
		envWith(body), [16]byte{0x44}))

	_, cancels, _, markets := repo.snapshot()
	require.Len(t, cancels, 1)
	require.Nil(t, cancels[0].MarketTypeID, "match-level cancel has no specific market_type_id")

	require.Equal(t, settlement.StatusCancelled, markets[marketKey{42, 1, ""}].current)
	require.Equal(t, settlement.StatusCancelled, markets[marketKey{42, 2, "hcap=-1"}].current)
	require.Equal(t, settlement.StatusCancelled, markets[marketKey{42, 3, ""}].current)
	require.Equal(t, settlement.StatusActive, markets[marketKey{99, 1, ""}].current,
		"foreign match must not be cancelled")
}
