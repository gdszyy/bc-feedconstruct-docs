import type {
  BetCancelAppliedPayload,
  BetCancelRolledBackPayload,
  BetSettlementAppliedPayload,
  BetSettlementRolledBackPayload,
  Certainty,
} from "@/contract/events";

// ---------------------------------------------------------------------------
// SettlementStore — bet_settlement.applied / bet_settlement.rolled_back
//
// Per-market record indexed by (match_id, market_id). Outcomes preserve the
// payload order. Strict version monotonicity (payload.version > existing
// only). A rolled_back event at a higher version deletes the record.
// ---------------------------------------------------------------------------

export type SettlementResult =
  | "win"
  | "lose"
  | "void"
  | "half_win"
  | "half_lose";

export interface SettlementOutcome {
  outcome_id: string;
  result: SettlementResult;
  void_factor?: number;
  dead_heat_factor?: number;
}

export interface MarketSettlementRecord {
  match_id: string;
  market_id: string;
  outcomes: SettlementOutcome[];
  certainty: Certainty;
  version: number;
}

export class SettlementStore {
  private readonly byMatch = new Map<
    string,
    Map<string, MarketSettlementRecord>
  >();
  private readonly listeners = new Set<() => void>();

  applyBetSettlementApplied(p: BetSettlementAppliedPayload): boolean {
    const bucket = this.bucketFor(p.match_id);
    const existing = bucket.get(p.market_id);
    if (existing && p.version <= existing.version) return false;

    bucket.set(p.market_id, {
      match_id: p.match_id,
      market_id: p.market_id,
      outcomes: p.outcomes.map((o) => ({ ...o })),
      certainty: p.certainty,
      version: p.version,
    });
    this.notify();
    return true;
  }

  applyBetSettlementRolledBack(p: BetSettlementRolledBackPayload): boolean {
    const bucket = this.byMatch.get(p.match_id);
    const existing = bucket?.get(p.market_id);
    if (!bucket || !existing) return false;
    if (p.version <= existing.version) return false;

    bucket.delete(p.market_id);
    if (bucket.size === 0) this.byMatch.delete(p.match_id);
    this.notify();
    return true;
  }

  selectSettlement(
    matchId: string,
    marketId: string,
  ): MarketSettlementRecord | undefined {
    const r = this.byMatch.get(matchId)?.get(marketId);
    return r ? cloneSettlement(r) : undefined;
  }

  selectOutcomeSettlement(
    matchId: string,
    marketId: string,
    outcomeId: string,
  ): SettlementOutcome | undefined {
    const r = this.byMatch.get(matchId)?.get(marketId);
    const o = r?.outcomes.find((x) => x.outcome_id === outcomeId);
    return o ? { ...o } : undefined;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private bucketFor(
    matchId: string,
  ): Map<string, MarketSettlementRecord> {
    let b = this.byMatch.get(matchId);
    if (!b) {
      b = new Map();
      this.byMatch.set(matchId, b);
    }
    return b;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneSettlement(r: MarketSettlementRecord): MarketSettlementRecord {
  return {
    match_id: r.match_id,
    market_id: r.market_id,
    outcomes: r.outcomes.map((o) => ({ ...o })),
    certainty: r.certainty,
    version: r.version,
  };
}

// ---------------------------------------------------------------------------
// CancelStore — bet_cancel.applied / bet_cancel.rolled_back
//
// Scope is determined by market_id presence: undefined ⇒ match-scope,
// otherwise market-scope. Match and market scope records coexist on the
// same match.
//
// Cancel events carry no `version` field. Idempotency is by field-equality
// over { void_reason, start_time, end_time, superceded_by }: if the
// incoming event is field-equal to the stored record, the call is a no-op;
// otherwise the record is replaced and listeners notified.
// ---------------------------------------------------------------------------

export interface CancelRecord {
  match_id: string;
  market_id?: string;
  void_reason: string;
  start_time?: string;
  end_time?: string;
  superceded_by?: string;
}

const MATCH_SCOPE_KEY = "";

export class CancelStore {
  private readonly byMatch = new Map<string, Map<string, CancelRecord>>();
  private readonly listeners = new Set<() => void>();

  applyBetCancelApplied(p: BetCancelAppliedPayload): boolean {
    const scopeKey = p.market_id ?? MATCH_SCOPE_KEY;
    const bucket = this.bucketFor(p.match_id);
    const existing = bucket.get(scopeKey);
    const next: CancelRecord = {
      match_id: p.match_id,
      market_id: p.market_id,
      void_reason: p.void_reason,
      start_time: p.start_time,
      end_time: p.end_time,
      superceded_by: p.superceded_by,
    };
    if (existing && sameCancel(existing, next)) return false;
    bucket.set(scopeKey, next);
    this.notify();
    return true;
  }

  applyBetCancelRolledBack(p: BetCancelRolledBackPayload): boolean {
    const bucket = this.byMatch.get(p.match_id);
    if (!bucket) return false;
    const scopeKey = p.market_id ?? MATCH_SCOPE_KEY;
    if (!bucket.delete(scopeKey)) return false;
    if (bucket.size === 0) this.byMatch.delete(p.match_id);
    this.notify();
    return true;
  }

  selectMatchCancelled(matchId: string): CancelRecord | undefined {
    const r = this.byMatch.get(matchId)?.get(MATCH_SCOPE_KEY);
    return r ? { ...r } : undefined;
  }

  selectMarketCancelled(
    matchId: string,
    marketId: string,
  ): CancelRecord | undefined {
    const r = this.byMatch.get(matchId)?.get(marketId);
    return r ? { ...r } : undefined;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private bucketFor(matchId: string): Map<string, CancelRecord> {
    let b = this.byMatch.get(matchId);
    if (!b) {
      b = new Map();
      this.byMatch.set(matchId, b);
    }
    return b;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function sameCancel(a: CancelRecord, b: CancelRecord): boolean {
  return (
    a.void_reason === b.void_reason &&
    a.start_time === b.start_time &&
    a.end_time === b.end_time &&
    a.superceded_by === b.superceded_by
  );
}
