// ---------------------------------------------------------------------------
// M16 — TelemetryStore
//
// Structured queue + batch shipper. Pure logic — the actual transport
// (HTTP, Sentry, etc.) is injected via TelemetryShipper.
//
// Locked decisions (PR thread):
//   3. Default PII redaction list = auth/identity fields only
//      (token, password, email, username, phone, api_key). Monetary fields
//      (stake, amount, balance, payout) are NOT redacted by default; callers
//      can extend via opts.redactKeys.
//   4. Ship failure → retain & retry indefinitely. The shipper is responsible
//      for its own backoff; the store just tracks shipFailed counter and
//      keeps events queued for the next flush.
//
//   + correlation_id is REQUIRED on every event (acceptance §1: "100% 携带").
//     If a caller omits it, the store synthesises one and increments
//     missingCorrelationId so missing-context bugs are visible.
//   + Auto-ship triggers: queue size threshold OR interval timer.
//   + Overflow: queue is bounded; oldest events are dropped and the overflow
//     counter is incremented (preserves recency for incident debugging).
// ---------------------------------------------------------------------------

export type TelemetryKind = "log" | "metric" | "error" | "audit";
export type LogLevel = "debug" | "info" | "warn" | "error";

export interface TelemetryEvent {
  id: string;
  kind: TelemetryKind;
  level?: LogLevel;
  occurred_at: string;
  correlation_id: string;
  session_id?: string;
  user_id?: string;
  payload: Record<string, unknown>;
}

export interface TelemetryShipper {
  ship(batch: TelemetryEvent[]): Promise<void>;
}

export interface TelemetryStoreOptions {
  shipper: TelemetryShipper;
  batchSize?: number;
  batchIntervalMs?: number;
  maxQueueSize?: number;
  redactKeys?: string[];
  generateId?: () => string;
  now?: () => string;
  generateCorrelationId?: () => string;
  scheduleTimeout?: (cb: () => void, ms: number) => unknown;
  cancelTimeout?: (h: unknown) => void;
}

export interface TelemetryCounters {
  enqueued: number;
  shipped: number;
  shipFailed: number;
  overflow: number;
  missingCorrelationId: number;
}

export interface TelemetryIdentity {
  session_id?: string;
  user_id?: string;
}

const DEFAULT_REDACT_KEYS = [
  "token",
  "password",
  "email",
  "username",
  "phone",
  "api_key",
];
const DEFAULT_BATCH_SIZE = 20;
const DEFAULT_BATCH_INTERVAL_MS = 5000;
const DEFAULT_MAX_QUEUE = 1000;
const REDACTED = "[REDACTED]";

export class TelemetryStore {
  private readonly shipper: TelemetryShipper;
  private readonly batchSize: number;
  private readonly intervalMs: number;
  private readonly maxQueue: number;
  private readonly redactKeys: Set<string>;
  private readonly idGen: () => string;
  private readonly now: () => string;
  private readonly corrGen: () => string;
  private readonly schedule: (cb: () => void, ms: number) => unknown;
  private readonly cancel: (h: unknown) => void;

  private queue: TelemetryEvent[] = [];
  private identity: TelemetryIdentity = {};
  private counters: TelemetryCounters = {
    enqueued: 0,
    shipped: 0,
    shipFailed: 0,
    overflow: 0,
    missingCorrelationId: 0,
  };
  private autoSeq = 0;
  private intervalHandle?: unknown;
  private shipPending = false;
  private readonly listeners = new Set<() => void>();

  constructor(opts: TelemetryStoreOptions) {
    this.shipper = opts.shipper;
    this.batchSize = opts.batchSize ?? DEFAULT_BATCH_SIZE;
    this.intervalMs = opts.batchIntervalMs ?? DEFAULT_BATCH_INTERVAL_MS;
    this.maxQueue = opts.maxQueueSize ?? DEFAULT_MAX_QUEUE;
    this.redactKeys = new Set([
      ...DEFAULT_REDACT_KEYS,
      ...(opts.redactKeys ?? []),
    ]);
    this.idGen = opts.generateId ?? (() => `evt-${++this.autoSeq}`);
    this.now = opts.now ?? (() => new Date().toISOString());
    this.corrGen =
      opts.generateCorrelationId ?? (() => `corr-synth-${++this.autoSeq}`);
    this.schedule =
      opts.scheduleTimeout ?? ((cb, ms) => setTimeout(cb, ms));
    this.cancel =
      opts.cancelTimeout ??
      ((h) => clearTimeout(h as ReturnType<typeof setTimeout>));
  }

  // -------------------------------------------------------------------------
  // Identity
  // -------------------------------------------------------------------------

  setIdentity(identity: TelemetryIdentity): void {
    this.identity = { ...identity };
  }

  getIdentity(): TelemetryIdentity {
    return { ...this.identity };
  }

  // -------------------------------------------------------------------------
  // Public emit surface
  // -------------------------------------------------------------------------

  log(args: {
    level: LogLevel;
    message: string;
    correlation_id?: string;
    props?: Record<string, unknown>;
  }): void {
    this.enqueue({
      kind: "log",
      level: args.level,
      correlation_id: args.correlation_id,
      payload: { message: args.message, ...(args.props ?? {}) },
    });
  }

  metric(args: {
    name: string;
    value: number;
    unit?: string;
    tags?: Record<string, string>;
    correlation_id?: string;
  }): void {
    this.enqueue({
      kind: "metric",
      correlation_id: args.correlation_id,
      payload: {
        name: args.name,
        value: args.value,
        unit: args.unit,
        tags: args.tags,
      },
    });
  }

  error(args: {
    kind: string;
    message: string;
    stack?: string;
    correlation_id?: string;
    props?: Record<string, unknown>;
  }): void {
    this.enqueue({
      kind: "error",
      level: "error",
      correlation_id: args.correlation_id,
      payload: {
        error_kind: args.kind,
        message: args.message,
        stack: args.stack,
        ...(args.props ?? {}),
      },
    });
  }

  audit(args: {
    action: string;
    correlation_id?: string;
    props?: Record<string, unknown>;
  }): void {
    this.enqueue({
      kind: "audit",
      correlation_id: args.correlation_id,
      payload: { action: args.action, ...(args.props ?? {}) },
    });
  }

  // -------------------------------------------------------------------------
  // Flush / batch
  // -------------------------------------------------------------------------

  flush(): Promise<void> {
    return this.flushInternal();
  }

  // -------------------------------------------------------------------------
  // Observability
  // -------------------------------------------------------------------------

  getCounters(): TelemetryCounters {
    return { ...this.counters };
  }

  getQueueSize(): number {
    return this.queue.length;
  }

  /** Test helper — returns a defensive copy. */
  getQueueSnapshot(): TelemetryEvent[] {
    return this.queue.map((e) => ({ ...e, payload: { ...e.payload } }));
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

  private enqueue(args: {
    kind: TelemetryKind;
    level?: LogLevel;
    correlation_id?: string;
    payload: Record<string, unknown>;
  }): void {
    let corr = args.correlation_id;
    if (!corr) {
      this.counters.missingCorrelationId++;
      corr = this.corrGen();
    }

    const event: TelemetryEvent = {
      id: this.idGen(),
      kind: args.kind,
      level: args.level,
      occurred_at: this.now(),
      correlation_id: corr,
      session_id: this.identity.session_id,
      user_id: this.identity.user_id,
      payload: this.redact(args.payload),
    };

    this.queue.push(event);
    this.counters.enqueued++;

    while (this.queue.length > this.maxQueue) {
      this.queue.shift();
      this.counters.overflow++;
    }

    this.notify();

    if (this.queue.length >= this.batchSize) {
      void this.flushInternal();
    } else {
      this.ensureInterval();
    }
  }

  private async flushInternal(): Promise<void> {
    if (this.queue.length === 0 || this.shipPending) return;
    this.cancelInterval();
    const batch = this.queue.slice();
    this.shipPending = true;
    try {
      await this.shipper.ship(batch);
      const ids = new Set(batch.map((e) => e.id));
      this.queue = this.queue.filter((e) => !ids.has(e.id));
      this.counters.shipped += batch.length;
    } catch {
      this.counters.shipFailed++;
      // Retain queue (locked decision #4).
    } finally {
      this.shipPending = false;
      this.notify();
      if (this.queue.length > 0) this.ensureInterval();
    }
  }

  private ensureInterval(): void {
    if (this.intervalHandle !== undefined) return;
    this.intervalHandle = this.schedule(() => {
      this.intervalHandle = undefined;
      void this.flushInternal();
    }, this.intervalMs);
  }

  private cancelInterval(): void {
    if (this.intervalHandle !== undefined) {
      this.cancel(this.intervalHandle);
      this.intervalHandle = undefined;
    }
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }

  private redact(value: unknown): Record<string, unknown> {
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      return value as Record<string, unknown>;
    }
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      if (this.redactKeys.has(k)) {
        out[k] = REDACTED;
      } else if (v && typeof v === "object" && !Array.isArray(v)) {
        out[k] = this.redact(v);
      } else {
        out[k] = v;
      }
    }
    return out;
  }
}
