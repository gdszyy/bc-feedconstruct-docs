// frontend/src/myBets/api.test.ts

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type { GetMyBetsResponse, MyBet } from "@/contract/rest";

import { fetchMyBetById, fetchMyBets } from "./api";

describe("fetchMyBets: status filter forwarded as array query", () => {
  it("when multi-valued status is provided then it is forwarded as an array to RestClient", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: { bets: [], count: 0 } satisfies GetMyBetsResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMyBetsResponse>,
    });
    await fetchMyBets(stub.asClient(), {
      status: ["Pending", "Accepted"],
      limit: 10,
    });
    expect(stub.calls[0]?.path).toBe("/api/v1/my-bets");
    expect(stub.calls[0]?.options.query).toEqual({
      user_id: undefined,
      limit: 10,
      status: ["Pending", "Accepted"],
    });
  });

  it("when no query is provided then the path stays /my-bets and undefined keys are skipped by RestClient", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: { bets: [], count: 0 } satisfies GetMyBetsResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMyBetsResponse>,
    });
    await fetchMyBets(stub.asClient());
    expect(stub.calls[0]?.path).toBe("/api/v1/my-bets");
  });
});

describe("fetchMyBetById: routes to /my-bets/{id}", () => {
  it("when fetchMyBetById is invoked then the path interpolates the bet id", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          id: "bet-42",
          user_id: "u1",
          placed_at: "t",
          stake: 10,
          currency: "USD",
          bet_type: "single",
          state: "Accepted",
          selections: [],
          history: [],
        } satisfies MyBet,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<MyBet>,
    });
    await fetchMyBetById(stub.asClient(), "bet 42/x");
    expect(stub.calls[0]?.path).toBe("/api/v1/my-bets/bet%2042%2Fx");
  });
});

describe("fetchMyBetById: 404 surfaced as error", () => {
  it("when the bet is unknown then the adapter returns status=error", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "error",
        error: {
          code: "HTTP_404",
          message: "not found",
          retriable: false,
          correlation_id: "c",
        },
        http_status: 404,
        correlation_id: "c",
      } as ApiResult<MyBet>,
    });
    const result = await fetchMyBetById(stub.asClient(), "unknown");
    expect(result.status).toBe("error");
  });
});
