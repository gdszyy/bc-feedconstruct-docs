import { describe, expect, it, vi } from "vitest";

import {
  BetStopStore,
  deriveBettable,
  deriveFrozen,
} from "./store";

// ---------------------------------------------------------------------------
// M07 — Bet Stop overlay
//
// bet_stop.applied / bet_stop.lifted carry market_groups: string[].
// Empty array ⇒ full-match scope; non-empty ⇒ those groups only.
// Decisions locked with product (see PR thread):
//   - applied([]) over groups          ⇒ full-match wins, groups dropped
//   - lifted([a]) while fullMatch=true ⇒ NO-OP (full stop must be lifted with [])
//   - selectFrozen(market) (M05)       ⇒ FSM-only view, preserved
//   - deriveFrozen({status, inBetStop})⇒ overlay-aware view, lives in M07
//   - deriveBettable({status, inBetStop, connection}) ⇒ pure fn, per M06 doc
// ---------------------------------------------------------------------------

// =================== Empty store ===================

// Given an empty BetStopStore
// When selectStopped("sr:match:1", "main_markets") is queried
// Then it returns false (no stop in effect)
describe("M07 baseline: empty store reports no stop", () => {
  it("when no bet_stop has been applied then any group on any match is not stopped", () => {
    const store = new BetStopStore();
    expect(store.selectStopped("sr:match:1", "main_markets")).toBe(false);
    expect(store.selectStopped("sr:match:1", "props")).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(false);
  });
});

// =================== Full-match stop ===================

// Given an empty store
// When bet_stop.applied { match_id, market_groups: [] } arrives
// Then ANY group on that match is reported as stopped
describe("M07 applied: empty market_groups → full-match stop", () => {
  it("when bet_stop.applied carries an empty group array then every group on the match is stopped", () => {
    const store = new BetStopStore();
    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: [] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "main_markets")).toBe(true);
    expect(store.selectStopped("sr:match:1", "props")).toBe(true);
    expect(store.selectStopped("sr:match:1", "anything")).toBe(true);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(true);
  });
});

// Given a store already in full-match stop
// When bet_stop.applied { market_groups: [] } arrives again
// Then it is a no-op (returns false, listeners not notified)
describe("M07 applied: re-applying full-match stop is idempotent", () => {
  it("when full-match stop is re-applied then the call is a no-op and listeners do not fire", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: [] });
    const listener = vi.fn();
    store.subscribe(listener);

    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: [] }),
    ).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Group-scoped stop ===================

// Given an empty store
// When bet_stop.applied { market_groups: ["main_markets"] } arrives
// Then "main_markets" on that match is stopped, other groups are NOT
describe("M07 applied: group-scoped stop", () => {
  it("when bet_stop.applied carries specific groups then only those groups are stopped", () => {
    const store = new BetStopStore();
    expect(
      store.applyApplied({
        match_id: "sr:match:1",
        market_groups: ["main_markets"],
      }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "main_markets")).toBe(true);
    expect(store.selectStopped("sr:match:1", "props")).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(false);
  });
});

// Given a store stopping {a}
// When bet_stop.applied { market_groups: ["b"] } arrives
// Then both {a, b} are stopped (union semantics)
describe("M07 applied: union of group-scoped stops", () => {
  it("when a second group-scoped stop is applied then the stopped set is the union", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: ["b"] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "a")).toBe(true);
    expect(store.selectStopped("sr:match:1", "b")).toBe(true);
    expect(store.selectStopped("sr:match:1", "c")).toBe(false);
  });
});

// Given a store stopping {a}
// When bet_stop.applied { market_groups: ["a"] } arrives (group already stopped)
// Then it is a no-op (listeners do not fire)
describe("M07 applied: re-applying the same group is idempotent", () => {
  it("when an already-stopped group is re-applied then the call is a no-op", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    const listener = vi.fn();
    store.subscribe(listener);
    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] }),
    ).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Full-match supersedes group ===================

// Given a store stopping {a}
// When bet_stop.applied { market_groups: [] } arrives
// Then the match becomes fully stopped (any group reads as stopped)
describe("M07 applied: full-match supersedes group stop", () => {
  it("when full-match stop applies after a group stop then the entire match is stopped", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: [] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "a")).toBe(true);
    expect(store.selectStopped("sr:match:1", "totally_other_group")).toBe(true);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(true);
  });
});

// Given a store in full-match stop
// When bet_stop.applied { market_groups: ["a"] } arrives
// Then full-match scope is preserved (no demotion)
describe("M07 applied: group stop while full-match is active is a no-op", () => {
  it("when a group-scoped stop applies on top of full-match stop then full-match scope is preserved", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: [] });
    expect(
      store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] }),
    ).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(true);
  });
});

// =================== lifted (full) ===================

// Given a store fully stopping a match
// When bet_stop.lifted { market_groups: [] } arrives
// Then the match is fully cleared (no group is stopped)
describe("M07 lifted: empty market_groups clears everything", () => {
  it("when bet_stop.lifted carries an empty group array then all bet_stop scope on the match is cleared", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: [] });
    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: [] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "main_markets")).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(false);
  });
});

// Given an empty store (no stop in effect)
// When bet_stop.lifted { market_groups: [] } arrives
// Then it is a no-op (returns false, listeners do not fire)
describe("M07 lifted: lift on an unstopped match is a no-op", () => {
  it("when bet_stop.lifted arrives but no stop is in effect then the call is a no-op", () => {
    const store = new BetStopStore();
    const listener = vi.fn();
    store.subscribe(listener);
    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: [] }),
    ).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== lifted (group) ===================

// Given a store stopping {a, b}
// When bet_stop.lifted { market_groups: ["a"] } arrives
// Then only "a" is removed; "b" remains stopped
describe("M07 lifted: group-scoped lift removes only the listed groups", () => {
  it("when bet_stop.lifted carries specific groups then only those groups are removed from the stopped set", () => {
    const store = new BetStopStore();
    store.applyApplied({
      match_id: "sr:match:1",
      market_groups: ["a", "b"],
    });
    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: ["a"] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "a")).toBe(false);
    expect(store.selectStopped("sr:match:1", "b")).toBe(true);
  });
});

// Given a store stopping {a}
// When bet_stop.lifted { market_groups: ["a"] } arrives
// Then "a" is removed and the match has no remaining stop
describe("M07 lifted: removing the last group clears the match", () => {
  it("when bet_stop.lifted removes the only stopped group then the match is no longer stopped", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: ["a"] }),
    ).toBe(true);
    expect(store.selectStopped("sr:match:1", "a")).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(false);
  });
});

// Given a store with fullMatch stop in effect
// When bet_stop.lifted { market_groups: ["a"] } arrives
// Then the match remains fully stopped (locked decision: group-scoped lift
// cannot partially lift a full stop)
describe("M07 lifted: group-scoped lift against full-match stop is a no-op", () => {
  it("when group-scoped lift arrives while the match is fully stopped then the full stop is preserved", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: [] });
    const listener = vi.fn();
    store.subscribe(listener);

    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: ["a"] }),
    ).toBe(false);
    expect(store.isMatchFullyStopped("sr:match:1")).toBe(true);
    expect(store.selectStopped("sr:match:1", "a")).toBe(true);
    expect(listener).not.toHaveBeenCalled();
  });
});

// Given a store stopping {a}
// When bet_stop.lifted { market_groups: ["x"] } arrives (group not currently stopped)
// Then the store is unchanged
describe("M07 lifted: unknown group is a no-op", () => {
  it("when bet_stop.lifted targets a group that was never stopped then the store is unchanged", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    const listener = vi.fn();
    store.subscribe(listener);

    expect(
      store.applyLifted({ match_id: "sr:match:1", market_groups: ["x"] }),
    ).toBe(false);
    expect(store.selectStopped("sr:match:1", "a")).toBe(true);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Multi-match isolation ===================

// Given a store with sr:match:1 fully stopped
// When sr:match:2 is queried
// Then sr:match:2 is not stopped
describe("M07 isolation: stops on one match don't bleed into others", () => {
  it("when one match is stopped then other matches are unaffected", () => {
    const store = new BetStopStore();
    store.applyApplied({ match_id: "sr:match:1", market_groups: [] });
    store.applyApplied({ match_id: "sr:match:2", market_groups: ["a"] });

    expect(store.isMatchFullyStopped("sr:match:1")).toBe(true);
    expect(store.isMatchFullyStopped("sr:match:2")).toBe(false);
    expect(store.selectStopped("sr:match:2", "a")).toBe(true);
    expect(store.selectStopped("sr:match:2", "b")).toBe(false);
    expect(store.selectStopped("sr:match:3", "a")).toBe(false);
  });
});

// =================== Listener notifications ===================

// Given a subscribed listener
// When bet_stop.applied / lifted actually mutate the store
// Then the listener is notified once per mutation
describe("M07 listeners: notified once per real mutation", () => {
  it("when applied and lifted mutate the store then listeners fire once per change", () => {
    const store = new BetStopStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyApplied({ match_id: "sr:match:1", market_groups: ["a"] });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyApplied({ match_id: "sr:match:1", market_groups: ["b"] });
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyLifted({ match_id: "sr:match:1", market_groups: ["a"] });
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyLifted({ match_id: "sr:match:1", market_groups: ["b"] });
    expect(listener).toHaveBeenCalledTimes(4);
  });
});

// =================== Derivations: frozen ===================

// Given (market.status=active, inBetStop=true)
// When deriveFrozen is queried
// Then it returns true (overlay freeze even when FSM says active)
//
// Given (market.status=suspended, inBetStop=false)
// When deriveFrozen is queried
// Then it returns true (M06 FSM freeze)
//
// Given (market.status=active, inBetStop=false)
// When deriveFrozen is queried
// Then it returns false
describe("M07 derive: frozen = status===suspended || inBetStop", () => {
  it("when bet_stop overlay is on then any status (incl. active) reads as frozen", () => {
    expect(deriveFrozen({ status: "active", inBetStop: true })).toBe(true);
    expect(deriveFrozen({ status: "deactivated", inBetStop: true })).toBe(true);
  });

  it("when FSM is suspended then frozen=true even without bet_stop", () => {
    expect(deriveFrozen({ status: "suspended", inBetStop: false })).toBe(true);
  });

  it("when status is active and no bet_stop overlay then frozen=false", () => {
    expect(deriveFrozen({ status: "active", inBetStop: false })).toBe(false);
  });

  it("when status is non-suspended non-active and no bet_stop then frozen=false", () => {
    expect(deriveFrozen({ status: "deactivated", inBetStop: false })).toBe(
      false,
    );
    expect(deriveFrozen({ status: "settled", inBetStop: false })).toBe(false);
    expect(deriveFrozen({ status: "cancelled", inBetStop: false })).toBe(false);
    expect(deriveFrozen({ status: "handed_over", inBetStop: false })).toBe(
      false,
    );
  });
});

// =================== Derivations: bettable ===================

// Given (market.status=active, inBetStop=false, connection="Open")
// When deriveBettable is queried
// Then it returns true
describe("M07 derive: bettable happy path", () => {
  it("when status=active, no bet_stop, connection=Open then bettable=true", () => {
    expect(
      deriveBettable({
        status: "active",
        inBetStop: false,
        connection: "Open",
      }),
    ).toBe(true);
  });

  it("when connection is Degraded then bettable still true (read-only allowed)", () => {
    expect(
      deriveBettable({
        status: "active",
        inBetStop: false,
        connection: "Degraded",
      }),
    ).toBe(true);
  });
});

// Given any of (status !== active) OR inBetStop OR connection === Reconnecting
// When deriveBettable is queried
// Then it returns false
describe("M07 derive: bettable blockers", () => {
  it("when status is anything but active then bettable=false", () => {
    for (const status of [
      "suspended",
      "deactivated",
      "settled",
      "cancelled",
      "handed_over",
    ] as const) {
      expect(
        deriveBettable({
          status,
          inBetStop: false,
          connection: "Open",
        }),
      ).toBe(false);
    }
  });

  it("when bet_stop overlay is on then bettable=false", () => {
    expect(
      deriveBettable({
        status: "active",
        inBetStop: true,
        connection: "Open",
      }),
    ).toBe(false);
  });

  it("when connection is Reconnecting then bettable=false", () => {
    expect(
      deriveBettable({
        status: "active",
        inBetStop: false,
        connection: "Reconnecting",
      }),
    ).toBe(false);
  });
});
