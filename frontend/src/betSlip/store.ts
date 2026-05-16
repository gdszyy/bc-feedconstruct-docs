import type { OddsChangedPayload } from "@/contract/events";
import type { BetSelection } from "@/contract/rest";

// ---------------------------------------------------------------------------
// M13 — BetSlipStore
//
// FSM (see docs/07_frontend_architecture/04_state_machines.md §6):
//
//   empty → editing → ready → submitting → submitted
//                 ⇄ needs_review (price drift)
//
// The store is a pure data model for the slip. REST/WS adapters
// (betSlipValidator, betSlipSubmitter, priceChangeWatcher,
// availabilityWatcher) wrap it. Disallowed transitions are dropped silently
// — listeners fire only on real state changes (mirrors M08/M09/M11).
//
// Idempotency notes:
//   - addSelection rejects a duplicate (match_id, market_id, outcome_id)
//   - setStake/setBetType/setCurrency with identical values are no-ops
//   - applyOddsChange only acts on 'ready' and only when at least one
//     outcome's odds drift from its captured locked_odds
//   - place/markAccepted/markRejected are gated on the appropriate state
//   - discard is a no-op when the slip is already at the default baseline
// ---------------------------------------------------------------------------

export type BetType = "single" | "multiple" | "system";

export type BetSlipState =
  | "empty"
  | "editing"
  | "ready"
  | "needs_review"
  | "submitting"
  | "submitted";

export interface BetSlipSelection {
  match_id: string;
  market_id: string;
  outcome_id: string;
  locked_odds: number;
}

export interface BetSlipReason {
  code: string;
  selection_position?: number;
  message: string;
}

export interface BetSlipPriceChange {
  position: number;
  match_id: string;
  market_id: string;
  outcome_id: string;
  from: number;
  to: number;
}

export interface PlaceBetIntent {
  selections: BetSelection[];
  stake: number;
  currency: string;
  bet_type: BetType;
  idempotency_key: string;
}

export class BetSlipStore {
  private state: BetSlipState = "empty";
  private selections: BetSlipSelection[] = [];
  private stake = 0;
  private currency = "";
  private bet_type: BetType = "single";
  private reasons: BetSlipReason[] = [];
  private priceChanges: BetSlipPriceChange[] = [];
  private idempotency_key?: string;
  private bet_id?: string;
  private readonly listeners = new Set<() => void>();

  // -------- Selectors --------

  selectState(): BetSlipState {
    return this.state;
  }
  selectSelections(): BetSlipSelection[] {
    return this.selections.map(cloneSel);
  }
  selectStake(): number {
    return this.stake;
  }
  selectCurrency(): string {
    return this.currency;
  }
  selectBetType(): BetType {
    return this.bet_type;
  }
  selectReasons(): BetSlipReason[] {
    return this.reasons.map(cloneReason);
  }
  selectPriceChanges(): BetSlipPriceChange[] {
    return this.priceChanges.map(cloneChange);
  }
  selectIdempotencyKey(): string | undefined {
    return this.idempotency_key;
  }
  selectBetId(): string | undefined {
    return this.bet_id;
  }

  // -------- Mutations --------

  addSelection(sel: BetSlipSelection): boolean {
    if (this.state === "submitting" || this.state === "submitted") return false;
    if (this.indexOfSelection(sel.match_id, sel.market_id, sel.outcome_id) >= 0) {
      return false;
    }
    this.selections.push(cloneSel(sel));
    this.demoteAfterEdit();
    if (this.state === "empty") this.state = "editing";
    this.notify();
    return true;
  }

  removeSelection(matchId: string, marketId: string, outcomeId: string): boolean {
    if (this.state === "submitting" || this.state === "submitted") return false;
    const idx = this.indexOfSelection(matchId, marketId, outcomeId);
    if (idx < 0) return false;
    this.selections.splice(idx, 1);
    this.priceChanges = this.priceChanges.filter(
      (pc) => !(pc.match_id === matchId && pc.market_id === marketId && pc.outcome_id === outcomeId),
    );
    if (this.selections.length === 0) {
      this.state = "empty";
      this.priceChanges = [];
    } else if (this.state === "ready" || this.state === "needs_review") {
      this.state = "editing";
      this.priceChanges = [];
    }
    this.notify();
    return true;
  }

  setStake(stake: number): boolean {
    if (this.state === "submitting" || this.state === "submitted") return false;
    if (this.stake === stake) return false;
    this.stake = stake;
    this.demoteAfterEdit();
    this.notify();
    return true;
  }

  setBetType(type: BetType): boolean {
    if (this.state === "submitting" || this.state === "submitted") return false;
    if (this.bet_type === type) return false;
    this.bet_type = type;
    this.demoteAfterEdit();
    this.notify();
    return true;
  }

  setCurrency(currency: string): boolean {
    if (this.state === "submitting" || this.state === "submitted") return false;
    if (this.currency === currency) return false;
    this.currency = currency;
    this.demoteAfterEdit();
    this.notify();
    return true;
  }

  markValidated(): boolean {
    if (this.state !== "editing") return false;
    this.state = "ready";
    this.reasons = [];
    this.priceChanges = [];
    this.notify();
    return true;
  }

  markValidationFailed(reasons: BetSlipReason[]): boolean {
    if (
      this.state !== "editing" &&
      this.state !== "ready" &&
      this.state !== "needs_review"
    ) {
      return false;
    }
    this.state = "editing";
    this.reasons = reasons.map(cloneReason);
    this.priceChanges = [];
    this.notify();
    return true;
  }

  applyOddsChange(p: OddsChangedPayload): boolean {
    if (this.state !== "ready") return false;
    const drifts: BetSlipPriceChange[] = [];
    for (const o of p.outcomes) {
      const idx = this.indexOfSelection(p.match_id, p.market_id, o.outcome_id);
      if (idx < 0) continue;
      const sel = this.selections[idx];
      if (sel.locked_odds === o.odds) continue;
      drifts.push({
        position: idx + 1,
        match_id: p.match_id,
        market_id: p.market_id,
        outcome_id: o.outcome_id,
        from: sel.locked_odds,
        to: o.odds,
      });
    }
    if (drifts.length === 0) return false;
    this.state = "needs_review";
    this.priceChanges = drifts;
    this.notify();
    return true;
  }

  acceptPriceChanges(): boolean {
    if (this.state !== "needs_review") return false;
    for (const pc of this.priceChanges) {
      const idx = this.indexOfSelection(pc.match_id, pc.market_id, pc.outcome_id);
      if (idx >= 0) {
        this.selections[idx] = { ...this.selections[idx], locked_odds: pc.to };
      }
    }
    this.priceChanges = [];
    this.state = "editing";
    this.notify();
    return true;
  }

  place(idempotencyKey: string): boolean {
    if (this.state !== "ready") return false;
    this.state = "submitting";
    this.idempotency_key = idempotencyKey;
    this.notify();
    return true;
  }

  markAccepted(p: { bet_id: string }): boolean {
    if (this.state !== "submitting") return false;
    this.state = "submitted";
    this.bet_id = p.bet_id;
    this.notify();
    return true;
  }

  markRejected(p: BetSlipReason): boolean {
    if (this.state !== "submitting") return false;
    this.state = "editing";
    this.reasons = [cloneReason(p)];
    this.idempotency_key = undefined;
    this.notify();
    return true;
  }

  discard(): boolean {
    if (this.isDefault()) return false;
    this.state = "empty";
    this.selections = [];
    this.stake = 0;
    this.currency = "";
    this.bet_type = "single";
    this.reasons = [];
    this.priceChanges = [];
    this.idempotency_key = undefined;
    this.bet_id = undefined;
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

  // -------- Internals --------

  private indexOfSelection(matchId: string, marketId: string, outcomeId: string): number {
    return this.selections.findIndex(
      (s) =>
        s.match_id === matchId &&
        s.market_id === marketId &&
        s.outcome_id === outcomeId,
    );
  }

  private demoteAfterEdit(): void {
    if (this.state === "ready" || this.state === "needs_review") {
      this.state = "editing";
      this.priceChanges = [];
    }
  }

  private isDefault(): boolean {
    return (
      this.state === "empty" &&
      this.selections.length === 0 &&
      this.stake === 0 &&
      this.currency === "" &&
      this.bet_type === "single" &&
      this.reasons.length === 0 &&
      this.priceChanges.length === 0 &&
      this.idempotency_key === undefined &&
      this.bet_id === undefined
    );
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneSel(s: BetSlipSelection): BetSlipSelection {
  return {
    match_id: s.match_id,
    market_id: s.market_id,
    outcome_id: s.outcome_id,
    locked_odds: s.locked_odds,
  };
}

function cloneReason(r: BetSlipReason): BetSlipReason {
  const out: BetSlipReason = { code: r.code, message: r.message };
  if (r.selection_position !== undefined) out.selection_position = r.selection_position;
  return out;
}

function cloneChange(c: BetSlipPriceChange): BetSlipPriceChange {
  return {
    position: c.position,
    match_id: c.match_id,
    market_id: c.market_id,
    outcome_id: c.outcome_id,
    from: c.from,
    to: c.to,
  };
}
