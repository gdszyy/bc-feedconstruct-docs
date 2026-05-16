// frontend/src/betSlip/api.test.ts
//
// BetSlipApi — adapter over RestClient for /bet-slip/validate and /place.

import { describe, expect, it } from "vitest";

import type { ApiResult } from "@/api/client";
import { StubRestClient } from "@/api/testing";
import type {
  PlaceBetResponse,
  ValidateBetSlipRequest,
  ValidateBetSlipResponse,
} from "@/contract/rest";

import { BetSlipApi } from "./api";

describe("BetSlipApi.validate: posts request to /bet-slip/validate", () => {
  it("when validate is invoked then the client posts the slip body", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: { valid: true } satisfies ValidateBetSlipResponse,
        correlation_id: "c-1",
        http_status: 200,
      } as ApiResult<ValidateBetSlipResponse>,
    });
    const api = new BetSlipApi(stub.asClient());
    const req: ValidateBetSlipRequest = {
      selections: [
        {
          position: 1,
          match_id: "m1",
          market_id: "mk1",
          outcome_id: "home",
          locked_odds: 2.0,
        },
      ],
      stake: 10,
      currency: "USD",
      bet_type: "single",
    };
    const result = await api.validate(req, "corr-explicit");
    expect(stub.calls).toEqual([
      {
        method: "POST",
        path: "/api/v1/bet-slip/validate",
        options: { body: req, correlationId: "corr-explicit" },
      },
    ]);
    expect(result.status).toBe("ok");
  });
});

describe("BetSlipApi.place: posts with idempotency key", () => {
  it("when place is invoked then the request carries the idempotency key", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "ok",
        body: {
          bet_id: "bet-42",
          state: "Accepted",
          deduped: false,
        } satisfies PlaceBetResponse,
        correlation_id: "c-1",
        http_status: 200,
      } as ApiResult<PlaceBetResponse>,
    });
    const api = new BetSlipApi(stub.asClient());
    await api.place(
      {
        selections: [],
        stake: 10,
        currency: "USD",
        bet_type: "single",
      },
      "idem-1",
    );
    expect(stub.calls[0]?.method).toBe("POST");
    expect(stub.calls[0]?.path).toBe("/api/v1/bet-slip/place");
    expect(stub.calls[0]?.options.idempotencyKey).toBe("idem-1");
  });
});

describe("BetSlipApi.place: 4xx error surfaced as result.error", () => {
  it("when the server rejects then the adapter returns status=error without throwing", async () => {
    const stub = new StubRestClient({
      defaultResponse: {
        status: "error",
        error: {
          code: "PRICE_CHANGED",
          message: "Odds moved",
          retriable: false,
          correlation_id: "c-1",
        },
        http_status: 400,
        correlation_id: "c-1",
      } as ApiResult<PlaceBetResponse>,
    });
    const api = new BetSlipApi(stub.asClient());
    const result = await api.place(
      { selections: [], stake: 1, currency: "USD", bet_type: "single" },
      "idem-1",
    );
    expect(result.status).toBe("error");
    if (result.status === "error") {
      expect(result.error.code).toBe("PRICE_CHANGED");
    }
  });
});
