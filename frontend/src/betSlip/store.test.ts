import { describe, expect, it, vi } from "vitest";

import { BetSlipStore } from "./store";

// ---------------------------------------------------------------------------
// M13 — BetSlipStore
//
// Locked decisions (PR thread):
//   FSM: empty → editing → ready → submitting → submitted
//        ready ⇄ needs_review on price drift (acceptPriceChanges → editing)
//        rejected returns to editing with reasons (caller re-validates +
//        re-places with a fresh idempotency key)
//
//   Mutating intents (return true iff state actually changed):
//     addSelection            empty → editing; demotes ready/needs_review
//     removeSelection         → empty if last; else demotes ready/needs_review
//     setStake/setBetType/setCurrency  demote ready/needs_review → editing
//     markValidated           editing → ready
//     markValidationFailed    editing | ready | needs_review → editing
//     applyOddsChange         ready → needs_review when locked_odds drifts
//     acceptPriceChanges      needs_review → editing; locked_odds refreshed
//     place                   ready → submitting
//     markAccepted            submitting → submitted with bet_id
//     markRejected            submitting → editing with reasons
//     discard                 any → empty (clears everything)
//
//   Idempotency mirrors M08/M09/M11:
//     - duplicate selections / no-op sets / illegal transitions return false
//     - listeners fire only when state actually changes
// ---------------------------------------------------------------------------

const SEL_A = {
  match_id: "sr:match:1",
  market_id: "sr:market:1",
  outcome_id: "1",
  locked_odds: 1.85,
};
const SEL_B = {
  match_id: "sr:match:2",
  market_id: "sr:market:18",
  outcome_id: "2",
  locked_odds: 2.1,
};

function readied(): BetSlipStore {
  const store = new BetSlipStore();
  store.addSelection(SEL_A);
  store.setStake(10);
  store.setCurrency("EUR");
  store.setBetType("single");
  store.markValidated();
  return store;
}

// =================== Empty baseline ===================

// Given a freshly constructed BetSlipStore
// When selectors are queried
// Then state='empty'; selections=[]; stake=0; reasons/priceChanges are empty
describe("M13 baseline: empty slip", () => {
  it("when the store is created then state is 'empty' with no selections", () => {
    const store = new BetSlipStore();
    expect(store.selectState()).toBe("empty");
    expect(store.selectSelections()).toEqual([]);
    expect(store.selectStake()).toBe(0);
    expect(store.selectBetType()).toBe("single");
    expect(store.selectReasons()).toEqual([]);
    expect(store.selectPriceChanges()).toEqual([]);
    expect(store.selectIdempotencyKey()).toBeUndefined();
    expect(store.selectBetId()).toBeUndefined();
  });
});

// =================== addSelection ===================

// Given an empty slip
// When addSelection(selection) is invoked
// Then state becomes 'editing' and the selection is present
describe("M13 add: empty → editing on first selection", () => {
  it("when the first selection is added then state moves to editing", () => {
    const store = new BetSlipStore();
    const ok = store.addSelection(SEL_A);
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectSelections()).toEqual([SEL_A]);
  });
});

// Given a slip already containing a selection (matchId, marketId, outcomeId)
// When the same selection key is added again
// Then the call returns false (dedup) and selections are unchanged
describe("M13 add: duplicate selection is a no-op", () => {
  it("when the same outcome is added twice then the slip rejects the duplicate", () => {
    const store = new BetSlipStore();
    store.addSelection(SEL_A);
    const ok = store.addSelection({ ...SEL_A, locked_odds: 1.95 });
    expect(ok).toBe(false);
    expect(store.selectSelections()).toEqual([SEL_A]);
  });
});

// Given an empty slip
// When two selections from different markets are added
// Then both are present in the order they were added
describe("M13 add: multiple distinct selections coexist", () => {
  it("when distinct outcomes are added then all appear in the selections list", () => {
    const store = new BetSlipStore();
    store.addSelection(SEL_A);
    store.addSelection(SEL_B);
    expect(store.selectSelections()).toEqual([SEL_A, SEL_B]);
  });
});

// =================== removeSelection ===================

// Given a slip with two selections
// When one selection is removed
// Then state remains 'editing' and only that selection is gone
describe("M13 remove: drops a selection but stays editing", () => {
  it("when one of multiple selections is removed then state stays editing", () => {
    const store = new BetSlipStore();
    store.addSelection(SEL_A);
    store.addSelection(SEL_B);
    const ok = store.removeSelection(SEL_A.match_id, SEL_A.market_id, SEL_A.outcome_id);
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectSelections()).toEqual([SEL_B]);
  });
});

// Given a slip with exactly one selection
// When that selection is removed
// Then state transitions back to 'empty' and selections=[]
describe("M13 remove: last selection drops back to empty", () => {
  it("when the final selection is removed then state returns to empty", () => {
    const store = new BetSlipStore();
    store.addSelection(SEL_A);
    const ok = store.removeSelection(SEL_A.match_id, SEL_A.market_id, SEL_A.outcome_id);
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("empty");
    expect(store.selectSelections()).toEqual([]);
  });
});

// =================== Field changes demote ready ===================

// Given a slip in 'ready'
// When setStake is invoked with a different stake value
// Then state demotes back to 'editing' (revalidation required)
describe("M13 setStake: changing stake demotes ready → editing", () => {
  it("when stake changes on a validated slip then the slip returns to editing", () => {
    const store = readied();
    expect(store.selectState()).toBe("ready");
    const ok = store.setStake(25);
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectStake()).toBe(25);
  });
});

// Given a slip in 'ready'
// When setBetType is invoked with a different bet type
// Then state demotes back to 'editing'
describe("M13 setBetType: changing bet type demotes ready → editing", () => {
  it("when bet type changes on a validated slip then the slip returns to editing", () => {
    const store = readied();
    const ok = store.setBetType("multiple");
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectBetType()).toBe("multiple");
  });
});

// =================== Validation ===================

// Given a slip in 'editing'
// When markValidated() is invoked (optionally with currentOdds)
// Then state becomes 'ready'; reasons are cleared
describe("M13 markValidated: editing → ready", () => {
  it("when validation succeeds then the slip enters ready and reasons are cleared", () => {
    const store = new BetSlipStore();
    store.addSelection(SEL_A);
    store.markValidationFailed([{ code: "STALE_ODDS", message: "stale" }]);
    expect(store.selectReasons()).toHaveLength(1);
    const ok = store.markValidated();
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("ready");
    expect(store.selectReasons()).toEqual([]);
  });
});

// Given a slip in 'editing' / 'ready' / 'needs_review'
// When markValidationFailed(reasons) is invoked
// Then state is 'editing' with reasons populated
describe("M13 markValidationFailed: stores reasons and stays/moves to editing", () => {
  it("when validation fails then reasons are stored and state is editing", () => {
    const store = readied();
    const ok = store.markValidationFailed([
      { code: "LIMIT_EXCEEDED", message: "stake too high", selection_position: 1 },
    ]);
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectReasons()).toEqual([
      { code: "LIMIT_EXCEEDED", message: "stake too high", selection_position: 1 },
    ]);
  });
});

// =================== Price change handling ===================

// Given a slip in 'ready' containing selection(outcomeId, lockedOdds=1.85)
// When applyOddsChange arrives with outcomeId at odds=2.10
// Then state transitions to 'needs_review' and priceChanges captures {from:1.85, to:2.10}
describe("M13 applyOddsChange: ready → needs_review on drift", () => {
  it("when a selection's odds drift on a ready slip then state enters needs_review with priceChanges", () => {
    const store = readied();
    const ok = store.applyOddsChange({
      match_id: SEL_A.match_id,
      market_id: SEL_A.market_id,
      outcomes: [{ outcome_id: SEL_A.outcome_id, odds: 2.1, active: true }],
      version: 1,
    });
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("needs_review");
    expect(store.selectPriceChanges()).toEqual([
      {
        position: 1,
        match_id: SEL_A.match_id,
        market_id: SEL_A.market_id,
        outcome_id: SEL_A.outcome_id,
        from: 1.85,
        to: 2.1,
      },
    ]);
  });
});

// Given a slip in 'ready'
// When applyOddsChange arrives where every outcome's odds match lockedOdds
// Then no transition occurs (returns false); listener does not fire
describe("M13 applyOddsChange: no drift is a no-op", () => {
  it("when odds match lockedOdds then the slip stays ready", () => {
    const store = readied();
    const ok = store.applyOddsChange({
      match_id: SEL_A.match_id,
      market_id: SEL_A.market_id,
      outcomes: [{ outcome_id: SEL_A.outcome_id, odds: 1.85, active: true }],
      version: 1,
    });
    expect(ok).toBe(false);
    expect(store.selectState()).toBe("ready");
    expect(store.selectPriceChanges()).toEqual([]);
  });
});

// Given a slip in 'needs_review' with priceChanges populated
// When acceptPriceChanges() is invoked
// Then state goes back to 'editing'; lockedOdds is refreshed; priceChanges cleared
describe("M13 acceptPriceChanges: needs_review → editing with refreshed lockedOdds", () => {
  it("when the user accepts the new prices then lockedOdds are updated and revalidation is required", () => {
    const store = readied();
    store.applyOddsChange({
      match_id: SEL_A.match_id,
      market_id: SEL_A.market_id,
      outcomes: [{ outcome_id: SEL_A.outcome_id, odds: 2.1, active: true }],
      version: 1,
    });

    const ok = store.acceptPriceChanges();
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectPriceChanges()).toEqual([]);
    expect(store.selectSelections()[0].locked_odds).toBe(2.1);
  });
});

// =================== Place ===================

// Given a slip in 'ready'
// When place('idem-1') is invoked
// Then state becomes 'submitting' and idempotencyKey='idem-1' is recorded
describe("M13 place: ready → submitting with idempotency key", () => {
  it("when place() is called on a ready slip then the slip enters submitting and stores the key", () => {
    const store = readied();
    const ok = store.place("idem-1");
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("submitting");
    expect(store.selectIdempotencyKey()).toBe("idem-1");
  });
});

// Given a slip in any non-'ready' state (e.g. editing, needs_review)
// When place(...) is invoked
// Then the call returns false; state unchanged
describe("M13 place: disallowed outside ready", () => {
  it("when place() is called outside ready then the FSM does not move", () => {
    const store = new BetSlipStore();
    expect(store.place("idem-1")).toBe(false);
    store.addSelection(SEL_A);
    expect(store.selectState()).toBe("editing");
    expect(store.place("idem-1")).toBe(false);
    expect(store.selectState()).toBe("editing");
    expect(store.selectIdempotencyKey()).toBeUndefined();
  });
});

// =================== Place result ===================

// Given a slip in 'submitting' with idempotencyKey
// When markAccepted({bet_id:'b1'}) is invoked
// Then state becomes 'submitted'; bet_id='b1' is recorded
describe("M13 markAccepted: submitting → submitted with bet_id", () => {
  it("when the BFF accepts the bet then the slip records bet_id and moves to submitted", () => {
    const store = readied();
    store.place("idem-1");
    const ok = store.markAccepted({ bet_id: "b-1" });
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("submitted");
    expect(store.selectBetId()).toBe("b-1");
  });
});

// Given a slip in 'submitting'
// When markRejected({code:'BET_REJECTED_*', message:'...'}) is invoked
// Then state returns to 'editing' with reasons populated
describe("M13 markRejected: submitting → editing with reasons", () => {
  it("when the BFF rejects the bet then the slip drops back to editing and surfaces reasons", () => {
    const store = readied();
    store.place("idem-1");
    const ok = store.markRejected({
      code: "BET_REJECTED_PRICE_CHANGED",
      message: "odds changed",
    });
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("editing");
    expect(store.selectReasons()).toEqual([
      { code: "BET_REJECTED_PRICE_CHANGED", message: "odds changed" },
    ]);
    expect(store.selectIdempotencyKey()).toBeUndefined();
  });
});

// =================== Discard ===================

// Given a slip with selections in any non-empty state
// When discard() is invoked
// Then state becomes 'empty'; selections / stake / reasons / bet_id / idempotency_key are all cleared
describe("M13 discard: any state → empty, wipes payload", () => {
  it("when the slip is discarded then everything is reset to the empty baseline", () => {
    const store = readied();
    store.place("idem-1");
    store.markAccepted({ bet_id: "b-1" });

    const ok = store.discard();
    expect(ok).toBe(true);
    expect(store.selectState()).toBe("empty");
    expect(store.selectSelections()).toEqual([]);
    expect(store.selectStake()).toBe(0);
    expect(store.selectCurrency()).toBe("");
    expect(store.selectBetType()).toBe("single");
    expect(store.selectReasons()).toEqual([]);
    expect(store.selectPriceChanges()).toEqual([]);
    expect(store.selectIdempotencyKey()).toBeUndefined();
    expect(store.selectBetId()).toBeUndefined();
  });
});

// =================== Listener notifications ===================

// Given a subscribed listener
// When transitions actually change state, listener fires once per change
// And duplicate / illegal calls are dropped silently and listener does NOT fire
describe("M13 listeners: notified only on actual transitions", () => {
  it("when transitions change state the listener fires; duplicate or illegal calls do not notify", () => {
    const store = new BetSlipStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.addSelection(SEL_A);
    expect(listener).toHaveBeenCalledTimes(1);

    store.addSelection({ ...SEL_A, locked_odds: 1.95 }); // duplicate key
    expect(listener).toHaveBeenCalledTimes(1);

    store.setStake(10);
    expect(listener).toHaveBeenCalledTimes(2);

    store.setStake(10); // same value
    expect(listener).toHaveBeenCalledTimes(2);

    store.markValidated();
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyOddsChange({
      match_id: SEL_A.match_id,
      market_id: SEL_A.market_id,
      outcomes: [{ outcome_id: SEL_A.outcome_id, odds: 1.85, active: true }],
      version: 1,
    }); // no drift
    expect(listener).toHaveBeenCalledTimes(3);

    store.applyOddsChange({
      match_id: SEL_A.match_id,
      market_id: SEL_A.market_id,
      outcomes: [{ outcome_id: SEL_A.outcome_id, odds: 2.1, active: true }],
      version: 2,
    });
    expect(listener).toHaveBeenCalledTimes(4);

    store.place("idem-1"); // illegal: state is needs_review
    expect(listener).toHaveBeenCalledTimes(4);

    store.discard();
    expect(listener).toHaveBeenCalledTimes(5);

    store.discard(); // already default
    expect(listener).toHaveBeenCalledTimes(5);
  });
});
