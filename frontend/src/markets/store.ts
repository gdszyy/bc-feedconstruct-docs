import type {
  MarketStatus,
  OddsChangedPayload,
} from "@/contract/events";

export interface OutcomeRecord {
  outcome_id: string;
  odds: number;
  active: boolean;
}

export interface MarketRecord {
  match_id: string;
  market_id: string;
  specifiers: Record<string, string>;
  status: MarketStatus;
  outcomes: OutcomeRecord[];
  version: number;
}

/**
 * Two-level (matchId → marketId) cache for markets and their outcomes.
 *
 * Merge policy:
 * - hydrateMatchMarkets only ADDS unseen markets per match; it never overrides
 *   an entry already populated by a live increment (mirrors M10's snapshot
 *   rule, parallel to CatalogStore / MatchStore).
 * - applyOddsChanged respects per-market version monotonicity. An older
 *   version is silently dropped.
 * - When odds.changed is the first observation for a (match, market), the
 *   record is created with status="active" by default; status transitions
 *   are owned by M06 (market.status_changed) and live elsewhere.
 * - When odds.changed updates an existing market, the outcome list is merged
 *   by outcome_id: incoming entries replace existing ones in-place, new ids
 *   are appended at the tail; outcomes the increment did not mention are
 *   preserved untouched.
 * - specifiers are part of market identity; if a payload omits them the
 *   existing value is preserved.
 *
 * Out of scope (handled in sibling modules):
 * - market.status_changed FSM transitions (M06)
 * - bet_stop frozen / bettable derivations (M07)
 */
export class MarketsStore {
  private readonly markets = new Map<string, Map<string, MarketRecord>>();
  private readonly listeners = new Set<() => void>();

  hydrateMatchMarkets(matchId: string, markets: MarketRecord[]): void {
    const bucket = this.bucketFor(matchId);
    let changed = false;
    for (const m of markets) {
      if (bucket.has(m.market_id)) continue;
      bucket.set(m.market_id, cloneMarket(m));
      changed = true;
    }
    if (changed) this.notify();
  }

  applyOddsChanged(p: OddsChangedPayload): boolean {
    const bucket = this.bucketFor(p.match_id);
    const existing = bucket.get(p.market_id);
    if (existing && p.version <= existing.version) return false;

    if (!existing) {
      bucket.set(p.market_id, {
        match_id: p.match_id,
        market_id: p.market_id,
        specifiers: p.specifiers ? { ...p.specifiers } : {},
        status: "active",
        outcomes: p.outcomes.map((o) => ({ ...o })),
        version: p.version,
      });
      this.notify();
      return true;
    }

    const incoming = new Map(p.outcomes.map((o) => [o.outcome_id, o]));
    const seen = new Set<string>();
    const merged: OutcomeRecord[] = [];
    for (const o of existing.outcomes) {
      const next = incoming.get(o.outcome_id);
      if (next) {
        merged.push({ ...next });
        seen.add(o.outcome_id);
      } else {
        merged.push({ ...o });
      }
    }
    for (const o of p.outcomes) {
      if (seen.has(o.outcome_id)) continue;
      merged.push({ ...o });
    }

    bucket.set(p.market_id, {
      ...existing,
      specifiers: p.specifiers ? { ...p.specifiers } : existing.specifiers,
      outcomes: merged,
      version: p.version,
    });
    this.notify();
    return true;
  }

  getMarket(matchId: string, marketId: string): MarketRecord | undefined {
    const r = this.markets.get(matchId)?.get(marketId);
    return r ? cloneMarket(r) : undefined;
  }

  listMarkets(matchId: string): MarketRecord[] {
    const bucket = this.markets.get(matchId);
    if (!bucket) return [];
    return [...bucket.values()].map(cloneMarket);
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private bucketFor(matchId: string): Map<string, MarketRecord> {
    let bucket = this.markets.get(matchId);
    if (!bucket) {
      bucket = new Map();
      this.markets.set(matchId, bucket);
    }
    return bucket;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneMarket(m: MarketRecord): MarketRecord {
  return {
    match_id: m.match_id,
    market_id: m.market_id,
    specifiers: { ...m.specifiers },
    status: m.status,
    outcomes: m.outcomes.map((o) => ({ ...o })),
    version: m.version,
  };
}

// ---------------------------------------------------------------------------
// Local selectors (status-only dimension; bet_stop / connection live in M07/M01)
// ---------------------------------------------------------------------------

const ODDS_VISIBLE: ReadonlySet<MarketStatus> = new Set([
  "active",
  "suspended",
]);

export function selectDisplayOdds(
  market: MarketRecord,
): OutcomeRecord[] | null {
  if (!ODDS_VISIBLE.has(market.status)) return null;
  return market.outcomes.map((o) => ({ ...o }));
}

export function selectFrozen(market: MarketRecord): boolean {
  return market.status === "suspended";
}
