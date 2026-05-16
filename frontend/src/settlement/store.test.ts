import { describe, expect, it, vi } from "vitest";

import { CancelStore, SettlementStore } from "./store";

// ---------------------------------------------------------------------------
// M08 — Settlement & Cancel
//
// Locked decisions (see PR thread):
//   - Two separate classes: SettlementStore (per-market record, strict
//     version monotonicity) + CancelStore (match/market scope, field-equal
//     idempotency, no version field on the wire).
//   - FSM (M06 Settled / Cancelled) is driven by M02 dispatcher fan-out;
//     these stores own data only.
//
// Contract enums (frontend/src/contract/events.ts):
//   result   ∈ {win, lose, void, half_win, half_lose}
//   certainty ∈ {certain, settled_after_confirmation}
// ---------------------------------------------------------------------------

// =================== SettlementStore: empty / first-seen ===================

// Given an empty SettlementStore
// When selectSettlement / selectOutcomeSettlement are queried
// Then both return undefined
describe("M08 settlement baseline: empty store", () => {
  it("when no settlement has been applied then any lookup returns undefined", () => {
    const store = new SettlementStore();
    expect(store.selectSettlement("sr:match:1", "sr:market:1")).toBeUndefined();
    expect(
      store.selectOutcomeSettlement("sr:match:1", "sr:market:1", "home"),
    ).toBeUndefined();
  });
});

// Given an empty SettlementStore
// When bet_settlement.applied lists two outcomes (home=win, away=lose), certainty=settled_after_confirmation, version=10
// Then each outcome is stored against its outcome_id, and metadata is preserved
describe("M08 settlement first-seen: multi-outcome payload", () => {
  it("when bet_settlement.applied lists multiple outcomes then each is stored against its outcome_id", () => {
    const store = new SettlementStore();
    const ok = store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", result: "win" },
        { outcome_id: "away", result: "lose" },
      ],
      certainty: "settled_after_confirmation",
      version: 10,
    });
    expect(ok).toBe(true);

    const rec = store.selectSettlement("sr:match:1", "sr:market:1")!;
    expect(rec.certainty).toBe("settled_after_confirmation");
    expect(rec.version).toBe(10);
    expect(rec.outcomes.map((o) => o.outcome_id)).toEqual(["home", "away"]);
    expect(
      store.selectOutcomeSettlement("sr:match:1", "sr:market:1", "home")
        ?.result,
    ).toBe("win");
    expect(
      store.selectOutcomeSettlement("sr:match:1", "sr:market:1", "away")
        ?.result,
    ).toBe("lose");
  });
});

// =================== SettlementStore: void_factor / dead_heat ===================

// Given an empty store
// When bet_settlement.applied carries void_factor / dead_heat_factor
// Then the stored record reflects them
describe("M08 settlement: void_factor and dead_heat_factor preserved", () => {
  it("when payload carries void_factor / dead_heat_factor then the stored record reflects them", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", result: "win", dead_heat_factor: 0.5 },
        { outcome_id: "away", result: "void", void_factor: 0.5 },
      ],
      certainty: "certain",
      version: 1,
    });

    const home = store.selectOutcomeSettlement(
      "sr:match:1",
      "sr:market:1",
      "home",
    );
    const away = store.selectOutcomeSettlement(
      "sr:match:1",
      "sr:market:1",
      "away",
    );
    expect(home?.dead_heat_factor).toBe(0.5);
    expect(home?.void_factor).toBeUndefined();
    expect(away?.void_factor).toBe(0.5);
    expect(away?.dead_heat_factor).toBeUndefined();
  });
});

// =================== SettlementStore: certainty upgrade ===================

// Given (sr:match:1, sr:market:1) settled at version=5, certainty=settled_after_confirmation
// When bet_settlement.applied arrives at version=8, certainty=certain
// Then the prior record is replaced
describe("M08 settlement upgrade: settled_after_confirmation → certain", () => {
  it("when a higher-version settlement upgrades certainty then the prior record is replaced", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "settled_after_confirmation",
      version: 5,
    });

    const ok = store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 8,
    });
    expect(ok).toBe(true);

    const rec = store.selectSettlement("sr:match:1", "sr:market:1")!;
    expect(rec.certainty).toBe("certain");
    expect(rec.version).toBe(8);
  });
});

// =================== SettlementStore: version guard ===================

// Given (sr:match:1, sr:market:1) settled at version=10
// When bet_settlement.applied arrives at version=5
// Then the older event is dropped
describe("M08 settlement version guard: older event dropped", () => {
  it("when a stale settlement arrives then it is silently dropped", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 10,
    });

    const ok = store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "lose" }],
      certainty: "settled_after_confirmation",
      version: 5,
    });
    expect(ok).toBe(false);
    const rec = store.selectSettlement("sr:match:1", "sr:market:1")!;
    expect(rec.version).toBe(10);
    expect(rec.outcomes[0].result).toBe("win");
  });
});

// Given (sr:match:1, sr:market:1) settled at version=10
// When bet_settlement.applied arrives at version=10
// Then the duplicate is dropped (strict > guard)
describe("M08 settlement version guard: equal version dropped", () => {
  it("when a same-version settlement arrives then the duplicate is dropped", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "settled_after_confirmation",
      version: 10,
    });
    const ok = store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 10,
    });
    expect(ok).toBe(false);
    expect(store.selectSettlement("sr:match:1", "sr:market:1")?.certainty).toBe(
      "settled_after_confirmation",
    );
  });
});

// =================== SettlementStore: rollback ===================

// Given (sr:match:1, sr:market:1) settled at version=10
// When bet_settlement.rolled_back arrives at version=11
// Then the settlement record is removed
describe("M08 settlement rollback: drops the market settlement", () => {
  it("when bet_settlement.rolled_back arrives at a higher version then the settlement is removed", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 10,
    });
    const ok = store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      version: 11,
    });
    expect(ok).toBe(true);
    expect(store.selectSettlement("sr:match:1", "sr:market:1")).toBeUndefined();
  });
});

// Given (sr:match:1, sr:market:1) settled at version=10
// When bet_settlement.rolled_back arrives at version=5
// Then the rollback is dropped (settlement preserved)
describe("M08 settlement rollback version guard: older rollback dropped", () => {
  it("when a stale rolled_back arrives then the existing settlement is preserved", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 10,
    });
    const ok = store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      version: 5,
    });
    expect(ok).toBe(false);
    expect(
      store.selectSettlement("sr:match:1", "sr:market:1")?.version,
    ).toBe(10);
  });
});

// Given an empty store
// When bet_settlement.rolled_back arrives for an unknown market
// Then the call is a no-op
describe("M08 settlement rollback: no-op on unknown market", () => {
  it("when bet_settlement.rolled_back targets a market with no settlement then the call is a no-op", () => {
    const store = new SettlementStore();
    const ok = store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:99",
      version: 5,
    });
    expect(ok).toBe(false);
  });
});

// =================== SettlementStore: per-market isolation ===================

// Given two markets in the same match are both settled
// When bet_settlement.rolled_back targets only market:1
// Then market:2's settlement is preserved
describe("M08 settlement isolation: market-level rollback only touches its market", () => {
  it("when one market is rolled back then sibling markets in the same match remain settled", () => {
    const store = new SettlementStore();
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 5,
    });
    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:2",
      outcomes: [{ outcome_id: "over", result: "win" }],
      certainty: "certain",
      version: 5,
    });

    store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      version: 6,
    });

    expect(store.selectSettlement("sr:match:1", "sr:market:1")).toBeUndefined();
    expect(
      store.selectSettlement("sr:match:1", "sr:market:2")?.version,
    ).toBe(5);
  });
});

// =================== SettlementStore: listeners ===================

// Given a subscribed listener
// When real mutations occur, listener is invoked once per mutation; older / no-op events do NOT notify
describe("M08 settlement listeners: notified only on real changes", () => {
  it("when settlement state actually changes the listener fires; otherwise it does not", () => {
    const store = new SettlementStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "win" }],
      certainty: "certain",
      version: 5,
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyBetSettlementApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", result: "lose" }],
      certainty: "settled_after_confirmation",
      version: 3,
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:99",
      version: 9,
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyBetSettlementRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      version: 9,
    });
    expect(listener).toHaveBeenCalledTimes(2);
  });
});

// =================== CancelStore: empty / scopes ===================

// Given an empty CancelStore
// When select queries are issued
// Then both return undefined
describe("M08 cancel baseline: empty store", () => {
  it("when no cancel has been applied then both scope queries return undefined", () => {
    const store = new CancelStore();
    expect(store.selectMatchCancelled("sr:match:1")).toBeUndefined();
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1"),
    ).toBeUndefined();
  });
});

// Given an empty store
// When bet_cancel.applied arrives WITHOUT market_id (match scope)
// Then selectMatchCancelled returns the record
describe("M08 cancel applied: match-scope (no market_id)", () => {
  it("when bet_cancel.applied omits market_id then a match-scope cancel record is stored", () => {
    const store = new CancelStore();
    const ok = store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "match_abandoned",
      start_time: "2026-05-16T10:00:00Z",
      end_time: "2026-05-16T12:00:00Z",
      superceded_by: undefined,
    });
    expect(ok).toBe(true);

    const r = store.selectMatchCancelled("sr:match:1")!;
    expect(r.void_reason).toBe("match_abandoned");
    expect(r.start_time).toBe("2026-05-16T10:00:00Z");
    expect(r.end_time).toBe("2026-05-16T12:00:00Z");
    expect(r.market_id).toBeUndefined();
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1"),
    ).toBeUndefined();
  });
});

// Given an empty store
// When bet_cancel.applied arrives WITH market_id (market scope)
// Then selectMarketCancelled returns the record; selectMatchCancelled returns undefined
describe("M08 cancel applied: market-scope (with market_id)", () => {
  it("when bet_cancel.applied carries market_id then only that market's record is stored", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1")?.void_reason,
    ).toBe("market_void");
    expect(store.selectMatchCancelled("sr:match:1")).toBeUndefined();
  });
});

// Given both scopes have records
// Then they are independently retrievable
describe("M08 cancel coexistence: match + market scopes", () => {
  it("when both scopes have records then they are independently retrievable", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "match_void",
    });
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });
    expect(store.selectMatchCancelled("sr:match:1")?.void_reason).toBe(
      "match_void",
    );
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1")?.void_reason,
    ).toBe("market_void");
  });
});

// =================== CancelStore: idempotency / replace ===================

// Given a market cancel exists with superceded_by undefined
// When the same payload arrives again
// Then it is a no-op
describe("M08 cancel idempotency: identical re-apply", () => {
  it("when bet_cancel.applied is identical to existing record then it is a no-op", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// Given a market cancel exists with superceded_by undefined
// When bet_cancel.applied arrives with superceded_by="X"
// Then the record is replaced and listener notified
describe("M08 cancel update: superceded_by added", () => {
  it("when bet_cancel.applied changes a tracked field then the record is replaced", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });
    const listener = vi.fn();
    store.subscribe(listener);

    const ok = store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
      superceded_by: "X",
    });
    expect(ok).toBe(true);
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1")?.superceded_by,
    ).toBe("X");
    expect(listener).toHaveBeenCalledTimes(1);
  });
});

// =================== CancelStore: rollback ===================

// Given a match-scope cancel exists
// When bet_cancel.rolled_back arrives WITHOUT market_id
// Then only the match-scope record is removed
describe("M08 cancel rollback: match-scope rollback only removes match record", () => {
  it("when bet_cancel.rolled_back targets the match scope then sibling market records are preserved", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "match_void",
    });
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "market_void",
    });

    const ok = store.applyBetCancelRolledBack({ match_id: "sr:match:1" });
    expect(ok).toBe(true);
    expect(store.selectMatchCancelled("sr:match:1")).toBeUndefined();
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1")?.void_reason,
    ).toBe("market_void");
  });
});

// Given a market-scope cancel exists
// When bet_cancel.rolled_back arrives WITH market_id
// Then only that market record is removed
describe("M08 cancel rollback: market-scope rollback removes only its market", () => {
  it("when bet_cancel.rolled_back targets a market then other markets and the match scope are preserved", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "match_void",
    });
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      void_reason: "m1_void",
    });
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      market_id: "sr:market:2",
      void_reason: "m2_void",
    });

    const ok = store.applyBetCancelRolledBack({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
    });
    expect(ok).toBe(true);
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:1"),
    ).toBeUndefined();
    expect(store.selectMatchCancelled("sr:match:1")?.void_reason).toBe(
      "match_void",
    );
    expect(
      store.selectMarketCancelled("sr:match:1", "sr:market:2")?.void_reason,
    ).toBe("m2_void");
  });
});

// Given an empty store
// When bet_cancel.rolled_back arrives
// Then the call is a no-op
describe("M08 cancel rollback: no-op on unknown record", () => {
  it("when bet_cancel.rolled_back targets a record that does not exist then the call is a no-op", () => {
    const store = new CancelStore();
    expect(
      store.applyBetCancelRolledBack({ match_id: "sr:match:1" }),
    ).toBe(false);
    expect(
      store.applyBetCancelRolledBack({
        match_id: "sr:match:1",
        market_id: "sr:market:1",
      }),
    ).toBe(false);
  });
});

// =================== CancelStore: per-match isolation ===================

// Given cancels exist for sr:match:1
// When sr:match:2 is queried
// Then sr:match:2 is not cancelled
describe("M08 cancel isolation: cross-match independence", () => {
  it("when one match has cancels then other matches are unaffected", () => {
    const store = new CancelStore();
    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "x",
    });
    expect(store.selectMatchCancelled("sr:match:1")).not.toBeUndefined();
    expect(store.selectMatchCancelled("sr:match:2")).toBeUndefined();
  });
});

// =================== CancelStore: listeners ===================

// Given a subscribed listener
// When cancel state actually changes the listener fires; otherwise it does not
describe("M08 cancel listeners: notified only on real changes", () => {
  it("when cancel state actually changes the listener fires; otherwise it does not", () => {
    const store = new CancelStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "x",
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyBetCancelApplied({
      match_id: "sr:match:1",
      void_reason: "x",
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyBetCancelRolledBack({ match_id: "sr:match:1" });
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyBetCancelRolledBack({ match_id: "sr:match:1" });
    expect(listener).toHaveBeenCalledTimes(2);
  });
});
