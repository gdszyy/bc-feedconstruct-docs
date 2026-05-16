import type { Envelope } from "@/contract/events";
import type { Dispatcher } from "@/dispatch/dispatcher";

// ---------------------------------------------------------------------------
// Centralised event → store routing.
//
// Locked design decisions:
//   * Centralised wireDispatcher (single audit point for every event route)
//   * Multiple handlers per event — fan-out events like `bet.accepted` register
//     once per consumer so a thrown error in one consumer does not stop the
//     others (the dispatcher already isolates handler errors via try/catch).
//
// All store fields are optional: an empty bundle is a valid no-op (useful
// for tests and for partial bring-up during refactors).
// ---------------------------------------------------------------------------

// Structural interfaces — each one matches the public surface of a domain
// store without importing the store class. This lets wireDispatcher be
// unit-tested with vi.fn() stubs and keeps wiring decoupled from store
// internals.

export interface CatalogLike {
  applySportUpserted(p: unknown): void;
  applySportRemoved(p: unknown): void;
  applyTournamentUpserted(p: unknown): void;
  applyTournamentRemoved(p: unknown): void;
}

export interface MatchLike {
  applyMatchUpserted(p: unknown): boolean | void;
  applyMatchStatusChanged(p: unknown): boolean | void;
}

export interface MarketsLike {
  applyOddsChanged(p: unknown): boolean | void;
  applyMarketStatusChanged(p: unknown): boolean | void;
}

export interface BetStopLike {
  applyApplied(p: unknown): boolean | void;
  applyLifted(p: unknown): boolean | void;
}

export interface SettlementLike {
  applyBetSettlementApplied(p: unknown): boolean | void;
  applyBetSettlementRolledBack(p: unknown): boolean | void;
}

export interface CancelLike {
  applyBetCancelApplied(p: unknown): boolean | void;
  applyBetCancelRolledBack(p: unknown): boolean | void;
}

export interface RollbackLike {
  recordSettlementRollback(p: unknown): boolean | void;
  recordCancelRollback(p: unknown): boolean | void;
}

export interface SubscriptionLike {
  applySubscriptionChanged(p: unknown): boolean | void;
}

export interface BetSlipLike {
  applyBetAccepted(p: unknown): boolean | void;
  applyBetRejected(p: unknown): boolean | void;
  applyOddsChanged(p: unknown): boolean | void;
  applyMarketStatusChanged(p: unknown): boolean | void;
  applyBetStopApplied(p: unknown): boolean | void;
  applyBetStopLifted(p: unknown): boolean | void;
}

export interface MyBetsLike {
  // M14 reducers take the full envelope so they can record correlation_id +
  // occurred_at into the bet history.
  applyBetAccepted(env: Envelope): boolean | void;
  applyBetRejected(env: Envelope): boolean | void;
  applyBetStateChanged(env: Envelope): boolean | void;
}

export interface HealthLike {
  applyProducerStatus(p: unknown): boolean | void;
}

export interface WiredStores {
  catalog?: CatalogLike;
  match?: MatchLike;
  markets?: MarketsLike;
  betStop?: BetStopLike;
  settlement?: SettlementLike;
  cancel?: CancelLike;
  rollback?: RollbackLike;
  subscription?: SubscriptionLike;
  betSlip?: BetSlipLike;
  myBets?: MyBetsLike;
  health?: HealthLike;
}

export function wireDispatcher(
  dispatcher: Dispatcher,
  stores: WiredStores,
): () => void {
  const unsubs: Array<() => void> = [];
  const on = (type: string, h: (env: Envelope) => void) => {
    unsubs.push(dispatcher.on(type, h));
  };

  // M03 Catalog
  if (stores.catalog) {
    const c = stores.catalog;
    on("sport.upserted", (env) => c.applySportUpserted(env.payload));
    on("sport.removed", (env) => c.applySportRemoved(env.payload));
    on("tournament.upserted", (env) => c.applyTournamentUpserted(env.payload));
    on("tournament.removed", (env) => c.applyTournamentRemoved(env.payload));
  }

  // M04 Match
  if (stores.match) {
    const m = stores.match;
    on("match.upserted", (env) => m.applyMatchUpserted(env.payload));
    on("match.status_changed", (env) =>
      m.applyMatchStatusChanged(env.payload),
    );
  }

  // M05 Markets
  if (stores.markets) {
    const mk = stores.markets;
    on("odds.changed", (env) => mk.applyOddsChanged(env.payload));
    on("market.status_changed", (env) =>
      mk.applyMarketStatusChanged(env.payload),
    );
  }

  // M06 BetStop
  if (stores.betStop) {
    const bs = stores.betStop;
    on("bet_stop.applied", (env) => bs.applyApplied(env.payload));
    on("bet_stop.lifted", (env) => bs.applyLifted(env.payload));
  }

  // M07 Settlement
  if (stores.settlement) {
    const s = stores.settlement;
    on("bet_settlement.applied", (env) =>
      s.applyBetSettlementApplied(env.payload),
    );
    on("bet_settlement.rolled_back", (env) =>
      s.applyBetSettlementRolledBack(env.payload),
    );
  }

  // M07b Cancel (separate store from Settlement; see settlement/store.ts)
  if (stores.cancel) {
    const c = stores.cancel;
    on("bet_cancel.applied", (env) => c.applyBetCancelApplied(env.payload));
    on("bet_cancel.rolled_back", (env) =>
      c.applyBetCancelRolledBack(env.payload),
    );
  }

  // M08 Rollback history — listens to the rolled_back variants too
  if (stores.rollback) {
    const r = stores.rollback;
    on("bet_settlement.rolled_back", (env) =>
      r.recordSettlementRollback(env.payload),
    );
    on("bet_cancel.rolled_back", (env) =>
      r.recordCancelRollback(env.payload),
    );
  }

  // M11 Subscription
  if (stores.subscription) {
    const sub = stores.subscription;
    on("subscription.changed", (env) =>
      sub.applySubscriptionChanged(env.payload),
    );
  }

  // M13 BetSlip
  if (stores.betSlip) {
    const slip = stores.betSlip;
    on("bet.accepted", (env) => slip.applyBetAccepted(env.payload));
    on("bet.rejected", (env) => slip.applyBetRejected(env.payload));
    on("odds.changed", (env) => slip.applyOddsChanged(env.payload));
    on("market.status_changed", (env) =>
      slip.applyMarketStatusChanged(env.payload),
    );
    on("bet_stop.applied", (env) => slip.applyBetStopApplied(env.payload));
    on("bet_stop.lifted", (env) => slip.applyBetStopLifted(env.payload));
  }

  // M14 MyBets — needs full envelope for correlation_id / occurred_at trail
  if (stores.myBets) {
    const mb = stores.myBets;
    on("bet.accepted", (env) => mb.applyBetAccepted(env));
    on("bet.rejected", (env) => mb.applyBetRejected(env));
    on("bet.state_changed", (env) => mb.applyBetStateChanged(env));
  }

  // M15 Health
  if (stores.health) {
    const h = stores.health;
    on("system.producer_status", (env) => h.applyProducerStatus(env.payload));
  }

  return () => {
    while (unsubs.length > 0) {
      const u = unsubs.pop();
      if (u) u();
    }
  };
}
