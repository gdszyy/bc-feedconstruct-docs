// frontend/src/react/StoresProvider.test.tsx

import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import {
  createDefaultStores,
  StoresProvider,
  useStores,
  type Stores,
} from "./StoresProvider";

function Probe({ onRead }: { onRead: (s: Stores) => void }) {
  const stores = useStores();
  onRead(stores);
  return null;
}

// Given <StoresProvider> mounted without props
// When the tree renders
// Then a complete Stores bundle is available via useStores()
describe("StoresProvider: default bundle construction", () => {
  it("when StoresProvider mounts then useStores() yields a complete bundle", () => {
    let captured: Stores | null = null;
    render(
      <StoresProvider>
        <Probe onRead={(s) => (captured = s)} />
      </StoresProvider>,
    );
    expect(captured).not.toBeNull();
    const s = captured!;
    expect(s.catalog).toBeDefined();
    expect(s.match).toBeDefined();
    expect(s.markets).toBeDefined();
    expect(s.betStop).toBeDefined();
    expect(s.settlement).toBeDefined();
    expect(s.cancel).toBeDefined();
    expect(s.rollback).toBeDefined();
    expect(s.subscription).toBeDefined();
    expect(s.favorites).toBeDefined();
    expect(s.descriptions).toBeDefined();
    expect(s.betSlip).toBeDefined();
    expect(s.myBets).toBeDefined();
    expect(s.health).toBeDefined();
    expect(s.telemetry).toBeDefined();
    expect(s.dispatcher).toBeDefined();
  });
});

// Given <StoresProvider value={preBuiltBundle}>
// When the tree renders
// Then useStores() returns the same object reference passed in (test override)
describe("StoresProvider: value override for tests", () => {
  it("when a pre-built bundle is passed via value then useStores() returns it", () => {
    const bundle = createDefaultStores();
    let captured: Stores | null = null;
    render(
      <StoresProvider value={bundle}>
        <Probe onRead={(s) => (captured = s)} />
      </StoresProvider>,
    );
    expect(captured).toBe(bundle);
  });
});

// Given a StoresProvider mounted with a dispatcher in the bundle
// When the provider's effect runs after mount
// Then wireDispatcher is invoked with (dispatcher, stores) once
describe("StoresProvider: wires dispatcher to stores on mount", () => {
  it("when the provider mounts then dispatcher → store wiring is established", () => {
    const wire = vi.fn().mockReturnValue(() => {});
    const bundle = createDefaultStores();
    render(
      <StoresProvider value={bundle} wire={wire}>
        <Probe onRead={() => {}} />
      </StoresProvider>,
    );
    expect(wire).toHaveBeenCalledTimes(1);
    expect(wire).toHaveBeenCalledWith(bundle.dispatcher, bundle);
  });
});

// Given a StoresProvider that wired the dispatcher on mount
// When the provider unmounts
// Then the wireDispatcher unsub closure is called (no leaked handlers)
describe("StoresProvider: unwires on unmount", () => {
  it("when the provider unmounts then the wiring's unsub closure is called", () => {
    const unsub = vi.fn();
    const wire = vi.fn().mockReturnValue(unsub);
    const bundle = createDefaultStores();
    const { unmount } = render(
      <StoresProvider value={bundle} wire={wire}>
        <Probe onRead={() => {}} />
      </StoresProvider>,
    );
    expect(unsub).not.toHaveBeenCalled();
    unmount();
    expect(unsub).toHaveBeenCalledTimes(1);
  });
});

// Given a component calling useStores() outside any provider
// When the hook runs
// Then it throws an explicit error (catches forgot-to-wrap mistakes)
describe("useStores: outside-provider error", () => {
  it("when useStores is called outside a provider then a descriptive error is thrown", () => {
    // Silence React's act warnings for this intentional throw
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    expect(() => render(<Probe onRead={() => {}} />)).toThrow(
      /useStores\(\) called outside <StoresProvider>/,
    );
    spy.mockRestore();
  });
});
