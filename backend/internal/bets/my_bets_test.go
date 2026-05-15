package bets_test

// 验收 16 — 我的投注（M14）
//
// 参考文档：
//   docs/07_frontend_architecture/modules/M14_my_bets.md
//   docs/07_frontend_architecture/04_state_machines.md §4 Bet FSM
//   docs/07_frontend_architecture/03_backend_data_contract.md §5
//     GET /api/v1/my-bets?status=...
//     GET /api/v1/my-bets/{betId}

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bets"
)

// placeBet seeds a Pending bet via the manager so the test starts from
// a realistic state, not a hand-crafted DB row.
func placeBet(t *testing.T, mgr *bets.Manager, repo *fakeRepo, userID, key string) string {
	t.Helper()
	resp, err := mgr.Place(context.Background(), bets.PlaceRequest{
		UserID: userID, IdempotencyKey: key,
		Selections: sampleSelections(),
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	_ = repo
	return resp.BetID
}

func float64Ptr(v float64) *float64 { return &v }

// Given a bet was placed (Pending) and the upstream BetGuard acknowledgement event arrives with bet.accepted.
// When the bets manager applies the event.
// Then the Bet transitions Pending → Accepted and the transition row records the source event_id (not the wall-clock alone).
func TestGiven_PendingBet_When_AcceptedEvent_Then_FsmAdvancedAndHistoryAppended(t *testing.T) {
	mgr, repo, _, logger := newSlipManager(t)
	betID := placeBet(t, mgr, repo, "u1", "key-1")

	written, applied, err := mgr.ApplyEvent(context.Background(), bets.EventInput{
		BetID:         betID,
		Kind:          bets.EventBetAccepted,
		EventID:       "evt-accept-1",
		CorrelationID: "corr-1",
		OccurredAt:    time.Date(2026, 5, 14, 12, 1, 0, 0, time.UTC),
		Reason:        "betguard.accepted",
	})
	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, bets.StatePending, written.From)
	require.Equal(t, bets.StateAccepted, written.To)
	require.Equal(t, "evt-accept-1", written.EventID)

	stored, ok, err := repo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, bets.StateAccepted, stored.State)
	require.Len(t, stored.Transitions, 2,
		"history is append-only: initial Pending + Accepted")
	require.Equal(t, bets.StatePending, stored.Transitions[1].From)
	require.Equal(t, bets.StateAccepted, stored.Transitions[1].To)
	require.Equal(t, "evt-accept-1", stored.Transitions[1].EventID)

	require.Len(t, logger.applied, 1)
	require.Equal(t, bets.EventBetAccepted, logger.applied[0].Event)
}

// Given a bet in Accepted state and a bet_settlement.applied event for the matching outcome with result=won.
// When the manager processes the event.
// Then the Bet transitions Accepted → Settled and the payout columns are populated from the event payload.
func TestGiven_AcceptedBet_When_SettlementApplied_Then_BetSettledWithPayout(t *testing.T) {
	mgr, repo, _, _ := newSlipManager(t)
	betID := placeBet(t, mgr, repo, "u1", "key-1")

	_, _, err := mgr.ApplyEvent(context.Background(), bets.EventInput{
		BetID: betID, Kind: bets.EventBetAccepted,
		EventID: "evt-accept",
	})
	require.NoError(t, err)

	_, applied, err := mgr.ApplyEvent(context.Background(), bets.EventInput{
		BetID:          betID,
		Kind:           bets.EventSettlementApplied,
		EventID:        "evt-settle",
		PayoutGross:    float64Ptr(21.0),
		PayoutCurrency: "USD",
	})
	require.NoError(t, err)
	require.True(t, applied)

	stored, _, err := repo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	require.Equal(t, bets.StateSettled, stored.State)
	require.NotNil(t, stored.PayoutGross)
	require.InDelta(t, 21.0, *stored.PayoutGross, 0.0001)
	require.Equal(t, "USD", stored.PayoutCurrency)
	require.Len(t, stored.Transitions, 3, "Pending → Accepted → Settled")
}

// Given a Settled bet whose settlement is rolled back upstream.
// When the manager processes bet_settlement.rolled_back.
// Then a new RolledBack transition is appended (history is NOT mutated in place) and the live state reverts to Accepted.
func TestGiven_SettledBet_When_SettlementRolledBack_Then_AppendedRollbackTransition(t *testing.T) {
	mgr, repo, _, _ := newSlipManager(t)
	betID := placeBet(t, mgr, repo, "u1", "key-1")

	for _, ev := range []bets.EventInput{
		{BetID: betID, Kind: bets.EventBetAccepted, EventID: "evt-accept"},
		{BetID: betID, Kind: bets.EventSettlementApplied, EventID: "evt-settle",
			PayoutGross: float64Ptr(21.0), PayoutCurrency: "USD"},
	} {
		_, _, err := mgr.ApplyEvent(context.Background(), ev)
		require.NoError(t, err)
	}

	historyBefore := repo.transitionCount(betID)

	_, applied, err := mgr.ApplyEvent(context.Background(), bets.EventInput{
		BetID: betID, Kind: bets.EventSettlementRolledBack, EventID: "evt-rb",
	})
	require.NoError(t, err)
	require.True(t, applied)

	stored, _, err := repo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	require.Equal(t, bets.StateAccepted, stored.State, "live state reverts to Accepted")
	require.Equal(t, historyBefore+1, repo.transitionCount(betID),
		"history grows by exactly one row; nothing is mutated in place")
	last := stored.Transitions[len(stored.Transitions)-1]
	require.Equal(t, bets.StateSettled, last.From)
	require.Equal(t, bets.StateAccepted, last.To)
	require.Nil(t, stored.PayoutGross, "rollback clears payout")
	require.Empty(t, stored.PayoutCurrency)
}

// Given the same bet.accepted event is re-delivered (replay scenario)
// When the manager applies it twice
// Then only one Accepted transition exists (idempotent on event_id)
func TestGiven_DuplicateBetEvent_When_AppliedTwice_Then_IdempotentByEventId(t *testing.T) {
	mgr, repo, _, logger := newSlipManager(t)
	betID := placeBet(t, mgr, repo, "u1", "key-1")

	for i := 0; i < 2; i++ {
		_, _, err := mgr.ApplyEvent(context.Background(), bets.EventInput{
			BetID: betID, Kind: bets.EventBetAccepted, EventID: "evt-accept-1",
		})
		require.NoError(t, err)
	}

	require.Equal(t, 2, repo.transitionCount(betID),
		"initial Pending + one Accepted; second delivery is a no-op")

	require.Len(t, logger.applied, 1, "Applied logged exactly once")
	require.Len(t, logger.skipped, 1, "second delivery logged as duplicate")
	require.Equal(t, "duplicate_event", logger.skipped[0].Reason)
}

// Given a user has one Pending and one Settled bet.
// When GET /api/v1/my-bets?status=settled is called.
// Then only the Settled bet is returned, with its full transition history attached.
func TestGiven_MixedBets_When_FilteredList_Then_OnlyMatchingStatusReturned(t *testing.T) {
	mgr, repo, _, _ := newSlipManager(t)

	pendingID := placeBet(t, mgr, repo, "u1", "key-pending")
	settledID := placeBet(t, mgr, repo, "u1", "key-settled")
	for _, ev := range []bets.EventInput{
		{BetID: settledID, Kind: bets.EventBetAccepted, EventID: "evt-a"},
		{BetID: settledID, Kind: bets.EventSettlementApplied, EventID: "evt-s",
			PayoutGross: float64Ptr(13.5), PayoutCurrency: "USD"},
	} {
		_, _, err := mgr.ApplyEvent(context.Background(), ev)
		require.NoError(t, err)
	}

	out, err := mgr.List(context.Background(), bets.ListFilter{
		UserID: "u1", States: []bets.State{bets.StateSettled},
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, settledID, out[0].ID)
	require.Equal(t, bets.StateSettled, out[0].State)
	require.NotNil(t, out[0].PayoutGross)
	// Sanity: pending bet is still in storage but excluded.
	all, err := mgr.List(context.Background(), bets.ListFilter{UserID: "u1"})
	require.NoError(t, err)
	require.Len(t, all, 2)
	for _, b := range all {
		if b.ID == pendingID {
			require.Equal(t, bets.StatePending, b.State)
		}
	}
}

// Given a bet has gone through Pending → Accepted → Settled → RolledBack (Accepted) → re-Settled.
// When GET /api/v1/my-bets/{betId} is called.
// Then the response carries every transition in chronological order with from/to/reason/event_id preserved.
func TestGiven_BetWithFullLifecycle_When_GetById_Then_AppendOnlyHistoryReturned(t *testing.T) {
	mgr, repo, _, _ := newSlipManager(t)
	betID := placeBet(t, mgr, repo, "u1", "key-1")

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	steps := []bets.EventInput{
		{BetID: betID, Kind: bets.EventBetAccepted, EventID: "e1",
			OccurredAt: now.Add(1 * time.Minute), Reason: "accept"},
		{BetID: betID, Kind: bets.EventSettlementApplied, EventID: "e2",
			OccurredAt: now.Add(2 * time.Minute), Reason: "settle",
			PayoutGross: float64Ptr(21.0), PayoutCurrency: "USD"},
		{BetID: betID, Kind: bets.EventSettlementRolledBack, EventID: "e3",
			OccurredAt: now.Add(3 * time.Minute), Reason: "rollback"},
		{BetID: betID, Kind: bets.EventSettlementApplied, EventID: "e4",
			OccurredAt: now.Add(4 * time.Minute), Reason: "re-settle",
			PayoutGross: float64Ptr(0.0), PayoutCurrency: "USD"},
	}
	for _, ev := range steps {
		_, _, err := mgr.ApplyEvent(context.Background(), ev)
		require.NoError(t, err)
	}

	got, ok, err := mgr.Get(context.Background(), betID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, bets.StateSettled, got.State)
	require.Len(t, got.Transitions, 5,
		"initial Pending + 4 events (each appended, none mutated)")

	wantTo := []bets.State{bets.StatePending, bets.StateAccepted, bets.StateSettled, bets.StateAccepted, bets.StateSettled}
	for i, tr := range got.Transitions {
		require.Equal(t, wantTo[i], tr.To, "transition %d: to", i)
	}
	wantEvent := []string{"", "e1", "e2", "e3", "e4"}
	for i, tr := range got.Transitions {
		require.Equal(t, wantEvent[i], tr.EventID, "transition %d: event_id", i)
	}
	for i := 1; i < len(got.Transitions); i++ {
		require.False(t, got.Transitions[i].At.Before(got.Transitions[i-1].At),
			"transitions must be chronological")
	}
}
