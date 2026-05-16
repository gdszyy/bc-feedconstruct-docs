// frontend/src/wiring/dispatchToStores.test.ts

import { describe, expect, it, vi } from "vitest";

import type { Envelope } from "@/contract/events";
import { Dispatcher } from "@/dispatch/dispatcher";

import { wireDispatcher, type WiredStores } from "./dispatchToStores";

function envelope(
  type: string,
  payload: unknown,
  overrides: Partial<Envelope> = {},
): Envelope {
  return {
    event_id: overrides.event_id ?? `evt-${type}-${Math.random()}`,
    type,
    occurred_at: overrides.occurred_at ?? "2026-05-16T08:00:00Z",
    correlation_id: overrides.correlation_id ?? "c-1",
    payload,
    ...overrides,
  } as Envelope;
}

function stubCatalog() {
  return {
    applySportUpserted: vi.fn(),
    applySportRemoved: vi.fn(),
    applyTournamentUpserted: vi.fn(),
    applyTournamentRemoved: vi.fn(),
  };
}

function stubMatch() {
  return {
    applyMatchUpserted: vi.fn(),
    applyMatchStatusChanged: vi.fn(),
  };
}

function stubMarkets() {
  return {
    applyOddsChanged: vi.fn(),
    applyMarketStatusChanged: vi.fn(),
  };
}

function stubBetStop() {
  return {
    applyApplied: vi.fn(),
    applyLifted: vi.fn(),
  };
}

function stubRollback() {
  return {
    recordSettlementRollback: vi.fn(),
    recordCancelRollback: vi.fn(),
  };
}

function stubSubscription() {
  return { applySubscriptionChanged: vi.fn() };
}

function stubBetSlip() {
  return {
    applyBetAccepted: vi.fn(),
    applyBetRejected: vi.fn(),
    applyOddsChanged: vi.fn(),
    applyMarketStatusChanged: vi.fn(),
    applyBetStopApplied: vi.fn(),
    applyBetStopLifted: vi.fn(),
  };
}

function stubMyBets() {
  return {
    applyBetAccepted: vi.fn(),
    applyBetRejected: vi.fn(),
    applyBetStateChanged: vi.fn(),
  };
}

function stubHealth() {
  return { applyProducerStatus: vi.fn() };
}

// Given a Dispatcher and a partial WiredStores bundle
// When wireDispatcher(dispatcher, stores) is invoked
// Then handler registrations are returned via an unsub closure
describe("wireDispatcher: returns an unsub closure", () => {
  it("when wireDispatcher is invoked then a function is returned", () => {
    const dispatcher = new Dispatcher();
    const unsub = wireDispatcher(dispatcher, {});
    expect(typeof unsub).toBe("function");
  });
});

// Given a Dispatcher + CatalogStore
// When dispatcher receives a sport.upserted envelope
// Then catalog.applySportUpserted is called with the payload
describe("wireDispatcher: catalog wiring (sport / tournament events)", () => {
  it("when catalog events arrive then the matching catalog reducers fire", () => {
    const dispatcher = new Dispatcher();
    const catalog = stubCatalog();
    wireDispatcher(dispatcher, { catalog } as WiredStores);

    const sportUp = { sport_id: "s1", name_translations: { en: "Soccer" } };
    const sportRm = { sport_id: "s1" };
    const tournUp = {
      tournament_id: "t1",
      sport_id: "s1",
      name_translations: { en: "EPL" },
    };
    const tournRm = { tournament_id: "t1" };

    dispatcher.dispatch(envelope("sport.upserted", sportUp));
    dispatcher.dispatch(envelope("sport.removed", sportRm));
    dispatcher.dispatch(envelope("tournament.upserted", tournUp));
    dispatcher.dispatch(envelope("tournament.removed", tournRm));

    expect(catalog.applySportUpserted).toHaveBeenCalledWith(sportUp);
    expect(catalog.applySportRemoved).toHaveBeenCalledWith(sportRm);
    expect(catalog.applyTournamentUpserted).toHaveBeenCalledWith(tournUp);
    expect(catalog.applyTournamentRemoved).toHaveBeenCalledWith(tournRm);
  });
});

// Given a Dispatcher + MatchStore
// When match.upserted / match.status_changed envelopes arrive
// Then match.applyMatchUpserted / applyMatchStatusChanged are called
describe("wireDispatcher: match wiring", () => {
  it("when match events arrive then the match reducers fire", () => {
    const dispatcher = new Dispatcher();
    const match = stubMatch();
    wireDispatcher(dispatcher, { match } as WiredStores);

    const upserted = { match_id: "m1", version: 1 };
    const statusChanged = { match_id: "m1", version: 2, status: "Live" };
    dispatcher.dispatch(envelope("match.upserted", upserted));
    dispatcher.dispatch(envelope("match.status_changed", statusChanged));

    expect(match.applyMatchUpserted).toHaveBeenCalledWith(upserted);
    expect(match.applyMatchStatusChanged).toHaveBeenCalledWith(statusChanged);
  });
});

// Given a Dispatcher + MarketStore
// When odds.changed and market.status_changed arrive
// Then market.applyOddsChanged / applyMarketStatusChanged are called
describe("wireDispatcher: markets wiring (odds + status)", () => {
  it("when odds / market.status events arrive then the markets reducers fire", () => {
    const dispatcher = new Dispatcher();
    const markets = stubMarkets();
    wireDispatcher(dispatcher, { markets } as WiredStores);

    const odds = {
      match_id: "m1",
      market_id: "mk1",
      version: 1,
      outcomes: [],
    };
    const status = {
      match_id: "m1",
      market_id: "mk1",
      version: 2,
      status: "Active",
    };
    dispatcher.dispatch(envelope("odds.changed", odds));
    dispatcher.dispatch(envelope("market.status_changed", status));

    expect(markets.applyOddsChanged).toHaveBeenCalledWith(odds);
    expect(markets.applyMarketStatusChanged).toHaveBeenCalledWith(status);
  });
});

// Given a Dispatcher + BetStopStore
// When bet_stop.applied and bet_stop.lifted arrive
// Then betStop.applyApplied / applyLifted are called
describe("wireDispatcher: betStop wiring", () => {
  it("when bet_stop events arrive then the betStop reducers fire", () => {
    const dispatcher = new Dispatcher();
    const betStop = stubBetStop();
    wireDispatcher(dispatcher, { betStop } as WiredStores);

    const applied = { match_id: "m1", reason: "manual", started_at: "t" };
    const lifted = { match_id: "m1", lifted_at: "t" };
    dispatcher.dispatch(envelope("bet_stop.applied", applied));
    dispatcher.dispatch(envelope("bet_stop.lifted", lifted));

    expect(betStop.applyApplied).toHaveBeenCalledWith(applied);
    expect(betStop.applyLifted).toHaveBeenCalledWith(lifted);
  });
});

// Given a Dispatcher + SettlementStore + RollbackHistoryStore
// When bet_settlement.applied, bet_settlement.rolled_back, bet_cancel.applied,
// bet_cancel.rolled_back arrive
// Then settlement reducers fire AND rolled_back events ALSO record into RollbackHistoryStore
describe("wireDispatcher: settlement + rollback wiring", () => {
  it("when settlement events arrive then settlement reducers fire and rollback history is recorded", () => {
    const dispatcher = new Dispatcher();
    const settlement = {
      applyBetSettlementApplied: vi.fn(),
      applyBetSettlementRolledBack: vi.fn(),
    };
    const cancel = {
      applyBetCancelApplied: vi.fn(),
      applyBetCancelRolledBack: vi.fn(),
    };
    const rollback = stubRollback();
    wireDispatcher(dispatcher, { settlement, cancel, rollback } as WiredStores);

    const setApplied = {
      match_id: "m1",
      market_id: "mk1",
      version: 1,
      outcomes: [],
    };
    const setRolledBack = { match_id: "m1", market_id: "mk1", version: 1 };
    const cancelApplied = { match_id: "m1", market_id: "mk1", version: 1 };
    const cancelRolledBack = { match_id: "m1", market_id: "mk1" };

    dispatcher.dispatch(envelope("bet_settlement.applied", setApplied));
    dispatcher.dispatch(
      envelope("bet_settlement.rolled_back", setRolledBack),
    );
    dispatcher.dispatch(envelope("bet_cancel.applied", cancelApplied));
    dispatcher.dispatch(envelope("bet_cancel.rolled_back", cancelRolledBack));

    expect(settlement.applyBetSettlementApplied).toHaveBeenCalledWith(
      setApplied,
    );
    expect(settlement.applyBetSettlementRolledBack).toHaveBeenCalledWith(
      setRolledBack,
    );
    expect(cancel.applyBetCancelApplied).toHaveBeenCalledWith(cancelApplied);
    expect(cancel.applyBetCancelRolledBack).toHaveBeenCalledWith(
      cancelRolledBack,
    );

    expect(rollback.recordSettlementRollback).toHaveBeenCalledWith(
      setRolledBack,
    );
    expect(rollback.recordCancelRollback).toHaveBeenCalledWith(
      cancelRolledBack,
    );
  });
});

// Given a Dispatcher + SubscriptionStore
// When subscription.changed arrives
// Then subscription.applySubscriptionChanged is called
describe("wireDispatcher: subscription wiring", () => {
  it("when subscription.changed arrives then the subscription reducer fires", () => {
    const dispatcher = new Dispatcher();
    const subscription = stubSubscription();
    wireDispatcher(dispatcher, { subscription } as WiredStores);

    const payload = {
      subscription_id: "sub1",
      match_id: "m1",
      state: "active",
    };
    dispatcher.dispatch(envelope("subscription.changed", payload));

    expect(subscription.applySubscriptionChanged).toHaveBeenCalledWith(
      payload,
    );
  });
});

// Given a Dispatcher + BetSlipStore + MyBetsStore
// When bet.accepted arrives
// Then BOTH betSlip.applyBetAccepted AND myBets.applyBetAccepted fire
describe("wireDispatcher: bet.accepted fans out to slip + my-bets", () => {
  it("when bet.accepted arrives then both slip and my-bets reducers fire (independent handlers)", () => {
    const dispatcher = new Dispatcher();
    const betSlip = stubBetSlip();
    const myBets = stubMyBets();
    wireDispatcher(dispatcher, { betSlip, myBets } as WiredStores);

    const payload = { bet_id: "bet-1", placed_at: "t" };
    dispatcher.dispatch(envelope("bet.accepted", payload));

    expect(betSlip.applyBetAccepted).toHaveBeenCalledWith(payload);
    expect(myBets.applyBetAccepted).toHaveBeenCalledWith(
      expect.objectContaining({ type: "bet.accepted", payload }),
    );
  });
});

// Given a Dispatcher + BetSlipStore + MyBetsStore
// When bet.rejected arrives
// Then BOTH betSlip.applyBetRejected AND myBets.applyBetRejected fire
describe("wireDispatcher: bet.rejected fans out to slip + my-bets", () => {
  it("when bet.rejected arrives then both slip and my-bets reducers fire (independent handlers)", () => {
    const dispatcher = new Dispatcher();
    const betSlip = stubBetSlip();
    const myBets = stubMyBets();
    wireDispatcher(dispatcher, { betSlip, myBets } as WiredStores);

    const payload = { reason: "PRICE_CHANGED", message: "moved" };
    dispatcher.dispatch(envelope("bet.rejected", payload));

    expect(betSlip.applyBetRejected).toHaveBeenCalledWith(payload);
    expect(myBets.applyBetRejected).toHaveBeenCalledWith(
      expect.objectContaining({ type: "bet.rejected", payload }),
    );
  });
});

// Given a Dispatcher + MyBetsStore
// When bet.state_changed arrives
// Then myBets.applyBetStateChanged is called (slip ignores this — it's already Submitted)
describe("wireDispatcher: bet.state_changed only routes to my-bets", () => {
  it("when bet.state_changed arrives then only my-bets reducer fires", () => {
    const dispatcher = new Dispatcher();
    const betSlip = stubBetSlip();
    const myBets = stubMyBets();
    wireDispatcher(dispatcher, { betSlip, myBets } as WiredStores);

    const payload = { bet_id: "bet-1", state: "Settled", changed_at: "t" };
    dispatcher.dispatch(envelope("bet.state_changed", payload));

    expect(myBets.applyBetStateChanged).toHaveBeenCalledWith(
      expect.objectContaining({ type: "bet.state_changed", payload }),
    );
    // slip has no .applyBetStateChanged
  });
});

// Given a Dispatcher + BetSlipStore
// When odds.changed / market.status_changed / bet_stop arrive
// Then the slip availability reducers fire
describe("wireDispatcher: bet slip availability gating", () => {
  it("when odds / market.status / bet_stop events arrive then betSlip availability reducers fire", () => {
    const dispatcher = new Dispatcher();
    const betSlip = stubBetSlip();
    wireDispatcher(dispatcher, { betSlip } as WiredStores);

    const odds = {
      match_id: "m1",
      market_id: "mk1",
      version: 1,
      outcomes: [],
    };
    const status = {
      match_id: "m1",
      market_id: "mk1",
      version: 2,
      status: "Suspended",
    };
    const stopApplied = { match_id: "m1", reason: "x", started_at: "t" };
    const stopLifted = { match_id: "m1", lifted_at: "t" };
    dispatcher.dispatch(envelope("odds.changed", odds));
    dispatcher.dispatch(envelope("market.status_changed", status));
    dispatcher.dispatch(envelope("bet_stop.applied", stopApplied));
    dispatcher.dispatch(envelope("bet_stop.lifted", stopLifted));

    expect(betSlip.applyOddsChanged).toHaveBeenCalledWith(odds);
    expect(betSlip.applyMarketStatusChanged).toHaveBeenCalledWith(status);
    expect(betSlip.applyBetStopApplied).toHaveBeenCalledWith(stopApplied);
    expect(betSlip.applyBetStopLifted).toHaveBeenCalledWith(stopLifted);
  });
});

// Given a Dispatcher + HealthStore
// When system.producer_status arrives
// Then health.applyProducerStatus is called
describe("wireDispatcher: health producer status wiring", () => {
  it("when system.producer_status arrives then the health reducer fires", () => {
    const dispatcher = new Dispatcher();
    const health = stubHealth();
    wireDispatcher(dispatcher, { health } as WiredStores);

    const payload = {
      product: "live",
      is_down: true,
      last_message_at: "t",
      down_since: "t-1",
    };
    dispatcher.dispatch(envelope("system.producer_status", payload));

    expect(health.applyProducerStatus).toHaveBeenCalledWith(payload);
  });
});

// Given a Dispatcher + a stores bundle without a particular store
// When events for the missing store arrive
// Then no handler runs and no error is thrown
describe("wireDispatcher: missing stores are skipped", () => {
  it("when a store is absent from the bundle then events targeting it are silently ignored", () => {
    const dispatcher = new Dispatcher();
    const recordUnknownType = vi.fn();
    const dispatcherWithTelemetry = new Dispatcher({
      telemetry: {
        recordUnknownType,
        recordHandlerError: vi.fn(),
        recordDuplicate: vi.fn(),
        recordStale: vi.fn(),
      },
    });
    // wire only catalog
    wireDispatcher(dispatcherWithTelemetry, {
      catalog: stubCatalog(),
    } as WiredStores);

    expect(() =>
      dispatcherWithTelemetry.dispatch(
        envelope("system.producer_status", {
          product: "live",
          is_down: false,
          last_message_at: "t",
        }),
      ),
    ).not.toThrow();
    // No handler for that type → dispatcher records it as unknown
    expect(recordUnknownType).toHaveBeenCalledTimes(1);

    // Also: wiring with an empty bundle must not throw
    expect(() => wireDispatcher(dispatcher, {})).not.toThrow();
  });
});

// Given the unsub function returned by wireDispatcher
// When unsub is invoked
// Then subsequent envelopes no longer reach the stores
describe("wireDispatcher: unsub stops further routing", () => {
  it("when the unsub closure is invoked then further envelopes do not reach any reducer", () => {
    const dispatcher = new Dispatcher();
    const match = stubMatch();
    const unsub = wireDispatcher(dispatcher, { match } as WiredStores);

    dispatcher.dispatch(
      envelope("match.upserted", { match_id: "m1", version: 1 }),
    );
    expect(match.applyMatchUpserted).toHaveBeenCalledTimes(1);

    unsub();
    dispatcher.dispatch(
      envelope("match.upserted", { match_id: "m2", version: 1 }),
    );
    expect(match.applyMatchUpserted).toHaveBeenCalledTimes(1);
  });
});
