import type { RestClient } from "@/api/client";

import type { TelemetryEvent, TelemetryShipper } from "./store";

// ---------------------------------------------------------------------------
// M16 — HTTP TelemetryShipper factory
//
// Concrete TelemetryShipper that POSTs batches via the central RestClient
// (so the telemetry endpoint inherits auth + correlation_id + Idempotency-Key
// machinery). Empty batches are a no-op. Non-2xx / network failures throw
// so the TelemetryStore retains the queue (locked decision: retain & retry).
// ---------------------------------------------------------------------------

export interface HttpTelemetryShipperOptions {
  client: RestClient;
  /** Endpoint path. Defaults to `/api/v1/telemetry/events`. */
  endpoint?: string;
}

const DEFAULT_ENDPOINT = "/api/v1/telemetry/events";

export function createHttpTelemetryShipper(
  opts: HttpTelemetryShipperOptions,
): TelemetryShipper {
  const endpoint = opts.endpoint ?? DEFAULT_ENDPOINT;
  return {
    async ship(batch: TelemetryEvent[]): Promise<void> {
      if (batch.length === 0) return;
      const result = await opts.client.post<unknown>(endpoint, {
        body: { events: batch },
      });
      if (result.status !== "ok") {
        const err =
          result.status === "error"
            ? result.error
            : { code: "UNEXPECTED", message: "non-ok telemetry response" };
        throw new Error(`telemetry ship failed: ${err.code} ${err.message}`);
      }
    },
  };
}
