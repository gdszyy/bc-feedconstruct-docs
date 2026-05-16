import type { BetType } from "@/betSlip/store";

// ---------------------------------------------------------------------------
// M14 — MyBetsStore
//
// Server-authoritative view of the user's bets. M13 hands a pending bet off
// via seedPending() at the moment of place(); REST snapshots seed history
// for already-existing bets; realtime events (bet.accepted / bet.rejected /
// bet.state_changed) drive transitions.
//
// FSM (see docs/07_frontend_architecture/04_state_machines.md §4):
//
//   pending ─bet.accepted──▶ accepted ─state_changed──▶ settled
//        ↘ bet.rejected ↘                          ▲
//                       rejected                   │
//                                          state_changed (rollback)
//                                                  │
//                       cancelled ◀─state_changed──┘
//
// The server is authoritative: the client only applies `state_changed` when
// `from === current state`. This single guard covers replay, out-of-order
// delivery, and stale events (mirrors M08/M09 idempotency).
// ---------------------------------------------------------------------------

export type BetState = "pending" | "accepted" | "rejected" | "settled" | "cancelled";

export interface MyBetSelection {
  match_id: string;
  market_id: string;
  outcome_id: string;
  odds: number;
}

export interface BetReason {
  code: string;
  message: string;
}

export interface BetTransition {
  from: BetState | "-";
  to: BetState;
  at: string;
  reason?: BetReason;
}

export interface MyBet {
  bet_id: string;
  user_id: string;
  stake: number;
  currency: string;
  bet_type: BetType;
  selections: MyBetSelection[];
  state: BetState;
  placed_at: string;
  accepted_at?: string;
  history: BetTransition[];
}

export interface SeedPendingPayload {
  bet_id: string;
  user_id: string;
  stake: number;
  currency: string;
  bet_type: BetType;
  selections: MyBetSelection[];
  placed_at: string;
}

export interface AcceptedPayload {
  bet_id: string;
  accepted_at: string;
  accepted_odds?: number;
}

export interface RejectedPayload {
  bet_id: string;
  code: string;
  message: string;
  at: string;
}

export interface StateChangedPayload {
  bet_id: string;
  from: BetState;
  to: BetState;
  at: string;
  reason?: BetReason;
}

export class MyBetsStore {
  private readonly bets = new Map<string, MyBet>();
  private readonly listeners = new Set<() => void>();

  // -------- Selectors --------

  selectById(betId: string): MyBet | undefined {
    const b = this.bets.get(betId);
    return b ? cloneBet(b) : undefined;
  }

  selectAll(): MyBet[] {
    return Array.from(this.bets.values()).map(cloneBet);
  }

  selectByState(state: BetState): MyBet[] {
    const out: MyBet[] = [];
    for (const b of this.bets.values()) {
      if (b.state === state) out.push(cloneBet(b));
    }
    return out;
  }

  selectHistory(betId: string): BetTransition[] {
    const b = this.bets.get(betId);
    return b ? b.history.map(cloneTransition) : [];
  }

  // -------- Mutations --------

  seedPending(p: SeedPendingPayload): boolean {
    if (this.bets.has(p.bet_id)) return false;
    const bet: MyBet = {
      bet_id: p.bet_id,
      user_id: p.user_id,
      stake: p.stake,
      currency: p.currency,
      bet_type: p.bet_type,
      selections: p.selections.map(cloneSel),
      state: "pending",
      placed_at: p.placed_at,
      history: [{ from: "-", to: "pending", at: p.placed_at }],
    };
    this.bets.set(p.bet_id, bet);
    this.notify();
    return true;
  }

  upsertFromSnapshot(snap: MyBet): boolean {
    const existing = this.bets.get(snap.bet_id);
    if (existing && betsEqual(existing, snap)) return false;
    this.bets.set(snap.bet_id, cloneBet(snap));
    this.notify();
    return true;
  }

  applyAccepted(p: AcceptedPayload): boolean {
    const bet = this.bets.get(p.bet_id);
    if (!bet) return false;
    if (bet.state !== "pending") return false;
    bet.state = "accepted";
    bet.accepted_at = p.accepted_at;
    bet.history.push({ from: "pending", to: "accepted", at: p.accepted_at });
    this.notify();
    return true;
  }

  applyRejected(p: RejectedPayload): boolean {
    const bet = this.bets.get(p.bet_id);
    if (!bet) return false;
    if (bet.state !== "pending") return false;
    bet.state = "rejected";
    bet.history.push({
      from: "pending",
      to: "rejected",
      at: p.at,
      reason: { code: p.code, message: p.message },
    });
    this.notify();
    return true;
  }

  applyStateChanged(p: StateChangedPayload): boolean {
    if (p.from === p.to) return false;
    const bet = this.bets.get(p.bet_id);
    if (!bet) return false;
    if (bet.state !== p.from) return false;
    bet.state = p.to;
    bet.history.push({
      from: p.from,
      to: p.to,
      at: p.at,
      ...(p.reason ? { reason: { ...p.reason } } : {}),
    });
    this.notify();
    return true;
  }

  // -------- Subscription --------

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneSel(s: MyBetSelection): MyBetSelection {
  return {
    match_id: s.match_id,
    market_id: s.market_id,
    outcome_id: s.outcome_id,
    odds: s.odds,
  };
}

function cloneTransition(t: BetTransition): BetTransition {
  const out: BetTransition = { from: t.from, to: t.to, at: t.at };
  if (t.reason) out.reason = { code: t.reason.code, message: t.reason.message };
  return out;
}

function cloneBet(b: MyBet): MyBet {
  return {
    bet_id: b.bet_id,
    user_id: b.user_id,
    stake: b.stake,
    currency: b.currency,
    bet_type: b.bet_type,
    selections: b.selections.map(cloneSel),
    state: b.state,
    placed_at: b.placed_at,
    accepted_at: b.accepted_at,
    history: b.history.map(cloneTransition),
  };
}

function selectionsEqual(a: MyBetSelection[], b: MyBetSelection[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const x = a[i];
    const y = b[i];
    if (
      x.match_id !== y.match_id ||
      x.market_id !== y.market_id ||
      x.outcome_id !== y.outcome_id ||
      x.odds !== y.odds
    ) {
      return false;
    }
  }
  return true;
}

function reasonsEqual(a: BetReason | undefined, b: BetReason | undefined): boolean {
  if (a === b) return true;
  if (!a || !b) return false;
  return a.code === b.code && a.message === b.message;
}

function transitionsEqual(a: BetTransition[], b: BetTransition[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const x = a[i];
    const y = b[i];
    if (x.from !== y.from || x.to !== y.to || x.at !== y.at) return false;
    if (!reasonsEqual(x.reason, y.reason)) return false;
  }
  return true;
}

function betsEqual(a: MyBet, b: MyBet): boolean {
  return (
    a.bet_id === b.bet_id &&
    a.user_id === b.user_id &&
    a.stake === b.stake &&
    a.currency === b.currency &&
    a.bet_type === b.bet_type &&
    a.state === b.state &&
    a.placed_at === b.placed_at &&
    a.accepted_at === b.accepted_at &&
    selectionsEqual(a.selections, b.selections) &&
    transitionsEqual(a.history, b.history)
  );
}
