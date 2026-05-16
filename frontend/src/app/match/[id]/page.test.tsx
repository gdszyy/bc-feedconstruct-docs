// frontend/src/app/match/[id]/page.test.tsx
//
// 页面 P03 — 赛事详情（MVP slice）
//
// Scope locked for this slice:
//   1. Mount calls fetchMatchSnapshot(client, id) and hydrates MatchStore +
//      MarketsStore; page header shows home/away team names and a market
//      count from MarketsStore.
//   2. After mount, an odds.changed event arriving via Transport updates the
//      MarketsStore version and the page re-renders to reflect the new
//      version of the affected market.
//   3. Unmount calls transport.unsubscribe({ match_ids: [id] }) (the matching
//      subscribe call was issued at mount).
//
// Out-of-scope for this slice (deferred): grouping markets by market_type,
// per-cell granular re-renders, bet_stop suspended overlay UI, descriptions
// / i18n. They will land in follow-up slices.

import { act, render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { StubRestClient } from "@/api/testing";
import type {
  Envelope,
  MarketStatusChangedPayload,
  OddsChangedPayload,
} from "@/contract/events";
import type { GetMatchSnapshotResponse } from "@/contract/rest";
import {
  Transport,
  type WebSocketLike,
} from "@/realtime/transport";
import {
  createDefaultStores,
  StoresProvider,
  type Stores,
} from "@/react/StoresProvider";

import MatchDetailPage from "./page";

// FakeWebSocket — local-only test double (same pattern as
// realtime/transport.test.ts and src/app/page.test.tsx).
class FakeWebSocket implements WebSocketLike {
  static instances: FakeWebSocket[] = [];
  static factory(url: string): FakeWebSocket {
    const ws = new FakeWebSocket(url);
    FakeWebSocket.instances.push(ws);
    return ws;
  }
  static reset(): void {
    FakeWebSocket.instances = [];
  }
  readyState = 0;
  sent: string[] = [];
  onopen: ((ev?: unknown) => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: ((ev: { code: number; reason?: string }) => void) | null = null;
  onerror: ((ev?: unknown) => void) | null = null;
  constructor(public url: string) {}
  send(data: string): void {
    this.sent.push(data);
  }
  close(code = 1000): void {
    this.readyState = 3;
    this.onclose?.({ code });
  }
  fireOpen(): void {
    this.readyState = 1;
    this.onopen?.();
  }
  fireMessage(env: unknown): void {
    this.onmessage?.({ data: JSON.stringify(env) });
  }
  parsedSent(): unknown[] {
    return this.sent.map((s) => JSON.parse(s));
  }
}

function makeSnapshot(): GetMatchSnapshotResponse {
  return {
    match: {
      match_id: "42",
      tournament_id: "t1",
      home_team: "Alpha",
      away_team: "Beta",
      scheduled_at: "2026-05-16T00:00:00Z",
      status: "live",
      is_live: true,
      version: 1,
    },
    markets: [
      {
        market_id: "m1",
        market_type_id: "1x2",
        status: "active",
        outcomes: [
          { outcome_id: "o1", odds: 1.5, active: true },
          { outcome_id: "o2", odds: 2.5, active: true },
        ],
        version: 1,
      },
      {
        market_id: "m2",
        market_type_id: "total",
        status: "active",
        outcomes: [{ outcome_id: "o3", odds: 1.85, active: true }],
        version: 1,
      },
    ],
  };
}

function makeBundle(snapshot: GetMatchSnapshotResponse): Stores {
  const stub = new StubRestClient({
    responses: [
      {
        match: {
          method: "GET",
          path: `/api/v1/matches/${snapshot.match.match_id}`,
        },
        response: {
          status: "ok",
          body: snapshot,
          correlation_id: "test-corr",
          http_status: 200,
        },
      },
    ],
  });
  const transport = new Transport({
    url: "ws://test/ws",
    webSocketFactory: FakeWebSocket.factory,
  });
  return createDefaultStores({ restClient: stub.asClient(), transport });
}

// Given /api/v1/matches/42 returns a snapshot with 2 markets and team
//   names "Alpha" vs "Beta"
// When the match-detail page renders for id=42
// Then the header shows "Alpha" and "Beta" and a markets count of 2
describe("given match snapshot for id=42", () => {
  it("when page renders then header shows team names and markets count", async () => {
    FakeWebSocket.reset();
    const bundle = makeBundle(makeSnapshot());
    const { findByText, getByTestId } = render(
      <StoresProvider value={bundle}>
        <MatchDetailPage params={{ id: "42" }} />
      </StoresProvider>,
    );

    // Header appears once the snapshot resolves and the page re-renders.
    // Team names live inside an <h1> alongside a "vs" <span>, so we match
    // via regex rather than exact text equality on the parent element.
    await findByText(/Alpha/);
    await findByText(/Beta/);
    await waitFor(() => {
      expect(getByTestId("markets-count").textContent).toBe("2");
    });
  });
});

// Given a subscribed match-detail page with the initial snapshot loaded
// When the Transport emits an odds.changed envelope for one of the markets
//   with a higher version
// Then the page re-renders and the markets list reflects the new version
//   for that market (MarketsStore.listMarkets sees the update)
describe("given subscribed match-detail page", () => {
  it("when ws odds_changed arrives then the markets list reflects the new version", async () => {
    FakeWebSocket.reset();
    const bundle = makeBundle(makeSnapshot());
    render(
      <StoresProvider value={bundle}>
        <MatchDetailPage params={{ id: "42" }} />
      </StoresProvider>,
    );

    // Wait for initial hydrate to finish.
    await waitFor(() => {
      expect(bundle.markets.listMarkets("42").length).toBe(2);
    });
    expect(bundle.markets.listMarkets("42").find((m) => m.market_id === "m1")
      ?.version).toBe(1);

    const ws = FakeWebSocket.instances[0]!;
    act(() => {
      ws.fireOpen();
    });

    const oddsEnv: Envelope<OddsChangedPayload> = {
      type: "odds.changed",
      schema_version: "1",
      event_id: "evt-odds-1",
      correlation_id: "corr-odds-1",
      product_id: "live",
      occurred_at: "2026-05-16T00:02:00Z",
      received_at: "2026-05-16T00:02:00Z",
      entity: { match_id: "42", market_id: "m1" },
      payload: {
        match_id: "42",
        market_id: "m1",
        outcomes: [
          { outcome_id: "o1", odds: 1.7, active: true },
          { outcome_id: "o2", odds: 2.2, active: true },
        ],
        version: 2,
      },
    };
    act(() => {
      ws.fireMessage(oddsEnv);
    });

    await waitFor(() => {
      const m1 = bundle.markets
        .listMarkets("42")
        .find((m) => m.market_id === "m1");
      expect(m1?.version).toBe(2);
      expect(m1?.outcomes.find((o) => o.outcome_id === "o1")?.odds).toBe(1.7);
    });
  });
});

// Given a mounted match-detail page that issued
//   transport.subscribe({ match_ids: ["42"] }) on mount
// When the page unmounts
// Then transport.unsubscribe({ match_ids: ["42"] }) is called exactly once
describe("given a mounted match-detail page with an active subscription", () => {
  it("when the page unmounts then transport.unsubscribe is called with the match scope", async () => {
    FakeWebSocket.reset();
    const bundle = makeBundle(makeSnapshot());
    const subscribeSpy = vi.spyOn(bundle.transport, "subscribe");
    const unsubscribeSpy = vi.spyOn(bundle.transport, "unsubscribe");

    const { unmount } = render(
      <StoresProvider value={bundle}>
        <MatchDetailPage params={{ id: "42" }} />
      </StoresProvider>,
    );

    // Wait for the mount effect to have called subscribe — it does so as
    // soon as the page-level effect runs (not gated on the snapshot
    // resolution).
    await waitFor(() => {
      expect(subscribeSpy).toHaveBeenCalledWith({ match_ids: ["42"] });
    });

    unmount();

    expect(unsubscribeSpy).toHaveBeenCalledWith({ match_ids: ["42"] });
    expect(unsubscribeSpy).toHaveBeenCalledTimes(1);
  });
});

// Given a subscribed match-detail page with a markets list showing market
//   m1 as "active"
// When the Transport emits a market.status_changed envelope for m1
//   transitioning active→suspended at a higher version
// Then the rendered markets list row for m1 shows status="suspended" without
//   a snapshot re-fetch
describe("given subscribed match-detail page with rendered market statuses", () => {
  it("when ws market.status_changed flips m1 active→suspended then the list row reflects suspended", async () => {
    FakeWebSocket.reset();
    const bundle = makeBundle(makeSnapshot());
    const { findByTestId } = render(
      <StoresProvider value={bundle}>
        <MatchDetailPage params={{ id: "42" }} />
      </StoresProvider>,
    );

    // Initial render — m1 status is "active" from the snapshot.
    const initialRow = await findByTestId("market-row-m1");
    expect(initialRow.textContent).toMatch(/active/);

    const ws = FakeWebSocket.instances[0]!;
    act(() => {
      ws.fireOpen();
    });

    const statusEnv: Envelope<MarketStatusChangedPayload> = {
      type: "market.status_changed",
      schema_version: "1",
      event_id: "evt-mstatus-1",
      correlation_id: "corr-mstatus-1",
      product_id: "live",
      occurred_at: "2026-05-16T00:03:00Z",
      received_at: "2026-05-16T00:03:00Z",
      entity: { match_id: "42", market_id: "m1" },
      payload: {
        match_id: "42",
        market_id: "m1",
        status: "suspended",
        version: 2,
      },
    };
    act(() => {
      ws.fireMessage(statusEnv);
    });

    await waitFor(() => {
      const row = bundle.markets
        .listMarkets("42")
        .find((m) => m.market_id === "m1");
      expect(row?.status).toBe("suspended");
    });

    // And the DOM row reflects it.
    const updatedRow = await findByTestId("market-row-m1");
    expect(updatedRow.textContent).toMatch(/suspended/);
    expect(updatedRow.textContent).not.toMatch(/active/);
  });
});
