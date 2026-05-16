"use client";

import { useEffect, useRef } from "react";

import { fetchSystemHealth } from "@/health/api";
import { useStore } from "@/react/useStore";
import { useStores } from "@/react/StoresProvider";

export default function HomePage() {
  const stores = useStores();

  // Select primitives — getBanner() returns a fresh object each call, so
  // selecting the object directly would break useSyncExternalStore's
  // referential-equality cache.
  const bannerLevel = useStore(stores.health, (h) => h.getBanner()?.level);
  const bannerMessage = useStore(stores.health, (h) => h.getBanner()?.message);

  // Startup hydrate. Subsequent updates flow via the Dispatcher into the
  // HealthStore (system.producer_status WS events).
  const didHydrate = useRef(false);
  useEffect(() => {
    if (didHydrate.current) return;
    didHydrate.current = true;
    void (async () => {
      const result = await fetchSystemHealth(stores.restClient);
      if (result.status === "ok") {
        stores.health.hydrate(result.body);
      }
    })();
  }, [stores]);

  return (
    <main style={{ padding: 24, fontFamily: "system-ui" }}>
      <h1>BC FeedConstruct Web</h1>

      {bannerLevel && bannerMessage ? (
        <div
          role="alert"
          data-testid="health-banner"
          data-level={bannerLevel}
          style={{
            margin: "12px 0",
            padding: "8px 12px",
            border: "1px solid",
            borderColor:
              bannerLevel === "error"
                ? "#c00"
                : bannerLevel === "warn"
                  ? "#e88"
                  : "#88c",
            background:
              bannerLevel === "error"
                ? "#fee"
                : bannerLevel === "warn"
                  ? "#ffe"
                  : "#eef",
          }}
        >
          {bannerMessage}
        </div>
      ) : null}

      <p>
        Live BFF consumer. Realtime updates flow over WebSocket; REST is used
        for initial snapshots.
      </p>
      <ul>
        <li>
          BFF HTTP: <code>{process.env.NEXT_PUBLIC_BFF_HTTP ?? "(unset)"}</code>
        </li>
        <li>
          BFF WS: <code>{process.env.NEXT_PUBLIC_BFF_WS ?? "(unset)"}</code>
        </li>
      </ul>
    </main>
  );
}
