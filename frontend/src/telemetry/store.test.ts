// frontend/src/telemetry/store.test.ts
//
// M16 — TelemetryStore: structured queue + batch shipper.
// Locked decisions per PR thread:
//   3. Default PII redaction list = auth/identity only
//      (token, password, email, username, phone, api_key)
//   4. Ship failure → retain & retry indefinitely
//   + correlation_id required on every event (synthesise + count if missing)
//   + Auto-ship: size threshold OR interval timer
//   + Overflow drops oldest, increments overflow counter

import { describe, expect, it, vi } from "vitest";

import {
  type TelemetryEvent,
  type TelemetryShipper,
  TelemetryStore,
} from "./store";

class FakeShipper implements TelemetryShipper {
  public batches: TelemetryEvent[][] = [];
  public reject = false;
  public resolveAfter?: () => void;

  ship(batch: TelemetryEvent[]): Promise<void> {
    this.batches.push(batch.map((e) => ({ ...e })));
    if (this.reject) return Promise.reject(new Error("ship failed"));
    if (this.resolveAfter) {
      return new Promise((res) => {
        this.resolveAfter = () => res();
      });
    }
    return Promise.resolve();
  }
}

interface FakeClock {
  schedule: (cb: () => void, ms: number) => unknown;
  cancel: (h: unknown) => void;
  advance: () => Promise<void>;
}

function fakeClock(): FakeClock {
  const pending = new Map<number, () => void>();
  let next = 1;
  return {
    schedule(cb) {
      const id = next++;
      pending.set(id, cb);
      return id;
    },
    cancel(h) {
      pending.delete(h as number);
    },
    async advance() {
      const callbacks = [...pending.values()];
      pending.clear();
      for (const cb of callbacks) cb();
      await Promise.resolve();
      await Promise.resolve();
    },
  };
}

function makeStore(
  shipper: TelemetryShipper,
  overrides: Partial<{
    batchSize: number;
    batchIntervalMs: number;
    maxQueueSize: number;
    redactKeys: string[];
    clock: FakeClock;
  }> = {},
): { store: TelemetryStore; clock: FakeClock } {
  const clock = overrides.clock ?? fakeClock();
  const store = new TelemetryStore({
    shipper,
    batchSize: overrides.batchSize ?? 20,
    batchIntervalMs: overrides.batchIntervalMs ?? 5000,
    maxQueueSize: overrides.maxQueueSize ?? 1000,
    redactKeys: overrides.redactKeys,
    scheduleTimeout: clock.schedule,
    cancelTimeout: clock.cancel,
    now: () => "2026-05-16T07:00:00Z",
  });
  return { store, clock };
}

// =================== Identity & enqueue ===================

describe("M16 enqueue: log() enriches with identity and correlation_id", () => {
  it("when log is called then the event is enqueued carrying session_id, user_id, correlation_id", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.setIdentity({ session_id: "sess-1", user_id: "u1" });
    store.log({
      level: "info",
      message: "hello",
      correlation_id: "corr-1",
      props: { extra: 7 },
    });
    const queued = store.getQueueSnapshot();
    expect(queued).toHaveLength(1);
    expect(queued[0]).toMatchObject({
      kind: "log",
      level: "info",
      session_id: "sess-1",
      user_id: "u1",
      correlation_id: "corr-1",
    });
    expect(queued[0]?.payload).toMatchObject({ message: "hello", extra: 7 });
  });
});

describe("M16 enqueue: metric() records a numeric sample", () => {
  it("when metric is called then a metric event is enqueued", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.metric({
      name: "place_latency_ms",
      value: 250,
      unit: "ms",
      tags: { product: "live" },
      correlation_id: "corr-1",
    });
    const queued = store.getQueueSnapshot();
    expect(queued[0]?.kind).toBe("metric");
    expect(queued[0]?.payload).toMatchObject({
      name: "place_latency_ms",
      value: 250,
      unit: "ms",
      tags: { product: "live" },
    });
  });
});

describe("M16 enqueue: error() retains stack", () => {
  it("when error is called then the stack is retained on the event payload", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.error({
      kind: "ws_close",
      message: "code 1006",
      stack: "Error\n  at foo",
      correlation_id: "corr-1",
    });
    const e = store.getQueueSnapshot()[0]!;
    expect(e.kind).toBe("error");
    expect(e.level).toBe("error");
    expect(e.payload).toMatchObject({
      error_kind: "ws_close",
      message: "code 1006",
      stack: "Error\n  at foo",
    });
  });
});

describe("M16 enqueue: audit() records user actions", () => {
  it("when audit is called then an audit event is enqueued", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.audit({
      action: "bet.placed",
      correlation_id: "corr-1",
      props: { bet_id: "bet-42" },
    });
    const e = store.getQueueSnapshot()[0]!;
    expect(e.kind).toBe("audit");
    expect(e.payload).toMatchObject({
      action: "bet.placed",
      bet_id: "bet-42",
    });
  });
});

// =================== correlation_id requirement ===================

describe("M16 correlation: missing correlation_id is synthesised + counted", () => {
  it("when an event arrives without correlation_id then one is synthesised and the missing counter increments", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.audit({ action: "x" });
    store.audit({ action: "y" });
    expect(store.getCounters().missingCorrelationId).toBe(2);
    const ids = store
      .getQueueSnapshot()
      .map((e) => e.correlation_id);
    expect(ids.every((id) => id && id.length > 0)).toBe(true);
    expect(new Set(ids).size).toBe(2);
  });
});

// =================== Redaction ===================

describe("M16 redaction: PII keys are stripped (auth/identity by default)", () => {
  it("when payload contains auth/identity keys then values are replaced with [REDACTED]", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper);
    store.audit({
      action: "x",
      correlation_id: "corr-1",
      props: {
        token: "secret-jwt",
        password: "hunter2",
        email: "a@b",
        username: "alice",
        phone: "+1",
        api_key: "k",
        // Monetary fields are NOT default-redacted (locked decision #3).
        stake: 100,
        amount: 50,
        nested: { token: "also-secret", harmless: "ok" },
      },
    });
    const payload = store.getQueueSnapshot()[0]?.payload as Record<
      string,
      unknown
    >;
    expect(payload.token).toBe("[REDACTED]");
    expect(payload.password).toBe("[REDACTED]");
    expect(payload.email).toBe("[REDACTED]");
    expect(payload.username).toBe("[REDACTED]");
    expect(payload.phone).toBe("[REDACTED]");
    expect(payload.api_key).toBe("[REDACTED]");
    // Monetary stays visible by default.
    expect(payload.stake).toBe(100);
    expect(payload.amount).toBe(50);
    // Nested object also redacted.
    expect(payload.nested).toEqual({
      token: "[REDACTED]",
      harmless: "ok",
    });
  });
});

describe("M16 redaction: custom keys augment the default list", () => {
  it("when extra keys are configured then those keys are also redacted", () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper, { redactKeys: ["stake", "balance"] });
    store.audit({
      action: "x",
      correlation_id: "corr-1",
      props: { stake: 100, balance: 999, token: "t" },
    });
    const payload = store.getQueueSnapshot()[0]?.payload as Record<
      string,
      unknown
    >;
    expect(payload.stake).toBe("[REDACTED]");
    expect(payload.balance).toBe("[REDACTED]");
    expect(payload.token).toBe("[REDACTED]");
  });
});

// =================== Batching ===================

describe("M16 batching: size threshold triggers ship", () => {
  it("when the batch size is reached then ship is invoked with the batch", async () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper, { batchSize: 3 });
    store.log({ level: "info", message: "a", correlation_id: "c" });
    store.log({ level: "info", message: "b", correlation_id: "c" });
    expect(shipper.batches).toHaveLength(0);
    store.log({ level: "info", message: "c", correlation_id: "c" });
    await Promise.resolve();
    await Promise.resolve();
    expect(shipper.batches).toHaveLength(1);
    expect(shipper.batches[0]).toHaveLength(3);
    expect(store.getQueueSize()).toBe(0);
    expect(store.getCounters().shipped).toBe(3);
  });
});

describe("M16 batching: interval triggers ship below threshold", () => {
  it("when the batch interval elapses then ship is invoked even below threshold", async () => {
    const shipper = new FakeShipper();
    const { store, clock } = makeStore(shipper, { batchSize: 10 });
    store.log({ level: "info", message: "a", correlation_id: "c" });
    expect(shipper.batches).toHaveLength(0);
    await clock.advance();
    expect(shipper.batches).toHaveLength(1);
    expect(store.getCounters().shipped).toBe(1);
  });
});

describe("M16 batching: flush ships all pending", () => {
  it("when flush is called then ship is invoked with all queued events", async () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper, { batchSize: 10 });
    store.log({ level: "info", message: "a", correlation_id: "c" });
    store.log({ level: "info", message: "b", correlation_id: "c" });
    await store.flush();
    expect(shipper.batches).toHaveLength(1);
    expect(shipper.batches[0]).toHaveLength(2);
  });
});

// =================== Overflow ===================

describe("M16 overflow: queue bounded and counts drops", () => {
  it("when the queue overflows then oldest events are dropped and the overflow counter increments", async () => {
    // batchSize larger than what we enqueue so no auto-flush fires.
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper, {
      maxQueueSize: 3,
      batchSize: 100,
    });
    for (let i = 0; i < 5; i++) {
      store.log({
        level: "info",
        message: `m${i}`,
        correlation_id: `c-${i}`,
      });
    }
    expect(store.getQueueSize()).toBe(3);
    expect(store.getCounters().overflow).toBe(2);
    // Most-recent retained.
    const messages = store
      .getQueueSnapshot()
      .map((e) => (e.payload as { message: string }).message);
    expect(messages).toEqual(["m2", "m3", "m4"]);
  });
});

// =================== Ship failure ===================

describe("M16 ship failure: retain & retry indefinitely", () => {
  it("when ship rejects then events remain queued and shipFailed counter increments", async () => {
    const shipper = new FakeShipper();
    shipper.reject = true;
    const { store } = makeStore(shipper, { batchSize: 100 });
    store.log({ level: "info", message: "a", correlation_id: "c" });
    store.log({ level: "info", message: "b", correlation_id: "c" });
    await store.flush();
    expect(store.getCounters().shipFailed).toBe(1);
    expect(store.getCounters().shipped).toBe(0);
    expect(store.getQueueSize()).toBe(2);

    // Next flush succeeds — events should ship.
    shipper.reject = false;
    await store.flush();
    expect(store.getCounters().shipped).toBe(2);
    expect(store.getQueueSize()).toBe(0);
  });
});

// =================== Listeners ===================

describe("M16 listeners: fire on enqueue and ship", () => {
  it("when an event is enqueued or shipped the listener fires", async () => {
    const shipper = new FakeShipper();
    const { store } = makeStore(shipper, { batchSize: 2 });
    const listener = vi.fn();
    store.subscribe(listener);
    store.log({ level: "info", message: "a", correlation_id: "c" });
    expect(listener).toHaveBeenCalledTimes(1);
    store.log({ level: "info", message: "b", correlation_id: "c" });
    await Promise.resolve();
    await Promise.resolve();
    expect(listener.mock.calls.length).toBeGreaterThanOrEqual(2);
  });
});
