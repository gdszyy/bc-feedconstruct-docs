"use client";

import { useEffect, useRef } from "react";

import { fetchMatchSnapshot } from "@/match/api";
import type { MatchRecord } from "@/match/store";
import type { MarketRecord } from "@/markets/store";
import { useStore } from "@/react/useStore";
import { useStores } from "@/react/StoresProvider";

interface MatchDetailPageProps {
  params: { id: string };
}

export default function MatchDetailPage({ params }: MatchDetailPageProps) {
  const { id } = params;
  const stores = useStores();

  // Primitive selectors — listMarkets() returns a fresh array each call,
  // which would break useSyncExternalStore's referential-equality cache if
  // we selected the array itself.
  const marketsCount = useStore(stores.markets, (m) =>
    m.listMarkets(id).length,
  );
  const marketsVersionHash = useStore(stores.markets, (m) =>
    m
      .listMarkets(id)
      .map((mk) => `${mk.market_id}:${mk.version}`)
      .join("|"),
  );
  const homeTeam = useStore(
    stores.match,
    (s) => s.getMatch(id)?.home_team,
  );
  const awayTeam = useStore(
    stores.match,
    (s) => s.getMatch(id)?.away_team,
  );

  // Startup hydration. Run once per (stores, id) — guard against React
  // StrictMode double-invoke. Subsequent updates flow through the
  // Dispatcher into MatchStore / MarketsStore.
  const didHydrate = useRef<string | null>(null);
  useEffect(() => {
    if (didHydrate.current === id) return;
    didHydrate.current = id;
    void (async () => {
      const result = await fetchMatchSnapshot(stores.restClient, id);
      if (result.status !== "ok") return;
      const body = result.body;
      const record: MatchRecord = {
        match_id: body.match.match_id,
        tournament_id: body.match.tournament_id,
        home_team: body.match.home_team,
        away_team: body.match.away_team,
        scheduled_at: body.match.scheduled_at,
        is_live: body.match.is_live,
        status: body.match.status,
        home_score: body.match.home_score,
        away_score: body.match.away_score,
        period: body.match.period,
        version: body.match.version,
      };
      stores.match.hydrateMatches([record]);
      const marketRecords: MarketRecord[] = body.markets.map((m) => ({
        match_id: id,
        market_id: m.market_id,
        specifiers: m.specifiers ?? {},
        status: m.status,
        outcomes: m.outcomes.map((o) => ({ ...o })),
        version: m.version,
      }));
      stores.markets.hydrateMatchMarkets(id, marketRecords);
    })();
  }, [stores, id]);

  // Subscribe to live updates for this match. Effect runs once on mount
  // (React StrictMode double-invoke triggers cleanup + re-subscribe; this
  // is the same scope so Transport's internal map deduplicates).
  useEffect(() => {
    const scope = { match_ids: [id] };
    stores.transport.subscribe(scope);
    return () => {
      stores.transport.unsubscribe(scope);
    };
  }, [stores, id]);

  return (
    <main style={{ padding: 24, fontFamily: "system-ui" }}>
      <header>
        <h1>
          {homeTeam ?? "—"} <span style={{ opacity: 0.6 }}>vs</span>{" "}
          {awayTeam ?? "—"}
        </h1>
        <p>
          Match id: <code>{id}</code>
        </p>
      </header>
      <section>
        <h2>Markets</h2>
        <p>
          Count: <span data-testid="markets-count">{marketsCount}</span>
        </p>
        {/* marketsVersionHash is read so React subscribes to version
            changes; surfaced as a hidden data attribute for tests. */}
        <p
          data-testid="markets-version-hash"
          style={{ display: "none" }}
        >
          {marketsVersionHash}
        </p>
      </section>
    </main>
  );
}
