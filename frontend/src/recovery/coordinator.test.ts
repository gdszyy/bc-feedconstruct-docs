import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  EntityRef,
  Envelope,
  EventType,
  TypedEnvelope,
} from "@/contract/events";
import type { ConnectionState } from "@/realtime/transport";

import {
  Coordinator,
  type DispatcherChannel,
  type RealtimeChannel,
  type SnapshotApi,
} from "./coordinator";

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

class FakeTransport implements RealtimeChannel {
  private msgListeners = new Set<(env: TypedEnvelope) => void>();
  private stateListeners = new Set<(s: ConnectionState) => void>();
  replayFromCalls: Array<[string, string]> = [];

  onMessage(handler: (env: TypedEnvelope) => void): () => void {
    this.msgListeners.add(handler);
    return () => this.msgListeners.delete(handler);
  }
  onState(handler: (s: ConnectionState) => void): () => void {
    this.stateListeners.add(handler);
    return () => this.stateListeners.delete(handler);
  }
  replayFrom(cursor: string, sessionId: string): void {
    this.replayFromCalls.push([cursor, sessionId]);
  }

  emit(env: TypedEnvelope): void {
    for (const l of this.msgListeners) l(env);
  }
  setState(state: ConnectionState): void {
    for (const l of this.stateListeners) l(state);
  }
}

class FakeDispatcher implements DispatcherChannel {
  private handlers = new Map<string, Set<(env: Envelope) => void>>();
  seededVersions: Envelope[] = [];

  on(pattern: string, handler: (env: Envelope) => void): () => void {
    if (!this.handlers.has(pattern)) this.handlers.set(pattern, new Set());
    this.handlers.get(pattern)!.add(handler);
    return () => this.handlers.get(pattern)?.delete(handler);
  }

  seedVersion(env: Envelope): void {
    this.seededVersions.push(env);
  }

  emit(pattern: string, env: Envelope): void {
    this.handlers.get(pattern)?.forEach((h) => h(env));
  }
}

interface ScheduledTask {
  id: number;
  cb: () => void;
  ms: number;
  cancelled: boolean;
}

function makeScheduler() {
  const tasks: ScheduledTask[] = [];
  let nextId = 0;
  return {
    tasks,
    schedule: (cb: () => void, ms: number): number => {
      const id = nextId++;
      tasks.push({ id, cb, ms, cancelled: false });
      return id;
    },
    cancel: (h: unknown): void => {
      const t = tasks.find((x) => x.id === h);
      if (t) t.cancelled = true;
    },
    fire: (id: number): void => {
      const t = tasks.find((x) => x.id === id);
      if (t && !t.cancelled) t.cb();
    },
    flushLast: (): void => {
      for (let i = tasks.length - 1; i >= 0; i--) {
        if (!tasks[i].cancelled) {
          tasks[i].cb();
          return;
        }
      }
    },
    activeCount: (): number => tasks.filter((t) => !t.cancelled).length,
  };
}

let seq = 0;

function makeEnvelope<T extends object>(opts: {
  type: EventType;
  event_id?: string;
  occurred_at?: string;
  entity?: EntityRef;
  payload?: T;
}): TypedEnvelope {
  seq += 1;
  const env: Envelope = {
    type: opts.type,
    schema_version: "1",
    event_id: opts.event_id ?? `evt-${seq}`,
    correlation_id: `corr-${seq}`,
    product_id: "live",
    occurred_at: opts.occurred_at ?? `2026-05-12T00:00:${String(seq).padStart(2, "0")}Z`,
    received_at: opts.occurred_at ?? `2026-05-12T00:00:${String(seq).padStart(2, "0")}Z`,
    entity: opts.entity ?? {},
    payload: opts.payload ?? {},
  };
  return env as TypedEnvelope;
}

interface Fixture {
  transport: FakeTransport;
  dispatcher: FakeDispatcher;
  snapshotApi: { fetchMatch: ReturnType<typeof vi.fn>; fetchMarkets: ReturnType<typeof vi.fn> };
  scheduler: ReturnType<typeof makeScheduler>;
  coord: Coordinator;
}

function makeFixture(opts: { fixtureChangeDebounceMs?: number } = {}): Fixture {
  const transport = new FakeTransport();
  const dispatcher = new FakeDispatcher();
  const snapshotApi: SnapshotApi & {
    fetchMatch: ReturnType<typeof vi.fn>;
    fetchMarkets: ReturnType<typeof vi.fn>;
  } = {
    fetchMatch: vi.fn(async () => undefined),
    fetchMarkets: vi.fn(async () => undefined),
  } as never;
  const scheduler = makeScheduler();
  const coord = new Coordinator({
    transport,
    dispatcher,
    snapshotApi,
    fixtureChangeDebounceMs: opts.fixtureChangeDebounceMs ?? 250,
    scheduleTimeout: scheduler.schedule,
    cancelTimeout: scheduler.cancel,
  });
  coord.start();
  return { transport, dispatcher, snapshotApi, scheduler, coord };
}

beforeEach(() => {
  seq = 0;
});

// ---------------------------------------------------------------------------
// lastEventId 追踪 + 重连时 replay_from
// ---------------------------------------------------------------------------

// Given dispatcher 已派发若干 envelope，coordinator 记录到最后的 event_id="evt-99"
// When transport 进入 Open（重连成功）
// Then coordinator 通过 transport 发送 control frame
//      { op: "replay_from", cursor: "evt-99", session_id: ... }
describe("given a coordinator that has tracked lastEventId from past dispatch", () => {
  it("when the transport reconnects to Open then it issues replay_from with the last cursor", () => {
    const f = makeFixture();

    f.dispatcher.emit(
      "system.hello",
      makeEnvelope({
        type: "system.hello",
        event_id: "evt-hello",
        payload: { session_id: "sess-1", heartbeat_interval_ms: 5000 },
      }),
    );
    // Raw stream advances lastEventId — even past the hello.
    f.transport.emit(
      makeEnvelope({
        type: "match.upserted",
        event_id: "evt-99",
        entity: { match_id: "m1" },
        payload: { match_id: "m1", version: 1 },
      }),
    );
    f.transport.setState("Reconnecting");
    f.transport.setState("Open");

    expect(f.transport.replayFromCalls).toEqual([["evt-99", "sess-1"]]);
  });
});

// Given 首次连接（尚无任何已记录 event_id）
// When transport 第一次进入 Open
// Then coordinator 不发送 replay_from，仅按视图依赖触发 snapshot
describe("given a fresh coordinator with no prior event_id", () => {
  it("when the transport first opens then no replay_from is issued", () => {
    const f = makeFixture();
    f.transport.setState("Connecting");
    f.transport.setState("Open");
    expect(f.transport.replayFromCalls).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// stale 生命周期
// ---------------------------------------------------------------------------

// Given coordinator 已注册 transport + dispatcher
// When 收到 system.replay_started
// Then staleTracker 进入 stale 状态，订阅者收到 stale=true 通知
describe("given a coordinator listening for replay lifecycle events", () => {
  it("when system.replay_started arrives then stale state becomes true", () => {
    const f = makeFixture();
    const observed: boolean[] = [];
    f.coord.onStaleChange((s) => observed.push(s));

    f.dispatcher.emit(
      "system.replay_started",
      makeEnvelope({
        type: "system.replay_started",
        payload: { from_cursor: "evt-50" },
      }),
    );

    expect(f.coord.isStale()).toBe(true);
    expect(observed).toEqual([true]);
  });
});

// Given staleTracker 已 stale=true
// When 收到 system.replay_completed
// Then stale=false，订阅者收到通知
describe("given the coordinator is currently stale", () => {
  it("when system.replay_completed arrives then stale is cleared", () => {
    const f = makeFixture();
    const observed: boolean[] = [];
    f.coord.onStaleChange((s) => observed.push(s));

    f.dispatcher.emit(
      "system.replay_started",
      makeEnvelope({
        type: "system.replay_started",
        payload: { from_cursor: "evt-50" },
      }),
    );
    f.dispatcher.emit(
      "system.replay_completed",
      makeEnvelope({
        type: "system.replay_completed",
        payload: { to_cursor: "evt-99" },
      }),
    );

    expect(f.coord.isStale()).toBe(false);
    expect(observed).toEqual([true, false]);
  });
});

// ---------------------------------------------------------------------------
// fixture.changed → 触发快照覆盖
// ---------------------------------------------------------------------------

// Given coordinator 已绑定 snapshotApi
// When 收到 type="fixture.changed" 且 entity.match_id="m1"
// Then snapshotApi.fetchMatch("m1") 被调用一次
describe("given a coordinator wired to snapshotApi", () => {
  it("when fixture.changed for match m1 arrives then snapshotApi.fetchMatch is called once for m1", () => {
    const f = makeFixture();
    f.dispatcher.emit(
      "fixture.changed",
      makeEnvelope({
        type: "fixture.changed",
        entity: { match_id: "m1" },
        payload: { match_id: "m1", change_type: "schedule", refetch_required: true },
      }),
    );

    expect(f.snapshotApi.fetchMatch).not.toHaveBeenCalled();
    f.scheduler.flushLast();
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledTimes(1);
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledWith("m1");
  });
});

// Given 同一 match 在很短时间内连续收到多次 fixture.changed
// When coordinator 处理这些事件
// Then 快照请求按 match 去重 / 节流，避免风暴
describe("given rapid duplicate fixture.changed for the same match", () => {
  it("when coordinator processes them then snapshot fetches are coalesced per match within the throttle window", () => {
    const f = makeFixture();
    for (let i = 0; i < 3; i++) {
      f.dispatcher.emit(
        "fixture.changed",
        makeEnvelope({
          type: "fixture.changed",
          entity: { match_id: "m1" },
          payload: { match_id: "m1", change_type: "schedule", refetch_required: true },
        }),
      );
    }

    // Three schedules total but only the latest is still active — the prior
    // two were cancelled when each new event arrived.
    expect(f.scheduler.tasks).toHaveLength(3);
    expect(f.scheduler.activeCount()).toBe(1);

    f.scheduler.flushLast();
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledTimes(1);
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledWith("m1");
  });
});

// ---------------------------------------------------------------------------
// 快照 vs 增量的版本对齐
// ---------------------------------------------------------------------------

// Given 一个 match 的 markets 已收到增量事件 version=12
// When snapshot 返回 version=10
// Then coordinator 不允许 snapshot 覆盖更新的增量；以增量为准
describe("given live increments have advanced beyond the snapshot version", () => {
  it("when an older snapshot arrives then coordinator does not regress the in-memory state", () => {
    const f = makeFixture();
    const applied: Array<{ match_id: string; version: number }> = [];
    f.coord.onSnapshotApplied((snap) => applied.push(snap));

    f.coord.recordObservedVersion("m1", 12);

    const result = f.coord.applySnapshot({ match_id: "m1", version: 10 });

    expect(result).toBe(false);
    expect(applied).toEqual([]);
    expect(f.dispatcher.seededVersions).toEqual([]);
  });
});

// Given snapshot 返回 version=20
// When 之后到达增量 version=15
// Then VersionGuard 丢弃旧增量（已由 M02 验证），coordinator 不重复报告
describe("given a snapshot has applied version=20", () => {
  it("when a stale increment version=15 arrives then it is dropped without further reporting", () => {
    const f = makeFixture();
    const applied: Array<{ match_id: string; version: number }> = [];
    f.coord.onSnapshotApplied((snap) => applied.push(snap));

    const result = f.coord.applySnapshot({ match_id: "m1", version: 20 });
    expect(result).toBe(true);
    expect(applied).toHaveLength(1);

    // The seedVersion side-effect is what M02's VersionGuard needs in order to
    // drop the next stale increment without M10 also surfacing it.
    expect(f.dispatcher.seededVersions).toHaveLength(1);
    expect(f.dispatcher.seededVersions[0].entity.match_id).toBe("m1");
    expect(
      (f.dispatcher.seededVersions[0].payload as { version: number }).version,
    ).toBe(20);

    // A second snapshot at v=15 must be rejected and must NOT re-emit a
    // snapshot-applied notification (no further reporting).
    const second = f.coord.applySnapshot({ match_id: "m1", version: 15 });
    expect(second).toBe(false);
    expect(applied).toHaveLength(1);
  });
});

// ---------------------------------------------------------------------------
// Hydration / 视图触发
// ---------------------------------------------------------------------------

// Given 用户进入赛事详情页（match m1）
// When coordinator.requestHydration({ match_ids: ["m1"] }) 被调用
// Then snapshotApi.fetchMatch("m1") 与 fetchMarkets("m1") 同时发起
describe("given the user navigates to a match detail view", () => {
  it("when coordinator.requestHydration is called then both match and markets snapshots are requested", () => {
    const f = makeFixture();
    f.coord.requestHydration({ match_ids: ["m1"] });
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledTimes(1);
    expect(f.snapshotApi.fetchMatch).toHaveBeenCalledWith("m1");
    expect(f.snapshotApi.fetchMarkets).toHaveBeenCalledTimes(1);
    expect(f.snapshotApi.fetchMarkets).toHaveBeenCalledWith("m1");
  });
});
