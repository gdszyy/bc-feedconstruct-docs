// frontend/src/react/StoresProvider.transport.test.tsx
//
// Behaviour: Transport ↔ Dispatcher wiring inside <StoresProvider>.
//
//   Given <stores bundle with a fake-WS-backed Transport>
//   When  <StoresProvider mounts the bundle>
//   Then  <transport.connect() is called and messages flow into the dispatcher>

import { act, render } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import type { Envelope, SystemProducerStatusPayload } from "@/contract/events";
import {
  Transport,
  type WebSocketLike,
} from "@/realtime/transport";

import {
  createDefaultStores,
  StoresProvider,
  type Stores,
} from "./StoresProvider";

// ---------------------------------------------------------------------------
// Test double — minimal FakeWebSocket mirroring the pattern used in
// realtime/transport.test.ts. We keep it local to avoid coupling test files.
// ---------------------------------------------------------------------------

class FakeWebSocket implements WebSocketLike {
  static instances: FakeWebSocket[] = [];

  static factory(url: string): FakeWebSocket {
    const ws = new FakeWebSocket(url);
    FakeWebSocket.instances.push(ws);
    return ws;
  }

  static reset(): void {
    FakeWebSocket.instances = [];
  }

  readyState = 0;
  sent: string[] = [];
  closed = false;
  onopen: ((ev?: unknown) => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: ((ev: { code: number; reason?: string }) => void) | null = null;
  onerror: ((ev?: unknown) => void) | null = null;

  constructor(public url: string) {}

  send(data: string): void {
    this.sent.push(data);
  }

  close(code = 1000): void {
    this.closed = true;
    this.readyState = 3;
    this.onclose?.({ code });
  }

  fireOpen(): void {
    this.readyState = 1;
    this.onopen?.();
  }

  fireMessage(env: unknown): void {
    this.onmessage?.({ data: JSON.stringify(env) });
  }
}

function makeBundle(): Stores {
  const transport = new Transport({
    url: "ws://test/ws",
    webSocketFactory: FakeWebSocket.factory,
  });
  return createDefaultStores({ transport });
}

function makeProducerStatusEnv(
  isDown: boolean,
): Envelope<SystemProducerStatusPayload> {
  return {
    type: "system.producer_status",
    schema_version: "1",
    event_id: `evt-${isDown ? "down" : "up"}-${Math.random()}`,
    correlation_id: "corr-transport-wire",
    product_id: "live",
    occurred_at: "2026-05-16T00:01:00Z",
    received_at: "2026-05-16T00:01:00Z",
    entity: {},
    payload: {
      product: "live",
      is_down: isDown,
      last_message_at: "2026-05-16T00:00:50Z",
      down_since: isDown ? "2026-05-16T00:01:00Z" : undefined,
    },
  };
}

beforeEach(() => {
  FakeWebSocket.reset();
});

// Given a Stores bundle whose Transport is built with a fake WebSocket
//   factory, and whose current ConnectionState is "Disconnected"
// When the StoresProvider renders the bundle
// Then the Transport receives a connect() call (exactly one WS instance is
//   constructed) and the state moves to "Connecting"
describe("given a disconnected Transport in the Stores bundle", () => {
  it("when the provider mounts then transport.connect() is invoked once", () => {
    const bundle = makeBundle();
    expect(bundle.transport.getState()).toBe("Disconnected");

    render(
      <StoresProvider value={bundle}>
        <div />
      </StoresProvider>,
    );

    expect(FakeWebSocket.instances.length).toBe(1);
    expect(bundle.transport.getState()).toBe("Connecting");
  });
});

// Given a mounted StoresProvider with a fake-WS-backed Transport that is open
// When the fake WebSocket emits a system.producer_status envelope marking
//   product=live is_down=true
// Then the dispatcher routes it into HealthStore.applyProducerStatus and the
//   HealthStore exposes a non-null banner
describe("given a mounted provider with an open Transport", () => {
  it("when the WS emits a producer_status envelope then HealthStore.banner becomes non-null", () => {
    const bundle = makeBundle();
    render(
      <StoresProvider value={bundle}>
        <div />
      </StoresProvider>,
    );
    const ws = FakeWebSocket.instances[0]!;

    act(() => {
      ws.fireOpen();
    });
    expect(bundle.transport.getState()).toBe("Open");

    act(() => {
      ws.fireMessage(makeProducerStatusEnv(true));
    });

    const banner = bundle.health.getBanner();
    expect(banner).toBeDefined();
    expect(banner?.message).toMatch(/live/i);
  });
});

// Given a mounted StoresProvider with a fake-WS-backed Transport
// When the provider unmounts and the fake WebSocket subsequently emits an
//   envelope
// Then the dispatcher does NOT receive the envelope (the provider's onMessage
//   subscription was cleaned up). Transport itself is NOT closed — it is
//   owned by the Stores bundle and may outlive React mount cycles.
describe("given a provider that has unmounted", () => {
  it("when the WS later emits an envelope then the dispatcher does not receive it", () => {
    const bundle = makeBundle();
    const { unmount } = render(
      <StoresProvider value={bundle}>
        <div />
      </StoresProvider>,
    );
    const ws = FakeWebSocket.instances[0]!;
    act(() => {
      ws.fireOpen();
    });

    unmount();

    // Transport must NOT be permanently closed — Transport.close() sets
    // stopped=true which would prevent any future reconnect. The bundle owns
    // its Transport across mount cycles.
    expect(ws.closed).toBe(false);

    // After unmount, the onMessage subscription installed by the provider is
    // gone, so a message arriving on the WS must not produce a HealthStore
    // banner update.
    act(() => {
      ws.fireMessage(makeProducerStatusEnv(true));
    });
    expect(bundle.health.getBanner()).toBeUndefined();
  });
});
