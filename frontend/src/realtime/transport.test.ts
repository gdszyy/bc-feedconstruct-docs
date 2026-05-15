import { beforeEach, describe, expect, it } from "vitest";

import {
  ORIGIN_REJECTED_CLOSE_CODE,
  Transport,
  type ConnectionState,
  type TransportError,
  type WebSocketLike,
} from "./transport";

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

class FakeWebSocket implements WebSocketLike {
  static instances: FakeWebSocket[] = [];

  static factory(url: string): FakeWebSocket {
    const ws = new FakeWebSocket(url);
    FakeWebSocket.instances.push(ws);
    return ws;
  }

  readyState = 0;
  sent: string[] = [];
  onopen: ((ev?: unknown) => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: ((ev: { code: number; reason?: string }) => void) | null = null;
  onerror: ((ev?: unknown) => void) | null = null;

  constructor(public url: string) {}

  send(data: string): void {
    this.sent.push(data);
  }

  close(code = 1000): void {
    this.readyState = 3;
    this.onclose?.({ code });
  }

  // helpers used only by tests
  fireOpen(): void {
    this.readyState = 1;
    this.onopen?.();
  }

  fireRemoteClose(code: number): void {
    this.readyState = 3;
    this.onclose?.({ code });
  }

  parsedFrames(): unknown[] {
    return this.sent.map((s) => JSON.parse(s));
  }
}

interface Scheduled {
  cb: () => void;
  ms: number;
}

function makeScheduler() {
  const scheduled: Scheduled[] = [];
  const cancelled = new Set<number>();
  const schedule = (cb: () => void, ms: number) => {
    const id = scheduled.length;
    scheduled.push({ cb, ms });
    return id;
  };
  const cancel = (h: unknown) => {
    cancelled.add(h as number);
  };
  const fire = (idx: number) => {
    if (cancelled.has(idx)) return;
    scheduled[idx].cb();
  };
  return { scheduled, cancelled, schedule, cancel, fire };
}

beforeEach(() => {
  FakeWebSocket.instances = [];
});

// ---------------------------------------------------------------------------
// Behaviour: Given NEXT_PUBLIC_BFF_WS configured
// When the realtime transport is initialised
// Then it opens a WebSocket to that URL and exposes a connected status
// ---------------------------------------------------------------------------
describe("given NEXT_PUBLIC_BFF_WS configured", () => {
  it("when transport initialises then websocket opens and status becomes connected", () => {
    const states: ConnectionState[] = [];
    const t = new Transport({
      url: "ws://bff.local/ws/v1/stream",
      webSocketFactory: FakeWebSocket.factory,
    });
    t.onState((s) => states.push(s));

    t.connect();

    expect(FakeWebSocket.instances).toHaveLength(1);
    expect(FakeWebSocket.instances[0].url).toBe("ws://bff.local/ws/v1/stream");
    expect(t.getState()).toBe("Connecting");
    expect(states).toContain("Connecting");

    FakeWebSocket.instances[0].fireOpen();

    expect(t.getState()).toBe("Open");
    expect(states.at(-1)).toBe("Open");
  });
});

// ---------------------------------------------------------------------------
// Behaviour: Given a dropped websocket
// When more than reconnectMinDelay elapses
// Then the transport reconnects with exponential backoff and replays the
// latest subscription set
// ---------------------------------------------------------------------------
describe("given a dropped websocket", () => {
  it("when reconnect delay elapses then it reconnects with backoff and re-subscribes", () => {
    const scheduler = makeScheduler();
    const t = new Transport({
      url: "ws://bff.local/ws/v1/stream",
      reconnectMinDelayMs: 1000,
      reconnectMaxDelayMs: 30000,
      webSocketFactory: FakeWebSocket.factory,
      scheduleTimeout: scheduler.schedule,
      cancelTimeout: scheduler.cancel,
    });

    t.connect();
    FakeWebSocket.instances[0].fireOpen();
    t.subscribe({ match_ids: ["m1"] });
    expect(FakeWebSocket.instances[0].parsedFrames()).toEqual([
      { op: "subscribe", scope: { match_ids: ["m1"] } },
    ]);

    // First drop with a generic abnormal close.
    FakeWebSocket.instances[0].fireRemoteClose(1006);

    expect(t.getState()).toBe("Reconnecting");
    expect(scheduler.scheduled).toHaveLength(1);
    expect(scheduler.scheduled[0].ms).toBe(1000);

    // Fire the timer → a new websocket is opened and the latest subscription
    // is replayed once the new socket reports Open.
    scheduler.fire(0);
    expect(FakeWebSocket.instances).toHaveLength(2);
    FakeWebSocket.instances[1].fireOpen();
    expect(FakeWebSocket.instances[1].parsedFrames()).toContainEqual({
      op: "subscribe",
      scope: { match_ids: ["m1"] },
    });

    // Second consecutive drop doubles the backoff delay.
    FakeWebSocket.instances[1].fireRemoteClose(1006);
    expect(scheduler.scheduled).toHaveLength(2);
    expect(scheduler.scheduled[1].ms).toBe(2000);
  });
});

// ---------------------------------------------------------------------------
// Behaviour: Given a 4401 origin-rejected close
// When the transport handles the close
// Then it does NOT reconnect and surfaces a non-retryable error to the UI
// ---------------------------------------------------------------------------
describe("given a 4401 origin-rejected close", () => {
  it("when transport observes the close then it stops reconnecting and surfaces the error", () => {
    const scheduler = makeScheduler();
    const errors: TransportError[] = [];
    const t = new Transport({
      url: "ws://bff.local/ws/v1/stream",
      webSocketFactory: FakeWebSocket.factory,
      scheduleTimeout: scheduler.schedule,
      cancelTimeout: scheduler.cancel,
    });
    t.onError((e) => errors.push(e));

    t.connect();
    FakeWebSocket.instances[0].fireOpen();
    FakeWebSocket.instances[0].fireRemoteClose(ORIGIN_REJECTED_CLOSE_CODE);

    expect(t.getState()).toBe("Closed");
    expect(scheduler.scheduled).toHaveLength(0);
    expect(errors).toHaveLength(1);
    expect(errors[0]).toMatchObject({
      code: "ORIGIN_REJECTED",
      retriable: false,
    });

    // A subsequent connect() must be a no-op — origin rejection is terminal.
    t.connect();
    expect(FakeWebSocket.instances).toHaveLength(1);
  });
});
