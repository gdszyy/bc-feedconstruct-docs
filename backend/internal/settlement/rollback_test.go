package settlement_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/settlement"
)

// 验收 9 — 回滚
//
// Given an existing settlements row for outcome (42,1,"",1)
// When a rollback_bet_settlement message arrives for the same outcome
// Then a rollbacks row is inserted (target='settlement', target_id=...)
//      AND settlements.rolled_back_at is set non-null
//      AND markets row (42,1,"") status reverts from settled to its prior status
func TestGiven_ExistingSettlement_When_RollbackArrives_Then_RollbackRecordedAndMarketReverts(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	// Market was active before the settlement put it in settled state.
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)
	ctx := context.Background()

	// Seed: settle outcome (42,1,"",1).
	settle := `{"matchId":42,"marketTypeId":1,"certainty":1,
		"outcomes":[{"id":1,"result":"win"}]}`
	require.NoError(t, h.HandleBetSettlement(ctx, feed.MsgBetSettlement, envWith(settle), [16]byte{0x01}))

	// Sanity: market is now settled.
	_, _, _, markets := repo.snapshot()
	require.Equal(t, settlement.StatusSettled, markets[marketKey{42, 1, ""}].current)

	// Now roll back.
	rollback := `{"matchId":42,"marketTypeId":1,"outcomeId":1,"target":"settlement"}`
	require.NoError(t, h.HandleRollback(ctx, feed.MsgRollback, envWith(rollback), [16]byte{0x09}))

	settlements, _, rollbacks, markets := repo.snapshot()
	require.Len(t, rollbacks, 1)
	require.Equal(t, settlement.TargetSettlement, rollbacks[0].Target)
	require.EqualValues(t, settlements[0].ID, rollbacks[0].TargetID)

	require.NotNil(t, settlements[0].RolledBackAt, "settlements.rolled_back_at must be set")
	require.Equal(t, settlement.StatusActive, markets[marketKey{42, 1, ""}].current,
		"acceptance 9: market reverts from settled to its prior operational status")
	require.EqualValues(t, 1, h.RollbackCount())
}

// Given an existing cancels row
// When a VoidNotification with VoidAction=2 (unvoid) arrives for the same target
// Then a rollbacks row is inserted (target='cancel')
//      AND cancels.rolled_back_at is set non-null
//      AND the market exits status=cancelled
func TestGiven_ExistingCancel_When_UnvoidArrives_Then_RollbackRecordedAndMarketRecovers(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)
	ctx := context.Background()

	// Seed: cancel market (42,1,"").
	cancel := `{"objectType":13,"matchId":42,"marketTypeId":1,"voidAction":1,"voidReason":"venue"}`
	require.NoError(t, h.HandleBetCancel(ctx, feed.MsgBetCancel, envWith(cancel), [16]byte{0x01}))

	_, cancels, _, markets := repo.snapshot()
	require.Equal(t, settlement.StatusCancelled, markets[marketKey{42, 1, ""}].current)
	require.Nil(t, cancels[0].RolledBackAt)

	// Unvoid via VoidAction=2 routed through HandleBetCancel.
	unvoid := `{"objectType":13,"matchId":42,"marketTypeId":1,"voidAction":2}`
	require.NoError(t, h.HandleBetCancel(ctx, feed.MsgBetCancel, envWith(unvoid), [16]byte{0x02}))

	_, cancels, rollbacks, markets := repo.snapshot()
	require.Len(t, rollbacks, 1)
	require.Equal(t, settlement.TargetCancel, rollbacks[0].Target)
	require.EqualValues(t, cancels[0].ID, rollbacks[0].TargetID)

	require.NotNil(t, cancels[0].RolledBackAt, "cancels.rolled_back_at must be set")
	require.NotEqual(t, settlement.StatusCancelled, markets[marketKey{42, 1, ""}].current,
		"market must exit status=cancelled after unvoid")
	require.Equal(t, settlement.StatusActive, markets[marketKey{42, 1, ""}].current)
}

// Given a rollback message arriving twice
// When both deliveries are processed
// Then rollbacks contains exactly one row (idempotent)
func TestGiven_DuplicateRollback_When_Handled_Then_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)
	ctx := context.Background()

	settle := `{"matchId":42,"marketTypeId":1,"certainty":1,
		"outcomes":[{"id":1,"result":"win"}]}`
	require.NoError(t, h.HandleBetSettlement(ctx, feed.MsgBetSettlement, envWith(settle), [16]byte{0x01}))

	rollback := `{"matchId":42,"marketTypeId":1,"outcomeId":1,"target":"settlement"}`
	rawID := [16]byte{0x09, 0x09, 0x09}
	require.NoError(t, h.HandleRollback(ctx, feed.MsgRollback, envWith(rollback), rawID))
	require.NoError(t, h.HandleRollback(ctx, feed.MsgRollback, envWith(rollback), rawID))

	_, _, rollbacks, _ := repo.snapshot()
	require.Len(t, rollbacks, 1, "acceptance 9: duplicate rollback must collapse to a single row")
	require.EqualValues(t, 1, h.RollbackCount())
	require.EqualValues(t, 1, h.DuplicateRollbacks(),
		"duplicate guard must increment the duplicate counter")
}
