// frontend/src/telemetry/httpShipper.test.ts
//
// HttpTelemetryShipper — concrete TelemetryShipper that POSTs batches via RestClient.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";

import { HttpTelemetryShipper } from "./httpShipper";
import type { TelemetryEvent } from "./store";

function makeEvent(id: string): TelemetryEvent {
  return {
    id,
    kind: "log",
    level: "info",
    occurred_at: "t",
    correlation_id: `corr-${id}`,
    payload: { message: id },
  };
}

describe("HttpTelemetryShipper.ship: posts batch to configured endpoint", () => {
  it("when ship is invoked then a POST goes to the configured endpoint with the batch", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: undefined,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<unknown>,
    });
    const shipper = new HttpTelemetryShipper({
      client: stub.asClient(),
      endpoint: "/api/v1/telemetry/events",
    });
    const batch = [makeEvent("a"), makeEvent("b")];
    await shipper.ship(batch);
    expect(stub.calls[0]?.method).toBe("POST");
    expect(stub.calls[0]?.path).toBe("/api/v1/telemetry/events");
    expect(stub.calls[0]?.options.body).toEqual({ events: batch });
  });

  it("when endpoint is omitted then the default /api/v1/telemetry/events is used", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: undefined,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<unknown>,
    });
    const shipper = new HttpTelemetryShipper({ client: stub.asClient() });
    await shipper.ship([makeEvent("a")]);
    expect(stub.calls[0]?.path).toBe("/api/v1/telemetry/events");
  });
});

describe("HttpTelemetryShipper.ship: rejects on error result", () => {
  it("when the client returns status=error then ship rejects (to engage retain & retry)", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "error",
        error: {
          code: "HTTP_503",
          message: "down",
          retriable: true,
          correlation_id: "c",
        },
        http_status: 503,
        correlation_id: "c",
      } as ApiResult<unknown>,
    });
    const shipper = new HttpTelemetryShipper({ client: stub.asClient() });
    await expect(shipper.ship([makeEvent("a")])).rejects.toThrow(/HTTP_503/);
  });
});

describe("HttpTelemetryShipper.ship: empty batch is a no-op", () => {
  it("when ship is invoked with [] then no fetch is made and the call resolves", async () => {
    const stub = new StubRestClient();
    const shipper = new HttpTelemetryShipper({ client: stub.asClient() });
    await shipper.ship([]);
    expect(stub.calls).toHaveLength(0);
  });
});
