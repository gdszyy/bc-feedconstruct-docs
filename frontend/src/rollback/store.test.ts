import { describe, expect, it, vi } from "vitest";

import { RollbackHistoryStore } from "./store";

// ---------------------------------------------------------------------------
// M09 — RollbackHistory append-only timeline
//
// Locked decisions (PR thread):
//   - Separate RollbackHistoryStore (M08 stores unchanged: current-view only)
//   - Chain granularity is (matchId, marketId); match-scope cancel rollbacks
//     are stored under a special match-scope key
//   - Bet FSM not driven here; M14 consumes this store to derive bet-level
//     state transitions
//   - record*() accepts optional `at` (default Date.now()) for determinism
//
// Idempotency (defensive against M10 replays):
//   - Settlement rollback: dedup if a prior entry for the same key already
//     carries target='settlement' AND the same version
//   - Cancel rollback: dedup if the LAST entry for the key is target='cancel'
//     and shares the same scope (consecutive duplicates collapse)
// ---------------------------------------------------------------------------

// =================== Empty store ===================

// Given an empty RollbackHistoryStore
// When selectChain / selectMatchScopeChain / selectAllForMatch are queried
// Then all return [] (empty timeline)
describe("M09 baseline: empty store yields empty timelines", () => {
  it("when no rollback has been recorded then all chain queries return []", () => {
    const store = new RollbackHistoryStore();
    expect(store.selectChain("sr:match:1", "sr:market:1")).toEqual([]);
    expect(store.selectMatchScopeChain("sr:match:1")).toEqual([]);
    expect(store.selectAllForMatch("sr:match:1")).toEqual([]);
  });
});

// =================== Settlement rollback ===================

// Given an empty store
// When recordSettlementRollback({match_id, market_id, version: 11}, at: 1000)
// Then selectChain returns [{target: 'settlement', match_id, market_id, version: 11, rolled_back_at: 1000}]
describe("M09 settlement rollback: appended to chain", () => {
  it("when a settlement rollback is recorded then it appears in the chain with target/version/at", () => {
    const store = new RollbackHistoryStore();
    const ok = store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 11 },
      1000,
    );
    expect(ok).toBe(true);
    expect(store.selectChain("sr:match:1", "sr:market:1")).toEqual([
      {
        target: "settlement",
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        version: 11,
        rolled_back_at: 1000,
      },
    ]);
  });
});

// Given a settlement rollback at version=11 already exists
// When recordSettlementRollback({version: 11}) is called again
// Then the chain remains length=1 (idempotent dedup)
describe("M09 settlement rollback: dedup on identical version", () => {
  it("when the same settlement rollback arrives twice then the chain appends only once", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 11 },
      1000,
    );
    const ok = store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 11 },
      2000,
    );
    expect(ok).toBe(false);
    expect(store.selectChain("sr:match:1", "sr:market:1")).toHaveLength(1);
    expect(store.selectChain("sr:match:1", "sr:market:1")[0].rolled_back_at).toBe(
      1000,
    );
  });
});

// Given a chain containing a settlement rollback at version=11
// When recordSettlementRollback at version=13 arrives (newer settlement was applied + rolled back again)
// Then the chain length grows to 2, preserving order
describe("M09 settlement rollback: distinct versions both retained", () => {
  it("when two settlement rollbacks at different versions are recorded then both stay in the chain in order", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 11 },
      1000,
    );
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 13 },
      2000,
    );
    const chain = store.selectChain("sr:match:1", "sr:market:1");
    expect(chain).toHaveLength(2);
    expect(chain.map((e) => (e.target === "settlement" ? e.version : -1))).toEqual(
      [11, 13],
    );
  });
});

// =================== Cancel rollback ===================

// Given an empty store
// When recordCancelRollback({match_id, market_id}, at: 1000) — market scope
// Then selectChain(matchId, marketId) returns [{target: 'cancel', market_id, rolled_back_at: 1000}]
describe("M09 cancel rollback: market-scope appended to chain", () => {
  it("when a market-scope cancel rollback is recorded then it appears in that market's chain", () => {
    const store = new RollbackHistoryStore();
    const ok = store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      1000,
    );
    expect(ok).toBe(true);
    expect(store.selectChain("sr:match:1", "sr:market:1")).toEqual([
      {
        target: "cancel",
        match_id: "sr:match:1",
        market_id: "sr:market:1",
        rolled_back_at: 1000,
      },
    ]);
    expect(store.selectMatchScopeChain("sr:match:1")).toEqual([]);
  });
});

// Given an empty store
// When recordCancelRollback({match_id}, at: 1000) — match scope, no market_id
// Then selectMatchScopeChain(matchId) returns the entry; selectChain(matchId, anyMarketId) returns []
describe("M09 cancel rollback: match-scope kept separate from market keys", () => {
  it("when a match-scope cancel rollback is recorded then only the match-scope chain reflects it", () => {
    const store = new RollbackHistoryStore();
    const ok = store.recordCancelRollback({ match_id: "sr:match:1" }, 1000);
    expect(ok).toBe(true);
    const matchChain = store.selectMatchScopeChain("sr:match:1");
    expect(matchChain).toHaveLength(1);
    expect(matchChain[0].target).toBe("cancel");
    if (matchChain[0].target === "cancel") {
      expect(matchChain[0].market_id).toBeUndefined();
      expect(matchChain[0].rolled_back_at).toBe(1000);
    }
    expect(store.selectChain("sr:match:1", "sr:market:1")).toEqual([]);
  });
});

// Given a chain ending with a cancel rollback at the same scope
// When recordCancelRollback arrives consecutively for the same scope
// Then the duplicate is dropped (chain length unchanged)
describe("M09 cancel rollback: dedup on consecutive duplicate", () => {
  it("when an identical cancel rollback follows immediately then the duplicate is collapsed", () => {
    const store = new RollbackHistoryStore();
    store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      1000,
    );
    const ok = store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      2000,
    );
    expect(ok).toBe(false);
    expect(store.selectChain("sr:match:1", "sr:market:1")).toHaveLength(1);
  });
});

// Given a chain containing a cancel rollback, then a settlement rollback (different scope = same key but different target)
// When recordCancelRollback arrives again
// Then the new entry IS appended (the last entry is not a cancel rollback anymore)
describe("M09 cancel rollback: cancel after intervening settlement entry is not deduped", () => {
  it("when a cancel rollback re-appears after another target intervened then it is appended", () => {
    const store = new RollbackHistoryStore();
    store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      1000,
    );
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 11 },
      2000,
    );
    const ok = store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      3000,
    );
    expect(ok).toBe(true);
    const chain = store.selectChain("sr:match:1", "sr:market:1");
    expect(chain).toHaveLength(3);
    expect(chain.map((e) => e.target)).toEqual([
      "cancel",
      "settlement",
      "cancel",
    ]);
  });
});

// =================== Interleaved settlement + cancel ===================

// Given a stream of mixed rollbacks for the same (matchId, marketId)
// When the chain is read
// Then entries are preserved in insertion order with correct targets
describe("M09 interleaved chain: settlement and cancel rollbacks coexist in order", () => {
  it("when settlement and cancel rollbacks interleave then the chain reflects insertion order", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 5 },
      100,
    );
    store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      200,
    );
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 9 },
      300,
    );
    const chain = store.selectChain("sr:match:1", "sr:market:1");
    expect(chain.map((e) => e.target)).toEqual([
      "settlement",
      "cancel",
      "settlement",
    ]);
    expect(chain.map((e) => e.rolled_back_at)).toEqual([100, 200, 300]);
  });
});

// =================== Scope isolation ===================

// Given rollbacks for (match:1, market:1) AND (match:1, market:2)
// When selectChain is queried per market
// Then each returns only its own entries
describe("M09 scope isolation: per-market chains don't bleed", () => {
  it("when rollbacks are recorded across markets then selectChain returns only that market's entries", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 5 },
      100,
    );
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:2", version: 7 },
      200,
    );
    expect(store.selectChain("sr:match:1", "sr:market:1")).toHaveLength(1);
    expect(store.selectChain("sr:match:1", "sr:market:2")).toHaveLength(1);
    expect(
      store
        .selectChain("sr:match:1", "sr:market:1")
        .every(
          (e) =>
            e.target === "settlement" &&
            e.market_id === "sr:market:1" &&
            e.version === 5,
        ),
    ).toBe(true);
  });
});

// Given rollbacks for (match:1, *) and (match:2, *)
// When selectAllForMatch is queried for each match
// Then each is isolated
describe("M09 cross-match isolation", () => {
  it("when rollbacks exist for multiple matches then selectAllForMatch returns only that match's entries", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 5 },
      100,
    );
    store.recordSettlementRollback(
      { match_id: "sr:match:2", market_id: "sr:market:1", version: 5 },
      200,
    );
    expect(store.selectAllForMatch("sr:match:1")).toHaveLength(1);
    expect(store.selectAllForMatch("sr:match:2")).toHaveLength(1);
    expect(
      store
        .selectAllForMatch("sr:match:1")
        .every((e) => e.match_id === "sr:match:1"),
    ).toBe(true);
  });
});

// =================== selectAllForMatch ===================

// Given mixed rollbacks across multiple markets on the same match AND a match-scope cancel rollback
// When selectAllForMatch(matchId) is called
// Then it returns ALL entries sorted by rolled_back_at ascending
describe("M09 selectAllForMatch: merged + sorted by time", () => {
  it("when multiple chains exist within one match then selectAllForMatch merges them in chronological order", () => {
    const store = new RollbackHistoryStore();
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:2", version: 5 },
      300,
    );
    store.recordCancelRollback({ match_id: "sr:match:1" }, 100);
    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 9 },
      200,
    );

    const merged = store.selectAllForMatch("sr:match:1");
    expect(merged.map((e) => e.rolled_back_at)).toEqual([100, 200, 300]);
    expect(merged.map((e) => e.target)).toEqual([
      "cancel",
      "settlement",
      "settlement",
    ]);
  });
});

// =================== Listener notifications ===================

// Given a subscribed listener
// When a real append happens, listener is notified; when a dedup drops an entry, listener is NOT notified
describe("M09 listeners: notified only on actual appends", () => {
  it("when rollback records actually append the listener fires; deduplicated calls do not notify", () => {
    const store = new RollbackHistoryStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 5 },
      100,
    );
    expect(listener).toHaveBeenCalledTimes(1);

    store.recordSettlementRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1", version: 5 },
      200,
    );
    expect(listener).toHaveBeenCalledTimes(1);

    store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      300,
    );
    expect(listener).toHaveBeenCalledTimes(2);

    store.recordCancelRollback(
      { match_id: "sr:match:1", market_id: "sr:market:1" },
      400,
    );
    expect(listener).toHaveBeenCalledTimes(2);
  });
});
