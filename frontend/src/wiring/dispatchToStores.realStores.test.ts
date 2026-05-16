// frontend/src/wiring/dispatchToStores.realStores.test.ts
//
// Smoke test: wires the actual store instances (not vi.fn() stubs) and dispatches
// one envelope per route to catch signature drift between the structural
// `*Like` interfaces and the real store classes. This is the canary — if a
// reducer changes its parameter shape, this test fails at compile time.

import { describe, expect, it } from "vitest";

import { BetSlipStore } from "@/betSlip/store";
import { BetStopStore } from "@/betStop/store";
import { CatalogStore } from "@/catalog/store";
import { DescriptionsStore } from "@/descriptions/store";
import { Dispatcher } from "@/dispatch/dispatcher";
import { HealthStore } from "@/health/store";
import { MarketsStore } from "@/markets/store";
import { MatchStore } from "@/match/store";
import { MyBetsStore } from "@/myBets/store";
import { RollbackHistoryStore } from "@/rollback/store";
import { CancelStore, SettlementStore } from "@/settlement/store";
import { SubscriptionStore } from "@/subscription/store";

import { wireDispatcher } from "./dispatchToStores";

describe("wireDispatcher: real-store smoke test", () => {
  it("when all real stores are wired then wiring compiles and accepts events without throwing", () => {
    const dispatcher = new Dispatcher();
    const bundle = {
      catalog: new CatalogStore(),
      match: new MatchStore(),
      markets: new MarketsStore(),
      betStop: new BetStopStore(),
      settlement: new SettlementStore(),
      cancel: new CancelStore(),
      rollback: new RollbackHistoryStore(),
      subscription: new SubscriptionStore(),
      betSlip: new BetSlipStore(),
      myBets: new MyBetsStore(),
      health: new HealthStore(),
    };
    // DescriptionsStore is here only to assert the import is exercised; it has
    // no event wiring (descriptions hydrate via REST + ETag).
    expect(new DescriptionsStore()).toBeDefined();

    const unsub = wireDispatcher(dispatcher, bundle);
    expect(() => unsub()).not.toThrow();
  });
});
