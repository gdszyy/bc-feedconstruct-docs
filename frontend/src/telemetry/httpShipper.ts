import type { RestClient } from "@/api/client";

import type { TelemetryEvent, TelemetryShipper } from "./store";

// ---------------------------------------------------------------------------
// M16 — HttpTelemetryShipper
//
// Concrete TelemetryShipper that POSTs batches to a configurable endpoint
// via the central RestClient. Empty batches are a no-op. Non-2xx /
// network failures reject the promise so the TelemetryStore retains the
// queue (locked decision #4 from M16).
// ---------------------------------------------------------------------------

export interface HttpTelemetryShipperOptions {
  client: RestClient;
  /** Endpoint path. Defaults to `/api/v1/telemetry/events`. */
  endpoint?: string;
}

export class HttpTelemetryShipper implements TelemetryShipper {
  private readonly client: RestClient;
  private readonly endpoint: string;

  constructor(opts: HttpTelemetryShipperOptions) {
    this.client = opts.client;
    this.endpoint = opts.endpoint ?? "/api/v1/telemetry/events";
  }

  async ship(batch: TelemetryEvent[]): Promise<void> {
    if (batch.length === 0) return;
    const result = await this.client.post<unknown>(this.endpoint, {
      body: { events: batch },
    });
    if (result.status !== "ok") {
      const err =
        result.status === "error"
          ? result.error
          : { code: "UNEXPECTED", message: "non-ok telemetry response" };
      throw new Error(`telemetry ship failed: ${err.code} ${err.message}`);
    }
  }
}
