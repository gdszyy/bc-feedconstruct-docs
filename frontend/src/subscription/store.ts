import type { SubscriptionChangedPayload } from "@/contract/events";

// ---------------------------------------------------------------------------
// M11 — SubscriptionStore
//
// Tracks the server-tracked subscription lifecycle per match. Implements the
// FSM in docs/07_frontend_architecture/04_state_machines.md §5:
//
//   absent ──markBooking──▶ booking ──markBooked / server 'active'──▶ subscribed
//                              │
//                              └──markBookFailed──▶ failed (book phase)
//
//   subscribed ──markUnbooking──▶ unbooking ──markUnbooked / server 'released'──▶ released
//                                     │
//                                     └──markUnbookFailed──▶ failed (unbook phase)
//
//   released | failed ──markBooking──▶ booking            (retry / re-subscribe)
//   subscribed ──server 'released' | 'cancelled'──▶ released  (BFF-initiated close)
//
// Idempotency:
//   - markBooking is rejected if the current status forbids it (booking,
//     subscribed, unbooking — anything but absent/released/failed)
//   - mark* ack/fail methods are rejected unless the entry is in the matching
//     in-flight phase
//   - applyServerChange dedups 'active' on entries already in 'subscribed' and
//     dedups 'released' / 'cancelled' on entries already in 'released'
//   - Listeners fire only on real transitions, never on dropped calls
//
// Boundaries:
//   - Favorites live in favorites.ts; this store has no opinion on local prefs
//   - REST/WebSocket I/O lives in a subscriptionService wrapper (out of scope)
// ---------------------------------------------------------------------------

export type SubscriptionStatus =
  | "booking"
  | "subscribed"
  | "unbooking"
  | "released"
  | "failed";

export interface SubscriptionError {
  code: string;
  message?: string;
}

export interface SubscriptionEntry {
  match_id: string;
  subscription_id?: string;
  status: SubscriptionStatus;
  last_transition_at: number;
  last_error?: SubscriptionError;
}

export class SubscriptionStore {
  private readonly entries = new Map<string, SubscriptionEntry>();
  private readonly listeners = new Set<() => void>();

  markBooking(matchId: string, at: number = Date.now()): boolean {
    const existing = this.entries.get(matchId);
    if (existing && existing.status !== "released" && existing.status !== "failed") {
      return false;
    }
    this.entries.set(matchId, {
      match_id: matchId,
      status: "booking",
      last_transition_at: at,
    });
    this.notify();
    return true;
  }

  markBooked(subscriptionId: string, matchId: string, at: number = Date.now()): boolean {
    const existing = this.entries.get(matchId);
    if (!existing || existing.status !== "booking") return false;
    this.entries.set(matchId, {
      match_id: matchId,
      subscription_id: subscriptionId,
      status: "subscribed",
      last_transition_at: at,
    });
    this.notify();
    return true;
  }

  markBookFailed(
    matchId: string,
    err: SubscriptionError,
    at: number = Date.now(),
  ): boolean {
    const existing = this.entries.get(matchId);
    if (!existing || existing.status !== "booking") return false;
    this.entries.set(matchId, {
      match_id: matchId,
      status: "failed",
      last_transition_at: at,
      last_error: { code: err.code, message: err.message },
    });
    this.notify();
    return true;
  }

  markUnbooking(matchId: string, at: number = Date.now()): boolean {
    const existing = this.entries.get(matchId);
    if (!existing || existing.status !== "subscribed") return false;
    this.entries.set(matchId, {
      match_id: matchId,
      subscription_id: existing.subscription_id,
      status: "unbooking",
      last_transition_at: at,
    });
    this.notify();
    return true;
  }

  markUnbooked(matchId: string, at: number = Date.now()): boolean {
    const existing = this.entries.get(matchId);
    if (!existing || existing.status !== "unbooking") return false;
    this.entries.set(matchId, {
      match_id: matchId,
      subscription_id: existing.subscription_id,
      status: "released",
      last_transition_at: at,
    });
    this.notify();
    return true;
  }

  markUnbookFailed(
    matchId: string,
    err: SubscriptionError,
    at: number = Date.now(),
  ): boolean {
    const existing = this.entries.get(matchId);
    if (!existing || existing.status !== "unbooking") return false;
    this.entries.set(matchId, {
      match_id: matchId,
      subscription_id: existing.subscription_id,
      status: "failed",
      last_transition_at: at,
      last_error: { code: err.code, message: err.message },
    });
    this.notify();
    return true;
  }

  applyServerChange(
    p: SubscriptionChangedPayload,
    at: number = Date.now(),
  ): boolean {
    const existing = this.entries.get(p.match_id);

    if (p.state === "active") {
      if (!existing) {
        this.entries.set(p.match_id, {
          match_id: p.match_id,
          subscription_id: p.subscription_id,
          status: "subscribed",
          last_transition_at: at,
        });
        this.notify();
        return true;
      }
      if (existing.status === "booking") {
        this.entries.set(p.match_id, {
          match_id: p.match_id,
          subscription_id: p.subscription_id,
          status: "subscribed",
          last_transition_at: at,
        });
        this.notify();
        return true;
      }
      return false;
    }

    // released | cancelled — both fold into the terminal Released state per
    // M11 acceptance ("比赛结束后 5 分钟内 UI 显示 Released").
    if (!existing) return false;
    if (existing.status === "released") return false;
    this.entries.set(p.match_id, {
      match_id: p.match_id,
      subscription_id: existing.subscription_id ?? p.subscription_id,
      status: "released",
      last_transition_at: at,
    });
    this.notify();
    return true;
  }

  selectByMatch(matchId: string): SubscriptionEntry | undefined {
    const e = this.entries.get(matchId);
    return e ? cloneSub(e) : undefined;
  }

  selectAll(): SubscriptionEntry[] {
    return Array.from(this.entries.values()).map(cloneSub);
  }

  selectFailed(): SubscriptionEntry[] {
    return Array.from(this.entries.values())
      .filter((e) => e.status === "failed")
      .map(cloneSub);
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

function cloneSub(e: SubscriptionEntry): SubscriptionEntry {
  const copy: SubscriptionEntry = {
    match_id: e.match_id,
    status: e.status,
    last_transition_at: e.last_transition_at,
  };
  if (e.subscription_id !== undefined) copy.subscription_id = e.subscription_id;
  if (e.last_error !== undefined) {
    copy.last_error = { code: e.last_error.code, message: e.last_error.message };
  }
  return copy;
}
