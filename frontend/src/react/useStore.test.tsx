// frontend/src/react/useStore.test.tsx

import { act, render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useStore } from "./useStore";

class CounterStore {
  private value = 0;
  private listeners = new Set<() => void>();
  get(): number {
    return this.value;
  }
  inc(): void {
    this.value += 1;
    for (const l of this.listeners) l();
  }
  notifyNoChange(): void {
    for (const l of this.listeners) l();
  }
  subscribe(h: () => void): () => void {
    this.listeners.add(h);
    return () => this.listeners.delete(h);
  }
  listenerCount(): number {
    return this.listeners.size;
  }
}

// Given a store with subscribe + a selector
// When the component first renders
// Then the selected slice is returned
describe("useStore: initial selection", () => {
  it("when the component renders then the selector result is returned", () => {
    const store = new CounterStore();
    function Probe() {
      const v = useStore(store, (s) => s.get());
      return <span data-testid="v">{v}</span>;
    }
    const { getByTestId } = render(<Probe />);
    expect(getByTestId("v").textContent).toBe("0");
  });
});

// Given a mounted component using useStore(store, selector)
// When store.notify() fires after a state mutation
// Then the component re-renders with the new selector result
describe("useStore: re-renders on store notify", () => {
  it("when the store notifies then the component re-renders with the new slice", () => {
    const store = new CounterStore();
    function Probe() {
      const v = useStore(store, (s) => s.get());
      return <span data-testid="v">{v}</span>;
    }
    const { getByTestId } = render(<Probe />);
    expect(getByTestId("v").textContent).toBe("0");
    act(() => {
      store.inc();
    });
    expect(getByTestId("v").textContent).toBe("1");
    act(() => {
      store.inc();
      store.inc();
    });
    expect(getByTestId("v").textContent).toBe("3");
  });
});

// Given a selector returning the same value across mutations (no slice change)
// When the store notifies
// Then the component does NOT re-render (referential equality short-circuits)
describe("useStore: stable selector skips re-render", () => {
  it("when the selector returns the same primitive then no extra render happens", () => {
    const store = new CounterStore();
    const renderCount = vi.fn();
    // Selector pinned to a constant — store state never changes the slice
    function Probe() {
      const v = useStore(store, () => "stable");
      renderCount();
      return <span>{v}</span>;
    }
    render(<Probe />);
    const initial = renderCount.mock.calls.length;
    act(() => {
      store.notifyNoChange();
      store.notifyNoChange();
    });
    // No re-render because the selector return is referentially equal.
    expect(renderCount.mock.calls.length).toBe(initial);
  });
});

// Given a component that unmounts
// When the component unmounts
// Then the underlying store subscription is removed (no leak)
describe("useStore: unsubscribes on unmount", () => {
  it("when the component unmounts then the store listener is removed", () => {
    const store = new CounterStore();
    function Probe() {
      const v = useStore(store, (s) => s.get());
      return <span>{v}</span>;
    }
    const { unmount } = render(<Probe />);
    expect(store.listenerCount()).toBe(1);
    unmount();
    expect(store.listenerCount()).toBe(0);
  });
});
