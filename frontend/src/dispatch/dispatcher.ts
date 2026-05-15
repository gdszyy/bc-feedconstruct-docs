import type { Envelope } from "@/contract/events";

export type EventHandler = (env: Envelope) => void;

export interface TelemetrySink {
  recordUnknownType(env: Envelope): void;
  recordHandlerError(env: Envelope, error: unknown): void;
  recordDuplicate(env: Envelope): void;
  recordStale(env: Envelope, reason: "version" | "occurred_at"): void;
}

export const noopTelemetry: TelemetrySink = {
  recordUnknownType() {},
  recordHandlerError() {},
  recordDuplicate() {},
  recordStale() {},
};

export interface DispatcherOptions {
  dedupCapacity?: number;
  telemetry?: TelemetrySink;
}

type Registration =
  | { kind: "exact"; key: string; handler: EventHandler }
  | { kind: "prefix"; key: string; handler: EventHandler }
  | { kind: "suffix"; key: string; handler: EventHandler };

const DEFAULT_DEDUP_CAPACITY = 4096;

export class Dispatcher {
  private readonly regs: Registration[] = [];
  private readonly dedup: LruSet;
  private readonly versions = new Map<string, number>();
  private readonly timestamps = new Map<string, string>();
  private readonly telemetry: TelemetrySink;

  constructor(opts: DispatcherOptions = {}) {
    this.dedup = new LruSet(opts.dedupCapacity ?? DEFAULT_DEDUP_CAPACITY);
    this.telemetry = opts.telemetry ?? noopTelemetry;
  }

  on(pattern: string, handler: EventHandler): () => void {
    const reg = parsePattern(pattern, handler);
    this.regs.push(reg);
    return () => {
      const idx = this.regs.indexOf(reg);
      if (idx >= 0) this.regs.splice(idx, 1);
    };
  }

  dispatch(env: Envelope): void {
    if (this.dedup.has(env.event_id)) {
      this.telemetry.recordDuplicate(env);
      return;
    }
    this.dedup.add(env.event_id);

    if (!this.versionGuardPasses(env)) {
      return;
    }

    const exact = this.regs.filter(
      (r): r is Registration & { kind: "exact" } =>
        r.kind === "exact" && r.key === env.type,
    );
    const targets =
      exact.length > 0
        ? exact
        : this.regs.filter((r) => matchesPattern(r, env.type));

    if (targets.length === 0) {
      this.telemetry.recordUnknownType(env);
      return;
    }

    for (const reg of targets) {
      try {
        reg.handler(env);
      } catch (err) {
        this.telemetry.recordHandlerError(env, err);
      }
    }
  }

  private versionGuardPasses(env: Envelope): boolean {
    const key = versionGuardKey(env);
    const payload = env.payload as { version?: number } | null | undefined;

    if (payload && typeof payload.version === "number") {
      const prev = this.versions.get(key);
      if (prev !== undefined && payload.version <= prev) {
        this.telemetry.recordStale(env, "version");
        return false;
      }
      this.versions.set(key, payload.version);
      return true;
    }

    const prevTs = this.timestamps.get(key);
    if (prevTs !== undefined && env.occurred_at <= prevTs) {
      this.telemetry.recordStale(env, "occurred_at");
      return false;
    }
    this.timestamps.set(key, env.occurred_at);
    return true;
  }
}

function versionGuardKey(env: Envelope): string {
  const e = env.entity ?? {};
  return [
    env.type,
    e.sport_id ?? "",
    e.tournament_id ?? "",
    e.match_id ?? "",
    e.market_id ?? "",
    e.outcome_id ?? "",
  ].join("|");
}

function parsePattern(pattern: string, handler: EventHandler): Registration {
  if (pattern.endsWith(".*")) {
    return { kind: "prefix", key: pattern.slice(0, -2), handler };
  }
  if (pattern.startsWith("*.")) {
    return { kind: "suffix", key: pattern.slice(2), handler };
  }
  return { kind: "exact", key: pattern, handler };
}

function matchesPattern(reg: Registration, type: string): boolean {
  switch (reg.kind) {
    case "exact":
      return reg.key === type;
    case "prefix":
      return type === reg.key || type.startsWith(reg.key + ".");
    case "suffix":
      return type === reg.key || type.endsWith("." + reg.key);
  }
}

class LruSet {
  private readonly set = new Map<string, true>();
  constructor(private readonly capacity: number) {}

  has(key: string): boolean {
    return this.set.has(key);
  }

  add(key: string): void {
    if (this.set.has(key)) {
      this.set.delete(key);
      this.set.set(key, true);
      return;
    }
    if (this.set.size >= this.capacity) {
      const oldest = this.set.keys().next().value;
      if (oldest !== undefined) this.set.delete(oldest);
    }
    this.set.set(key, true);
  }
}
