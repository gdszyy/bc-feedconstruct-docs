"use client";

import { createContext, useContext, useEffect, useMemo, type ReactNode } from "react";

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
import { FavoritesStore } from "@/subscription/favorites";
import { SubscriptionStore } from "@/subscription/store";
import { TelemetryStore, type TelemetryShipper } from "@/telemetry/store";
import { wireDispatcher } from "@/wiring/dispatchToStores";

// No-op TelemetryShipper used as the default. Production callers should pass
// a real shipper (e.g. createHttpTelemetryShipper) via createDefaultStores.
const noopShipper: TelemetryShipper = {
  async ship() {},
};

// ---------------------------------------------------------------------------
// StoresProvider — single mount point that
//   1. instantiates the full domain-store bundle (or accepts a pre-built one
//      via the `value` prop for tests),
//   2. wires Dispatcher → stores via wireDispatcher on mount,
//   3. exposes the bundle via context so descendant hooks can read it.
//
// Locked design: internal construction with `value` override. Production code
// uses <StoresProvider>{children}</StoresProvider> with no props; tests pass
// pre-built bundles to isolate behaviour.
// ---------------------------------------------------------------------------

export interface Stores {
  catalog: CatalogStore;
  match: MatchStore;
  markets: MarketsStore;
  betStop: BetStopStore;
  settlement: SettlementStore;
  cancel: CancelStore;
  rollback: RollbackHistoryStore;
  subscription: SubscriptionStore;
  favorites: FavoritesStore;
  descriptions: DescriptionsStore;
  betSlip: BetSlipStore;
  myBets: MyBetsStore;
  health: HealthStore;
  telemetry: TelemetryStore;
  dispatcher: Dispatcher;
}

export interface CreateDefaultStoresOptions {
  /**
   * TelemetryShipper used by TelemetryStore. Production should pass
   * `createHttpTelemetryShipper({ client })`. Defaults to a no-op shipper.
   */
  telemetryShipper?: TelemetryShipper;
}

export function createDefaultStores(
  opts: CreateDefaultStoresOptions = {},
): Stores {
  return {
    catalog: new CatalogStore(),
    match: new MatchStore(),
    markets: new MarketsStore(),
    betStop: new BetStopStore(),
    settlement: new SettlementStore(),
    cancel: new CancelStore(),
    rollback: new RollbackHistoryStore(),
    subscription: new SubscriptionStore(),
    favorites: new FavoritesStore(),
    descriptions: new DescriptionsStore(),
    betSlip: new BetSlipStore(),
    myBets: new MyBetsStore(),
    health: new HealthStore(),
    telemetry: new TelemetryStore({
      shipper: opts.telemetryShipper ?? noopShipper,
    }),
    dispatcher: new Dispatcher(),
  };
}

const StoresContext = createContext<Stores | null>(null);

export interface StoresProviderProps {
  value?: Stores;
  children: ReactNode;
  /** Hook so tests can spy on wireDispatcher invocations. */
  wire?: typeof wireDispatcher;
}

export function StoresProvider({
  value,
  children,
  wire = wireDispatcher,
}: StoresProviderProps): JSX.Element {
  // Build once per mount when no override is provided. We deliberately do NOT
  // depend on `value` here — swapping bundles mid-tree would be a footgun.
  const stores = useMemo<Stores>(
    () => value ?? createDefaultStores(),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  useEffect(() => {
    const unsub = wire(stores.dispatcher, stores);
    return () => {
      unsub();
    };
  }, [stores, wire]);

  return (
    <StoresContext.Provider value={stores}>{children}</StoresContext.Provider>
  );
}

export function useStores(): Stores {
  const ctx = useContext(StoresContext);
  if (!ctx) {
    throw new Error(
      "useStores() called outside <StoresProvider>. Wrap your tree with <StoresProvider>.",
    );
  }
  return ctx;
}
