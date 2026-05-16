import { describe, expect, it, vi } from "vitest";

import {
  FAVORITES_STORAGE_KEY,
  FavoritesStore,
  type FavoritesStorage,
} from "./favorites";
import { SubscriptionStore } from "./store";

// ---------------------------------------------------------------------------
// M11 — FavoritesStore
//
// Purely local preference for "this user wants to keep an eye on match X".
// localStorage-backed, independent from the server-side Subscription FSM.
//
// Locked decisions (per PR thread):
//   - Storage key: "bc.favorites.v1"
//   - Serialised shape: { items: Array<{ matchId, addedAt }> }
//   - Anonymous favorites allowed
//   - list() ordering: ascending by addedAt (oldest first)
//   - Idempotent add(): preserves original addedAt, no listener fire
//   - Malformed JSON degrades to empty
// ---------------------------------------------------------------------------

function makeStorage(initial?: Record<string, string>): FavoritesStorage & {
  data: Map<string, string>;
} {
  const data = new Map<string, string>(
    initial ? Object.entries(initial) : undefined,
  );
  return {
    data,
    getItem(key) {
      return data.has(key) ? data.get(key)! : null;
    },
    setItem(key, value) {
      data.set(key, value);
    },
    removeItem(key) {
      data.delete(key);
    },
  };
}

// =================== Empty store ===================

// Given a brand-new FavoritesStore with no prior storage
// When isFavorite / list is queried before any mutation
// Then isFavorite is false and list is empty
describe("M11 favorites baseline: empty store", () => {
  it("when no favorites exist then isFavorite is false and list is empty", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    expect(store.isFavorite("sr:match:1")).toBe(false);
    expect(store.list()).toEqual([]);
  });
});

// =================== Add ===================

// Given an empty FavoritesStore
// When add(matchId) is called with addedAt=t0
// Then isFavorite(matchId) is true and list() contains exactly { matchId, addedAt: t0 }
describe("M11 favorites add: first add", () => {
  it("when add(matchId) is called then the match becomes favorited with the supplied timestamp", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    const ok = store.add("sr:match:1", 1_000);
    expect(ok).toBe(true);
    expect(store.isFavorite("sr:match:1")).toBe(true);
    expect(store.list()).toEqual([{ matchId: "sr:match:1", addedAt: 1_000 }]);
  });
});

// Given matchId already favorited at t0
// When add(matchId) is called again with t1 (>t0)
// Then the call is idempotent: no duplicate entry, addedAt stays at t0, no listener fires
describe("M11 favorites add: idempotent re-add preserves original addedAt", () => {
  it("when add is called for an already-favorited match then the original addedAt is preserved and no listener fires", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    store.add("sr:match:1", 1_000);
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.add("sr:match:1", 5_000);
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
    expect(store.list()).toEqual([{ matchId: "sr:match:1", addedAt: 1_000 }]);
  });
});

// =================== Remove ===================

// Given matchId already favorited
// When remove(matchId) is called
// Then isFavorite(matchId) is false and list() no longer contains it
describe("M11 favorites remove: existing entry", () => {
  it("when remove targets a favorited match then it is removed from the store", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    store.add("sr:match:1", 1_000);
    const ok = store.remove("sr:match:1");
    expect(ok).toBe(true);
    expect(store.isFavorite("sr:match:1")).toBe(false);
    expect(store.list()).toEqual([]);
  });
});

// Given matchId NOT favorited
// When remove(matchId) is called
// Then the call is a no-op (no listener fires)
describe("M11 favorites remove: missing entry is a no-op", () => {
  it("when remove targets a non-favorited match then no mutation occurs and no listener fires", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.remove("sr:match:1");
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// =================== Ordering ===================

// Given three matches added at t0 < t1 < t2 (in that order)
// When list() is queried
// Then results are ordered by addedAt ascending: [t0, t1, t2]
describe("M11 favorites ordering: list returns ascending by addedAt", () => {
  it("when multiple favorites exist then list() is ordered by addedAt ascending", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    store.add("c", 3_000);
    store.add("a", 1_000);
    store.add("b", 2_000);
    expect(store.list().map((e) => e.matchId)).toEqual(["a", "b", "c"]);
  });
});

// =================== Persistence ===================

// Given a FavoritesStore has previously persisted two entries to localStorage
// When a NEW FavoritesStore is constructed against the same Storage instance
// Then the new instance rehydrates the entries (isFavorite returns true; list matches)
describe("M11 favorites persistence: rehydrate from localStorage", () => {
  it("when a new store is constructed against existing storage then the prior favorites are restored", () => {
    const storage = makeStorage();
    const first = new FavoritesStore({ storage });
    first.add("sr:match:1", 1_000);
    first.add("sr:match:2", 2_000);

    const second = new FavoritesStore({ storage });
    expect(second.isFavorite("sr:match:1")).toBe(true);
    expect(second.isFavorite("sr:match:2")).toBe(true);
    expect(second.list()).toEqual([
      { matchId: "sr:match:1", addedAt: 1_000 },
      { matchId: "sr:match:2", addedAt: 2_000 },
    ]);
  });
});

// Given a fresh FavoritesStore wired to a Storage shim
// When add / remove mutations occur
// Then localStorage is written back with the canonical serialised shape
describe("M11 favorites persistence: mutations write through to localStorage", () => {
  it("when mutations occur then the storage shim receives the canonical serialised shape", () => {
    const storage = makeStorage();
    const store = new FavoritesStore({ storage });
    store.add("sr:match:1", 1_000);

    const raw = storage.getItem(FAVORITES_STORAGE_KEY)!;
    expect(JSON.parse(raw)).toEqual({
      items: [{ matchId: "sr:match:1", addedAt: 1_000 }],
    });

    store.remove("sr:match:1");
    expect(JSON.parse(storage.getItem(FAVORITES_STORAGE_KEY)!)).toEqual({
      items: [],
    });
  });
});

// Given localStorage holds malformed JSON for the favorites key
// When the store is constructed
// Then it starts empty (graceful degradation) and the malformed value is replaced on next mutation
describe("M11 favorites persistence: malformed storage falls back to empty", () => {
  it("when stored payload is malformed then the store starts empty without throwing", () => {
    const storage = makeStorage({ [FAVORITES_STORAGE_KEY]: "{not-json" });
    const store = new FavoritesStore({ storage });
    expect(store.list()).toEqual([]);
    store.add("sr:match:1", 1_000);
    expect(JSON.parse(storage.getItem(FAVORITES_STORAGE_KEY)!)).toEqual({
      items: [{ matchId: "sr:match:1", addedAt: 1_000 }],
    });
  });
});

// =================== Listeners ===================

// Given a subscribed listener
// When real mutations occur, the listener fires exactly once per mutation;
//      no-op calls (idempotent add / missing remove) do NOT fire
describe("M11 favorites listeners: fire only on real mutations", () => {
  it("when favorites state changes the listener fires; otherwise it does not", () => {
    const store = new FavoritesStore({ storage: makeStorage() });
    const listener = vi.fn();
    store.subscribe(listener);

    store.add("sr:match:1", 1_000);
    expect(listener).toHaveBeenCalledTimes(1);

    store.add("sr:match:1", 5_000);
    expect(listener).toHaveBeenCalledTimes(1);

    store.remove("sr:match:99");
    expect(listener).toHaveBeenCalledTimes(1);

    store.remove("sr:match:1");
    expect(listener).toHaveBeenCalledTimes(2);
  });
});

// =================== Isolation from Subscription FSM ===================

// Given a match is favorited locally (FavoritesStore)
// And the same match has NO server-side subscription (SubscriptionStore is empty for it)
// Then both lookups remain independent: favorite=true, subscription=undefined
describe("M11 favorites isolation: independent of server-side subscription state", () => {
  it("when a match is only favorited then SubscriptionStore lookup remains undefined", () => {
    const favorites = new FavoritesStore({ storage: makeStorage() });
    const subscriptions = new SubscriptionStore();
    favorites.add("sr:match:1", 1_000);
    expect(favorites.isFavorite("sr:match:1")).toBe(true);
    expect(subscriptions.selectByMatch("sr:match:1")).toBeUndefined();
  });
});
