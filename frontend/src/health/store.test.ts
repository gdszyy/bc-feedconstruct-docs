// frontend/src/health/store.test.ts
//
// M15 — HealthStore + degradation policy.
// Locked decisions per PR thread:
//   1. Strict per-product gating
//   2. Producer-down and stale stay independent signals (no auto-propagation)
//   + Banner severity: error > warn > info
//   + canSubmitBet blocks on Degraded / Reconnecting / Closed (acceptance §3)

import { describe, expect, it, vi } from "vitest";

import type { SystemProducerStatusPayload } from "@/contract/events";
import type { GetSystemHealthResponse } from "@/contract/rest";

import { HealthStore } from "./store";

function producerPayload(
  product: "live" | "prematch",
  isDown: boolean,
): SystemProducerStatusPayload {
  return {
    product,
    is_down: isDown,
    last_message_at: "2026-05-16T07:00:00Z",
    down_since: isDown ? "2026-05-16T06:59:00Z" : undefined,
  };
}

// =================== Baseline / hydrate ===================

describe("M15 health baseline: fresh store", () => {
  it("when no signals have arrived then connection is Disconnected and no banner is shown", () => {
    const s = new HealthStore();
    expect(s.getConnection()).toBe("Disconnected");
    expect(s.getBanner()).toBeUndefined();
    expect(s.getStaleScope()).toEqual({ kind: "none" });
  });
});

describe("M15 health hydrate: REST snapshot", () => {
  it("when hydrate is invoked then producers are populated and a derived banner reflects degraded state", () => {
    const s = new HealthStore({ now: () => 1000 });
    const snapshot: GetSystemHealthResponse = {
      producers: [
        {
          product: "live",
          is_down: true,
          last_message_at: "t",
          down_since: "t",
        },
        { product: "prematch", is_down: false, last_message_at: "t" },
      ],
      degraded: true,
    };
    s.hydrate(snapshot);
    expect(s.getProducer("live")).toBe("down");
    expect(s.getProducer("prematch")).toBe("up");
    const banner = s.getBanner();
    expect(banner?.level).toBe("warn");
    expect(banner?.message).toContain("Producer live is down");
  });
});

// =================== Banner policy ===================

describe("M15 banner: healthy state → no banner", () => {
  it("when connection is Open and all producers are up then no banner is shown", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", false));
    s.applyProducerStatus(producerPayload("prematch", false));
    expect(s.getBanner()).toBeUndefined();
  });
});

describe("M15 banner: Reconnecting → warn banner", () => {
  it("when the transport is reconnecting then getBanner returns level=warn", () => {
    const s = new HealthStore();
    s.applyConnectionState("Reconnecting");
    expect(s.getBanner()?.level).toBe("warn");
  });
});

describe("M15 banner: Degraded connection → warn banner", () => {
  it("when the connection is Degraded then getBanner returns level=warn", () => {
    const s = new HealthStore();
    s.applyConnectionState("Degraded");
    expect(s.getBanner()?.level).toBe("warn");
  });
});

describe("M15 banner: Closed connection → error banner", () => {
  it("when the connection is Closed then getBanner returns level=error", () => {
    const s = new HealthStore();
    s.applyConnectionState("Closed");
    expect(s.getBanner()?.level).toBe("error");
  });
});

describe("M15 banner: producer down → warn banner referencing product", () => {
  it("when a producer is down then getBanner returns level=warn with the product name", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    const banner = s.getBanner();
    expect(banner?.level).toBe("warn");
    expect(banner?.message).toContain("live");
  });
});

describe("M15 banner: global stale → info banner", () => {
  it("when stale scope is global then getBanner returns level=info", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.setStaleScope({ kind: "global" });
    expect(s.getBanner()?.level).toBe("info");
  });
});

describe("M15 banner severity: highest wins (error > warn > info)", () => {
  it("when Reconnecting (warn) + producer down (warn) + stale (info) co-occur then level is warn", () => {
    const s = new HealthStore();
    s.applyConnectionState("Reconnecting");
    s.applyProducerStatus(producerPayload("live", true));
    s.setStaleScope({ kind: "global" });
    expect(s.getBanner()?.level).toBe("warn");
  });

  it("when Closed (error) + warn signals co-occur then level is error", () => {
    const s = new HealthStore();
    s.applyConnectionState("Closed");
    s.applyProducerStatus(producerPayload("live", true));
    expect(s.getBanner()?.level).toBe("error");
  });
});

describe("M15 banner: producer.up + stale cleared → banner disappears", () => {
  it("when producers return up and stale is cleared then getBanner returns undefined", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    s.setStaleScope({ kind: "global" });
    expect(s.getBanner()).toBeDefined();
    s.applyProducerStatus(producerPayload("live", false));
    s.setStaleScope({ kind: "none" });
    expect(s.getBanner()).toBeUndefined();
  });
});

// =================== canSubmitBet gating ===================

describe("M15 gating: healthy → canSubmitBet true", () => {
  it("when all signals are healthy then canSubmitBet returns true", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", false));
    s.applyProducerStatus(producerPayload("prematch", false));
    expect(
      s.canSubmitBet({ product_id: "live", match_id: "m1" }),
    ).toEqual({ ok: true, reasons: [] });
  });
});

describe("M15 gating: Degraded connection blocks", () => {
  it("when the connection is Degraded then canSubmitBet returns false with connection_degraded", () => {
    const s = new HealthStore();
    s.applyConnectionState("Degraded");
    s.applyProducerStatus(producerPayload("live", false));
    const gate = s.canSubmitBet({ product_id: "live" });
    expect(gate.ok).toBe(false);
    expect(gate.reasons).toContainEqual({ kind: "connection_degraded" });
  });
});

describe("M15 gating: Reconnecting blocks", () => {
  it("when the transport is reconnecting then canSubmitBet returns false", () => {
    const s = new HealthStore();
    s.applyConnectionState("Reconnecting");
    expect(s.canSubmitBet().ok).toBe(false);
  });
});

describe("M15 gating: strict per-product (locked decision #1)", () => {
  it("when live producer is down but prematch is up then prematch bets still pass", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    s.applyProducerStatus(producerPayload("prematch", false));
    expect(s.canSubmitBet({ product_id: "live" }).ok).toBe(false);
    expect(s.canSubmitBet({ product_id: "prematch" }).ok).toBe(true);
  });

  it("when canSubmitBet is called without product_id then producer state is not consulted", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    expect(s.canSubmitBet().ok).toBe(true);
  });
});

describe("M15 gating: global stale blocks", () => {
  it("when stale is global then canSubmitBet returns false with stale_global", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.setStaleScope({ kind: "global" });
    const gate = s.canSubmitBet({ product_id: "live", match_id: "m1" });
    expect(gate.ok).toBe(false);
    expect(gate.reasons).toContainEqual({ kind: "stale_global" });
  });
});

describe("M15 gating: scoped stale blocks affected match only", () => {
  it("when a match is in the stale set then canSubmitBet for that match returns false", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.setStaleScope({ kind: "scoped", match_ids: ["m1", "m2"] });
    expect(s.canSubmitBet({ match_id: "m1" }).ok).toBe(false);
    expect(s.canSubmitBet({ match_id: "m1" }).reasons).toContainEqual({
      kind: "stale_match",
      match_id: "m1",
    });
  });

  it("when a match is not in the stale set then canSubmitBet for it returns true", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.setStaleScope({ kind: "scoped", match_ids: ["m1"] });
    expect(s.canSubmitBet({ match_id: "m99" }).ok).toBe(true);
  });
});

describe("M15 gating: stale + producer-down are independent signals (locked #2)", () => {
  it("when both producer is down and stale is set then canSubmitBet surfaces BOTH reasons", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    s.setStaleScope({ kind: "global" });
    const gate = s.canSubmitBet({ product_id: "live", match_id: "m1" });
    expect(gate.ok).toBe(false);
    expect(gate.reasons).toContainEqual({
      kind: "producer_down",
      product: "live",
    });
    expect(gate.reasons).toContainEqual({ kind: "stale_global" });
  });

  it("when producer is down but no stale signal is set then stale scope stays 'none' (no auto-propagation)", () => {
    const s = new HealthStore();
    s.applyConnectionState("Open");
    s.applyProducerStatus(producerPayload("live", true));
    expect(s.getStaleScope()).toEqual({ kind: "none" });
  });
});

// =================== Listeners ===================

describe("M15 listeners: notify only on real changes", () => {
  it("when state actually changes the listener fires; redundant signals do not notify", () => {
    const s = new HealthStore();
    const listener = vi.fn();
    s.subscribe(listener);

    s.applyConnectionState("Open");
    expect(listener).toHaveBeenCalledTimes(1);

    // Same state — no notify.
    s.applyConnectionState("Open");
    expect(listener).toHaveBeenCalledTimes(1);

    s.applyProducerStatus(producerPayload("live", true));
    expect(listener).toHaveBeenCalledTimes(2);

    // Same producer status — no notify.
    s.applyProducerStatus(producerPayload("live", true));
    expect(listener).toHaveBeenCalledTimes(2);

    s.setStaleScope({ kind: "global" });
    expect(listener).toHaveBeenCalledTimes(3);
  });
});
