// frontend/src/match/api.test.ts
//
// Behaviour: fetchMatchSnapshot — startup hydration for a single match.
//
//   Given a RestClient stub configured for /api/v1/matches/{id}
//   When  fetchMatchSnapshot is invoked
//   Then  GET targets the correct path, forwards correlationId, and returns
//         the parsed body unchanged.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type { GetMatchSnapshotResponse } from "@/contract/rest";

import { fetchMatchSnapshot } from "./api";

function makeSnapshot(): GetMatchSnapshotResponse {
  return {
    match: {
      match_id: "42",
      tournament_id: "t1",
      home_team: "Alpha",
      away_team: "Beta",
      scheduled_at: "2026-05-16T00:00:00Z",
      status: "live",
      is_live: true,
      version: 1,
    },
    markets: [],
  };
}

describe("given a RestClient stubbed for /api/v1/matches/{id}", () => {
  it("when fetchMatchSnapshot is called then it GETs the correct path and forwards correlationId", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: makeSnapshot(),
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMatchSnapshotResponse>,
    });
    await fetchMatchSnapshot(stub.asClient(), "42", "corr-m42");
    expect(stub.calls).toEqual([
      {
        method: "GET",
        path: "/api/v1/matches/42",
        options: { correlationId: "corr-m42" },
      },
    ]);
  });

  it("when matchId contains URL-reserved characters then they are percent-encoded", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: makeSnapshot(),
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMatchSnapshotResponse>,
    });
    await fetchMatchSnapshot(stub.asClient(), "sr:match/123 weird");
    expect(stub.calls[0]!.path).toBe(
      "/api/v1/matches/sr%3Amatch%2F123%20weird",
    );
  });
});

describe("given the upstream returns a snapshot body", () => {
  it("when fetchMatchSnapshot resolves then the body is returned unchanged", async () => {
    const expected = makeSnapshot();
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: expected,
        correlation_id: "c",
        http_status: 200,
      } as ApiResult<GetMatchSnapshotResponse>,
    });
    const result = await fetchMatchSnapshot(stub.asClient(), "42");
    expect(result.status).toBe("ok");
    if (result.status === "ok") expect(result.body).toEqual(expected);
  });
});
