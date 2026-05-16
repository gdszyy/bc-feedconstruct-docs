// frontend/src/descriptions/api.test.ts

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type {
  GetMarketDescriptionsResponse,
  GetOutcomeDescriptionsResponse,
} from "@/contract/rest";

import { fetchMarketDescriptions, fetchOutcomeDescriptions } from "./api";

describe("fetchMarketDescriptions: initial fetch", () => {
  it("when called without ifNoneMatch then no If-None-Match header is forwarded and lang is in query", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          version: "v1",
          descriptions: [],
        } satisfies GetMarketDescriptionsResponse,
        correlation_id: "c-1",
        http_status: 200,
      } as ApiResult<GetMarketDescriptionsResponse>,
    });
    await fetchMarketDescriptions(stub.asClient(), { lang: "en" });
    expect(stub.calls[0]?.path).toBe("/api/v1/descriptions/markets");
    expect(stub.calls[0]?.options.query).toEqual({ lang: "en" });
    expect(stub.calls[0]?.options.ifNoneMatch).toBeUndefined();
  });
});

describe("fetchMarketDescriptions: 304 round-trip", () => {
  it("when the cached ETag still matches then status=not-modified", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "not-modified",
        etag: "v1",
        correlation_id: "c-1",
        http_status: 304,
      } as ApiResult<GetMarketDescriptionsResponse>,
    });
    const result = await fetchMarketDescriptions(stub.asClient(), {
      lang: "en",
      ifNoneMatch: "v1",
    });
    expect(stub.calls[0]?.options.ifNoneMatch).toBe("v1");
    expect(result.status).toBe("not-modified");
  });
});

describe("fetchOutcomeDescriptions: routes correctly", () => {
  it("when fetchOutcomeDescriptions is invoked then the right URL is used", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          version: "v1",
          descriptions: [],
        } satisfies GetOutcomeDescriptionsResponse,
        correlation_id: "c-1",
        http_status: 200,
      } as ApiResult<GetOutcomeDescriptionsResponse>,
    });
    await fetchOutcomeDescriptions(stub.asClient(), { lang: "fr" });
    expect(stub.calls[0]?.path).toBe("/api/v1/descriptions/outcomes");
    expect(stub.calls[0]?.options.query).toEqual({ lang: "fr" });
  });
});
