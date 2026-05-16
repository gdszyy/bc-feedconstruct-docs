"use client";

import { useCallback, useSyncExternalStore } from "react";

// ---------------------------------------------------------------------------
// Generic store hook over the project's `subscribe(handler): unsub` shape.
// Locked decisions:
//   * Generic primitive only — per-module convenience hooks land in their own
//     modules if/when needed.
//   * Referential equality semantics inherited from useSyncExternalStore;
//     selectors that derive new objects/arrays must memoise on the caller side.
// ---------------------------------------------------------------------------

export interface SubscribableStore {
  subscribe(handler: () => void): () => void;
}

export function useStore<S extends SubscribableStore, T>(
  store: S,
  selector: (store: S) => T,
): T {
  const sub = useCallback(
    (cb: () => void) => store.subscribe(cb),
    [store],
  );
  const get = useCallback(() => selector(store), [store, selector]);
  return useSyncExternalStore(sub, get, get);
}
