import { describe, expect, it, vi } from "vitest";

import {
  type SubscriptionTelemetry,
  SubscriptionStore,
} from "./store";

// ---------------------------------------------------------------------------
// M11 — SubscriptionStore
//
// FSM (docs/.../04_state_machines.md §5):
//   Idle ──book──▶ Booking ──ok──▶ Subscribed ──unbook──▶ Unbooking ──ok──▶ Released
//                    │                                       │
//                    └───fail──▶ Failed                       └───fail──▶ Failed
//
// Locked decisions (per PR thread):
//   - Wire mapping: active→Subscribed, released→Released, cancelled→Released
//     with cancelled=true.
//   - Keying per-match: Released → requestBook returns to Booking.
//   - retryFromFailed is the only path out of Failed.
//   - active wire over Released is illegal → telemetry.illegalTransition.
//   - active wire over Idle / Booking / Unbooking / Failed → Subscribed
//     (server is canonical).
// ---------------------------------------------------------------------------

function recordingTelemetry(): SubscriptionTelemetry & {
  bookFailed: ReturnType<typeof vi.fn>;
  unbookFailed: ReturnType<typeof vi.fn>;
  illegalTransition: ReturnType<typeof vi.fn>;
} {
  return {
    bookFailed: vi.fn(),
    unbookFailed: vi.fn(),
    illegalTransition: vi.fn(),
  };
}

// =================== Empty store ===================

// Given a brand-new SubscriptionStore
// When selectByMatch is called before any mutation
// Then it returns undefined (the match is in implicit Idle state)
describe("M11 subscription baseline: empty store", () => {
  it("when no subscription exists for a match then selectByMatch returns undefined", () => {
    const store = new SubscriptionStore();
    expect(store.selectByMatch("sr:match:1")).toBeUndefined();
  });
});

// =================== requestBook → Booking ===================

// Given an empty SubscriptionStore
// When requestBook(matchId) is invoked
// Then a record exists with state=Booking, subscription_id=undefined (no server ack yet),
//      and a listener subscribed beforehand fires exactly once
describe("M11 subscription requestBook: Idle → Booking", () => {
  it("when requestBook is invoked from Idle then the record transitions to Booking with no subscription_id yet", () => {
    const store = new SubscriptionStore();
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.requestBook("sr:match:1", 1_000);
    expect(ok).toBe(true);
    expect(listener).toHaveBeenCalledTimes(1);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Booking");
    expect(rec.subscription_id).toBeUndefined();
    expect(rec.last_transition_at).toBe(1_000);
  });
});

// Given a match already in Subscribed
// When requestBook(matchId) is invoked again
// Then the call is a no-op (state stays Subscribed; no listener fires)
describe("M11 subscription requestBook: no-op when already Subscribed", () => {
  it("when requestBook targets an already-Subscribed match then it is a no-op", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.requestBook("sr:match:1", 3_000);
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
    expect(store.selectByMatch("sr:match:1")?.state).toBe("Subscribed");
  });
});

// =================== ack subscription.changed(active) ===================

// Given a match in state=Booking
// When subscription.changed(state=active, subscription_id=S1) is applied
// Then state transitions to Subscribed, subscription_id=S1, listener fires once
describe("M11 subscription ack: Booking → Subscribed on wire 'active'", () => {
  it("when subscription.changed(active) arrives during Booking then state becomes Subscribed", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    expect(ok).toBe(true);
    expect(listener).toHaveBeenCalledTimes(1);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Subscribed");
    expect(rec.subscription_id).toBe("S1");
    expect(rec.last_transition_at).toBe(2_000);
  });
});

// Given a match in state=Idle (no requestBook has fired locally)
// When subscription.changed(state=active, subscription_id=S1) arrives (e.g. server-initiated booking)
// Then the record is created in Subscribed with subscription_id=S1
describe("M11 subscription ack: server-initiated booking creates Subscribed", () => {
  it("when subscription.changed(active) arrives without a local Booking then a Subscribed record is created", () => {
    const store = new SubscriptionStore();
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S9", match_id: "sr:match:1", state: "active" },
      1_000,
    );
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Subscribed");
    expect(rec.subscription_id).toBe("S9");
  });
});

// =================== reportBookFailed → Failed ===================

// Given a match in state=Booking
// When reportBookFailed(matchId, reason) is invoked
// Then state transitions to Failed with failedKind='book', last_error=reason;
//      injected telemetry.bookFailed is invoked exactly once
describe("M11 subscription book fail: Booking → Failed + telemetry", () => {
  it("when reportBookFailed is invoked from Booking then state becomes Failed and telemetry fires", () => {
    const telemetry = recordingTelemetry();
    const store = new SubscriptionStore({ telemetry });
    store.requestBook("sr:match:1", 1_000);
    const ok = store.reportBookFailed("sr:match:1", "rate_limited", 2_000);
    expect(ok).toBe(true);
    expect(telemetry.bookFailed).toHaveBeenCalledTimes(1);
    expect(telemetry.unbookFailed).not.toHaveBeenCalled();
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Failed");
    expect(rec.failed_kind).toBe("book");
    expect(rec.last_error).toBe("rate_limited");
  });
});

// =================== retryFromFailed ===================

// Given a match in state=Failed with failedKind='book'
// When retryFromFailed(matchId) is invoked
// Then state transitions Failed → Booking; last_error cleared
describe("M11 subscription retry: Failed(book) → Booking", () => {
  it("when retryFromFailed is invoked after a book failure then the FSM returns to Booking", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.reportBookFailed("sr:match:1", "rate_limited", 2_000);
    const ok = store.retryFromFailed("sr:match:1", 3_000);
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Booking");
    expect(rec.failed_kind).toBeUndefined();
    expect(rec.last_error).toBeUndefined();
    expect(rec.last_transition_at).toBe(3_000);
  });
});

// Given a match in state=Failed with failedKind='unbook'
// When retryFromFailed(matchId) is invoked
// Then state transitions Failed → Unbooking; last_error cleared and subscription_id preserved
describe("M11 subscription retry: Failed(unbook) → Unbooking", () => {
  it("when retryFromFailed is invoked after an unbook failure then the FSM returns to Unbooking", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    store.requestUnbook("sr:match:1", 3_000);
    store.reportUnbookFailed("sr:match:1", "network", 4_000);

    const ok = store.retryFromFailed("sr:match:1", 5_000);
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Unbooking");
    expect(rec.subscription_id).toBe("S1");
    expect(rec.last_error).toBeUndefined();
  });
});

// Given a match NOT in Failed (e.g. Subscribed)
// When retryFromFailed(matchId) is invoked
// Then the call is a no-op
describe("M11 subscription retry: only allowed from Failed", () => {
  it("when retryFromFailed is invoked from a non-Failed state then it is a no-op", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    const listener = vi.fn();
    store.subscribe(listener);
    const ok = store.retryFromFailed("sr:match:1", 3_000);
    expect(ok).toBe(false);
    expect(listener).not.toHaveBeenCalled();
    expect(store.selectByMatch("sr:match:1")?.state).toBe("Subscribed");
  });
});

// =================== requestUnbook → Unbooking ===================

// Given a match in state=Subscribed
// When requestUnbook(matchId) is invoked
// Then state transitions to Unbooking; subscription_id preserved
describe("M11 subscription requestUnbook: Subscribed → Unbooking", () => {
  it("when requestUnbook is invoked from Subscribed then the FSM transitions to Unbooking", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    const ok = store.requestUnbook("sr:match:1", 3_000);
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Unbooking");
    expect(rec.subscription_id).toBe("S1");
  });
});

// Given a match in state=Idle / Booking / Released
// When requestUnbook(matchId) is invoked
// Then the call is a no-op
describe("M11 subscription requestUnbook: no-op outside Subscribed", () => {
  it("when requestUnbook is invoked from a non-Subscribed state then it is a no-op", () => {
    const store = new SubscriptionStore();
    expect(store.requestUnbook("sr:match:idle", 1_000)).toBe(false);

    store.requestBook("sr:match:booking", 1_000);
    expect(store.requestUnbook("sr:match:booking", 2_000)).toBe(false);
    expect(store.selectByMatch("sr:match:booking")?.state).toBe("Booking");

    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:released", state: "released" },
      3_000,
    );
    expect(store.requestUnbook("sr:match:released", 4_000)).toBe(false);
    expect(store.selectByMatch("sr:match:released")?.state).toBe("Released");
  });
});

// =================== ack subscription.changed(released) ===================

// Given a match in state=Unbooking
// When subscription.changed(state=released) arrives
// Then state transitions Unbooking → Released; listener fires once
describe("M11 subscription ack: Unbooking → Released on wire 'released'", () => {
  it("when subscription.changed(released) arrives during Unbooking then state becomes Released", () => {
    const store = new SubscriptionStore();
    store.requestBook("sr:match:1", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    store.requestUnbook("sr:match:1", 3_000);
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "released" },
      4_000,
    );
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Released");
    expect(rec.cancelled).toBe(false);
  });
});

// Given a match in state=Subscribed (no local unbook fired)
// When subscription.changed(state=released) arrives (e.g. match-end auto-release per doc)
// Then state transitions Subscribed → Released directly
describe("M11 subscription ack: Subscribed → Released on server-pushed match-end", () => {
  it("when subscription.changed(released) arrives during Subscribed then state becomes Released (auto-release)", () => {
    const store = new SubscriptionStore();
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      1_000,
    );
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "released" },
      2_000,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")?.state).toBe("Released");
  });
});

// =================== ack subscription.changed(cancelled) ===================

// Given a match in state=Subscribed
// When subscription.changed(state=cancelled) arrives
// Then state transitions to Released with cancelled=true; listener fires once
describe("M11 subscription ack: cancelled converges to Released with cancelled flag", () => {
  it("when subscription.changed(cancelled) arrives then state becomes Released with cancelled=true", () => {
    const store = new SubscriptionStore();
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      1_000,
    );
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "cancelled" },
      2_000,
    );
    expect(ok).toBe(true);
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Released");
    expect(rec.cancelled).toBe(true);
  });
});

// =================== reportUnbookFailed → Failed ===================

// Given a match in state=Unbooking
// When reportUnbookFailed(matchId, reason) is invoked
// Then state transitions to Failed with failedKind='unbook', last_error=reason;
//      injected telemetry.unbookFailed is invoked exactly once
describe("M11 subscription unbook fail: Unbooking → Failed + telemetry", () => {
  it("when reportUnbookFailed is invoked from Unbooking then state becomes Failed and telemetry fires", () => {
    const telemetry = recordingTelemetry();
    const store = new SubscriptionStore({ telemetry });
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      1_000,
    );
    store.requestUnbook("sr:match:1", 2_000);
    const ok = store.reportUnbookFailed("sr:match:1", "5xx", 3_000);
    expect(ok).toBe(true);
    expect(telemetry.unbookFailed).toHaveBeenCalledTimes(1);
    expect(telemetry.bookFailed).not.toHaveBeenCalled();
    const rec = store.selectByMatch("sr:match:1")!;
    expect(rec.state).toBe("Failed");
    expect(rec.failed_kind).toBe("unbook");
    expect(rec.last_error).toBe("5xx");
    expect(rec.subscription_id).toBe("S1");
  });
});

// =================== Illegal wire transitions ===================

// Given a match in state=Released (terminal)
// When subscription.changed(state=active) arrives (illegal: terminal → live)
// Then state remains Released; telemetry.illegalTransition is invoked exactly once
describe("M11 subscription illegal wire: active over Released → telemetry", () => {
  it("when subscription.changed(active) arrives over a Released record then it is rejected with telemetry", () => {
    const telemetry = recordingTelemetry();
    const store = new SubscriptionStore({ telemetry });
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "released" },
      1_000,
    );
    const ok = store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    expect(ok).toBe(false);
    expect(telemetry.illegalTransition).toHaveBeenCalledTimes(1);
    expect(telemetry.illegalTransition.mock.calls[0][0]).toMatchObject({
      match_id: "sr:match:1",
      from: "Released",
      wire_state: "active",
    });
    expect(store.selectByMatch("sr:match:1")?.state).toBe("Released");
  });
});

// =================== Per-match isolation ===================

// Given subscriptions for match A and match B in different states
// When mutations target match A
// Then match B's record is untouched
describe("M11 subscription isolation: per-match independence", () => {
  it("when one match transitions then sibling matches are unaffected", () => {
    const store = new SubscriptionStore();
    store.requestBook("A", 1_000);
    store.applySubscriptionChanged(
      { subscription_id: "Sa", match_id: "A", state: "active" },
      2_000,
    );
    store.requestBook("B", 3_000);
    store.applySubscriptionChanged(
      { subscription_id: "Sb", match_id: "B", state: "active" },
      4_000,
    );
    store.requestUnbook("A", 5_000);

    expect(store.selectByMatch("A")?.state).toBe("Unbooking");
    expect(store.selectByMatch("B")?.state).toBe("Subscribed");
    expect(store.selectByMatch("B")?.subscription_id).toBe("Sb");
  });
});

// =================== Listeners ===================

// Given a subscribed listener
// When real transitions occur the listener fires once per transition;
//      no-op calls (e.g. requestBook over Subscribed) do NOT fire
describe("M11 subscription listeners: fire only on real transitions", () => {
  it("when subscription state actually changes the listener fires; otherwise it does not", () => {
    const store = new SubscriptionStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.requestBook("sr:match:1", 1_000);
    expect(listener).toHaveBeenCalledTimes(1);

    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      2_000,
    );
    expect(listener).toHaveBeenCalledTimes(2);

    // no-op: already Subscribed with same subscription_id
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "active" },
      3_000,
    );
    expect(listener).toHaveBeenCalledTimes(2);

    // no-op: requestBook over Subscribed
    store.requestBook("sr:match:1", 4_000);
    expect(listener).toHaveBeenCalledTimes(2);

    // real: Subscribed → Released (auto)
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "released" },
      5_000,
    );
    expect(listener).toHaveBeenCalledTimes(3);

    // no-op: already Released
    store.applySubscriptionChanged(
      { subscription_id: "S1", match_id: "sr:match:1", state: "released" },
      6_000,
    );
    expect(listener).toHaveBeenCalledTimes(3);
  });
});
