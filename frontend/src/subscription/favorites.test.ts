import { describe, expect, it, vi } from "vitest";

import { FavoritesStore } from "./favorites";

// ---------------------------------------------------------------------------
// M11 — FavoritesStore
//
// Locked decisions (PR thread):
//   - Favorite is a LOCAL preference, NOT a server subscription
//   - Distinct from SubscriptionStore — the two coexist without overriding
//   - Persists to localStorage at the integration layer; this store is the
//     pure in-memory model and exposes a serializable snapshot
//   - Idempotent mutations: add()/remove() return true only when state
//     actually changes
//   - Listener notifications fire only on real mutations
// ---------------------------------------------------------------------------

// =================== Empty store ===================

// Given an empty FavoritesStore
// When list() / has(matchId) are queried
// Then list() returns [] and has(matchId) returns false
describe("M11 favorites baseline: empty store yields nothing", () => {
  it("when no favorite has been added then list() returns [] and has() returns false", () => {
    const store = new FavoritesStore();
    expect(store.list()).toEqual([]);
    expect(store.has("sr:match:1")).toBe(false);
  });
});

// =================== Add ===================

// Given an empty FavoritesStore
// When add(matchId, at: 1000) is invoked
// Then list() returns [{match_id: matchId, added_at: 1000}] and has(matchId) is true
describe("M11 favorites add: first add appends entry", () => {
  it("when a new match is added then it appears in list() with the supplied addedAt", () => {
    const store = new FavoritesStore();
    const ok = store.add("sr:match:1", 1000);
    expect(ok).toBe(true);
    expect(store.has("sr:match:1")).toBe(true);
    expect(store.list()).toEqual([{ match_id: "sr:match:1", added_at: 1000 }]);
  });
});

// Given a favorite already exists for matchId at 1000
// When add(matchId, at: 2000) is invoked again
// Then add returns false (dedup), the entry stays at addedAt=1000, list is unchanged
describe("M11 favorites add: duplicate add is a no-op", () => {
  it("when the same match is added twice then add() returns false and the first addedAt is preserved", () => {
    const store = new FavoritesStore();
    store.add("sr:match:1", 1000);
    const ok = store.add("sr:match:1", 2000);
    expect(ok).toBe(false);
    expect(store.list()).toEqual([{ match_id: "sr:match:1", added_at: 1000 }]);
  });
});

// =================== Remove ===================

// Given a favorite exists for matchId
// When remove(matchId) is invoked
// Then remove returns true, has(matchId) is false, list omits the entry
describe("M11 favorites remove: existing entry is dropped", () => {
  it("when an existing favorite is removed then has() flips to false and list() drops it", () => {
    const store = new FavoritesStore();
    store.add("sr:match:1", 1000);
    const ok = store.remove("sr:match:1");
    expect(ok).toBe(true);
    expect(store.has("sr:match:1")).toBe(false);
    expect(store.list()).toEqual([]);
  });
});

// Given no favorite exists for matchId
// When remove(matchId) is invoked
// Then remove returns false and the store remains empty
describe("M11 favorites remove: removing a missing entry is a no-op", () => {
  it("when remove() targets a match that was never added then it returns false", () => {
    const store = new FavoritesStore();
    const ok = store.remove("sr:match:1");
    expect(ok).toBe(false);
    expect(store.list()).toEqual([]);
  });
});

// =================== list() ordering ===================

// Given multiple favorites added out of chronological order
// When list() is called
// Then entries are returned sorted ascending by added_at
describe("M11 favorites list: sorted ascending by added_at", () => {
  it("when favorites are added with mixed timestamps then list() returns them in chronological order", () => {
    const store = new FavoritesStore();
    store.add("sr:match:1", 3000);
    store.add("sr:match:2", 1000);
    store.add("sr:match:3", 2000);
    expect(store.list().map((e) => e.match_id)).toEqual([
      "sr:match:2",
      "sr:match:3",
      "sr:match:1",
    ]);
    expect(store.list().map((e) => e.added_at)).toEqual([1000, 2000, 3000]);
  });
});

// =================== Snapshot / hydrate ===================

// Given a serializable snapshot of favorites (e.g. from localStorage)
// When the store is hydrated from that snapshot
// Then list()/has() reflect the snapshot entries exactly
describe("M11 favorites hydrate: snapshot round-trips through hydrate()/snapshot()", () => {
  it("when a snapshot is produced then re-hydrating yields an equivalent store", () => {
    const source = new FavoritesStore();
    source.add("sr:match:1", 1000);
    source.add("sr:match:2", 2000);

    const snap = source.snapshot();
    expect(snap).toEqual({
      entries: [
        { match_id: "sr:match:1", added_at: 1000 },
        { match_id: "sr:match:2", added_at: 2000 },
      ],
    });

    const target = new FavoritesStore();
    target.hydrate(snap);
    expect(target.has("sr:match:1")).toBe(true);
    expect(target.has("sr:match:2")).toBe(true);
    expect(target.list()).toEqual(snap.entries);
  });
});

// =================== Listener notifications ===================

// Given a subscribed listener
// When add(...) and remove(...) actually change state
// Then the listener fires once per change
// And duplicate add()/remove() that don't change state do NOT notify
describe("M11 favorites listeners: notified only on actual mutations", () => {
  it("when add/remove change state the listener fires; duplicate no-ops do not", () => {
    const store = new FavoritesStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.add("sr:match:1", 1000);
    expect(listener).toHaveBeenCalledTimes(1);

    store.add("sr:match:1", 2000); // dup
    expect(listener).toHaveBeenCalledTimes(1);

    store.remove("sr:match:1");
    expect(listener).toHaveBeenCalledTimes(2);

    store.remove("sr:match:1"); // missing
    expect(listener).toHaveBeenCalledTimes(2);
  });
});

// =================== Multi-match isolation ===================

// Given favorites for match:1 and match:2
// When remove(match:1) is invoked
// Then match:2 remains untouched
describe("M11 favorites multi-match isolation", () => {
  it("when one favorite is removed then others remain in the store", () => {
    const store = new FavoritesStore();
    store.add("sr:match:1", 1000);
    store.add("sr:match:2", 2000);
    store.remove("sr:match:1");
    expect(store.has("sr:match:1")).toBe(false);
    expect(store.has("sr:match:2")).toBe(true);
    expect(store.list().map((e) => e.match_id)).toEqual(["sr:match:2"]);
  });
});
