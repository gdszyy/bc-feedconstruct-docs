import { describe, expect, it, vi } from "vitest";

import type { MarketStatus } from "@/contract/events";

import {
  MarketsStore,
  selectDisplayOdds,
  type IllegalTransitionRecord,
  type MarketStatusTelemetry,
} from "./store";

// ---------------------------------------------------------------------------
// M06 — Market FSM
//
// Reducer lives on MarketsStore.applyMarketStatusChanged. Transition table is
// strict per docs/07_frontend_architecture/04_state_machines.md §3:
//   active     → suspended, deactivated, cancelled, handed_over
//   suspended  → active, cancelled, handed_over            (NOT deactivated)
//   deactivated→ settled, cancelled, handed_over
//   settled    → deactivated, cancelled, handed_over
//   cancelled  → active, suspended, deactivated, settled, handed_over
//   handed_over→ ∅  (terminal)
// Same-status status_changed at a higher version is accepted as a no-op:
// version bumps, but listeners are NOT notified.
// Older/equal version is silently dropped (no telemetry).
// Illegal transition: state unchanged, version unchanged, telemetry emitted.
// ---------------------------------------------------------------------------

function seed(
  store: MarketsStore,
  status: MarketStatus,
  version: number,
): void {
  if (status === "active") {
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.8, active: true }],
      version,
    });
    return;
  }
  // For non-active seed states we use lazy-create via status_changed itself
  // (it's the only legitimate way to bring a fresh store into a non-active
  // state without driving the FSM through legal hops we are about to test).
  store.applyMarketStatusChanged({
    match_id: "sr:match:1",
    market_id: "sr:market:1",
    status,
    version,
  });
}

function makeStore(): {
  store: MarketsStore;
  sink: ReturnType<typeof makeSink>;
} {
  const sink = makeSink();
  const store = new MarketsStore({ telemetry: sink });
  return { store, sink };
}

function makeSink() {
  const records: IllegalTransitionRecord[] = [];
  const telemetry: MarketStatusTelemetry = {
    illegalTransition(r) {
      records.push(r);
    },
  };
  return Object.assign(telemetry, { records });
}

// =================== Legal transitions ===================

// Given a market currently in status=active at version=5
// When market.status_changed { status: "suspended", version: 6 } arrives
// Then status becomes suspended, version=6, accepted=true, listener notified
describe("M06 legal: active → suspended", () => {
  it("when market.status_changed targets suspended then status becomes suspended and version advances", () => {
    const { store, sink } = makeStore();
    seed(store, "active", 5);
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 6,
    });

    expect(ok).toBe(true);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("suspended");
    expect(m.version).toBe(6);
    expect(listener).toHaveBeenCalledTimes(1);
    expect(sink.records).toHaveLength(0);
  });
});

// Given a market currently in status=active at version=5
// When market.status_changed { status: "deactivated", version: 6 } arrives
// Then status becomes deactivated, version=6
describe("M06 legal: active → deactivated", () => {
  it("when market.status_changed targets deactivated then status becomes deactivated", () => {
    const { store } = makeStore();
    seed(store, "active", 5);
    expect(
      store.applyMarketStatusChanged({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        status: "deactivated",
        version: 6,
      }),
    ).toBe(true);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
      "deactivated",
    );
  });
});

// Given a market currently in status=suspended at version=5
// When market.status_changed { status: "active", version: 6 } arrives (resume)
// Then status becomes active, version=6
describe("M06 legal: suspended → active (resume)", () => {
  it("when market.status_changed targets active then status becomes active", () => {
    const { store } = makeStore();
    seed(store, "suspended", 5);
    expect(
      store.applyMarketStatusChanged({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        status: "active",
        version: 6,
      }),
    ).toBe(true);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe("active");
  });
});

// Given a market currently in status=deactivated at version=5
// When market.status_changed { status: "settled", version: 6 } arrives
// Then status becomes settled
describe("M06 legal: deactivated → settled", () => {
  it("when market.status_changed targets settled then status becomes settled", () => {
    const { store } = makeStore();
    seed(store, "deactivated", 5);
    expect(
      store.applyMarketStatusChanged({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        status: "settled",
        version: 6,
      }),
    ).toBe(true);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
      "settled",
    );
  });
});

// Given a market currently in status=settled at version=5
// When market.status_changed { status: "deactivated", version: 6 } arrives (rollback_settle)
// Then status becomes deactivated
describe("M06 legal: settled → deactivated (rollback_settle)", () => {
  it("when market.status_changed targets deactivated then status becomes deactivated", () => {
    const { store } = makeStore();
    seed(store, "settled", 5);
    expect(
      store.applyMarketStatusChanged({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        status: "deactivated",
        version: 6,
      }),
    ).toBe(true);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
      "deactivated",
    );
  });
});

// Given a market that walks through each non-cancelled non-terminal state
// When market.status_changed { status: "cancelled" } is applied from each
// Then ANY of {active, suspended, deactivated, settled} legally cancels
describe("M06 legal: Any → cancelled (cancel)", () => {
  it.each<MarketStatus>(["active", "suspended", "deactivated", "settled"])(
    "when source is %s then market.status_changed → cancelled is accepted",
    (source) => {
      const { store, sink } = makeStore();
      seed(store, source, 5);
      expect(
        store.applyMarketStatusChanged({
          match_id: "sr:match:1",
          market_id: "sr:market:1",
          status: "cancelled",
          version: 6,
        }),
      ).toBe(true);
      expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
        "cancelled",
      );
      expect(sink.records).toHaveLength(0);
    },
  );
});

// Given a market currently in status=cancelled at version=5
// When market.status_changed { status: <prev>, version: 6 } arrives (rollback_cancel)
// Then any of {active, suspended, deactivated, settled, handed_over} is accepted
describe("M06 legal: cancelled → <prev> (rollback_cancel)", () => {
  it.each<MarketStatus>([
    "active",
    "suspended",
    "deactivated",
    "settled",
    "handed_over",
  ])(
    "when target is %s then rollback_cancel from cancelled is accepted",
    (target) => {
      const { store, sink } = makeStore();
      seed(store, "cancelled", 5);
      expect(
        store.applyMarketStatusChanged({
          match_id: "sr:match:1",
          market_id: "sr:market:1",
          status: target,
          version: 6,
        }),
      ).toBe(true);
      expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(target);
      expect(sink.records).toHaveLength(0);
    },
  );
});

// Given a market that walks through each non-terminal source state
// When market.status_changed { status: "handed_over" } is applied from each
// Then ANY non-terminal source legally hands over
describe("M06 legal: Any → handed_over (handover)", () => {
  it.each<MarketStatus>([
    "active",
    "suspended",
    "deactivated",
    "settled",
    "cancelled",
  ])(
    "when source is %s then market.status_changed → handed_over is accepted",
    (source) => {
      const { store, sink } = makeStore();
      seed(store, source, 5);
      expect(
        store.applyMarketStatusChanged({
          match_id: "sr:match:1",
          market_id: "sr:market:1",
          status: "handed_over",
          version: 6,
        }),
      ).toBe(true);
      expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
        "handed_over",
      );
      expect(sink.records).toHaveLength(0);
    },
  );
});

// =================== Illegal transitions ===================

// Given a market currently in status=active at version=5
// When market.status_changed { status: "settled", version: 6 } arrives (active cannot settle directly)
// Then state remains active, version remains 5, accepted=false, telemetry records the rejection
describe("M06 illegal: active → settled is rejected", () => {
  it("when market.status_changed skips deactivated then the transition is rejected and telemetry is emitted", () => {
    const { store, sink } = makeStore();
    seed(store, "active", 5);
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "settled",
      version: 6,
    });

    expect(ok).toBe(false);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("active");
    expect(m.version).toBe(5);
    expect(listener).not.toHaveBeenCalled();
    expect(sink.records).toEqual([
      {
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        from: "active",
        to: "settled",
        version: 6,
      },
    ]);
  });
});

// Given a market currently in status=suspended at version=5
// When market.status_changed { status: "deactivated", version: 6 } arrives
// Then per strict diagram the transition is rejected (must resume → active → deactivate)
describe("M06 illegal: suspended → deactivated is rejected (strict diagram)", () => {
  it("when market.status_changed skips the resume hop then the transition is rejected and telemetry is emitted", () => {
    const { store, sink } = makeStore();
    seed(store, "suspended", 5);

    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "deactivated",
      version: 6,
    });

    expect(ok).toBe(false);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
      "suspended",
    );
    expect(sink.records[0].from).toBe("suspended");
    expect(sink.records[0].to).toBe("deactivated");
  });
});

// Given a market currently in status=handed_over (terminal) at version=5
// When market.status_changed { status: "active", version: 6 } arrives
// Then the state remains handed_over, telemetry records the rejection
describe("M06 illegal: handed_over is terminal", () => {
  it.each<MarketStatus>([
    "active",
    "suspended",
    "deactivated",
    "settled",
    "cancelled",
  ])(
    "when target is %s then leaving handed_over is rejected and telemetry is emitted",
    (target) => {
      const { store, sink } = makeStore();
      seed(store, "handed_over", 5);
      const ok = store.applyMarketStatusChanged({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        status: target,
        version: 6,
      });
      expect(ok).toBe(false);
      expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe(
        "handed_over",
      );
      expect(sink.records).toHaveLength(1);
      expect(sink.records[0].from).toBe("handed_over");
      expect(sink.records[0].to).toBe(target);
    },
  );
});

// Given an illegal transition is rejected
// When the telemetry sink is inspected
// Then the record carries match_id / market_id / from / to / version
describe("M06 telemetry: rejected-transition record shape", () => {
  it("when an illegal transition is rejected then the telemetry record carries match_id, market_id, from, to, version", () => {
    const { store, sink } = makeStore();
    seed(store, "active", 5);
    store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "settled",
      version: 9,
    });
    expect(sink.records[0]).toEqual({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      from: "active",
      to: "settled",
      version: 9,
    });
  });
});

// =================== Idempotent self-transition ===================

// Given a market currently in status=suspended at version=5
// When market.status_changed { status: "suspended", version: 6 } arrives (same status, newer version)
// Then status remains suspended, version is bumped to 6, accepted=true, listener is NOT notified
describe("M06 idempotent: same-status status_changed", () => {
  it("when status equals current then it is a no-op transition that bumps version without notifying listeners", () => {
    const { store, sink } = makeStore();
    seed(store, "suspended", 5);
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 6,
    });

    expect(ok).toBe(true);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("suspended");
    expect(m.version).toBe(6);
    expect(listener).not.toHaveBeenCalled();
    expect(sink.records).toHaveLength(0);
  });
});

// =================== Version guard ===================

// Given a market with version=10
// When an older market.status_changed { version: 4 } arrives with a different status
// Then state unchanged, version unchanged, accepted=false, telemetry NOT emitted
describe("M06 version guard: older event silently dropped", () => {
  it("when version is older than current then the event is ignored and is NOT reported as illegal", () => {
    const { store, sink } = makeStore();
    seed(store, "active", 10);
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 4,
    });

    expect(ok).toBe(false);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("active");
    expect(m.version).toBe(10);
    expect(listener).not.toHaveBeenCalled();
    expect(sink.records).toHaveLength(0);
  });
});

// Given a market with version=10
// When market.status_changed { version: 10 } at the SAME version arrives
// Then the event is dropped silently (strict monotonicity ≤)
describe("M06 version guard: equal version dropped", () => {
  it("when version equals current then the event is dropped silently", () => {
    const { store, sink } = makeStore();
    seed(store, "active", 10);
    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 10,
    });
    expect(ok).toBe(false);
    expect(store.getMarket("sr:match:1", "sr:market:1")?.status).toBe("active");
    expect(sink.records).toHaveLength(0);
  });
});

// =================== Lazy create (first observation) ===================

// Given no market exists yet for (sr:match:1, sr:market:1)
// When market.status_changed { status: "suspended", version: 2 } arrives first
// Then the market is lazily created with status=suspended, version=2, outcomes=[]
describe("M06 first-seen: lazy create from market.status_changed", () => {
  it("when market.status_changed is the first observation then the market is created with the target status", () => {
    const { store } = makeStore();
    const ok = store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 2,
    });

    expect(ok).toBe(true);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("suspended");
    expect(m.version).toBe(2);
    expect(m.outcomes).toEqual([]);
    expect(m.specifiers).toEqual({});
  });
});

// =================== Selector parity ===================

// Given a market reaches status ∈ {deactivated, settled, cancelled, handed_over}
// When selectDisplayOdds is queried
// Then it returns null (odds only exposed in {active, suspended})
describe("M06 selector parity: displayOdds hidden outside active/suspended", () => {
  it.each<MarketStatus>(["deactivated", "settled", "cancelled", "handed_over"])(
    "when status is %s then selectDisplayOdds returns null",
    (status) => {
      const { store } = makeStore();
      seed(store, status, 1);
      const m = store.getMarket("sr:match:1", "sr:market:1")!;
      expect(selectDisplayOdds(m)).toBeNull();
    },
  );

  it("when status is active or suspended then selectDisplayOdds returns the outcome list", () => {
    const { store } = makeStore();
    seed(store, "active", 1);
    expect(
      selectDisplayOdds(store.getMarket("sr:match:1", "sr:market:1")!),
    ).not.toBeNull();
    store.applyMarketStatusChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      status: "suspended",
      version: 2,
    });
    expect(
      selectDisplayOdds(store.getMarket("sr:match:1", "sr:market:1")!),
    ).not.toBeNull();
  });
});
