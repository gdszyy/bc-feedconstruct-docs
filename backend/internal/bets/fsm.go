package bets

import "fmt"

// EventKind is the upstream event type the FSM understands. It maps
// 1:1 to the WS envelope `type` field documented in
// docs/07_frontend_architecture/03_backend_data_contract.md §3.6 and
// §3.4.
type EventKind string

// Recognised event kinds.
const (
	EventBetAccepted          EventKind = "bet.accepted"
	EventBetRejected          EventKind = "bet.rejected"
	EventSettlementApplied    EventKind = "bet_settlement.applied"
	EventSettlementRolledBack EventKind = "bet_settlement.rolled_back"
	EventCancelApplied        EventKind = "bet_cancel.applied"
	EventCancelRolledBack     EventKind = "bet_cancel.rolled_back"
)

// Apply returns the next state after applying ev to current. The
// boolean is false when the transition is not valid for the FSM
// (caller should drop the event with a "stale" log line, not error).
//
// Rollback events restore the prior state. Because the live state on
// the bets row is the only "current" we know, rollback uses a
// hard-coded inverse: settlement.rolled_back → Accepted (settlement
// always comes from Accepted), cancel.rolled_back → Accepted (cancel
// always comes from Accepted in our supported lifecycle).
func Apply(current State, ev EventKind) (State, bool) {
	switch ev {
	case EventBetAccepted:
		if current == StatePending {
			return StateAccepted, true
		}
	case EventBetRejected:
		if current == StatePending {
			return StateRejected, true
		}
	case EventSettlementApplied:
		if current == StateAccepted {
			return StateSettled, true
		}
	case EventSettlementRolledBack:
		if current == StateSettled {
			return StateAccepted, true
		}
	case EventCancelApplied:
		if current == StateAccepted || current == StateSettled {
			return StateCancelled, true
		}
	case EventCancelRolledBack:
		if current == StateCancelled {
			return StateAccepted, true
		}
	}
	return current, false
}

// String implements fmt.Stringer for log lines.
func (s State) String() string { return string(s) }

// EventKind String impl.
func (e EventKind) String() string { return string(e) }

// FSMRejectError describes a no-op event (applied to a state that
// does not accept it). The Manager uses this to decide between
// "warn + skip" and "promote to error".
type FSMRejectError struct {
	BetID string
	From  State
	Event EventKind
}

// Error implements error.
func (e *FSMRejectError) Error() string {
	return fmt.Sprintf("bets: event %s not valid in state %s for bet %s", e.Event, e.From, e.BetID)
}
