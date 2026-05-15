package bff_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bets"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bff"
)

// ── stub manager ────────────────────────────────────────────────────

type stubManager struct {
	validateResp bets.ValidateResponse
	placeResp    bets.PlaceResponse
	placeErr     error
	getBet       *bets.Bet
	listBets     []*bets.Bet

	placeCalls atomic.Int64
}

func (s *stubManager) Validate(_ context.Context, _ bets.ValidateRequest) (bets.ValidateResponse, error) {
	return s.validateResp, nil
}

func (s *stubManager) Place(_ context.Context, _ bets.PlaceRequest) (bets.PlaceResponse, error) {
	s.placeCalls.Add(1)
	if s.placeErr != nil {
		return bets.PlaceResponse{}, s.placeErr
	}
	return s.placeResp, nil
}

func (s *stubManager) Get(_ context.Context, betID string) (*bets.Bet, bool, error) {
	if s.getBet != nil && s.getBet.ID == betID {
		return s.getBet, true, nil
	}
	return nil, false, nil
}

func (s *stubManager) List(_ context.Context, _ bets.ListFilter) ([]*bets.Bet, error) {
	return s.listBets, nil
}

func newServer(stub *stubManager) *httptest.Server {
	mux := http.NewServeMux()
	bff.RegisterBetsRoutes(mux, stub)
	return httptest.NewServer(mux)
}

// Given POST /api/v1/bet-slip/place with no Idempotency-Key header
// When the request is processed
// Then 400 IDEMPOTENCY_KEY_REQUIRED is returned and the manager is
//
//	never invoked
func TestGiven_PlaceWithoutIdempotencyKey_When_POST_Then_400AndManagerNotInvoked(t *testing.T) {
	stub := &stubManager{}
	srv := newServer(stub)
	defer srv.Close()

	body := `{"user_id":"u1","stake":10,"currency":"USD","bet_type":"single",
	          "selections":[{"match_id":"m","market_id":"1","outcome_id":"1","locked_odds":2.0}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/bet-slip/place", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	errMap, _ := got["error"].(map[string]any)
	require.Equal(t, bets.CodeIdempotencyMissing, errMap["code"])
	require.Equal(t, int64(0), stub.placeCalls.Load())
}

// Given a valid POST /api/v1/bet-slip/place with Idempotency-Key
// When the manager returns a fresh bet
// Then 201 Created is returned with the bet_id and state in the body
func TestGiven_ValidPlace_When_POST_Then_201WithBetId(t *testing.T) {
	stub := &stubManager{
		placeResp: bets.PlaceResponse{BetID: "bet_0001", State: bets.StatePending},
	}
	srv := newServer(stub)
	defer srv.Close()

	body := `{"user_id":"u1","stake":10,"currency":"USD","bet_type":"single",
	          "selections":[{"match_id":"m","market_id":"1","outcome_id":"1","locked_odds":2.0}]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/bet-slip/place", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "abc-123")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var got bets.PlaceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	require.Equal(t, "bet_0001", got.BetID)
	require.Equal(t, bets.StatePending, got.State)
}

// Given a Place request whose Idempotency-Key matches an existing bet
// When the manager returns Deduped=true
// Then 200 OK is returned (not 201) so the client knows it was a replay
func TestGiven_DedupedPlace_When_POST_Then_200NotCreated(t *testing.T) {
	stub := &stubManager{
		placeResp: bets.PlaceResponse{BetID: "bet_0001", State: bets.StateAccepted, Deduped: true},
	}
	srv := newServer(stub)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/bet-slip/place",
		bytes.NewBufferString(`{"user_id":"u1","stake":10,"currency":"USD","bet_type":"single","selections":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "abc-123")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// Given GET /api/v1/my-bets/{id} for a known bet
// When the manager returns the populated Bet
// Then the JSON shape carries id, state, selections and history in
//
//	chronological order
func TestGiven_KnownBetId_When_GETById_Then_FullShapeReturned(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	stub := &stubManager{
		getBet: &bets.Bet{
			ID: "bet_0001", UserID: "u1", PlacedAt: now,
			Stake: 10, Currency: "USD", BetType: bets.BetTypeSingle,
			State: bets.StateAccepted,
			Selections: []bets.Selection{{
				Position: 1, MatchID: "m1", MarketID: "1", OutcomeID: "1", LockedOdds: 2.10,
			}},
			Transitions: []bets.Transition{
				{At: now, From: "", To: bets.StatePending, Reason: "place"},
				{At: now.Add(time.Minute), From: bets.StatePending, To: bets.StateAccepted, EventID: "e1"},
			},
		},
	}
	srv := newServer(stub)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/my-bets/bet_0001")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	require.Equal(t, "bet_0001", got["id"])
	require.Equal(t, "accepted", got["state"])
	hist := got["history"].([]any)
	require.Len(t, hist, 2)
	require.Equal(t, "pending", hist[0].(map[string]any)["to"])
	require.Equal(t, "accepted", hist[1].(map[string]any)["to"])
	require.Equal(t, "e1", hist[1].(map[string]any)["event_id"])
}

// Given GET /api/v1/my-bets/{id} for an unknown bet
// When the manager returns ok=false
// Then 404 NOT_FOUND is returned
func TestGiven_UnknownBetId_When_GETById_Then_404(t *testing.T) {
	srv := newServer(&stubManager{})
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/my-bets/missing")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Given GET /api/v1/my-bets without user_id and without X-User-Id
// When the request is processed
// Then 400 USER_REQUIRED is returned
func TestGiven_ListWithoutUser_When_GET_Then_400(t *testing.T) {
	srv := newServer(&stubManager{})
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/my-bets")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
