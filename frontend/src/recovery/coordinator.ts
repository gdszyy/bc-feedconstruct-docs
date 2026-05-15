import type { Envelope, EventType, TypedEnvelope } from "@/contract/events";
import type { ConnectionState } from "@/realtime/transport";

export interface RealtimeChannel {
  onState(handler: (s: ConnectionState) => void): () => void;
  onMessage(handler: (env: TypedEnvelope) => void): () => void;
  replayFrom(cursor: string, sessionId: string): void;
}

export interface DispatcherChannel {
  on(pattern: string, handler: (env: Envelope) => void): () => void;
  seedVersion(env: Envelope): void;
}

export interface SnapshotApi {
  fetchMatch(matchId: string): Promise<unknown>;
  fetchMarkets(matchId: string): Promise<unknown>;
}

export interface CoordinatorOptions {
  transport: RealtimeChannel;
  dispatcher: DispatcherChannel;
  snapshotApi: SnapshotApi;
  fixtureChangeDebounceMs?: number;
  scheduleTimeout?: (cb: () => void, ms: number) => unknown;
  cancelTimeout?: (handle: unknown) => void;
}

export interface SnapshotEnvelope {
  match_id: string;
  version: number;
  [extra: string]: unknown;
}

const DEFAULT_FIXTURE_DEBOUNCE_MS = 250;

export class Coordinator {
  private readonly opts: CoordinatorOptions;
  private readonly debounceMs: number;
  private readonly schedule: (cb: () => void, ms: number) => unknown;
  private readonly cancel: (h: unknown) => void;

  private lastEventId: string | null = null;
  private sessionId: string | null = null;
  private stale = false;
  private observedVersions = new Map<string, number>();

  private staleListeners = new Set<(stale: boolean) => void>();
  private snapshotListeners = new Set<(snap: SnapshotEnvelope) => void>();
  private fixtureTimers = new Map<string, unknown>();
  private unsubs: Array<() => void> = [];

  constructor(opts: CoordinatorOptions) {
    this.opts = opts;
    this.debounceMs = opts.fixtureChangeDebounceMs ?? DEFAULT_FIXTURE_DEBOUNCE_MS;
    this.schedule =
      opts.scheduleTimeout ?? ((cb, ms) => setTimeout(cb, ms));
    this.cancel =
      opts.cancelTimeout ??
      ((h) => clearTimeout(h as ReturnType<typeof setTimeout>));
  }

  start(): void {
    this.unsubs.push(
      this.opts.transport.onMessage((env) => {
        this.lastEventId = env.event_id;
      }),
    );
    this.unsubs.push(
      this.opts.transport.onState((s) => this.onStateChange(s)),
    );
    this.unsubs.push(
      this.opts.dispatcher.on("system.hello", (env) => {
        const payload = env.payload as { session_id?: string };
        if (payload?.session_id) this.sessionId = payload.session_id;
      }),
    );
    this.unsubs.push(
      this.opts.dispatcher.on("system.replay_started", () =>
        this.setStale(true),
      ),
    );
    this.unsubs.push(
      this.opts.dispatcher.on("system.replay_completed", () =>
        this.setStale(false),
      ),
    );
    this.unsubs.push(
      this.opts.dispatcher.on("fixture.changed", (env) => {
        const matchId = env.entity?.match_id;
        if (matchId) this.scheduleFixtureSnapshot(matchId);
      }),
    );
  }

  stop(): void {
    for (const u of this.unsubs) u();
    this.unsubs = [];
    for (const h of this.fixtureTimers.values()) this.cancel(h);
    this.fixtureTimers.clear();
  }

  requestHydration(scope: { match_ids?: string[] }): void {
    for (const matchId of scope.match_ids ?? []) {
      this.opts.snapshotApi.fetchMatch(matchId);
      this.opts.snapshotApi.fetchMarkets(matchId);
    }
  }

  /**
   * Apply a snapshot returned by snapshotApi. Returns false (and notifies no
   * listeners) when the snapshot's version is older than the highest version
   * already observed for that match — never regress past live increments.
   * On success the dispatcher's VersionGuard is seeded so subsequent stale
   * real-time events are dropped without further reporting from this layer.
   */
  applySnapshot(snapshot: SnapshotEnvelope): boolean {
    const observed = this.observedVersions.get(snapshot.match_id) ?? 0;
    if (snapshot.version < observed) return false;
    this.observedVersions.set(snapshot.match_id, snapshot.version);
    this.opts.dispatcher.seedVersion(
      this.synthesizeSnapshotEnvelope(snapshot),
    );
    for (const l of this.snapshotListeners) l(snapshot);
    return true;
  }

  recordObservedVersion(matchId: string, version: number): void {
    const cur = this.observedVersions.get(matchId) ?? 0;
    if (version > cur) this.observedVersions.set(matchId, version);
  }

  isStale(): boolean {
    return this.stale;
  }

  getLastEventId(): string | null {
    return this.lastEventId;
  }

  getSessionId(): string | null {
    return this.sessionId;
  }

  onStaleChange(handler: (stale: boolean) => void): () => void {
    this.staleListeners.add(handler);
    return () => this.staleListeners.delete(handler);
  }

  onSnapshotApplied(
    handler: (snap: SnapshotEnvelope) => void,
  ): () => void {
    this.snapshotListeners.add(handler);
    return () => this.snapshotListeners.delete(handler);
  }

  private onStateChange(state: ConnectionState): void {
    if (state !== "Open") return;
    if (!this.lastEventId || !this.sessionId) return;
    this.opts.transport.replayFrom(this.lastEventId, this.sessionId);
  }

  private setStale(next: boolean): void {
    if (this.stale === next) return;
    this.stale = next;
    for (const l of this.staleListeners) l(next);
  }

  private scheduleFixtureSnapshot(matchId: string): void {
    const existing = this.fixtureTimers.get(matchId);
    if (existing !== undefined) this.cancel(existing);
    const handle = this.schedule(() => {
      this.fixtureTimers.delete(matchId);
      this.opts.snapshotApi.fetchMatch(matchId);
    }, this.debounceMs);
    this.fixtureTimers.set(matchId, handle);
  }

  private synthesizeSnapshotEnvelope(
    snapshot: SnapshotEnvelope,
  ): Envelope {
    return {
      type: "match.upserted" as EventType,
      schema_version: "1",
      event_id: `snapshot:${snapshot.match_id}:${snapshot.version}`,
      correlation_id: `snapshot:${snapshot.match_id}`,
      product_id: "live",
      occurred_at: new Date().toISOString(),
      received_at: new Date().toISOString(),
      entity: { match_id: snapshot.match_id },
      payload: { ...snapshot, version: snapshot.version },
    };
  }
}
