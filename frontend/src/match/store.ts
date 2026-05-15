import type {
  FixtureChangedPayload,
  MatchStatus,
  MatchStatusChangedPayload,
  MatchUpsertedPayload,
} from "@/contract/events";

export interface MatchRecord {
  match_id: string;
  tournament_id: string;
  home_team: string;
  away_team: string;
  scheduled_at: string;
  is_live: boolean;
  status: MatchStatus;
  home_score?: number;
  away_score?: number;
  period?: string;
  version: number;
}

export type MatchRefetcher = (
  matchId: string,
) => Promise<MatchRecord | null | undefined>;

export interface MatchStoreError {
  kind:
    | "fixture_refetch_failed"
    | "fixture_refetch_empty"
    | "fixture_refetch_unconfigured";
  match_id: string;
  cause?: unknown;
}

export type MatchErrorSink = (err: MatchStoreError) => void;

/**
 * Match FSM rank for the regression guard.
 *
 *   not_started < live (≈ suspended) < ended < closed
 *   cancelled / abandoned are absorbing — see isAbsorbing().
 *
 * The wire-level MatchStatus union (contract/events.ts) also includes
 * "suspended" which is reserved for Market FSM but may appear on a match
 * status_changed payload; treat it at the same rank as live so that
 * live ↔ suspended movement is allowed without regression damage.
 */
const STATUS_RANK: Record<MatchStatus, number> = {
  not_started: 0,
  live: 1,
  suspended: 1,
  ended: 2,
  closed: 3,
  cancelled: 0,
  abandoned: 0,
};

function isAbsorbing(s: MatchStatus): boolean {
  return s === "cancelled" || s === "abandoned";
}

/**
 * Id-indexed cache for matches.
 *
 * Merge policy:
 * - hydrateMatches only ADDS unseen records; it never overrides an entry
 *   already populated by a live increment (mirrors M10's snapshot rule).
 * - applyMatchUpserted respects per-match version monotonicity.
 * - applyMatchStatusChanged enforces version monotonicity AND Match-FSM
 *   anti-regression: higher-ranked states are never downgraded; absorbing
 *   terminals (cancelled / abandoned) reject any further transition.
 *   Same-rank status updates pass through so that score / period can be
 *   refreshed inside a single FSM stage (e.g. live → live with new score).
 * - applyFixtureChanged invokes the configured refetcher and atomically
 *   replaces the stored entry; failures preserve the existing record and
 *   surface the error to the configured sink.
 */
export class MatchStore {
  private readonly matches = new Map<string, MatchRecord>();
  private readonly byTournament = new Map<string, Set<string>>();
  private readonly listeners = new Set<() => void>();
  private refetcher: MatchRefetcher | null = null;
  private errorSink: MatchErrorSink | null = null;

  setFixtureRefetcher(refetcher: MatchRefetcher | null): void {
    this.refetcher = refetcher;
  }

  setErrorSink(sink: MatchErrorSink | null): void {
    this.errorSink = sink;
  }

  hydrateMatches(matches: MatchRecord[]): void {
    let changed = false;
    for (const m of matches) {
      if (this.matches.has(m.match_id)) continue;
      this.matches.set(m.match_id, { ...m });
      this.indexTournament(m.tournament_id, m.match_id);
      changed = true;
    }
    if (changed) this.notify();
  }

  applyMatchUpserted(p: MatchUpsertedPayload): boolean {
    const existing = this.matches.get(p.match_id);
    if (existing && p.version <= existing.version) return false;

    const record: MatchRecord = existing
      ? {
          ...existing,
          tournament_id: p.tournament_id,
          home_team: p.home_team,
          away_team: p.away_team,
          scheduled_at: p.scheduled_at,
          is_live: p.is_live,
          version: p.version,
        }
      : {
          match_id: p.match_id,
          tournament_id: p.tournament_id,
          home_team: p.home_team,
          away_team: p.away_team,
          scheduled_at: p.scheduled_at,
          is_live: p.is_live,
          status: "not_started",
          version: p.version,
        };

    if (existing && existing.tournament_id !== p.tournament_id) {
      this.byTournament.get(existing.tournament_id)?.delete(p.match_id);
    }
    this.matches.set(p.match_id, record);
    this.indexTournament(record.tournament_id, record.match_id);
    this.notify();
    return true;
  }

  applyMatchStatusChanged(p: MatchStatusChangedPayload): boolean {
    const existing = this.matches.get(p.match_id);
    if (!existing) return false;
    if (p.version <= existing.version) return false;
    if (isAbsorbing(existing.status)) return false;
    if (
      !isAbsorbing(p.status) &&
      STATUS_RANK[p.status] < STATUS_RANK[existing.status]
    ) {
      return false;
    }

    const updated: MatchRecord = {
      ...existing,
      status: p.status,
      version: p.version,
    };
    if (p.home_score !== undefined) updated.home_score = p.home_score;
    if (p.away_score !== undefined) updated.away_score = p.away_score;
    if (p.period !== undefined) updated.period = p.period;

    this.matches.set(p.match_id, updated);
    this.notify();
    return true;
  }

  async applyFixtureChanged(p: FixtureChangedPayload): Promise<boolean> {
    if (!this.refetcher) {
      this.errorSink?.({
        kind: "fixture_refetch_unconfigured",
        match_id: p.match_id,
      });
      return false;
    }
    try {
      const fresh = await this.refetcher(p.match_id);
      if (!fresh) {
        this.errorSink?.({
          kind: "fixture_refetch_empty",
          match_id: p.match_id,
        });
        return false;
      }
      const previous = this.matches.get(p.match_id);
      if (previous && previous.tournament_id !== fresh.tournament_id) {
        this.byTournament.get(previous.tournament_id)?.delete(p.match_id);
      }
      this.matches.set(p.match_id, { ...fresh });
      this.indexTournament(fresh.tournament_id, fresh.match_id);
      this.notify();
      return true;
    } catch (cause) {
      this.errorSink?.({
        kind: "fixture_refetch_failed",
        match_id: p.match_id,
        cause,
      });
      return false;
    }
  }

  getMatch(matchId: string): MatchRecord | undefined {
    const record = this.matches.get(matchId);
    return record ? { ...record } : undefined;
  }

  listByTournament(tournamentId: string): MatchRecord[] {
    const ids = this.byTournament.get(tournamentId);
    if (!ids) return [];
    const out: MatchRecord[] = [];
    for (const id of ids) {
      const m = this.matches.get(id);
      if (m) out.push({ ...m });
    }
    return out;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private indexTournament(tournamentId: string, matchId: string): void {
    if (!this.byTournament.has(tournamentId)) {
      this.byTournament.set(tournamentId, new Set());
    }
    this.byTournament.get(tournamentId)!.add(matchId);
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}
