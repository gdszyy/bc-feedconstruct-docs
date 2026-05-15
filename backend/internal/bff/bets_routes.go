package bff

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bets"
)

// BetsManager is the slice of *bets.Manager BetsRoutes calls. Defined
// as an interface so HTTP tests can stub without spinning the whole
// domain up.
type BetsManager interface {
	Validate(ctx context.Context, req bets.ValidateRequest) (bets.ValidateResponse, error)
	Place(ctx context.Context, req bets.PlaceRequest) (bets.PlaceResponse, error)
	Get(ctx context.Context, betID string) (*bets.Bet, bool, error)
	List(ctx context.Context, f bets.ListFilter) ([]*bets.Bet, error)
}

// RegisterBetsRoutes wires the M13/M14 endpoints onto mux. The mux
// must be a Go 1.22+ ServeMux (we use {id} path patterns).
func RegisterBetsRoutes(mux *http.ServeMux, mgr BetsManager) {
	h := &betsHandlers{mgr: mgr}
	mux.HandleFunc("POST /api/v1/bet-slip/validate", h.validate)
	mux.HandleFunc("POST /api/v1/bet-slip/place", h.place)
	mux.HandleFunc("GET /api/v1/my-bets", h.list)
	mux.HandleFunc("GET /api/v1/my-bets/{id}", h.getByID)
}

type betsHandlers struct{ mgr BetsManager }

// validateBody mirrors the JSON request shape.
type validateBody struct {
	Selections []apiSelection `json:"selections"`
	Stake      float64        `json:"stake"`
	Currency   string         `json:"currency"`
	BetType    string         `json:"bet_type"`
}

type apiSelection struct {
	Position   int     `json:"position"`
	MatchID    string  `json:"match_id"`
	MarketID   string  `json:"market_id"`
	OutcomeID  string  `json:"outcome_id"`
	LockedOdds float64 `json:"locked_odds"`
}

func (h *betsHandlers) validate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body validateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	resp, err := h.mgr.Validate(r.Context(), bets.ValidateRequest{
		Selections: toDomainSelections(body.Selections),
		Stake:      body.Stake,
		Currency:   body.Currency,
		BetType:    bets.BetType(body.BetType),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

type placeBody struct {
	UserID     string         `json:"user_id"`
	Selections []apiSelection `json:"selections"`
	Stake      float64        `json:"stake"`
	Currency   string         `json:"currency"`
	BetType    string         `json:"bet_type"`
}

func (h *betsHandlers) place(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body placeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	idem := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idem == "" {
		writeError(w, http.StatusBadRequest, bets.CodeIdempotencyMissing, "Idempotency-Key header required")
		return
	}
	userID := body.UserID
	if userID == "" {
		userID = r.Header.Get("X-User-Id")
	}

	resp, err := h.mgr.Place(r.Context(), bets.PlaceRequest{
		UserID:         userID,
		IdempotencyKey: idem,
		Selections:     toDomainSelections(body.Selections),
		Stake:          body.Stake,
		Currency:       body.Currency,
		BetType:        bets.BetType(body.BetType),
	})
	if err != nil {
		var pe *bets.PlaceError
		if errors.As(err, &pe) {
			writeError(w, http.StatusBadRequest, pe.Code, pe.Message)
			return
		}
		if errors.Is(err, bets.ErrIdempotencyKeyMissing) {
			writeError(w, http.StatusBadRequest, bets.CodeIdempotencyMissing, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	status := http.StatusCreated
	if resp.Deduped {
		status = http.StatusOK
	}
	writeJSON(w, status, resp)
}

func (h *betsHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID == "" {
		userID = r.Header.Get("X-User-Id")
	}
	if userID == "" {
		writeError(w, http.StatusBadRequest, "USER_REQUIRED", "user_id query param or X-User-Id header required")
		return
	}
	var states []bets.State
	for _, s := range r.URL.Query()["status"] {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		states = append(states, bets.State(strings.ToLower(s)))
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.mgr.List(r.Context(), bets.ListFilter{
		UserID: userID, States: states, Limit: limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"bets":  toAPIBets(out),
		"count": len(out),
	})
}

func (h *betsHandlers) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "ID_REQUIRED", "bet id required")
		return
	}
	bet, ok, err := h.mgr.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "bet not found")
		return
	}
	writeJSON(w, http.StatusOK, toAPIBet(bet))
}

func toDomainSelections(in []apiSelection) []bets.Selection {
	out := make([]bets.Selection, len(in))
	for i, s := range in {
		out[i] = bets.Selection{
			Position:   s.Position,
			MatchID:    s.MatchID,
			MarketID:   s.MarketID,
			OutcomeID:  s.OutcomeID,
			LockedOdds: s.LockedOdds,
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": msg,
		},
	})
}

// apiBet is the wire shape returned by GET /my-bets and /my-bets/{id}.
type apiBet struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	PlacedAt       time.Time       `json:"placed_at"`
	Stake          float64         `json:"stake"`
	Currency       string          `json:"currency"`
	BetType        string          `json:"bet_type"`
	State          string          `json:"state"`
	Selections     []apiSelection  `json:"selections"`
	History        []apiTransition `json:"history"`
	PayoutGross    *float64        `json:"payout_gross,omitempty"`
	PayoutCurrency string          `json:"payout_currency,omitempty"`
	VoidFactor     *float64        `json:"void_factor,omitempty"`
	DeadHeatFactor *float64        `json:"dead_heat_factor,omitempty"`
}

type apiTransition struct {
	At            time.Time `json:"at"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Reason        string    `json:"reason,omitempty"`
	EventID       string    `json:"event_id,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

func toAPIBets(in []*bets.Bet) []apiBet {
	out := make([]apiBet, len(in))
	for i, b := range in {
		out[i] = toAPIBet(b)
	}
	return out
}

func toAPIBet(b *bets.Bet) apiBet {
	out := apiBet{
		ID: b.ID, UserID: b.UserID, PlacedAt: b.PlacedAt,
		Stake: b.Stake, Currency: b.Currency,
		BetType: string(b.BetType), State: string(b.State),
		PayoutGross: b.PayoutGross, PayoutCurrency: b.PayoutCurrency,
		VoidFactor: b.VoidFactor, DeadHeatFactor: b.DeadHeatFactor,
	}
	out.Selections = make([]apiSelection, len(b.Selections))
	for i, s := range b.Selections {
		out.Selections[i] = apiSelection{
			Position: s.Position, MatchID: s.MatchID, MarketID: s.MarketID,
			OutcomeID: s.OutcomeID, LockedOdds: s.LockedOdds,
		}
	}
	out.History = make([]apiTransition, len(b.Transitions))
	for i, t := range b.Transitions {
		out.History[i] = apiTransition{
			At: t.At, From: string(t.From), To: string(t.To),
			Reason: t.Reason, EventID: t.EventID, CorrelationID: t.CorrelationID,
		}
	}
	return out
}
