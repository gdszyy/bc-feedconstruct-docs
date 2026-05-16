// ---------------------------------------------------------------------------
// M11 — FavoritesStore
//
// Purely local preference for "this user wants to keep an eye on match X".
// Persisted to localStorage under STORAGE_KEY. Independent from the
// server-side SubscriptionStore (see ./store.ts) — favoriting a match never
// touches the subscription FSM.
//
// Locked decisions (see PR thread for M11):
//   - Storage key: "bc.favorites.v1"
//   - Serialised shape: { items: Array<{ matchId: string; addedAt: number }> }
//   - Anonymous users may favorite (no user gating; the store does not know
//     who the viewer is).
//   - list() ordering: ascending by addedAt (oldest first).
//   - Idempotency: add() preserves the original addedAt and does NOT notify.
//   - Malformed JSON in storage degrades gracefully to an empty store.
// ---------------------------------------------------------------------------

export const FAVORITES_STORAGE_KEY = "bc.favorites.v1";

export interface FavoriteEntry {
  matchId: string;
  addedAt: number;
}

export interface FavoritesStorage {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

export interface FavoritesStoreOptions {
  storage?: FavoritesStorage | null;
  now?: () => number;
}

interface PersistedShape {
  items: FavoriteEntry[];
}

export class FavoritesStore {
  private readonly items = new Map<string, number>();
  private readonly listeners = new Set<() => void>();
  private readonly storage: FavoritesStorage | null;
  private readonly now: () => number;

  constructor(opts: FavoritesStoreOptions = {}) {
    this.storage =
      opts.storage === null
        ? null
        : (opts.storage ?? resolveDefaultStorage());
    this.now = opts.now ?? (() => Date.now());
    this.rehydrate();
  }

  add(matchId: string, addedAt?: number): boolean {
    if (this.items.has(matchId)) return false;
    this.items.set(matchId, addedAt ?? this.now());
    this.persist();
    this.notify();
    return true;
  }

  remove(matchId: string): boolean {
    if (!this.items.delete(matchId)) return false;
    this.persist();
    this.notify();
    return true;
  }

  isFavorite(matchId: string): boolean {
    return this.items.has(matchId);
  }

  list(): FavoriteEntry[] {
    return Array.from(this.items.entries())
      .map(([matchId, addedAt]) => ({ matchId, addedAt }))
      .sort((a, b) => a.addedAt - b.addedAt);
  }

  subscribe(handler: () => void): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  private rehydrate(): void {
    if (!this.storage) return;
    const raw = this.storage.getItem(FAVORITES_STORAGE_KEY);
    if (!raw) return;
    try {
      const parsed = JSON.parse(raw) as PersistedShape;
      if (!parsed || !Array.isArray(parsed.items)) return;
      for (const entry of parsed.items) {
        if (
          typeof entry?.matchId === "string" &&
          typeof entry?.addedAt === "number" &&
          !this.items.has(entry.matchId)
        ) {
          this.items.set(entry.matchId, entry.addedAt);
        }
      }
    } catch {
      // malformed JSON — silently fall back to empty
    }
  }

  private persist(): void {
    if (!this.storage) return;
    const payload: PersistedShape = { items: this.list() };
    this.storage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(payload));
  }

  private notify(): void {
    for (const l of this.listeners) l();
  }
}

function resolveDefaultStorage(): FavoritesStorage | null {
  if (typeof globalThis === "undefined") return null;
  const candidate = (globalThis as { localStorage?: FavoritesStorage })
    .localStorage;
  return candidate ?? null;
}
