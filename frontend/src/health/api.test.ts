// frontend/src/health/api.test.ts
//
// HealthApi — adapter over RestClient for /system/health.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type { GetSystemHealthResponse } from "@/contract/rest";

import { HealthApi } from "./api";

describe("HealthApi.fetchSystemHealth: routes to /system/health", () => {
  it("when fetchSystemHealth is invoked then the GET targets /system/health", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          producers: [],
          degraded: false,
        } satisfies GetSystemHealthResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetSystemHealthResponse>,
    });
    const api = new HealthApi(stub.asClient());
    await api.fetchSystemHealth("corr-h");
    expect(stub.calls).toEqual([
      {
        method: "GET",
        path: "/api/v1/system/health",
        options: { correlationId: "corr-h" },
      },
    ]);
  });
});

describe("HealthApi.fetchSystemHealth: 200 returns producer states", () => {
  it("when the server responds 200 then the parsed body is returned", async () => {
    const expected: GetSystemHealthResponse = {
      producers: [
        {
          product: "live",
          is_down: true,
          last_message_at: "t",
          down_since: "t-1",
        },
      ],
      degraded: true,
    };
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: expected,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetSystemHealthResponse>,
    });
    const api = new HealthApi(stub.asClient());
    const result = await api.fetchSystemHealth();
    expect(result.status).toBe("ok");
    if (result.status === "ok") expect(result.body).toEqual(expected);
  });
});
