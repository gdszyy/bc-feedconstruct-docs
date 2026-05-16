import type { SubscriptionChangedPayload } from "@/contract/events";

// ---------------------------------------------------------------------------
// M11 — SubscriptionStore
//
// Server-side subscription state machine (docs/.../04_state_machines.md §5):
//
//   Idle ──book──▶ Booking ──ok──▶ Subscribed ──unbook──▶ Unbooking ──ok──▶ Released
//                    │                                       │
//                    └───fail──▶ Failed                       └───fail──▶ Failed
//
// Locked decisions (see PR thread for M11):
//   - Keying is per-match (`match_id`). A match that reaches Released can be
//     re-booked: requestBook() over Released transitions to Booking.
//   - Wire-state mapping:
//       "active"    → Subscribed
//       "released"  → Released  (e.g. match-end auto-release)
//       "cancelled" → Released with cancelled=true
//   - Server is canonical: an "active" wire ack over Unbooking / Failed
//     restores Subscribed; "active" over Released is illegal and reported
//     via SubscriptionTelemetry.illegalTransition().
//   - Failure handling:
//       reportBookFailed   only valid from Booking.
//       reportUnbookFailed only valid from Unbooking.
//       Both surface to SubscriptionTelemetry.{bookFailed,unbookFailed}.
//   - retryFromFailed() is the ONLY way out of Failed; it restores Booking
//     or Unbooking based on the recorded failed_kind. requestBook /
//     requestUnbook from Failed are no-ops.
// ---------------------------------------------------------------------------

export type SubscriptionState =
  | "Idle"
  | "Booking"
  | "Subscribed"
  | "Unbooking"
  | "Released"
  | "Failed";

export type FailedKind = "book" | "unbook";

export interface SubscriptionRecord {
  match_id: string;
  state: SubscriptionState;
  subscription_id?: string;
  failed_kind?: FailedKind;
  last_error?: string;
  cancelled?: boolean;
  last_transition_at: number;
}

export interface IllegalSubscriptionTransitionRecord {
  match_id: string;
  from: SubscriptionState;
  wire_state: SubscriptionChangedPayload["state"];
}

export interface SubscriptionTelemetry {
  bookFailed(record: SubscriptionRecord): void;
  unbookFailed(record: SubscriptionRecord): void;
  illegalTransition(record: IllegalSubscriptionTransitionRecord): void;
}

export interface SubscriptionStoreOptions {
  telemetry?: SubscriptionTelemetry;
  now?: () => number;
}

export class SubscriptionStore {
  private readonly byMatch = new Map<string, SubscriptionRecord>();
  private readonly listeners = new Set<() => void>();
  private readonly telemetry?: SubscriptionTelemetry;
  private readonly now: () => number;

  constructor(opts: SubscriptionStoreOptions = {}) {
    this.telemetry = opts.telemetry;
    this.now = opts.now ?? (() => Date.now());
  }

  // -------------------------------------------------------------------------
  // Local intent transitions
  // -------------------------------------------------------------------------

  requestBook(matchId: string, at?: number): boolean {
    const existing = this.byMatch.get(matchId);
    const fromState = existing?.state ?? "Idle";
    // Allowed only from Idle (no record) or Released (re-subscription).
    if (fromState !== "Idle" && fromState !== "Released") return false;

    this.byMatch.set(matchId, {
      match_id: matchId,
      state: "Booking",
      subscription_id: undefined,
      cancelled: false,
      last_transition_at: at ?? this.now(),
    });
    this.notify();
    return true;
  }

  requestUnbook(matchId: string, at?: number): boolean {
    const existing = this.byMatch.get(matchId);
    if (!existing || existing.state !== "Subscribed") return false;
    this.byMatch.set(matchId, {
      ...existing,
      state: "Unbooking",
      last_transition_at: at ?? this.now(),
    });
    this.notify();
    return true;
  }

  reportBookFailed(matchId: string, reason: string, at?: number): boolean {
    const existing = this.byMatch.get(matchId);
    if (!existing || existing.state !== "Booking") return false;
    const next: SubscriptionRecord = {
      ...existing,
      state: "Failed",
      failed_kind: "book",
      last_error: reason,
      last_transition_at: at ?? this.now(),
    };
    this.byMatch.set(matchId, next);
    this.telemetry?.bookFailed(cloneRecord(next));
    this.notify();
    return true;
  }

  reportUnbookFailed(matchId: string, reason: string, at?: number): boolean {
    const existing = this.byMatch.get(matchId);
    if (!existing || existing.state !== "Unbooking") return false;
    const next: SubscriptionRecord = {
      ...existing,
      state: "Failed",
      failed_kind: "unbook",
      last_error: reason,
      last_transition_at: at ?? this.now(),
    };
    this.byMatch.set(matchId, next);
    this.telemetry?.unbookFailed(cloneRecord(next));
    this.notify();
    return true;
  }

  retryFromFailed(matchId: string, at?: number): boolean {
    const existing = this.byMatch.get(matchId);
    if (!existing || existing.state !== "Failed" || !existing.failed_kind) {
      return false;
    }
    const nextState: SubscriptionState =
      existing.failed_kind === "book" ? "Booking" : "Unbooking";
    const next: SubscriptionRecord = {
      ...existing,
      state: nextState,
      failed_kind: undefined,
      last_error: undefined,
      last_transition_at: at ?? this.now(),
    };
    this.byMatch.set(matchId, next);
    this.notify();
    return true;
  }

  // -------------------------------------------------------------------------
  // Wire reducer (subscription.changed)
  // -------------------------------------------------------------------------

  applySubscriptionChanged(
    payload: SubscriptionChangedPayload,
    at?: number,
  ): boolean {
    const existing = this.byMatch.get(payload.match_id);
    const fromState: SubscriptionState = existing?.state ?? "Idle";
    const ts = at ?? this.now();

    if (payload.state === "active") {
      if (fromState === "Released") {
        this.telemetry?.illegalTransition({
          match_id: payload.match_id,
          from: fromState,
          wire_state: payload.state,
        });
        return false;
      }
      if (
        existing?.state === "Subscribed" &&
        existing.subscription_id === payload.subscription_id &&
        existing.cancelled === false
      ) {
        return false;
      }
      this.byMatch.set(payload.match_id, {
        match_id: payload.match_id,
        state: "Subscribed",
        subscription_id: payload.subscription_id,
        cancelled: false,
        last_transition_at: ts,
      });
      this.notify();
      return true;
    }

    // released | cancelled — both converge on terminal Released.
    if (fromState === "Released") return false;
    this.byMatch.set(payload.match_id, {
      match_id: payload.match_id,
      state: "Released",
      subscription_id: payload.subscription_id,
      cancelled: payload.state === "cancelled",
      last_transition_at: ts,
    });
    this.notify();
    return true;
  }

  // -------------------------------------------------------------------------
  // Selectors
  // -------------------------------------------------------------------------

  selectByMatch(matchId: string): SubscriptionRecord | undefined {
    const r = this.byMatch.get(matchId);
    return r ? cloneRecord(r) : undefined;
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function cloneRecord(r: SubscriptionRecord): SubscriptionRecord {
  return { ...r };
}
