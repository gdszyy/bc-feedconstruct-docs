// frontend/src/myBets/store.test.ts
//
// M14 — MyBetsStore: bet lifecycle index + append-only history reducer.
// Locked decisions per PR thread:
//   4. Unknown-bet on wire event: create record from event payload, origin='server-pushed'.
//   + From-state mismatch: reject + fire telemetry.fromStateMismatch().
//   + Idempotency (P6): each bet tracks applied event_ids; duplicates dropped.
//   + Pending → bet_id promotion: trackPending(idempotency_key, ...) keys by
//     idempotency_key, re-keyed under bet_id on applyBetAccepted.
//   + Append-only history (P9): rollback transitions appended, never overwrite.

import { describe, expect, it, vi } from "vitest";

import type {
  BetAcceptedPayload,
  BetRejectedPayload,
  BetStateChangedPayload,
  Envelope,
} from "@/contract/events";
import type { BetSelection, MyBet } from "@/contract/rest";

import { MyBetsStore } from "./store";

function envelope<T>(
  type: Envelope["type"],
  event_id: string,
  payload: T,
  overrides: Partial<Envelope<T>> = {},
): Envelope<T> {
  return {
    type,
    schema_version: "1",
    event_id,
    correlation_id: `corr-${event_id}`,
    product_id: "live",
    occurred_at: "2026-05-16T07:00:00Z",
    received_at: "2026-05-16T07:00:00Z",
    entity: {},
    payload,
    ...overrides,
  };
}

function makeBet(overrides: Partial<MyBet> = {}): MyBet {
  return {
    id: "bet-1",
    user_id: "u1",
    placed_at: "2026-05-16T06:00:00Z",
    stake: 10,
    currency: "USD",
    bet_type: "single",
    state: "Accepted",
    selections: [makeSelection()],
    history: [
      { at: "2026-05-16T06:00:00Z", from: "Pending", to: "Accepted" },
    ],
    ...overrides,
  };
}

function makeSelection(overrides: Partial<BetSelection> = {}): BetSelection {
  return {
    position: 1,
    match_id: "m1",
    market_id: "mk1",
    outcome_id: "home",
    locked_odds: 2.0,
    ...overrides,
  };
}

// =================== Baseline / hydration ===================

describe("M14 myBets baseline: empty", () => {
  it("when no bets have been added then list is empty", () => {
    const s = new MyBetsStore();
    expect(s.list()).toEqual([]);
    expect(s.selectById("anything")).toBeUndefined();
  });
});

describe("M14 myBets hydrate: REST snapshot", () => {
  it("when hydrateBets seeds the store then bets are retrievable by id with history intact", () => {
    const s = new MyBetsStore();
    s.hydrateBets([
      makeBet({ id: "bet-1", state: "Accepted" }),
      makeBet({
        id: "bet-2",
        state: "Settled",
        history: [
          { at: "t1", from: "Pending", to: "Accepted" },
          { at: "t2", from: "Accepted", to: "Settled" },
        ],
      }),
    ]);
    expect(s.list()).toHaveLength(2);
    const bet1 = s.selectById("bet-1");
    expect(bet1?.state).toBe("Accepted");
    expect(bet1?.history).toHaveLength(1);
    expect(bet1?.history[0]?.origin).toBe("rest");
    const bet2 = s.selectById("bet-2");
    expect(bet2?.history).toHaveLength(2);
  });
});

describe("M14 myBets selector: listByStatus filter", () => {
  it("when listByStatus is invoked then only matching bets are returned", () => {
    const s = new MyBetsStore();
    s.hydrateBets([
      makeBet({ id: "bet-1", state: "Accepted" }),
      makeBet({ id: "bet-2", state: "Settled" }),
      makeBet({ id: "bet-3", state: "Settled" }),
      makeBet({ id: "bet-4", state: "Cancelled" }),
    ]);
    expect(s.listByStatus("Settled")).toHaveLength(2);
    expect(s.listByStatus("Accepted")).toHaveLength(1);
    expect(s.listByStatus("Pending")).toHaveLength(0);
  });
});

// =================== bet.accepted ===================

describe("M14 myBets accept: known Pending bet → Accepted with history", () => {
  it("when bet.accepted arrives then state becomes Accepted and history records the transition", () => {
    const s = new MyBetsStore();
    s.trackPending({
      idempotency_key: "idem-1",
      bet_type: "single",
      stake: 10,
      currency: "USD",
      selections: [makeSelection()],
      user_id: "u1",
      placed_at: "2026-05-16T06:00:00Z",
    });
    const env = envelope<BetAcceptedPayload>(
      "bet.accepted",
      "evt-1",
      {
        bet_id: "bet-42",
        user_id: "u1",
        accepted_odds: 2.0,
        accepted_at: "2026-05-16T07:00:00Z",
      },
    );
    expect(s.applyBetAccepted(env)).toBe(true);
    // Re-keyed under bet_id; old key gone.
    expect(s.selectById("idem-1")).toBeUndefined();
    const record = s.selectById("bet-42");
    expect(record?.state).toBe("Accepted");
    expect(record?.history).toHaveLength(2);
    expect(record?.history[1]).toMatchObject({
      from: "Pending",
      to: "Accepted",
      event_id: "evt-1",
      origin: "server-pushed",
    });
  });
});

describe("M14 myBets accept: unknown bet creates Accepted record", () => {
  it("when bet.accepted arrives for an unknown bet then a new Accepted record is created from the event", () => {
    const s = new MyBetsStore();
    const env = envelope<BetAcceptedPayload>(
      "bet.accepted",
      "evt-1",
      {
        bet_id: "bet-99",
        user_id: "u2",
        accepted_odds: 1.5,
        accepted_at: "2026-05-16T07:00:00Z",
      },
    );
    expect(s.applyBetAccepted(env)).toBe(true);
    const record = s.selectById("bet-99");
    expect(record?.state).toBe("Accepted");
    expect(record?.user_id).toBe("u2");
    expect(record?.history).toEqual([
      expect.objectContaining({
        from: "",
        to: "Accepted",
        event_id: "evt-1",
        origin: "server-pushed",
      }),
    ]);
  });
});

// =================== bet.rejected ===================

describe("M14 myBets reject: known Pending bet → Rejected", () => {
  it("when bet.rejected arrives then state becomes Rejected with reason in history", () => {
    const s = new MyBetsStore();
    s.trackPending({
      idempotency_key: "idem-1",
      bet_type: "single",
      stake: 10,
      currency: "USD",
      selections: [makeSelection()],
      user_id: "u1",
    });
    const env = envelope<BetRejectedPayload>(
      "bet.rejected",
      "evt-1",
      {
        bet_id: "bet-42",
        user_id: "u1",
        code: "RISK_LIMIT",
        message: "exposure cap",
      },
    );
    expect(s.applyBetRejected(env)).toBe(true);
    const record = s.selectById("bet-42");
    expect(record?.state).toBe("Rejected");
    expect(record?.history.at(-1)?.reason).toBe("exposure cap");
  });
});

// =================== bet.state_changed ===================

describe("M14 myBets transition: Accepted → Settled", () => {
  it("when bet.state_changed Accepted→Settled arrives then state becomes Settled", () => {
    const s = new MyBetsStore();
    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    const env = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-2",
      {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Settled",
        at: "2026-05-16T07:00:00Z",
      },
    );
    expect(s.applyBetStateChanged(env)).toBe(true);
    expect(s.selectById("bet-1")?.state).toBe("Settled");
  });
});

describe("M14 myBets rollback: Settled → Accepted preserves history", () => {
  it("when a rollback transition arrives then state reverts but history is append-only", () => {
    const s = new MyBetsStore();
    s.hydrateBets([
      makeBet({
        id: "bet-1",
        state: "Settled",
        history: [
          { at: "t1", from: "Pending", to: "Accepted" },
          { at: "t2", from: "Accepted", to: "Settled" },
        ],
      }),
    ]);
    const rollback = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-rb",
      {
        bet_id: "bet-1",
        from: "Settled",
        to: "Accepted",
        at: "t3",
        reason: "settlement.rolled_back",
      },
    );
    expect(s.applyBetStateChanged(rollback)).toBe(true);
    const record = s.selectById("bet-1")!;
    expect(record.state).toBe("Accepted");
    expect(record.history).toHaveLength(3);
    // Original Settled transition preserved.
    expect(record.history[1]).toMatchObject({
      from: "Accepted",
      to: "Settled",
    });
    // Rollback appended on top.
    expect(record.history[2]).toMatchObject({
      from: "Settled",
      to: "Accepted",
      reason: "settlement.rolled_back",
      origin: "server-pushed",
    });
  });
});

describe("M14 myBets transition: Accepted → Cancelled records void_reason", () => {
  it("when bet is cancelled then state becomes Cancelled and the reason is in history", () => {
    const s = new MyBetsStore();
    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    const env = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-c",
      {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Cancelled",
        at: "t9",
        reason: "match_abandoned",
      },
    );
    expect(s.applyBetStateChanged(env)).toBe(true);
    const record = s.selectById("bet-1")!;
    expect(record.state).toBe("Cancelled");
    expect(record.history.at(-1)?.reason).toBe("match_abandoned");
  });
});

describe("M14 myBets transition: from-mismatch rejected with telemetry", () => {
  it("when an inbound transition's from-state does not match the current state then it is rejected and telemetry fires", () => {
    const telemetry = { fromStateMismatch: vi.fn() };
    const s = new MyBetsStore({ telemetry });
    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    const env = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-bad",
      {
        bet_id: "bet-1",
        from: "Pending",
        to: "Settled",
        at: "t1",
      },
    );
    expect(s.applyBetStateChanged(env)).toBe(false);
    expect(s.selectById("bet-1")?.state).toBe("Accepted");
    expect(telemetry.fromStateMismatch).toHaveBeenCalledWith({
      bet_id: "bet-1",
      expected_from: "Pending",
      current: "Accepted",
      event_id: "evt-bad",
    });
  });
});

// =================== Idempotency (P6) ===================

describe("M14 myBets idempotency: duplicate event_id ignored", () => {
  it("when the same event_id is applied twice then the second is dropped (no double history entry)", () => {
    const s = new MyBetsStore();
    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    const env = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-dup",
      {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Settled",
        at: "t1",
      },
    );
    expect(s.applyBetStateChanged(env)).toBe(true);
    const before = s.selectById("bet-1")!.history.length;
    expect(s.applyBetStateChanged(env)).toBe(false);
    expect(s.selectById("bet-1")!.history.length).toBe(before);
  });
});

// =================== Append-only history ===================

describe("M14 myBets append-only history: ordered transitions", () => {
  it("when multiple transitions occur then all are preserved in order", () => {
    const s = new MyBetsStore();
    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    s.applyBetStateChanged(
      envelope<BetStateChangedPayload>("bet.state_changed", "e1", {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Settled",
        at: "t1",
      }),
    );
    s.applyBetStateChanged(
      envelope<BetStateChangedPayload>("bet.state_changed", "e2", {
        bet_id: "bet-1",
        from: "Settled",
        to: "Accepted",
        at: "t2",
        reason: "rollback",
      }),
    );
    s.applyBetStateChanged(
      envelope<BetStateChangedPayload>("bet.state_changed", "e3", {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Settled",
        at: "t3",
      }),
    );
    const record = s.selectById("bet-1")!;
    expect(record.history.map((t) => `${t.from}→${t.to}`)).toEqual([
      "Pending→Accepted",
      "Accepted→Settled",
      "Settled→Accepted",
      "Accepted→Settled",
    ]);
  });
});

// =================== Cross-module link: slip → my-bets ===================

describe("M14 myBets link: trackPending registers a pending bet", () => {
  it("when trackPending is invoked then the bet is indexed in Pending state keyed by idempotency_key", () => {
    const s = new MyBetsStore();
    expect(
      s.trackPending({
        idempotency_key: "idem-77",
        bet_type: "single",
        stake: 20,
        currency: "EUR",
        selections: [makeSelection()],
      }),
    ).toBe(true);
    const record = s.selectById("idem-77");
    expect(record?.state).toBe("Pending");
    expect(record?.stake).toBe(20);
    expect(record?.history.at(-1)?.origin).toBe("local");
  });

  it("when trackPending is invoked with the same idempotency_key twice then the second call is a no-op", () => {
    const s = new MyBetsStore();
    const args = {
      idempotency_key: "idem-77",
      bet_type: "single",
      stake: 20,
      currency: "EUR",
      selections: [makeSelection()],
    };
    expect(s.trackPending(args)).toBe(true);
    expect(s.trackPending(args)).toBe(false);
  });
});

describe("M14 myBets link: idempotency_key promoted to bet_id on accept", () => {
  it("when bet.accepted arrives then a pending record is re-keyed under bet_id", () => {
    const s = new MyBetsStore();
    s.trackPending({
      idempotency_key: "idem-77",
      bet_type: "single",
      stake: 20,
      currency: "EUR",
      selections: [makeSelection()],
      user_id: "u1",
    });
    expect(s.selectById("idem-77")?.state).toBe("Pending");
    s.applyBetAccepted(
      envelope<BetAcceptedPayload>("bet.accepted", "evt-a", {
        bet_id: "bet-real-123",
        user_id: "u1",
        accepted_odds: 2.0,
        accepted_at: "t",
      }),
    );
    expect(s.selectById("idem-77")).toBeUndefined();
    const record = s.selectById("bet-real-123");
    expect(record?.state).toBe("Accepted");
    // Selections from trackPending should still be present.
    expect(record?.selections).toHaveLength(1);
  });
});

// =================== Listeners ===================

describe("M14 myBets listeners: fire only on real mutations", () => {
  it("when state actually changes the listener fires; duplicate events do not notify", () => {
    const s = new MyBetsStore();
    const listener = vi.fn();
    s.subscribe(listener);

    s.hydrateBets([makeBet({ id: "bet-1", state: "Accepted" })]);
    expect(listener).toHaveBeenCalledTimes(1);

    const env = envelope<BetStateChangedPayload>(
      "bet.state_changed",
      "evt-x",
      {
        bet_id: "bet-1",
        from: "Accepted",
        to: "Settled",
        at: "t1",
      },
    );
    s.applyBetStateChanged(env);
    expect(listener).toHaveBeenCalledTimes(2);

    // Duplicate envelope — no notify.
    s.applyBetStateChanged(env);
    expect(listener).toHaveBeenCalledTimes(2);
  });
});
