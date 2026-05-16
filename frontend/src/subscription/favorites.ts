// ---------------------------------------------------------------------------
// M11 — FavoritesStore
//
// Local-only "follow this match" preference. Distinct from SubscriptionStore:
// favoriting does NOT contact the BFF and has no FSM. The persistence layer
// (localStorage adapter in the Next.js app shell) consumes snapshot()/hydrate()
// to bridge to disk; this module stays pure.
//
// Idempotency: add() and remove() return true only when state actually changes.
// Listeners are notified only on real mutations — matching the M08/M09 pattern.
// ---------------------------------------------------------------------------

export interface FavoriteEntry {
  match_id: string;
  added_at: number;
}

export interface FavoritesSnapshot {
  entries: FavoriteEntry[];
}

export class FavoritesStore {
  private readonly entries = new Map<string, FavoriteEntry>();
  private readonly listeners = new Set<() => void>();

  add(matchId: string, at: number = Date.now()): boolean {
    if (this.entries.has(matchId)) return false;
    this.entries.set(matchId, { match_id: matchId, added_at: at });
    this.notify();
    return true;
  }

  remove(matchId: string): boolean {
    const removed = this.entries.delete(matchId);
    if (removed) this.notify();
    return removed;
  }

  has(matchId: string): boolean {
    return this.entries.has(matchId);
  }

  list(): FavoriteEntry[] {
    const entries = Array.from(this.entries.values()).map(cloneFav);
    entries.sort((a, b) => a.added_at - b.added_at);
    return entries;
  }

  snapshot(): FavoritesSnapshot {
    return { entries: this.list() };
  }

  hydrate(snap: FavoritesSnapshot): void {
    this.entries.clear();
    for (const e of snap.entries) {
      this.entries.set(e.match_id, { match_id: e.match_id, added_at: e.added_at });
    }
    this.notify();
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

function cloneFav(e: FavoriteEntry): FavoriteEntry {
  return { match_id: e.match_id, added_at: e.added_at };
}
