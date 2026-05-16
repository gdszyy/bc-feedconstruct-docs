import type {
  BetAcceptedPayload,
  BetRejectedPayload,
  BetStateChangedPayload,
  Envelope,
} from "@/contract/events";
import type { BetSelection, MyBet, MyBetTransition } from "@/contract/rest";

// ---------------------------------------------------------------------------
// M14 — MyBetsStore
//
// Index of the user's bets across the full Bet FSM (Pending → Accepted →
// Settled → Cancelled → RolledBack), with append-only history.
//
// Locked decisions (PR thread):
//   4. Unknown-bet on wire event: create record from event payload, with
//      history entry origin='server-pushed'. Server is authoritative.
//      Selections/stake details are filled in by REST refresh.
//   + From-state validation: when bet.state_changed claims from=X but the
//      local state is Y ≠ X, the event is rejected and telemetry fires
//      (fromStateMismatch). Mirrors M11 illegal-transition handling.
//   + Idempotency (P6): each bet tracks the set of applied event_ids;
//      duplicates are dropped silently.
//   + Pending → bet_id promotion: trackPending(idempotency_key, ...) creates
//      a Pending record keyed by idempotency_key; applyBetAccepted then
//      re-keys it under bet_id and transitions to Accepted.
//   + Append-only history (P9): rollback transitions are appended, not
//      reverse-applied; original transitions remain visible in UI.
// ---------------------------------------------------------------------------

export type MyBetState =
  | "Pending"
  | "Accepted"
  | "Rejected"
  | "Settled"
  | "Cancelled";

export type TransitionOrigin = "rest" | "server-pushed" | "local";

export interface MyBetTransitionRecord extends MyBetTransition {
  origin: TransitionOrigin;
}

export interface MyBetRecord {
  id: string;
  user_id?: string;
  placed_at: string;
  stake?: number;
  currency?: string;
  bet_type?: string;
  state: MyBetState;
  selections: BetSelection[];
  history: MyBetTransitionRecord[];
  payout_gross?: number;
  payout_currency?: string;
  void_factor?: number;
  dead_heat_factor?: number;
  idempotency_key?: string;
  applied_event_ids: Set<string>;
}

export interface FromStateMismatch {
  bet_id: string;
  expected_from: string;
  current: MyBetState;
  event_id?: string;
}

export interface MyBetsTelemetry {
  fromStateMismatch(record: FromStateMismatch): void;
}

export interface MyBetsStoreOptions {
  telemetry?: MyBetsTelemetry;
  now?: () => string;
}

export interface TrackPendingArgs {
  idempotency_key: string;
  bet_type: string;
  stake: number;
  currency: string;
  selections: BetSelection[];
  user_id?: string;
  placed_at?: string;
}

const WIRE_TO_STATE: Record<string, MyBetState | undefined> = {
  Pending: "Pending",
  pending: "Pending",
  Accepted: "Accepted",
  accepted: "Accepted",
  Rejected: "Rejected",
  rejected: "Rejected",
  Settled: "Settled",
  settled: "Settled",
  Cancelled: "Cancelled",
  cancelled: "Cancelled",
};

function toMyBetState(raw: string): MyBetState | undefined {
  return WIRE_TO_STATE[raw];
}

export class MyBetsStore {
  private readonly byId = new Map<string, MyBetRecord>();
  private readonly byIdempotencyKey = new Map<string, string>();
  private readonly listeners = new Set<() => void>();
  private readonly telemetry?: MyBetsTelemetry;
  private readonly now: () => string;

  constructor(opts: MyBetsStoreOptions = {}) {
    this.telemetry = opts.telemetry;
    this.now = opts.now ?? (() => new Date().toISOString());
  }

  // -------------------------------------------------------------------------
  // Hydration
  // -------------------------------------------------------------------------

  hydrateBets(bets: ReadonlyArray<MyBet>): void {
    for (const b of bets) {
      const state = toMyBetState(b.state);
      if (!state) continue;
      const history: MyBetTransitionRecord[] = b.history.map((t) => ({
        ...t,
        origin: "rest",
      }));
      const applied = new Set<string>();
      for (const t of history) if (t.event_id) applied.add(t.event_id);
      this.byId.set(b.id, {
        id: b.id,
        user_id: b.user_id,
        placed_at: b.placed_at,
        stake: b.stake,
        currency: b.currency,
        bet_type: b.bet_type,
        state,
        selections: b.selections.map((s) => ({ ...s })),
        history,
        payout_gross: b.payout_gross,
        payout_currency: b.payout_currency,
        void_factor: b.void_factor,
        dead_heat_factor: b.dead_heat_factor,
        idempotency_key: undefined,
        applied_event_ids: applied,
      });
    }
    this.notify();
  }

  // -------------------------------------------------------------------------
  // Pending tracking (slip → my-bets link)
  // -------------------------------------------------------------------------

  trackPending(args: TrackPendingArgs): boolean {
    if (this.byIdempotencyKey.has(args.idempotency_key)) return false;
    const placedAt = args.placed_at ?? this.now();
    const record: MyBetRecord = {
      id: args.idempotency_key,
      user_id: args.user_id,
      placed_at: placedAt,
      stake: args.stake,
      currency: args.currency,
      bet_type: args.bet_type,
      state: "Pending",
      selections: args.selections.map((s) => ({ ...s })),
      history: [
        {
          at: placedAt,
          from: "",
          to: "Pending",
          origin: "local",
        },
      ],
      idempotency_key: args.idempotency_key,
      applied_event_ids: new Set<string>(),
    };
    this.byId.set(args.idempotency_key, record);
    this.byIdempotencyKey.set(args.idempotency_key, args.idempotency_key);
    this.notify();
    return true;
  }

  // -------------------------------------------------------------------------
  // Wire reducers
  // -------------------------------------------------------------------------

  applyBetAccepted(envelope: Envelope<BetAcceptedPayload>): boolean {
    const { payload, event_id, correlation_id } = envelope;
    let record = this.resolveByPayload({
      bet_id: payload.bet_id,
      idempotency_key: undefined,
    });

    // Promote a pending record (keyed by idempotency_key) to bet_id when
    // the slip emitted trackPending earlier this session.
    if (!record) {
      const fromAnyPending = this.findPendingByUser(payload.user_id);
      if (fromAnyPending && fromAnyPending.state === "Pending") {
        record = fromAnyPending;
      }
    }

    if (!record) {
      // Unknown bet → create from server (locked decision #4).
      record = this.createFromServerEvent({
        bet_id: payload.bet_id,
        user_id: payload.user_id,
        placed_at: payload.accepted_at,
        initial_state: "Accepted",
        event_id,
        correlation_id,
        at: payload.accepted_at,
      });
      this.notify();
      return true;
    }

    if (record.applied_event_ids.has(event_id)) return false;
    if (record.state === "Accepted") {
      record.applied_event_ids.add(event_id);
      return false;
    }

    const from = record.state;
    if (record.idempotency_key && record.id === record.idempotency_key) {
      this.byIdempotencyKey.set(record.idempotency_key, payload.bet_id);
      this.byId.delete(record.id);
      record.id = payload.bet_id;
      this.byId.set(payload.bet_id, record);
    }
    record.state = "Accepted";
    record.user_id = record.user_id ?? payload.user_id;
    record.history = [
      ...record.history,
      {
        at: payload.accepted_at,
        from,
        to: "Accepted",
        event_id,
        correlation_id,
        origin: "server-pushed",
      },
    ];
    record.applied_event_ids.add(event_id);
    this.notify();
    return true;
  }

  applyBetRejected(envelope: Envelope<BetRejectedPayload>): boolean {
    const { payload, event_id, correlation_id } = envelope;
    let record = this.byId.get(payload.bet_id);

    if (!record) {
      const fromAnyPending = this.findPendingByUser(payload.user_id);
      if (fromAnyPending && fromAnyPending.state === "Pending") {
        record = fromAnyPending;
      }
    }

    if (!record) {
      record = this.createFromServerEvent({
        bet_id: payload.bet_id,
        user_id: payload.user_id,
        placed_at: this.now(),
        initial_state: "Rejected",
        reason: payload.message,
        event_id,
        correlation_id,
        at: this.now(),
      });
      this.notify();
      return true;
    }

    if (record.applied_event_ids.has(event_id)) return false;
    if (record.state === "Rejected") {
      record.applied_event_ids.add(event_id);
      return false;
    }

    const from = record.state;
    if (record.idempotency_key && record.id === record.idempotency_key) {
      this.byIdempotencyKey.set(record.idempotency_key, payload.bet_id);
      this.byId.delete(record.id);
      record.id = payload.bet_id;
      this.byId.set(payload.bet_id, record);
    }
    record.state = "Rejected";
    record.user_id = record.user_id ?? payload.user_id;
    record.history = [
      ...record.history,
      {
        at: this.now(),
        from,
        to: "Rejected",
        reason: payload.message,
        event_id,
        correlation_id,
        origin: "server-pushed",
      },
    ];
    record.applied_event_ids.add(event_id);
    this.notify();
    return true;
  }

  applyBetStateChanged(envelope: Envelope<BetStateChangedPayload>): boolean {
    const { payload, event_id, correlation_id } = envelope;
    const toState = toMyBetState(payload.to);
    if (!toState) return false;
    const record = this.byId.get(payload.bet_id);

    if (!record) {
      this.createFromServerEvent({
        bet_id: payload.bet_id,
        placed_at: payload.at,
        initial_state: toState,
        reason: payload.reason,
        event_id,
        correlation_id,
        at: payload.at,
      });
      this.notify();
      return true;
    }

    if (record.applied_event_ids.has(event_id)) return false;

    const expectedFrom = toMyBetState(payload.from);
    if (expectedFrom && expectedFrom !== record.state) {
      this.telemetry?.fromStateMismatch({
        bet_id: payload.bet_id,
        expected_from: payload.from,
        current: record.state,
        event_id,
      });
      return false;
    }

    if (record.state === toState) {
      record.applied_event_ids.add(event_id);
      return false;
    }

    const from = record.state;
    record.state = toState;
    record.history = [
      ...record.history,
      {
        at: payload.at,
        from,
        to: toState,
        reason: payload.reason,
        event_id,
        correlation_id,
        origin: "server-pushed",
      },
    ];
    record.applied_event_ids.add(event_id);
    this.notify();
    return true;
  }

  // -------------------------------------------------------------------------
  // Selectors
  // -------------------------------------------------------------------------

  selectById(id: string): MyBetRecord | undefined {
    const r = this.byId.get(id);
    return r ? cloneRecord(r) : undefined;
  }

  list(): MyBetRecord[] {
    return Array.from(this.byId.values()).map(cloneRecord);
  }

  listByStatus(state: MyBetState): MyBetRecord[] {
    return this.list().filter((b) => b.state === state);
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

  private resolveByPayload(args: {
    bet_id: string;
    idempotency_key?: string;
  }): MyBetRecord | undefined {
    if (args.idempotency_key) {
      const known = this.byIdempotencyKey.get(args.idempotency_key);
      if (known) return this.byId.get(known);
    }
    return this.byId.get(args.bet_id);
  }

  private findPendingByUser(userId: string): MyBetRecord | undefined {
    for (const r of this.byId.values()) {
      if (r.state !== "Pending") continue;
      if (r.user_id && userId && r.user_id !== userId) continue;
      return r;
    }
    return undefined;
  }

  private createFromServerEvent(args: {
    bet_id: string;
    user_id?: string;
    placed_at: string;
    initial_state: MyBetState;
    reason?: string;
    event_id: string;
    correlation_id: string;
    at: string;
  }): MyBetRecord {
    const record: MyBetRecord = {
      id: args.bet_id,
      user_id: args.user_id,
      placed_at: args.placed_at,
      state: args.initial_state,
      selections: [],
      history: [
        {
          at: args.at,
          from: "",
          to: args.initial_state,
          reason: args.reason,
          event_id: args.event_id,
          correlation_id: args.correlation_id,
          origin: "server-pushed",
        },
      ],
      applied_event_ids: new Set<string>([args.event_id]),
    };
    this.byId.set(args.bet_id, record);
    return record;
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneRecord(r: MyBetRecord): MyBetRecord {
  return {
    ...r,
    selections: r.selections.map((s) => ({ ...s })),
    history: r.history.map((t) => ({ ...t })),
    applied_event_ids: new Set(r.applied_event_ids),
  };
}
