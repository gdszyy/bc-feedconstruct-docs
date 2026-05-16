import type {
  ProductID,
  SystemProducerStatusPayload,
} from "@/contract/events";
import type { GetSystemHealthResponse } from "@/contract/rest";
import type { ConnectionState } from "@/realtime/transport";

// ---------------------------------------------------------------------------
// M15 — HealthStore + degradation policy
//
// Aggregates four signals into a single derived banner + gating selector:
//
//   * connection: ConnectionState (M01 Transport)
//   * producers: per-product up/down (system.producer_status + REST hydrate)
//   * staleScope: M10 staleTracker output (none | global | scoped match_ids)
//   * + REST hydrate at startup
//
// Locked decisions (PR thread):
//   1. Strict per-product gating. canSubmitBet({ product_id: 'live' }) blocks
//      only if the live producer is down; prematch outage doesn't block live
//      bets. Match-level stale set still scoped by match_id.
//   2. Stale and producer-down stay INDEPENDENT signals. canSubmitBet
//      aggregates them; the banner lists both reasons; M10 owns stale scope
//      and M15 owns producer state. No auto-propagation.
//   + Banner severity: error > warn > info.
//   + canSubmitBet blocks on Degraded / Reconnecting / Closed (acceptance §3).
// ---------------------------------------------------------------------------

export type ProducerHealth = "up" | "down";

export type StaleScope =
  | { kind: "none" }
  | { kind: "global" }
  | { kind: "scoped"; match_ids: string[] };

export interface HealthBanner {
  level: "info" | "warn" | "error";
  message: string;
  since: number;
}

export type BlockReason =
  | { kind: "connection_closed" }
  | { kind: "connection_reconnecting" }
  | { kind: "connection_degraded" }
  | { kind: "connection_not_open"; state: ConnectionState }
  | { kind: "producer_down"; product: ProductID }
  | { kind: "stale_global" }
  | { kind: "stale_match"; match_id: string };

export interface CanSubmitGate {
  ok: boolean;
  reasons: BlockReason[];
}

export interface CanSubmitArgs {
  product_id?: ProductID;
  match_id?: string;
}

export interface HealthStoreOptions {
  now?: () => number;
}

export class HealthStore {
  private connection: ConnectionState = "Disconnected";
  private readonly producers = new Map<ProductID, ProducerHealth>();
  private staleScope: StaleScope = { kind: "none" };
  private banner?: HealthBanner;
  private readonly listeners = new Set<() => void>();
  private readonly now: () => number;

  constructor(opts: HealthStoreOptions = {}) {
    this.now = opts.now ?? (() => Date.now());
  }

  // -------------------------------------------------------------------------
  // Reducers
  // -------------------------------------------------------------------------

  applyConnectionState(state: ConnectionState): boolean {
    if (this.connection === state) return false;
    this.connection = state;
    this.recomputeBanner();
    this.notify();
    return true;
  }

  applyProducerStatus(payload: SystemProducerStatusPayload): boolean {
    const next: ProducerHealth = payload.is_down ? "down" : "up";
    if (this.producers.get(payload.product) === next) return false;
    this.producers.set(payload.product, next);
    this.recomputeBanner();
    this.notify();
    return true;
  }

  setStaleScope(scope: StaleScope): boolean {
    if (scopesEqual(this.staleScope, scope)) return false;
    this.staleScope = cloneScope(scope);
    this.recomputeBanner();
    this.notify();
    return true;
  }

  hydrate(snapshot: GetSystemHealthResponse): void {
    let mutated = false;
    for (const p of snapshot.producers) {
      const next: ProducerHealth = p.is_down ? "down" : "up";
      if (this.producers.get(p.product) !== next) {
        this.producers.set(p.product, next);
        mutated = true;
      }
    }
    if (mutated) {
      this.recomputeBanner();
      this.notify();
    }
  }

  // -------------------------------------------------------------------------
  // Selectors
  // -------------------------------------------------------------------------

  getConnection(): ConnectionState {
    return this.connection;
  }

  getProducer(product: ProductID): ProducerHealth | undefined {
    return this.producers.get(product);
  }

  getStaleScope(): StaleScope {
    return cloneScope(this.staleScope);
  }

  getBanner(): HealthBanner | undefined {
    return this.banner ? { ...this.banner } : undefined;
  }

  canSubmitBet(args: CanSubmitArgs = {}): CanSubmitGate {
    const reasons: BlockReason[] = [];

    // Connection gate (acceptance §3).
    switch (this.connection) {
      case "Closed":
        reasons.push({ kind: "connection_closed" });
        break;
      case "Reconnecting":
        reasons.push({ kind: "connection_reconnecting" });
        break;
      case "Degraded":
        reasons.push({ kind: "connection_degraded" });
        break;
      case "Open":
        break;
      default:
        reasons.push({ kind: "connection_not_open", state: this.connection });
    }

    // Strict per-product gate (locked decision #1).
    if (args.product_id) {
      if (this.producers.get(args.product_id) === "down") {
        reasons.push({ kind: "producer_down", product: args.product_id });
      }
    }

    // Stale gate (independent signal).
    if (this.staleScope.kind === "global") {
      reasons.push({ kind: "stale_global" });
    } else if (this.staleScope.kind === "scoped" && args.match_id) {
      if (this.staleScope.match_ids.includes(args.match_id)) {
        reasons.push({ kind: "stale_match", match_id: args.match_id });
      }
    }

    return { ok: reasons.length === 0, reasons };
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  // -------------------------------------------------------------------------
  // Internals
  // -------------------------------------------------------------------------

  private recomputeBanner(): void {
    let level: HealthBanner["level"] | null = null;
    const messages: string[] = [];

    const bump = (next: HealthBanner["level"]) => {
      if (level === null) {
        level = next;
        return;
      }
      // error > warn > info
      if (next === "error") level = "error";
      else if (next === "warn" && level === "info") level = "warn";
    };

    switch (this.connection) {
      case "Closed":
        bump("error");
        messages.push("Connection closed");
        break;
      case "Reconnecting":
        bump("warn");
        messages.push("Reconnecting…");
        break;
      case "Degraded":
        bump("warn");
        messages.push("Connection degraded");
        break;
      default:
        break;
    }

    for (const [product, state] of this.producers) {
      if (state === "down") {
        bump("warn");
        messages.push(`Producer ${product} is down`);
      }
    }

    switch (this.staleScope.kind) {
      case "global":
        bump("info");
        messages.push("Data may be stale");
        break;
      case "scoped":
        bump("info");
        messages.push(`${this.staleScope.match_ids.length} match(es) may be stale`);
        break;
      default:
        break;
    }

    if (level === null) {
      this.banner = undefined;
      return;
    }
    const message = messages.join("; ");
    if (
      this.banner &&
      this.banner.level === level &&
      this.banner.message === message
    ) {
      return;
    }
    this.banner = { level, message, since: this.now() };
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function scopesEqual(a: StaleScope, b: StaleScope): boolean {
  if (a.kind !== b.kind) return false;
  if (a.kind === "scoped" && b.kind === "scoped") {
    if (a.match_ids.length !== b.match_ids.length) return false;
    const aset = new Set(a.match_ids);
    for (const id of b.match_ids) if (!aset.has(id)) return false;
  }
  return true;
}

function cloneScope(scope: StaleScope): StaleScope {
  if (scope.kind === "scoped") {
    return { kind: "scoped", match_ids: [...scope.match_ids] };
  }
  return { kind: scope.kind };
}
