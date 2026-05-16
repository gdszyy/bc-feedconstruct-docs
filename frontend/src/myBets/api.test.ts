// frontend/src/myBets/api.test.ts
//
// MyBetsApi — adapter over RestClient for /my-bets.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type { GetMyBetsResponse, MyBet } from "@/contract/rest";

import { MyBetsApi } from "./api";

describe("MyBetsApi.fetchMyBets: encodes status filter as repeated params", () => {
  it("when multi-valued status is provided then status appears once per value", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: { bets: [], count: 0 } satisfies GetMyBetsResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMyBetsResponse>,
    });
    const api = new MyBetsApi(stub.asClient());
    await api.fetchMyBets({ status: ["Pending", "Accepted"], limit: 10 });
    expect(stub.calls[0]?.path).toBe(
      "/api/v1/my-bets?limit=10&status=Pending&status=Accepted",
    );
  });

  it("when no query is provided then the path is bare /my-bets", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: { bets: [], count: 0 } satisfies GetMyBetsResponse,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMyBetsResponse>,
    });
    const api = new MyBetsApi(stub.asClient());
    await api.fetchMyBets();
    expect(stub.calls[0]?.path).toBe("/api/v1/my-bets");
  });
});

describe("MyBetsApi.fetchMyBetById: routes to /my-bets/{id}", () => {
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
    const api = new MyBetsApi(stub.asClient());
    await api.fetchMyBetById("bet 42/x");
    expect(stub.calls[0]?.path).toBe("/api/v1/my-bets/bet%2042%2Fx");
  });
});

describe("MyBetsApi.fetchMyBetById: 404 surfaced as error", () => {
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
    const api = new MyBetsApi(stub.asClient());
    const result = await api.fetchMyBetById("unknown");
    expect(result.status).toBe("error");
  });
});
