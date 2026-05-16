// frontend/src/app/page.test.tsx
//
// 页面 P01 — 首页 / 大厅
//
// Scope locked for this turn: only the HealthStore-driven degradation banner.
// Sport list / live counts depend on a snapshot endpoint that does not exist
// yet and are deferred to a follow-up PR.

import { act, render, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StubRestClient } from "@/api/testing";
import type { Envelope, SystemProducerStatusPayload } from "@/contract/events";
import type { GetSystemHealthResponse } from "@/contract/rest";
import {
  createDefaultStores,
  StoresProvider,
  type Stores,
} from "@/react/StoresProvider";

import HomePage from "./page";

function makeBundle(healthBody: GetSystemHealthResponse): Stores {
  const stub = new StubRestClient({
    responses: [
      {
        match: { method: "GET", path: "/api/v1/system/health" },
        response: {
          status: "ok",
          body: healthBody,
          correlation_id: "test-corr",
          http_status: 200,
        },
      },
    ],
  });
  return createDefaultStores({ restClient: stub.asClient() });
}

// Given the BFF returns a healthy GET /api/v1/system/health snapshot
// (every producer is_down=false) on mount
// When the home page renders and the snapshot resolves
// Then no degradation banner is shown
describe("home page: healthy hydrate", () => {
  it("when system health snapshot is clean then no degradation banner is shown", async () => {
    const bundle = makeBundle({
      producers: [
        { product: "live", is_down: false, last_message_at: "2026-05-16T00:00:00Z" },
        { product: "prematch", is_down: false, last_message_at: "2026-05-16T00:00:00Z" },
      ],
      degraded: false,
    });
    const { queryByTestId, findByText } = render(
      <StoresProvider value={bundle}>
        <HomePage />
      </StoresProvider>,
    );
    // Wait for the async hydrate effect to settle.
    await findByText(/BC FeedConstruct Web/);
    // After hydrate completes, the HealthStore still has no banner.
    await waitFor(() => {
      expect(bundle.health.getBanner()).toBeUndefined();
    });
    expect(queryByTestId("health-banner")).toBeNull();
  });
});

// Given the BFF returns GET /api/v1/system/health with at least one
// is_down=true producer on mount
// When the home page renders and the snapshot resolves
// Then a degradation banner is shown reflecting the HealthStore's
// computed banner level + message
describe("home page: degraded hydrate", () => {
  it("when a producer is down at hydrate then the degradation banner is shown", async () => {
    const bundle = makeBundle({
      producers: [
        {
          product: "live",
          is_down: true,
          last_message_at: "2026-05-16T00:00:00Z",
          down_since: "2026-05-16T00:00:10Z",
        },
        { product: "prematch", is_down: false, last_message_at: "2026-05-16T00:00:00Z" },
      ],
      degraded: true,
    });
    const { findByTestId } = render(
      <StoresProvider value={bundle}>
        <HomePage />
      </StoresProvider>,
    );
    const banner = await findByTestId("health-banner");
    expect(banner.textContent).toMatch(/live/i);
    // The HealthStore-computed level should appear as a data attribute.
    expect(banner.getAttribute("data-level")).toMatch(/info|warn|error/);
  });
});

// Given the home page is mounted with a clean snapshot
// When a system.producer_status event arrives via Dispatcher with is_down=true
// Then the page re-renders to show the degradation banner without a refetch
describe("home page: live degradation", () => {
  it("when a producer_status event marks a producer down then the banner appears live", async () => {
    const bundle = makeBundle({
      producers: [
        { product: "live", is_down: false, last_message_at: "2026-05-16T00:00:00Z" },
        { product: "prematch", is_down: false, last_message_at: "2026-05-16T00:00:00Z" },
      ],
      degraded: false,
    });
    const { queryByTestId, findByText } = render(
      <StoresProvider value={bundle}>
        <HomePage />
      </StoresProvider>,
    );
    await findByText(/BC FeedConstruct Web/);
    // No banner after the clean hydrate.
    await waitFor(() => expect(bundle.health.getBanner()).toBeUndefined());
    expect(queryByTestId("health-banner")).toBeNull();

    // Push a producer_status event through the dispatcher; wireDispatcher
    // routes it into HealthStore.applyProducerStatus.
    const env: Envelope<SystemProducerStatusPayload> = {
      type: "system.producer_status",
      schema_version: "1",
      event_id: "evt-live-down-1",
      correlation_id: "corr-live-down-1",
      product_id: "live",
      occurred_at: "2026-05-16T00:01:00Z",
      received_at: "2026-05-16T00:01:00Z",
      entity: {},
      payload: {
        product: "live",
        is_down: true,
        last_message_at: "2026-05-16T00:00:50Z",
        down_since: "2026-05-16T00:01:00Z",
      },
    };
    act(() => {
      bundle.dispatcher.dispatch(env);
    });

    // Banner now visible without re-running fetchSystemHealth.
    await waitFor(() => {
      expect(queryByTestId("health-banner")).not.toBeNull();
    });
  });
});
