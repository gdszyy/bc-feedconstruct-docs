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

export interface ApplyResult {
  applied: boolean;
  reason?:
    | "stale_version"
    | "fsm_downgrade"
    | "terminal_state"
    | "unknown_match";
}

const STATUS_RANK: Record<MatchStatus, number> = {
  not_started: 0,
  live: 1,
  suspended: 1, // transitional — same rank as live
  ended: 2,
  closed: 3,
  cancelled: -1, // marker; terminal handled separately
  abandoned: -1,
};

const TERMINAL_STATES: ReadonlySet<MatchStatus> = new Set<MatchStatus>([
  "cancelled",
  "abandoned",
]);

function canTransition(
  current: MatchStatus,
  next: MatchStatus,
): { ok: true } | { ok: false; reason: "fsm_downgrade" | "terminal_state" } {
  if (TERMINAL_STATES.has(current)) {
    return { ok: false, reason: "terminal_state" };
  }
  if (TERMINAL_STATES.has(next)) {
    return { ok: true };
  }
  if (STATUS_RANK[next] < STATUS_RANK[current]) {
    return { ok: false, reason: "fsm_downgrade" };
  }
  return { ok: true };
}

/**
 * Per-match store.
 *
 * Concurrency model: events are applied serially by the dispatcher (M02).
 * The store enforces two independent guards:
 *
 *   1. version monotonicity — stale versions are discarded silently.
 *   2. FSM anti-regression — high-rank states (Live, Ended, Closed) and
 *      terminal states (Cancelled, Abandoned) cannot be downgraded.
 */
export class MatchStore {
  private readonly matches = new Map<string, MatchRecord>();
  private readonly listeners = new Set<() => void>();

  applyUpserted(payload: MatchUpsertedPayload): ApplyResult {
    const existing = this.matches.get(payload.match_id);
    if (existing && payload.version < existing.version) {
      return { applied: false, reason: "stale_version" };
    }
    const next: MatchRecord = existing
      ? {
          ...existing,
          tournament_id: payload.tournament_id,
          home_team: payload.home_team,
          away_team: payload.away_team,
          scheduled_at: payload.scheduled_at,
          is_live: payload.is_live,
          version: payload.version,
        }
      : {
          match_id: payload.match_id,
          tournament_id: payload.tournament_id,
          home_team: payload.home_team,
          away_team: payload.away_team,
          scheduled_at: payload.scheduled_at,
          is_live: payload.is_live,
          status: "not_started",
          version: payload.version,
        };
    this.matches.set(payload.match_id, next);
    this.notify();
    return { applied: true };
  }

  applyStatusChanged(payload: MatchStatusChangedPayload): ApplyResult {
    const existing = this.matches.get(payload.match_id);
    if (!existing) {
      return { applied: false, reason: "unknown_match" };
    }
    if (payload.version < existing.version) {
      return { applied: false, reason: "stale_version" };
    }
    const transition = canTransition(existing.status, payload.status);
    if (!transition.ok) {
      return { applied: false, reason: transition.reason };
    }
    this.matches.set(payload.match_id, {
      ...existing,
      status: payload.status,
      home_score: payload.home_score ?? existing.home_score,
      away_score: payload.away_score ?? existing.away_score,
      period: payload.period ?? existing.period,
      version: payload.version,
    });
    this.notify();
    return { applied: true };
  }

  /**
   * Snapshot overlay used by fixtureChangeHandler. Schedule / teams / tournament
   * fields are taken from the snapshot, but status / score / version are
   * preserved when the live store already holds a higher-rank state — the
   * snapshot must never regress observed live signals.
   */
  applySnapshot(snapshot: MatchRecord): void {
    const existing = this.matches.get(snapshot.match_id);
    if (!existing) {
      this.matches.set(snapshot.match_id, snapshot);
      this.notify();
      return;
    }

    const keepStatus =
      TERMINAL_STATES.has(existing.status) ||
      STATUS_RANK[existing.status] > STATUS_RANK[snapshot.status];

    this.matches.set(snapshot.match_id, {
      match_id: snapshot.match_id,
      tournament_id: snapshot.tournament_id,
      home_team: snapshot.home_team,
      away_team: snapshot.away_team,
      scheduled_at: snapshot.scheduled_at,
      is_live: snapshot.is_live,
      status: keepStatus ? existing.status : snapshot.status,
      home_score: keepStatus
        ? existing.home_score
        : (snapshot.home_score ?? existing.home_score),
      away_score: keepStatus
        ? existing.away_score
        : (snapshot.away_score ?? existing.away_score),
      period: keepStatus ? existing.period : snapshot.period,
      version: Math.max(existing.version, snapshot.version),
    });
    this.notify();
  }

  get(matchId: string): MatchRecord | undefined {
    return this.matches.get(matchId);
  }

  listByTournament(tournamentId: string): MatchRecord[] {
    const out: MatchRecord[] = [];
    for (const m of this.matches.values()) {
      if (m.tournament_id === tournamentId) out.push(m);
    }
    return out;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => this.listeners.delete(handler);
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

export type SnapshotFetcher = (matchId: string) => Promise<MatchRecord>;

/**
 * Bridges fixture.changed events to a REST refetch + store overlay.
 * The fetcher is injected so tests (and app boot) can plug a real HTTP call.
 */
export function createFixtureChangeHandler(
  store: MatchStore,
  fetcher: SnapshotFetcher,
): (payload: FixtureChangedPayload) => Promise<void> {
  return async (payload) => {
    const snapshot = await fetcher(payload.match_id);
    store.applySnapshot(snapshot);
  };
}
