// frontend/src/subscription/api.test.ts
//
// SubscriptionApi — adapter over RestClient for /subscriptions.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type { CreateSubscriptionResponse } from "@/contract/rest";

import { SubscriptionApi } from "./api";

describe("SubscriptionApi.createSubscription: posts to /subscriptions", () => {
  it("when createSubscription is invoked then the POST carries the match_id", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          subscription_id: "sub-1",
          match_id: "m1",
          state: "active",
        } satisfies CreateSubscriptionResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<CreateSubscriptionResponse>,
    });
    const api = new SubscriptionApi(stub.asClient());
    await api.createSubscription({ match_id: "m1" });
    expect(stub.calls[0]?.method).toBe("POST");
    expect(stub.calls[0]?.path).toBe("/api/v1/subscriptions");
    expect(stub.calls[0]?.options.body).toEqual({ match_id: "m1" });
  });
});

describe("SubscriptionApi.deleteSubscription: routes to /subscriptions/{id}", () => {
  it("when deleteSubscription is invoked then the DELETE targets /subscriptions/{id}", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: undefined,
        correlation_id: "c",
        http_status: 204,
      } as ApiResult<void>,
    });
    const api = new SubscriptionApi(stub.asClient());
    await api.deleteSubscription("sub-1");
    expect(stub.calls[0]?.method).toBe("DELETE");
    expect(stub.calls[0]?.path).toBe("/api/v1/subscriptions/sub-1");
  });
});
