"use client";

import { createContext, useContext, useEffect, useMemo, type ReactNode } from "react";

import { RestClient } from "@/api/client";
import { BetSlipStore } from "@/betSlip/store";
import { BetStopStore } from "@/betStop/store";
import { CatalogStore } from "@/catalog/store";
import { DescriptionsStore } from "@/descriptions/store";
import { Dispatcher } from "@/dispatch/dispatcher";
import { HealthStore } from "@/health/store";
import { MarketsStore } from "@/markets/store";
import { MatchStore } from "@/match/store";
import { MyBetsStore } from "@/myBets/store";
import {
  Transport,
  type WebSocketFactory,
  type WebSocketLike,
} from "@/realtime/transport";
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
  restClient: RestClient;
  transport: Transport;
}

export interface CreateDefaultStoresOptions {
  /**
   * TelemetryShipper used by TelemetryStore. Production should pass
   * `createHttpTelemetryShipper({ client })`. Defaults to a no-op shipper.
   */
  telemetryShipper?: TelemetryShipper;
  /**
   * RestClient used to talk to the Go BFF. Defaults to a client built from
   * `NEXT_PUBLIC_BFF_HTTP` (or `http://localhost:8080`) using the global
   * fetch — sufficient for the dev server; production should inject a
   * client with auth + telemetry already wired.
   */
  restClient?: RestClient;
  /**
   * WS Transport. Defaults to one pointed at `NEXT_PUBLIC_BFF_WS`
   * (or `ws://localhost:8080/ws`). Tests should inject a Transport built
   * with a fake `webSocketFactory`.
   */
  transport?: Transport;
  /**
   * Override the WebSocket factory used by the default Transport. Ignored
   * when `transport` is provided. Useful when production wants to instrument
   * the WS but is fine with the default URL + reconnect policy.
   */
  webSocketFactory?: WebSocketFactory;
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
    restClient: opts.restClient ?? createDefaultRestClient(),
    transport:
      opts.transport ??
      createDefaultTransport({ webSocketFactory: opts.webSocketFactory }),
  };
}

function createDefaultRestClient(): RestClient {
  const baseUrl =
    (typeof process !== "undefined" && process.env?.NEXT_PUBLIC_BFF_HTTP) ||
    "http://localhost:8080";
  return new RestClient({
    baseUrl,
    fetch: (input, init) => fetch(input, init),
  });
}

function createDefaultTransport(opts: {
  webSocketFactory?: WebSocketFactory;
}): Transport {
  const url =
    (typeof process !== "undefined" && process.env?.NEXT_PUBLIC_BFF_WS) ||
    "ws://localhost:8080/ws";
  // In environments without a global WebSocket (Node / jsdom default), use an
  // inert factory so the Transport instance exists but never actually fires
  // open/message/close. Production hits the real browser global.
  const factory =
    opts.webSocketFactory ??
    (typeof WebSocket !== "undefined" ? undefined : createInertWsFactory());
  return new Transport({ url, webSocketFactory: factory });
}

function createInertWsFactory(): WebSocketFactory {
  return (_url: string): WebSocketLike => ({
    readyState: 0,
    send() {},
    close() {},
    onopen: null,
    onmessage: null,
    onclose: null,
    onerror: null,
  });
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
    const unsubWire = wire(stores.dispatcher, stores);
    const unsubMsg = stores.transport.onMessage((env) => {
      stores.dispatcher.dispatch(env);
    });
    // Idempotent — Transport.connect() short-circuits if already
    // Connecting/Open. Skipped for Closed (stopped) bundles so we never
    // accidentally restart a deliberately-closed Transport.
    if (stores.transport.getState() === "Disconnected") {
      stores.transport.connect();
    }
    return () => {
      unsubMsg();
      unsubWire();
      // Intentional: we do NOT call stores.transport.close(). The Transport
      // is owned by the Stores bundle and is meant to outlive React mount /
      // unmount cycles (React StrictMode runs effects twice in dev, and
      // Transport.close() sets stopped=true permanently).
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
