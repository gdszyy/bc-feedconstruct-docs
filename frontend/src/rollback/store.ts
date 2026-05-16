import type {
  BetCancelRolledBackPayload,
  BetSettlementRolledBackPayload,
} from "@/contract/events";

// ---------------------------------------------------------------------------
// M09 — RollbackHistoryStore
//
// Append-only timeline of supplier-driven rollbacks. Keyed by
// (matchId, marketId); match-scope cancel rollbacks live under the special
// MATCH_SCOPE_KEY. Consumers (M14) join this store with M08 stores to
// derive bet-level FSM transitions.
//
// Idempotency (defensive against M10 replays):
//   - settlement rollback: dedup if any prior entry for the same key has
//     target='settlement' AND the same version
//   - cancel rollback: dedup if the LAST entry for the key is target='cancel'
//     and shares the same scope (consecutive duplicates collapse)
// ---------------------------------------------------------------------------

export type RollbackTarget = "settlement" | "cancel";

export type RollbackHistoryEntry =
  | {
      target: "settlement";
      match_id: string;
      market_id: string;
      version: number;
      rolled_back_at: number;
    }
  | {
      target: "cancel";
      match_id: string;
      market_id?: string;
      rolled_back_at: number;
    };

const MATCH_SCOPE_KEY = "";

export class RollbackHistoryStore {
  private readonly chains = new Map<
    string,
    Map<string, RollbackHistoryEntry[]>
  >();
  private readonly listeners = new Set<() => void>();

  recordSettlementRollback(
    p: BetSettlementRolledBackPayload,
    at: number = Date.now(),
  ): boolean {
    const chain = this.chainFor(p.match_id, p.market_id);
    if (
      chain.some(
        (e) => e.target === "settlement" && e.version === p.version,
      )
    ) {
      return false;
    }
    chain.push({
      target: "settlement",
      match_id: p.match_id,
      market_id: p.market_id,
      version: p.version,
      rolled_back_at: at,
    });
    this.notify();
    return true;
  }

  recordCancelRollback(
    p: BetCancelRolledBackPayload,
    at: number = Date.now(),
  ): boolean {
    const scopeKey = p.market_id ?? MATCH_SCOPE_KEY;
    const chain = this.chainFor(p.match_id, scopeKey);
    const last = chain[chain.length - 1];
    if (last && last.target === "cancel") {
      const lastScopeKey = last.market_id ?? MATCH_SCOPE_KEY;
      if (lastScopeKey === scopeKey) return false;
    }
    chain.push({
      target: "cancel",
      match_id: p.match_id,
      market_id: p.market_id,
      rolled_back_at: at,
    });
    this.notify();
    return true;
  }

  selectChain(matchId: string, marketId: string): RollbackHistoryEntry[] {
    const chain = this.chains.get(matchId)?.get(marketId);
    return chain ? chain.map(cloneEntry) : [];
  }

  selectMatchScopeChain(matchId: string): RollbackHistoryEntry[] {
    const chain = this.chains.get(matchId)?.get(MATCH_SCOPE_KEY);
    return chain ? chain.map(cloneEntry) : [];
  }

  selectAllForMatch(matchId: string): RollbackHistoryEntry[] {
    const matchBucket = this.chains.get(matchId);
    if (!matchBucket) return [];
    const merged: RollbackHistoryEntry[] = [];
    for (const chain of matchBucket.values()) {
      for (const e of chain) merged.push(cloneEntry(e));
    }
    merged.sort((a, b) => a.rolled_back_at - b.rolled_back_at);
    return merged;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private chainFor(matchId: string, key: string): RollbackHistoryEntry[] {
    let bucket = this.chains.get(matchId);
    if (!bucket) {
      bucket = new Map();
      this.chains.set(matchId, bucket);
    }
    let chain = bucket.get(key);
    if (!chain) {
      chain = [];
      bucket.set(key, chain);
    }
    return chain;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneEntry(e: RollbackHistoryEntry): RollbackHistoryEntry {
  return { ...e };
}
