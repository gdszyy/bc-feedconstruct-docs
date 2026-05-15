package bets_test

// 验收 15 — 投注单（M13）
//
// 参考文档：
//   docs/07_frontend_architecture/modules/M13_bet_slip.md
//   docs/07_frontend_architecture/04_state_machines.md §6 BetSlip FSM
//   docs/07_frontend_architecture/03_backend_data_contract.md §5
//     POST /api/v1/bet-slip/validate
//     POST /api/v1/bet-slip/place

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bets"
)

func sampleSelections() []bets.Selection {
	return []bets.Selection{
		{Position: 1, MatchID: "m1", MarketID: "1", OutcomeID: "1", LockedOdds: 2.10},
	}
}

func newSlipManager(t *testing.T) (*bets.Manager, *fakeRepo, *fakeOutcomes, *captureLogger) {
	t.Helper()
	repo := newFakeRepo()
	outcomes := newFakeOutcomes()
	ids := &counterIDs{}
	mgr := bets.New(repo, outcomes, ids)
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) }
	logger := &captureLogger{}
	mgr.Logger = logger
	return mgr, repo, outcomes, logger
}

// Given a bet slip with a single Active outcome whose lockedOdds matches the latest cached odds.
// When validate() is called.
// Then it returns ok=true and no priceChanges array.
func TestGiven_StableOdds_When_Validate_Then_OK(t *testing.T) {
	mgr, _, outcomes, logger := newSlipManager(t)
	outcomes.set("m1", "1", "1", bets.OutcomeView{
		MarketActive: true, OutcomeActive: true, CurrentOdds: 2.10,
	})

	resp, err := mgr.Validate(context.Background(), bets.ValidateRequest{
		Selections: sampleSelections(),
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	require.True(t, resp.OK)
	require.Empty(t, resp.PriceChanges)
	require.Empty(t, resp.Unavailable)
	require.Empty(t, resp.Code)
	require.Len(t, logger.validate, 1)
	require.True(t, logger.validate[0].OK)
}

// Given a bet slip whose outcome was suspended after the user added it.
// When validate() is called.
// Then it returns ok=false with code=OUTCOME_UNAVAILABLE and the offending outcomeId is identified.
func TestGiven_SuspendedOutcome_When_Validate_Then_Unavailable(t *testing.T) {
	mgr, _, outcomes, _ := newSlipManager(t)
	outcomes.set("m1", "1", "1", bets.OutcomeView{
		MarketActive: false, OutcomeActive: true, CurrentOdds: 2.10,
	})

	resp, err := mgr.Validate(context.Background(), bets.ValidateRequest{
		Selections: sampleSelections(),
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	require.False(t, resp.OK)
	require.Equal(t, bets.CodeOutcomeUnavailable, resp.Code)
	require.Equal(t, []string{"1"}, resp.Unavailable)
	require.NotEmpty(t, resp.Message)
}

// Given a bet slip whose outcome's odds moved between user view and submission.
// When validate() is called with the now-stale lockedOdds.
// Then it returns ok=false, code=PRICE_CHANGED, and a priceChanges entry with from/to so the UI can route to NeedsReview.
func TestGiven_PriceMoved_When_Validate_Then_NeedsReview(t *testing.T) {
	mgr, _, outcomes, _ := newSlipManager(t)
	outcomes.set("m1", "1", "1", bets.OutcomeView{
		MarketActive: true, OutcomeActive: true, CurrentOdds: 2.25,
	})

	resp, err := mgr.Validate(context.Background(), bets.ValidateRequest{
		Selections: sampleSelections(), // locked at 2.10
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	require.False(t, resp.OK)
	require.Equal(t, bets.CodePriceChanged, resp.Code)
	require.Len(t, resp.PriceChanges, 1)
	require.Equal(t, "1", resp.PriceChanges[0].OutcomeID)
	require.InDelta(t, 2.10, resp.PriceChanges[0].From, 0.0001)
	require.InDelta(t, 2.25, resp.PriceChanges[0].To, 0.0001)
}

// Given a fresh place() request with a never-seen Idempotency-Key.
// When the BFF processes the request.
// Then a new Bet row is persisted in state=Pending and a Pending transition is appended to bet_transitions.
func TestGiven_NewIdempotencyKey_When_Place_Then_BetCreatedAndTransitionAppended(t *testing.T) {
	mgr, repo, _, logger := newSlipManager(t)

	resp, err := mgr.Place(context.Background(), bets.PlaceRequest{
		UserID:         "u1",
		IdempotencyKey: "key-1",
		Selections:     sampleSelections(),
		Stake:          10, Currency: "usd", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	require.False(t, resp.Deduped)
	require.Equal(t, "bet_0001", resp.BetID)
	require.Equal(t, bets.StatePending, resp.State)

	stored, ok, err := repo.GetByID(context.Background(), resp.BetID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "USD", stored.Currency, "currency normalised to upper")
	require.Equal(t, bets.StatePending, stored.State)
	require.Len(t, stored.Selections, 1)
	require.Len(t, stored.Transitions, 1)
	require.Equal(t, bets.StatePending, stored.Transitions[0].To)
	require.Equal(t, bets.State(""), stored.Transitions[0].From)

	require.Len(t, logger.placed, 1)
	require.False(t, logger.placed[0].Deduped)
}

// Given a place() request whose Idempotency-Key matches an existing bet.
// When the BFF re-runs the request.
// Then it returns the original betId without creating a duplicate row (no extra bet_transitions either).
func TestGiven_RepeatedIdempotencyKey_When_Place_Then_DedupedSameBet(t *testing.T) {
	mgr, repo, _, logger := newSlipManager(t)

	first, err := mgr.Place(context.Background(), bets.PlaceRequest{
		UserID: "u1", IdempotencyKey: "key-1",
		Selections: sampleSelections(),
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)

	second, err := mgr.Place(context.Background(), bets.PlaceRequest{
		UserID: "u1", IdempotencyKey: "key-1",
		Selections: sampleSelections(),
		Stake:      10, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.NoError(t, err)
	require.Equal(t, first.BetID, second.BetID)
	require.True(t, second.Deduped)
	require.Equal(t, 1, repo.transitionCount(first.BetID),
		"dedup must NOT append a second transition")

	require.Len(t, logger.placed, 2)
	require.False(t, logger.placed[0].Deduped)
	require.True(t, logger.placed[1].Deduped)
}

// Given a bet slip with a stake that exceeds the configured per-bet limit.
// When place() is called.
// Then it returns code=STAKE_EXCEEDS_LIMIT and no Bet row is written.
func TestGiven_StakeOverLimit_When_Place_Then_Rejected(t *testing.T) {
	mgr, repo, _, _ := newSlipManager(t)
	mgr.Limits.MaxStake = 100

	_, err := mgr.Place(context.Background(), bets.PlaceRequest{
		UserID: "u1", IdempotencyKey: "key-over",
		Selections: sampleSelections(),
		Stake:      150, Currency: "USD", BetType: bets.BetTypeSingle,
	})
	require.Error(t, err)
	var pe *bets.PlaceError
	require.True(t, errors.As(err, &pe))
	require.Equal(t, bets.CodeStakeOverLimit, pe.Code)

	out, err := repo.List(context.Background(), bets.ListFilter{UserID: "u1"})
	require.NoError(t, err)
	require.Empty(t, out, "no bet row must be written on stake-limit rejection")
}
