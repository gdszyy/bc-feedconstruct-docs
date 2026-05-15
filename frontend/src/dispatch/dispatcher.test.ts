import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  EntityRef,
  Envelope,
  EventType,
} from "@/contract/events";

import { Dispatcher, type TelemetrySink } from "./dispatcher";

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

let seq = 0;

function makeEnvelope<T extends object>(opts: {
  type: EventType;
  event_id?: string;
  occurred_at?: string;
  entity?: EntityRef;
  payload?: T;
}): Envelope<T> {
  seq += 1;
  return {
    type: opts.type,
    schema_version: "1",
    event_id: opts.event_id ?? `evt-${seq}`,
    correlation_id: `corr-${seq}`,
    product_id: "live",
    occurred_at: opts.occurred_at ?? `2026-05-12T00:00:${pad(seq)}Z`,
    received_at: opts.occurred_at ?? `2026-05-12T00:00:${pad(seq)}Z`,
    entity: opts.entity ?? {},
    payload: (opts.payload ?? ({} as T)) as T,
  };
}

function pad(n: number): string {
  return n.toString().padStart(2, "0");
}

function makeTelemetry(): TelemetrySink & {
  unknown: ReturnType<typeof vi.fn>;
  handlerError: ReturnType<typeof vi.fn>;
  duplicate: ReturnType<typeof vi.fn>;
  stale: ReturnType<typeof vi.fn>;
} {
  const unknown = vi.fn();
  const handlerError = vi.fn();
  const duplicate = vi.fn();
  const stale = vi.fn();
  return {
    unknown,
    handlerError,
    duplicate,
    stale,
    recordUnknownType: unknown,
    recordHandlerError: handlerError,
    recordDuplicate: duplicate,
    recordStale: stale,
  };
}

beforeEach(() => {
  seq = 0;
});

// ---------------------------------------------------------------------------
// 路由（type → handler 注册表）
// ---------------------------------------------------------------------------

// Given 一个已注册 match.* 前缀 handler 的 dispatcher
// When 派发一条 type="match.upserted" 的 envelope
// Then 该 handler 被恰好调用一次且收到原始 envelope
describe("given a dispatcher with a match.* handler registered", () => {
  it("when a match.upserted envelope is dispatched then the handler receives it exactly once", () => {
    const d = new Dispatcher();
    const handler = vi.fn();
    d.on("match.*", handler);

    const env = makeEnvelope({
      type: "match.upserted",
      entity: { match_id: "m1" },
      payload: { match_id: "m1", version: 1 },
    });
    d.dispatch(env);

    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith(env);
  });
});

// Given 一个同时注册了 match.* 与 match.status_changed 精确 handler 的 dispatcher
// When 派发 type="match.status_changed"
// Then 精确 handler 优先，前缀 handler 不被调用
describe("given both prefix and exact-type handlers for match.status_changed", () => {
  it("when match.status_changed dispatches then the exact handler wins and the prefix handler is skipped", () => {
    const d = new Dispatcher();
    const prefixHandler = vi.fn();
    const exactHandler = vi.fn();
    d.on("match.*", prefixHandler);
    d.on("match.status_changed", exactHandler);

    d.dispatch(
      makeEnvelope({
        type: "match.status_changed",
        entity: { match_id: "m1" },
        payload: { match_id: "m1", status: "live", version: 1 },
      }),
    );

    expect(exactHandler).toHaveBeenCalledTimes(1);
    expect(prefixHandler).not.toHaveBeenCalled();
  });
});

// Given dispatcher 收到 type="match.upserted" 同时影响 M04 catalog 与 M05 markets
// When 该 envelope 被派发
// Then 路由必须以确定性顺序（按注册顺序）通知所有命中通道
describe("given cross-entity routing for a single event", () => {
  it("when the envelope hits multiple registered channels then they fire in registration order", () => {
    const d = new Dispatcher();
    const calls: string[] = [];
    d.on("match.*", () => calls.push("catalog"));
    d.on("match.*", () => calls.push("markets"));

    d.dispatch(
      makeEnvelope({
        type: "match.upserted",
        entity: { match_id: "m1" },
        payload: { match_id: "m1", version: 1 },
      }),
    );

    expect(calls).toEqual(["catalog", "markets"]);
  });
});

// Given 一个带有 *.rolled_back 通配 handler 的 dispatcher
// When 派发 type="bet_settlement.rolled_back"
// Then 后缀通配 handler 命中（与 bet_settlement.* 前缀 handler 共存）
describe("given a *.rolled_back suffix handler is registered", () => {
  it("when bet_settlement.rolled_back dispatches then both the prefix and suffix handlers fire", () => {
    const d = new Dispatcher();
    const prefixHandler = vi.fn();
    const suffixHandler = vi.fn();
    d.on("bet_settlement.*", prefixHandler);
    d.on("*.rolled_back", suffixHandler);

    d.dispatch(
      makeEnvelope({
        type: "bet_settlement.rolled_back",
        entity: { match_id: "m1", market_id: "mk1" },
        payload: { match_id: "m1", market_id: "mk1", version: 1 },
      }),
    );

    expect(prefixHandler).toHaveBeenCalledTimes(1);
    expect(suffixHandler).toHaveBeenCalledTimes(1);
  });
});

// ---------------------------------------------------------------------------
// EventDedup — 幂等（按 event_id）
// ---------------------------------------------------------------------------

// Given dispatcher 已经派发过 event_id="01HX...A" 的事件
// When 同一 event_id 的事件再次进入 dispatcher
// Then handler 不再被触发，dedup 计数增加
describe("given an event_id was already dispatched", () => {
  it("when the same event_id arrives again then the handler is not triggered and a dedup metric is recorded", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry });
    const handler = vi.fn();
    d.on("match.*", handler);

    const env = makeEnvelope({
      type: "match.upserted",
      event_id: "evt-A",
      entity: { match_id: "m1" },
      payload: { match_id: "m1", version: 1 },
    });
    d.dispatch(env);
    d.dispatch(env);

    expect(handler).toHaveBeenCalledTimes(1);
    expect(telemetry.duplicate).toHaveBeenCalledTimes(1);
    expect(telemetry.duplicate).toHaveBeenCalledWith(env);
  });
});

// Given dedup 缓存到达上限（环形 / LRU）
// When 一个新的 event_id 进入
// Then 最早缓存的 event_id 被淘汰，新事件正常派发
describe("given the dedup cache reached its capacity", () => {
  it("when a new event_id is dispatched then the oldest id is evicted and the new event still routes", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry, dedupCapacity: 2 });
    const handler = vi.fn();
    d.on("match.*", handler);

    d.dispatch(
      makeEnvelope({
        type: "match.upserted",
        event_id: "a",
        entity: { match_id: "m1" },
        payload: { match_id: "m1", version: 1 },
      }),
    );
    d.dispatch(
      makeEnvelope({
        type: "match.upserted",
        event_id: "b",
        entity: { match_id: "m2" },
        payload: { match_id: "m2", version: 1 },
      }),
    );
    d.dispatch(
      makeEnvelope({
        type: "match.upserted",
        event_id: "c",
        entity: { match_id: "m3" },
        payload: { match_id: "m3", version: 1 },
      }),
    );

    expect(handler).toHaveBeenCalledTimes(3);

    // Re-issue event_id "a" against a fresh entity so VersionGuard passes —
    // because "a" was evicted from the dedup cache the handler must fire again.
    d.dispatch(
      makeEnvelope({
        type: "match.upserted",
        event_id: "a",
        entity: { match_id: "m4" },
        payload: { match_id: "m4", version: 1 },
      }),
    );

    expect(handler).toHaveBeenCalledTimes(4);
    expect(telemetry.duplicate).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// VersionGuard — 版本/时间单调
// ---------------------------------------------------------------------------

// Given match_id+market_id 维度的最新已知 version=10
// When 一条 version=9 的 odds.changed 到达
// Then VersionGuard 丢弃该事件，handler 不被触发，stale 计数增加
describe("given a newer odds.changed has already been dispatched", () => {
  it("when an older version arrives then VersionGuard drops it and counts the stale event", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry });
    const handler = vi.fn();
    d.on("odds.*", handler);

    d.dispatch(
      makeEnvelope({
        type: "odds.changed",
        entity: { match_id: "m1", market_id: "mk1" },
        payload: { match_id: "m1", market_id: "mk1", version: 10 },
      }),
    );
    const stale = makeEnvelope({
      type: "odds.changed",
      entity: { match_id: "m1", market_id: "mk1" },
      payload: { match_id: "m1", market_id: "mk1", version: 9 },
    });
    d.dispatch(stale);

    expect(handler).toHaveBeenCalledTimes(1);
    expect(telemetry.stale).toHaveBeenCalledTimes(1);
    expect(telemetry.stale).toHaveBeenCalledWith(stale, "version");
  });
});

// Given 不携带 payload.version 的事件（例如 system.heartbeat）
// When 该事件到达 dispatcher
// Then 退化为 occurred_at 单调比较，旧的 occurred_at 被丢弃
describe("given an event type without payload.version", () => {
  it("when older occurred_at arrives then VersionGuard drops it via timestamp fallback", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry });
    const handler = vi.fn();
    d.on("system.*", handler);

    d.dispatch(
      makeEnvelope({
        type: "system.heartbeat",
        occurred_at: "2026-05-12T00:00:10Z",
        payload: { server_time: "2026-05-12T00:00:10Z" },
      }),
    );
    const stale = makeEnvelope({
      type: "system.heartbeat",
      occurred_at: "2026-05-12T00:00:05Z",
      payload: { server_time: "2026-05-12T00:00:05Z" },
    });
    d.dispatch(stale);

    expect(handler).toHaveBeenCalledTimes(1);
    expect(telemetry.stale).toHaveBeenCalledTimes(1);
    expect(telemetry.stale).toHaveBeenCalledWith(stale, "occurred_at");
  });
});

// ---------------------------------------------------------------------------
// ErrorRouter — 未知 type / 解析失败
// ---------------------------------------------------------------------------

// Given dispatcher 没有任何 handler 命中前缀 "alien.*"
// When 一条 type="alien.beep" envelope 到达
// Then 不抛异常，错误被路由到 telemetry sink，业务路径不受影响
describe("given an unknown event type", () => {
  it("when the envelope has no matching handler then it is forwarded to the telemetry sink without throwing", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry });
    const matchHandler = vi.fn();
    d.on("match.*", matchHandler);

    // "alien.beep" is intentionally outside the EventType union — cast it for
    // the test only so we can prove ErrorRouter handles it gracefully.
    const env = {
      ...makeEnvelope({ type: "match.upserted", payload: { version: 1 } }),
      type: "alien.beep" as unknown as EventType,
    };

    expect(() => d.dispatch(env)).not.toThrow();
    expect(matchHandler).not.toHaveBeenCalled();
    expect(telemetry.unknown).toHaveBeenCalledTimes(1);
    expect(telemetry.unknown).toHaveBeenCalledWith(env);
  });
});

// Given 某个 handler 在处理时抛出异常
// When dispatcher 调用该 handler
// Then 异常被捕获并路由到 telemetry，后续注册的 handler 仍然被调用
describe("given a handler throws during dispatch", () => {
  it("when dispatcher invokes it then the error is captured to telemetry and downstream handlers still fire", () => {
    const telemetry = makeTelemetry();
    const d = new Dispatcher({ telemetry });
    const boom = new Error("boom");
    const failing = vi.fn(() => {
      throw boom;
    });
    const downstream = vi.fn();
    d.on("match.*", failing);
    d.on("match.*", downstream);

    const env = makeEnvelope({
      type: "match.upserted",
      entity: { match_id: "m1" },
      payload: { match_id: "m1", version: 1 },
    });
    expect(() => d.dispatch(env)).not.toThrow();

    expect(failing).toHaveBeenCalledTimes(1);
    expect(downstream).toHaveBeenCalledTimes(1);
    expect(telemetry.handlerError).toHaveBeenCalledTimes(1);
    expect(telemetry.handlerError).toHaveBeenCalledWith(env, boom);
  });
});
