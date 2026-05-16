import { describe, expect, it, vi } from "vitest";

import { SubscriptionStore } from "./store";

// ---------------------------------------------------------------------------
// M11 — SubscriptionStore
//
// Locked decisions (PR thread):
//
//   FSM: absent → booking → subscribed → unbooking → released
//        booking ─fail→ failed,  unbooking ─fail→ failed
//        released | failed ─markBooking→ booking (retry / re-subscribe)
//
//   Local intents drive direct transitions:
//     markBooking()       absent | released | failed → booking
//     markBooked()        booking                    → subscribed
//     markBookFailed()    booking                    → failed
//     markUnbooking()     subscribed                 → unbooking
//     markUnbooked()      unbooking                  → released
//     markUnbookFailed()  unbooking                  → failed
//
//   Server-driven reducer applies subscription.changed:
//     state=active     absent | booking → subscribed
//     state=active     subscribed       → no-op (idempotent)
//     state=active     other            → no-op (no FSM regress)
//     state=released   any non-released → released
//     state=cancelled  any non-released → released (folded; per M11 acceptance)
//     state=released | cancelled  on absent → no-op (we don't track ghosts)
//
//   Idempotency mirrors M08 / M09:
//     - Disallowed transitions are dropped silently (return false)
//     - Listeners fire only when state actually changes
// ---------------------------------------------------------------------------

// =================== Empty store ===================

// Given an empty SubscriptionStore
// When selectByMatch / selectAll / selectFailed are queried
// Then they return undefined / [] / [] respectively
describe("M11 subscription baseline: empty store yields nothing", () => {
  it("when no subscription has been recorded then selectByMatch is undefined and selectAll is []", () => {
    const store = new SubscriptionStore();
    expect(store.selectByMatch("sr:match:1")).toBeUndefined();
    expect(store.selectAll()).toEqual([]);
    expect(store.selectFailed()).toEqual([]);
  });
});

// =================== Local intent: book ===================

// Given an empty store
// When markBooking(matchId, at: 1000) is invoked
// Then selectByMatch returns {match_id, status: 'booking', last_transition_at: 1000} and listener fires
describe("M11 booking intent: idle → booking", () => {
  it("when markBooking is invoked on an unknown match then the entry becomes booking", () => {
    const store = new SubscriptionStore();
    const ok = store.markBooking("sr:match:1", 1000);
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      status: "booking",
      last_transition_at: 1000,
    });
  });
});

// Given an entry already in 'booking'
// When markBooking is invoked again with a later timestamp
// Then the call returns false, last_transition_at stays at the original, listener does not fire
describe("M11 booking intent: duplicate markBooking is a no-op", () => {
  it("when markBooking is invoked twice in a row then the duplicate is suppressed and last_transition_at is preserved", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    const ok = store.markBooking("sr:match:1", 2000);
    expect(ok).toBe(false);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      status: "booking",
      last_transition_at: 1000,
    });
  });
});

// Given an entry already in 'subscribed'
// When markBooking is invoked
// Then the call returns false; status remains 'subscribed'
describe("M11 booking intent: subscribed entries reject re-book", () => {
  it("when markBooking is invoked on a subscribed entry then the FSM does not regress", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);
    const ok = store.markBooking("sr:match:1", 2000);
    expect(ok).toBe(false);
    const entry = store.selectByMatch("sr:match:1");
    expect(entry?.status).toBe("subscribed");
    expect(entry?.subscription_id).toBe("srv-sub-1");
    expect(entry?.last_transition_at).toBe(1500);
  });
});

// Given an entry in 'released' (terminal from a prior lifecycle)
// When markBooking is invoked
// Then the entry transitions back to 'booking' (re-subscribe is permitted)
describe("M11 booking intent: released entries can be re-booked", () => {
  it("when markBooking is invoked on a released entry then the FSM re-enters booking", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);
    store.markUnbooking("sr:match:1", 2000);
    store.markUnbooked("sr:match:1", 2500);

    expect(store.selectByMatch("sr:match:1")?.status).toBe("released");

    const ok = store.markBooking("sr:match:1", 3000);
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      status: "booking",
      last_transition_at: 3000,
    });
  });
});

// =================== Local intent: book ack / fail ===================

// Given an entry in 'booking'
// When markBooked(subscriptionId, at) is invoked
// Then status becomes 'subscribed', subscription_id is recorded, listener fires
describe("M11 book ack: booking → subscribed", () => {
  it("when markBooked is invoked while booking then the entry becomes subscribed and records subscription_id", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    const ok = store.markBooked("srv-sub-1", "sr:match:1", 1500);
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "subscribed",
      last_transition_at: 1500,
    });
  });
});

// Given an entry NOT in 'booking' (e.g. idle, subscribed, released)
// When markBooked is invoked
// Then the call returns false (illegal transition is dropped); state unchanged
describe("M11 book ack: illegal transitions are dropped", () => {
  it("when markBooked is invoked outside the booking state then the FSM does not move", () => {
    const store = new SubscriptionStore();
    // absent
    expect(store.markBooked("srv-sub-1", "sr:match:1", 1000)).toBe(false);
    expect(store.selectByMatch("sr:match:1")).toBeUndefined();

    // subscribed
    store.markBooking("sr:match:1", 1100);
    store.markBooked("srv-sub-1", "sr:match:1", 1200);
    expect(store.markBooked("srv-sub-2", "sr:match:1", 1300)).toBe(false);
    expect(store.selectByMatch("sr:match:1")?.subscription_id).toBe("srv-sub-1");
    expect(store.selectByMatch("sr:match:1")?.last_transition_at).toBe(1200);
  });
});

// Given an entry in 'booking'
// When markBookFailed({code, message}, at) is invoked
// Then status becomes 'failed' with last_error populated; listener fires
describe("M11 book fail: booking → failed with last_error", () => {
  it("when markBookFailed is invoked while booking then the entry becomes failed and stores last_error", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    const ok = store.markBookFailed(
      "sr:match:1",
      { code: "RATE_LIMIT", message: "too many" },
      1500,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      status: "failed",
      last_transition_at: 1500,
      last_error: { code: "RATE_LIMIT", message: "too many" },
    });
    expect(store.selectFailed()).toHaveLength(1);
  });
});

// Given an entry in 'failed' from a prior book attempt
// When markBooking is invoked again
// Then the entry transitions back to 'booking' and last_error is cleared
describe("M11 book retry: failed → booking clears last_error", () => {
  it("when markBooking re-attempts after a failure then last_error is dropped and status returns to booking", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBookFailed("sr:match:1", { code: "RATE_LIMIT" }, 1500);

    const ok = store.markBooking("sr:match:1", 2000);
    expect(ok).toBe(true);
    const entry = store.selectByMatch("sr:match:1");
    expect(entry?.status).toBe("booking");
    expect(entry?.last_transition_at).toBe(2000);
    expect(entry?.last_error).toBeUndefined();
    expect(store.selectFailed()).toEqual([]);
  });
});

// =================== Local intent: unbook ===================

// Given an entry in 'subscribed'
// When markUnbooking(at) is invoked
// Then status becomes 'unbooking'; listener fires
describe("M11 unbook intent: subscribed → unbooking", () => {
  it("when markUnbooking is invoked while subscribed then the entry transitions to unbooking", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);

    const ok = store.markUnbooking("sr:match:1", 2000);
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "unbooking",
      last_transition_at: 2000,
    });
  });
});

// Given an entry NOT in 'subscribed'
// When markUnbooking is invoked
// Then the call returns false; status unchanged
describe("M11 unbook intent: illegal transitions are dropped", () => {
  it("when markUnbooking is invoked outside the subscribed state then the FSM does not move", () => {
    const store = new SubscriptionStore();
    // absent
    expect(store.markUnbooking("sr:match:1", 1000)).toBe(false);

    // booking
    store.markBooking("sr:match:1", 1100);
    expect(store.markUnbooking("sr:match:1", 1200)).toBe(false);
    expect(store.selectByMatch("sr:match:1")?.status).toBe("booking");
  });
});

// Given an entry in 'unbooking'
// When markUnbooked(at) is invoked
// Then status becomes 'released' (terminal); listener fires
describe("M11 unbook ack: unbooking → released", () => {
  it("when markUnbooked is invoked while unbooking then the entry becomes released", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);
    store.markUnbooking("sr:match:1", 2000);

    const ok = store.markUnbooked("sr:match:1", 2500);
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "released",
      last_transition_at: 2500,
    });
  });
});

// Given an entry in 'unbooking'
// When markUnbookFailed({code, message}, at) is invoked
// Then status becomes 'failed' with last_error populated; listener fires
describe("M11 unbook fail: unbooking → failed with last_error", () => {
  it("when markUnbookFailed is invoked while unbooking then the entry becomes failed and stores last_error", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);
    store.markUnbooking("sr:match:1", 2000);

    const ok = store.markUnbookFailed(
      "sr:match:1",
      { code: "TIMEOUT", message: "no response" },
      2500,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "failed",
      last_transition_at: 2500,
      last_error: { code: "TIMEOUT", message: "no response" },
    });
  });
});

// =================== Server reducer: subscription.changed ===================

// Given an empty store
// When applyServerChange({subscription_id, match_id, state: 'active'}, at) is invoked
// Then the store synthesizes an entry in 'subscribed' with the server-issued subscription_id
describe("M11 server reducer: synth subscribed on first 'active'", () => {
  it("when subscription.changed with state=active arrives for an unknown match then a subscribed entry is synthesized", () => {
    const store = new SubscriptionStore();
    const ok = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "active",
      },
      1000,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "subscribed",
      last_transition_at: 1000,
    });
  });
});

// Given an entry in 'booking' (local intent in flight)
// When applyServerChange({state: 'active'}, at) is invoked
// Then the entry transitions to 'subscribed' and records subscription_id
describe("M11 server reducer: 'active' confirms booking", () => {
  it("when subscription.changed with state=active arrives during booking then the entry becomes subscribed", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    const ok = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "active",
      },
      1500,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "subscribed",
      last_transition_at: 1500,
    });
  });
});

// Given an entry already in 'subscribed' with the same subscription_id
// When applyServerChange({state: 'active'}, at) is invoked again
// Then the call is dropped (field-equal idempotency); listener does not fire
describe("M11 server reducer: duplicate 'active' is a no-op", () => {
  it("when subscription.changed with state=active arrives twice for the same id then the second is suppressed", () => {
    const store = new SubscriptionStore();
    store.applyServerChange(
      { subscription_id: "srv-sub-1", match_id: "sr:match:1", state: "active" },
      1000,
    );
    const ok = store.applyServerChange(
      { subscription_id: "srv-sub-1", match_id: "sr:match:1", state: "active" },
      2000,
    );
    expect(ok).toBe(false);
    expect(store.selectByMatch("sr:match:1")?.last_transition_at).toBe(1000);
  });
});

// Given an entry in 'subscribed'
// When applyServerChange({state: 'released'}, at) is invoked
// Then the entry transitions to 'released' (terminal); listener fires
describe("M11 server reducer: 'released' moves subscribed → released", () => {
  it("when subscription.changed with state=released arrives for a subscribed entry then it becomes released", () => {
    const store = new SubscriptionStore();
    store.applyServerChange(
      { subscription_id: "srv-sub-1", match_id: "sr:match:1", state: "active" },
      1000,
    );
    const ok = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      2000,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")).toEqual({
      match_id: "sr:match:1",
      subscription_id: "srv-sub-1",
      status: "released",
      last_transition_at: 2000,
    });
  });
});

// Given an entry in 'unbooking' (local intent in flight)
// When applyServerChange({state: 'released'}, at) is invoked
// Then the entry transitions to 'released'; listener fires
describe("M11 server reducer: 'released' confirms unbooking", () => {
  it("when subscription.changed with state=released arrives during unbooking then the entry becomes released", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);
    store.markUnbooking("sr:match:1", 2000);

    const ok = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      2500,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")?.status).toBe("released");
    expect(store.selectByMatch("sr:match:1")?.last_transition_at).toBe(2500);
  });
});

// Given an entry in 'subscribed'
// When applyServerChange({state: 'cancelled'}, at) is invoked
// Then the entry transitions to 'released' (terminal; cancelled and released collapse to the same UI state per M11 acceptance)
describe("M11 server reducer: 'cancelled' folds to released", () => {
  it("when subscription.changed with state=cancelled arrives then the entry becomes released", () => {
    const store = new SubscriptionStore();
    store.applyServerChange(
      { subscription_id: "srv-sub-1", match_id: "sr:match:1", state: "active" },
      1000,
    );
    const ok = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "cancelled",
      },
      2000,
    );
    expect(ok).toBe(true);
    expect(store.selectByMatch("sr:match:1")?.status).toBe("released");
  });
});

// Given an entry already in 'released'
// When applyServerChange({state: 'released' | 'cancelled'}) arrives again
// Then it's a no-op; listener does not fire
describe("M11 server reducer: duplicate terminal apply is suppressed", () => {
  it("when a released entry receives another released/cancelled event then no listener fires", () => {
    const store = new SubscriptionStore();
    store.applyServerChange(
      { subscription_id: "srv-sub-1", match_id: "sr:match:1", state: "active" },
      1000,
    );
    store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      2000,
    );

    const okReleased = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      3000,
    );
    const okCancelled = store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "cancelled",
      },
      4000,
    );
    expect(okReleased).toBe(false);
    expect(okCancelled).toBe(false);
    expect(store.selectByMatch("sr:match:1")?.last_transition_at).toBe(2000);
  });
});

// =================== Selectors ===================

// Given mixed entries across multiple matches and statuses
// When selectAll() / selectFailed() / selectByMatch() are queried
// Then each returns the expected subset
describe("M11 selectors: per-match and per-status views", () => {
  it("when entries span multiple statuses then selectFailed returns only failed; selectByMatch returns the targeted entry", () => {
    const store = new SubscriptionStore();
    // match:1 — subscribed
    store.markBooking("sr:match:1", 1000);
    store.markBooked("srv-sub-1", "sr:match:1", 1100);
    // match:2 — booking
    store.markBooking("sr:match:2", 1200);
    // match:3 — failed
    store.markBooking("sr:match:3", 1300);
    store.markBookFailed("sr:match:3", { code: "RATE_LIMIT" }, 1400);

    expect(store.selectAll()).toHaveLength(3);
    expect(store.selectFailed().map((e) => e.match_id)).toEqual(["sr:match:3"]);
    expect(store.selectByMatch("sr:match:2")?.status).toBe("booking");
  });
});

// =================== Cross-match isolation ===================

// Given entries for match:1 and match:2
// When a transition is applied to match:1
// Then match:2 remains untouched
describe("M11 cross-match isolation", () => {
  it("when one match transitions then sibling matches are not affected", () => {
    const store = new SubscriptionStore();
    store.markBooking("sr:match:1", 1000);
    store.markBooking("sr:match:2", 1100);
    store.markBooked("srv-sub-1", "sr:match:1", 1500);

    expect(store.selectByMatch("sr:match:1")?.status).toBe("subscribed");
    expect(store.selectByMatch("sr:match:2")?.status).toBe("booking");
    expect(store.selectByMatch("sr:match:2")?.last_transition_at).toBe(1100);
  });
});

// =================== Listener notifications ===================

// Given a subscribed listener
// When transitions actually change state, listener fires once per change
// And when duplicates / illegal transitions are dropped, listener does NOT fire
describe("M11 listeners: notified only on actual transitions", () => {
  it("when transitions change state the listener fires; idempotent or illegal calls do not notify", () => {
    const store = new SubscriptionStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.markBooking("sr:match:1", 1000);
    expect(listener).toHaveBeenCalledTimes(1);

    store.markBooking("sr:match:1", 2000); // duplicate
    expect(listener).toHaveBeenCalledTimes(1);

    store.markBooked("srv-sub-1", "sr:match:1", 3000);
    expect(listener).toHaveBeenCalledTimes(2);

    store.markBooked("srv-sub-2", "sr:match:1", 4000); // illegal: already subscribed
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "active",
      },
      5000,
    ); // duplicate
    expect(listener).toHaveBeenCalledTimes(2);

    store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      6000,
    );
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyServerChange(
      {
        subscription_id: "srv-sub-1",
        match_id: "sr:match:1",
        state: "released",
      },
      7000,
    ); // duplicate terminal
    expect(listener).toHaveBeenCalledTimes(3);
  });
});
