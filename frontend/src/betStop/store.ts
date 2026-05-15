import type {
  BetStopAppliedPayload,
  BetStopLiftedPayload,
  MarketStatus,
} from "@/contract/events";
import type { ConnectionState } from "@/realtime/transport";

interface BetStopRecord {
  fullMatch: boolean;
  groups: Set<string>;
}

/**
 * M07 bet_stop overlay.
 *
 * `bet_stop.applied` / `bet_stop.lifted` carry `market_groups: string[]`. An
 * empty array means full-match scope; a non-empty array narrows scope to the
 * listed groups.
 *
 * Semantics decided with product (see PR thread):
 *   - applied([])               ⇒ fullMatch=true, groups irrelevant
 *   - applied([a,b])            ⇒ groups ∪= {a,b} (additive)
 *   - applied([a]) over full    ⇒ fullMatch stays true (no demotion)
 *   - lifted([])                ⇒ fullMatch=false, groups cleared
 *   - lifted([a]) over full     ⇒ NO-OP (full stop must be lifted with empty)
 *   - lifted([a]) over groups   ⇒ groups -= {a}; record dropped when empty
 *
 * The store owns no per-market metadata; consumers supply `(matchId, group)`
 * via `selectStopped`. Pure derivations live alongside the store.
 */
export class BetStopStore {
  private readonly byMatch = new Map<string, BetStopRecord>();
  private readonly listeners = new Set<() => void>();

  applyApplied(p: BetStopAppliedPayload): boolean {
    const existing = this.byMatch.get(p.match_id);

    if (p.market_groups.length === 0) {
      if (existing?.fullMatch) return false;
      this.byMatch.set(p.match_id, {
        fullMatch: true,
        groups: new Set<string>(),
      });
      this.notify();
      return true;
    }

    if (existing?.fullMatch) return false;

    const groups = new Set<string>(existing?.groups ?? []);
    let added = false;
    for (const g of p.market_groups) {
      if (!groups.has(g)) {
        groups.add(g);
        added = true;
      }
    }
    if (!added && existing) return false;

    this.byMatch.set(p.match_id, { fullMatch: false, groups });
    this.notify();
    return true;
  }

  applyLifted(p: BetStopLiftedPayload): boolean {
    const existing = this.byMatch.get(p.match_id);
    if (!existing) return false;

    if (p.market_groups.length === 0) {
      this.byMatch.delete(p.match_id);
      this.notify();
      return true;
    }

    if (existing.fullMatch) return false;

    const groups = new Set<string>(existing.groups);
    let removed = false;
    for (const g of p.market_groups) {
      if (groups.delete(g)) removed = true;
    }
    if (!removed) return false;

    if (groups.size === 0) {
      this.byMatch.delete(p.match_id);
    } else {
      this.byMatch.set(p.match_id, { fullMatch: false, groups });
    }
    this.notify();
    return true;
  }

  selectStopped(matchId: string, group: string): boolean {
    const r = this.byMatch.get(matchId);
    if (!r) return false;
    return r.fullMatch || r.groups.has(group);
  }

  isMatchFullyStopped(matchId: string): boolean {
    return this.byMatch.get(matchId)?.fullMatch === true;
  }

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

// ---------------------------------------------------------------------------
// Pure derivations — M06 FSM + M07 overlay + M01 connection
// ---------------------------------------------------------------------------

export function deriveFrozen(input: {
  status: MarketStatus;
  inBetStop: boolean;
}): boolean {
  return input.status === "suspended" || input.inBetStop;
}

export function deriveBettable(input: {
  status: MarketStatus;
  inBetStop: boolean;
  connection: ConnectionState;
}): boolean {
  return (
    input.status === "active" &&
    !input.inBetStop &&
    input.connection !== "Reconnecting"
  );
}
