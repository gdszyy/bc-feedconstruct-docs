import { describe, expect, it, vi } from "vitest";

import { MyBetsStore, type MyBet } from "./store";

// ---------------------------------------------------------------------------
// M14 — MyBetsStore
//
// Locked decisions (per PR thread):
//   FSM (server-authoritative; client only applies if `from === current`):
//     pending ─bet.accepted→ accepted ─state_changed→ settled
//          ↘ bet.rejected ↘
//                         rejected
//                                       cancelled ◀─state_changed─ accepted
//                                       <prev>    ◀─state_changed─ settled/cancelled (rollback)
//
//   M13 hands a pending bet off via seedPending() at the moment of place().
//   REST snapshots seed history for already-existing bets verbatim.
//   The single `state !== from` guard covers replay, out-of-order, and stale
//   events (mirrors M08/M09).
// ---------------------------------------------------------------------------

const SEL = {
  match_id: "sr:match:1",
  market_id: "sr:market:1",
  outcome_id: "1",
  odds: 1.85,
};

const PENDING_SEED = {
  bet_id: "b-1",
  user_id: "u-1",
  stake: 10,
  currency: "EUR",
  bet_type: "single" as const,
  selections: [SEL],
  placed_at: "2026-05-16T10:00:00Z",
};

function snapshot(overrides: Partial<MyBet> = {}): MyBet {
  return {
    bet_id: "b-1",
    user_id: "u-1",
    stake: 10,
    currency: "EUR",
    bet_type: "single",
    selections: [SEL],
    state: "accepted",
    placed_at: "2026-05-16T10:00:00Z",
    accepted_at: "2026-05-16T10:00:01Z",
    history: [
      { from: "-", to: "pending", at: "2026-05-16T10:00:00Z" },
      { from: "pending", to: "accepted", at: "2026-05-16T10:00:01Z" },
    ],
    ...overrides,
  };
}

// =================== Empty store ===================

describe("M14 baseline: empty store yields nothing", () => {
  it("when no bet has been recorded then selectors return undefined or empty arrays", () => {
    const store = new MyBetsStore();
    expect(store.selectById("b-1")).toBeUndefined();
    expect(store.selectAll()).toEqual([]);
    expect(store.selectByState("pending")).toEqual([]);
    expect(store.selectHistory("b-1")).toEqual([]);
  });
});

// =================== seedPending ===================

describe("M14 seedPending: M13 handoff creates pending entry", () => {
  it("when M13 hands off via seedPending then the bet appears in pending state with an initial history entry", () => {
    const store = new MyBetsStore();
    const ok = store.seedPending(PENDING_SEED);
    expect(ok).toBe(true);
    const bet = store.selectById("b-1");
    expect(bet?.state).toBe("pending");
    expect(bet?.placed_at).toBe("2026-05-16T10:00:00Z");
    expect(bet?.selections).toEqual([SEL]);
    expect(bet?.history).toEqual([
      { from: "-", to: "pending", at: "2026-05-16T10:00:00Z" },
    ]);
  });

  it("when seedPending is called twice with the same bet_id then the second call is a no-op", () => {
    const store = new MyBetsStore();
    store.seedPending(PENDING_SEED);
    const ok = store.seedPending({ ...PENDING_SEED, stake: 999 });
    expect(ok).toBe(false);
    expect(store.selectById("b-1")?.stake).toBe(10);
  });
});

// =================== upsertFromSnapshot ===================

describe("M14 upsertFromSnapshot: REST seeds a complete bet", () => {
  it("when a REST snapshot arrives for a new bet then the store records it with the snapshot history", () => {
    const store = new MyBetsStore();
    const snap = snapshot();
    const ok = store.upsertFromSnapshot(snap);
    expect(ok).toBe(true);
    expect(store.selectById("b-1")).toEqual(snap);
  });
});

describe("M14 upsertFromSnapshot: idempotent on identical content", () => {
  it("when a REST snapshot matches the current entry then the store skips the update", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot());
    const ok = store.upsertFromSnapshot(snapshot());
    expect(ok).toBe(false);
  });
});

describe("M14 upsertFromSnapshot: applies newer server state", () => {
  it("when a REST snapshot advances the bet's state then the store updates and listener fires", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot());
    const newer = snapshot({
      state: "settled",
      history: [
        { from: "-", to: "pending", at: "2026-05-16T10:00:00Z" },
        { from: "pending", to: "accepted", at: "2026-05-16T10:00:01Z" },
        { from: "accepted", to: "settled", at: "2026-05-16T11:00:00Z" },
      ],
    });
    const ok = store.upsertFromSnapshot(newer);
    expect(ok).toBe(true);
    expect(store.selectById("b-1")?.state).toBe("settled");
    expect(store.selectHistory("b-1")).toHaveLength(3);
  });
});

// =================== applyAccepted ===================

describe("M14 applyAccepted: pending → accepted with history append", () => {
  it("when bet.accepted arrives for a pending bet then state moves to accepted and history grows", () => {
    const store = new MyBetsStore();
    store.seedPending(PENDING_SEED);
    const ok = store.applyAccepted({
      bet_id: "b-1",
      accepted_at: "2026-05-16T10:00:01Z",
    });
    expect(ok).toBe(true);
    const bet = store.selectById("b-1");
    expect(bet?.state).toBe("accepted");
    expect(bet?.accepted_at).toBe("2026-05-16T10:00:01Z");
    expect(bet?.history).toEqual([
      { from: "-", to: "pending", at: "2026-05-16T10:00:00Z" },
      { from: "pending", to: "accepted", at: "2026-05-16T10:00:01Z" },
    ]);
  });
});

describe("M14 applyAccepted: duplicate replay is a no-op", () => {
  it("when bet.accepted replays at the same accepted_at then the store skips it", () => {
    const store = new MyBetsStore();
    store.seedPending(PENDING_SEED);
    store.applyAccepted({ bet_id: "b-1", accepted_at: "2026-05-16T10:00:01Z" });
    const ok = store.applyAccepted({
      bet_id: "b-1",
      accepted_at: "2026-05-16T10:00:01Z",
    });
    expect(ok).toBe(false);
    expect(store.selectHistory("b-1")).toHaveLength(2);
  });
});

describe("M14 applyAccepted: unknown bet_id is ignored", () => {
  it("when bet.accepted arrives for a bet the store doesn't track then it is dropped", () => {
    const store = new MyBetsStore();
    const ok = store.applyAccepted({
      bet_id: "missing",
      accepted_at: "2026-05-16T10:00:01Z",
    });
    expect(ok).toBe(false);
    expect(store.selectAll()).toEqual([]);
  });
});

// =================== applyRejected ===================

describe("M14 applyRejected: pending → rejected with reason captured", () => {
  it("when bet.rejected arrives for a pending bet then state moves to rejected and reason is stored", () => {
    const store = new MyBetsStore();
    store.seedPending(PENDING_SEED);
    const ok = store.applyRejected({
      bet_id: "b-1",
      code: "BET_REJECTED_PRICE_CHANGED",
      message: "odds changed",
      at: "2026-05-16T10:00:02Z",
    });
    expect(ok).toBe(true);
    const bet = store.selectById("b-1");
    expect(bet?.state).toBe("rejected");
    expect(bet?.history[1]).toEqual({
      from: "pending",
      to: "rejected",
      at: "2026-05-16T10:00:02Z",
      reason: { code: "BET_REJECTED_PRICE_CHANGED", message: "odds changed" },
    });
  });
});

// =================== applyStateChanged ===================

describe("M14 applyStateChanged: generic FSM transition with append-only history", () => {
  it("when bet.state_changed advances accepted → settled then the store updates and history records the move", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot());
    const ok = store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
      reason: { code: "SETTLED_WIN", message: "settled" },
    });
    expect(ok).toBe(true);
    expect(store.selectById("b-1")?.state).toBe("settled");
    expect(store.selectHistory("b-1")).toHaveLength(3);
    expect(store.selectHistory("b-1")[2]).toEqual({
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
      reason: { code: "SETTLED_WIN", message: "settled" },
    });
  });
});

describe("M14 applyStateChanged: replay with field-equal transition is a no-op", () => {
  it("when bet.state_changed replays the same transition at the same time then the store skips it", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot());
    store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
    });
    const ok = store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
    });
    expect(ok).toBe(false);
    expect(store.selectHistory("b-1")).toHaveLength(3);
  });
});

describe("M14 applyStateChanged: stale `from` is rejected", () => {
  it("when bet.state_changed's `from` does not match the current state then the transition is dropped", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot({ state: "settled" }));
    const ok = store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "cancelled",
      at: "2026-05-16T12:00:00Z",
    });
    expect(ok).toBe(false);
    expect(store.selectById("b-1")?.state).toBe("settled");
  });
});

// =================== Rollback ===================

describe("M14 applyStateChanged: rollback settled → accepted is allowed", () => {
  it("when bet.state_changed rolls a settled bet back to accepted then the store records the rollback", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot({ state: "settled" }));
    const ok = store.applyStateChanged({
      bet_id: "b-1",
      from: "settled",
      to: "accepted",
      at: "2026-05-16T12:00:00Z",
      reason: { code: "ROLLBACK", message: "settlement rolled back" },
    });
    expect(ok).toBe(true);
    expect(store.selectById("b-1")?.state).toBe("accepted");
    const history = store.selectHistory("b-1");
    expect(history[history.length - 1]).toEqual({
      from: "settled",
      to: "accepted",
      at: "2026-05-16T12:00:00Z",
      reason: { code: "ROLLBACK", message: "settlement rolled back" },
    });
  });
});

// =================== Selectors ===================

describe("M14 selectors: per-state filtering", () => {
  it("when bets span multiple states then selectByState returns only the matching subset", () => {
    const store = new MyBetsStore();
    store.upsertFromSnapshot(snapshot({ bet_id: "b-1", state: "accepted" }));
    store.upsertFromSnapshot(snapshot({ bet_id: "b-2", state: "settled" }));
    store.upsertFromSnapshot(snapshot({ bet_id: "b-3", state: "accepted" }));
    const accepted = store.selectByState("accepted");
    expect(accepted.map((b) => b.bet_id).sort()).toEqual(["b-1", "b-3"]);
    expect(store.selectByState("settled").map((b) => b.bet_id)).toEqual(["b-2"]);
    expect(store.selectByState("rejected")).toEqual([]);
  });
});

describe("M14 selectors: selectHistory returns the append-only chain", () => {
  it("when a bet has transitioned multiple times then selectHistory returns the full chain in order", () => {
    const store = new MyBetsStore();
    store.seedPending(PENDING_SEED);
    store.applyAccepted({ bet_id: "b-1", accepted_at: "2026-05-16T10:00:01Z" });
    store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
    });
    const history = store.selectHistory("b-1");
    expect(history.map((t) => `${t.from}→${t.to}`)).toEqual([
      "-→pending",
      "pending→accepted",
      "accepted→settled",
    ]);
  });
});

// =================== Listener notifications ===================

describe("M14 listeners: notified only on actual mutations", () => {
  it("when transitions actually change state the listener fires; duplicates do not notify", () => {
    const store = new MyBetsStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.seedPending(PENDING_SEED);
    expect(listener).toHaveBeenCalledTimes(1);

    store.seedPending(PENDING_SEED);
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyAccepted({ bet_id: "b-1", accepted_at: "2026-05-16T10:00:01Z" });
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyAccepted({ bet_id: "b-1", accepted_at: "2026-05-16T10:00:01Z" });
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyAccepted({ bet_id: "missing", accepted_at: "x" });
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "settled",
      at: "2026-05-16T11:00:00Z",
    });
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyStateChanged({
      bet_id: "b-1",
      from: "accepted",
      to: "cancelled",
      at: "2026-05-16T11:30:00Z",
    });
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyStateChanged({
      bet_id: "b-1",
      from: "settled",
      to: "accepted",
      at: "2026-05-16T12:00:00Z",
      reason: { code: "ROLLBACK", message: "rolled back" },
    });
    expect(listener).toHaveBeenCalledTimes(4);
  });
});
