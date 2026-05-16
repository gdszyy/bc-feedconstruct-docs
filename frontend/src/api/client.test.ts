// frontend/src/api/client.test.ts
//
// RestClient — central /api/v1 wrapper.
// Locked decisions per PR thread:
//   1. Result discriminated union (no throw)
//   2. Async tokenProvider
//   3. Client-generated, canonical correlation_id
//   4. Caller-managed ETag (ifNoneMatch in, etag out)

import { describe, expect, it, vi } from "vitest";

import {
  type FetchLike,
  type FetchResponseLike,
  RestClient,
} from "./client";

interface Recorded {
  url: string;
  method: string;
  headers: Record<string, string>;
  body?: string;
}

function makeFetch(
  responder: (req: Recorded) => Partial<FetchResponseLike> | Promise<Partial<FetchResponseLike>>,
): { fetch: FetchLike; calls: Recorded[] } {
  const calls: Recorded[] = [];
  const fetch: FetchLike = async (url, init = {}) => {
    const recorded: Recorded = {
      url,
      method: init.method ?? "GET",
      headers: (init.headers ?? {}) as Record<string, string>,
      body: init.body,
    };
    calls.push(recorded);
    const partial = await responder(recorded);
    const headers = (partial.headers ?? { get: () => null }) as {
      get(name: string): string | null;
    };
    return {
      status: partial.status ?? 200,
      headers,
      text: partial.text ?? (() => Promise.resolve("")),
    };
  };
  return { fetch, calls };
}

function jsonResp(
  status: number,
  body: unknown,
  extraHeaders: Record<string, string> = {},
): Partial<FetchResponseLike> {
  const headers = new Map(Object.entries(extraHeaders));
  return {
    status,
    headers: { get: (n) => headers.get(n) ?? null },
    text: () => Promise.resolve(JSON.stringify(body)),
  };
}

function client(opts: {
  fetch: FetchLike;
  baseUrl?: string;
  tokenProvider?: () => Promise<string | undefined>;
  generateCorrelationId?: () => string;
  now?: () => number;
  telemetry?: ConstructorParameters<typeof RestClient>[0]["telemetry"];
}): RestClient {
  return new RestClient({
    baseUrl: opts.baseUrl ?? "https://api.example.com",
    fetch: opts.fetch,
    tokenProvider: opts.tokenProvider,
    generateCorrelationId: opts.generateCorrelationId,
    now: opts.now,
    telemetry: opts.telemetry,
  });
}

// =================== Request shape ===================

describe("RestClient request: path joined with base URL", () => {
  it("when get is invoked with a path then fetch sees base + path", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, { ok: true }));
    const c = client({ fetch });
    await c.get("/api/v1/foo");
    expect(calls[0]?.url).toBe("https://api.example.com/api/v1/foo");
    expect(calls[0]?.method).toBe("GET");
  });

  it("when query params are provided then they are appended", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({ fetch });
    await c.get("/api/v1/matches", {
      query: { filter: "live", limit: 10, cursor: undefined, sport_id: null },
    });
    expect(calls[0]?.url).toBe(
      "https://api.example.com/api/v1/matches?filter=live&limit=10",
    );
  });
});

// =================== Auth ===================

describe("RestClient auth: token provider injects Authorization header", () => {
  it("when an auth token is available then requests carry Bearer token", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({
      fetch,
      tokenProvider: async () => "jwt-1",
    });
    await c.get("/x");
    expect(calls[0]?.headers.Authorization).toBe("Bearer jwt-1");
  });
});

describe("RestClient auth: missing token sends no header", () => {
  it("when no token provider is configured then no Authorization header is sent", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({ fetch });
    await c.get("/x");
    expect(calls[0]?.headers.Authorization).toBeUndefined();
  });

  it("when token provider returns undefined then no Authorization header is sent", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({ fetch, tokenProvider: async () => undefined });
    await c.get("/x");
    expect(calls[0]?.headers.Authorization).toBeUndefined();
  });
});

// =================== Correlation ===================

describe("RestClient correlation: generates and attaches X-Correlation-Id", () => {
  it("when a request is dispatched then a correlation_id header is generated", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({
      fetch,
      generateCorrelationId: () => "corr-gen-1",
    });
    const result = await c.get("/x");
    expect(calls[0]?.headers["X-Correlation-Id"]).toBe("corr-gen-1");
    expect(result.correlation_id).toBe("corr-gen-1");
  });
});

describe("RestClient correlation: caller-supplied id wins over generator", () => {
  it("when correlationId is explicitly passed then it is used unchanged", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({
      fetch,
      generateCorrelationId: () => "should-not-be-used",
    });
    const result = await c.get("/x", { correlationId: "explicit-1" });
    expect(calls[0]?.headers["X-Correlation-Id"]).toBe("explicit-1");
    expect(result.correlation_id).toBe("explicit-1");
  });
});

// =================== Idempotency ===================

describe("RestClient idempotency: POST with idempotencyKey sets header", () => {
  it("when idempotencyKey is passed to a POST then Idempotency-Key header is attached", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, { bet_id: "b1" }));
    const c = client({ fetch });
    await c.post("/api/v1/bet-slip/place", {
      idempotencyKey: "idem-1",
      body: { selections: [] },
    });
    expect(calls[0]?.headers["Idempotency-Key"]).toBe("idem-1");
    expect(calls[0]?.method).toBe("POST");
    expect(JSON.parse(calls[0]?.body ?? "{}")).toEqual({ selections: [] });
  });
});

// =================== ETag ===================

describe("RestClient ETag: ifNoneMatch attaches If-None-Match header", () => {
  it("when ifNoneMatch is provided then If-None-Match header is sent", async () => {
    const { fetch, calls } = makeFetch(() => jsonResp(200, {}));
    const c = client({ fetch });
    await c.get("/x", { ifNoneMatch: "v1" });
    expect(calls[0]?.headers["If-None-Match"]).toBe("v1");
  });
});

describe("RestClient ETag: 304 returns not-modified", () => {
  it("when the server responds 304 then status is not-modified and body is undefined", async () => {
    const { fetch } = makeFetch(() => ({
      status: 304,
      headers: { get: (n) => (n === "ETag" ? "v1" : null) },
      text: () => Promise.resolve(""),
    }));
    const c = client({ fetch });
    const result = await c.get("/x", { ifNoneMatch: "v1" });
    expect(result.status).toBe("not-modified");
    if (result.status === "not-modified") {
      expect(result.etag).toBe("v1");
      expect(result.http_status).toBe(304);
    }
  });
});

describe("RestClient ETag: 200 returns parsed body + ETag value", () => {
  it("when the server responds 200 with ETag then etag and body are both returned", async () => {
    const { fetch } = makeFetch(() => jsonResp(200, { hello: "world" }, { ETag: "v2" }));
    const c = client({ fetch });
    const result = await c.get<{ hello: string }>("/x");
    expect(result.status).toBe("ok");
    if (result.status === "ok") {
      expect(result.body).toEqual({ hello: "world" });
      expect(result.etag).toBe("v2");
      expect(result.http_status).toBe(200);
    }
  });
});

// =================== Response ===================

describe("RestClient response: 204 returns ok with undefined body", () => {
  it("when the server responds 204 then status is ok and body is undefined", async () => {
    const { fetch } = makeFetch(() => ({
      status: 204,
      headers: { get: () => null },
      text: () => Promise.resolve(""),
    }));
    const c = client({ fetch });
    const result = await c.delete("/x");
    expect(result.status).toBe("ok");
    if (result.status === "ok") {
      expect(result.body).toBeUndefined();
    }
  });
});

// =================== Errors ===================

describe("RestClient error: ApiError-shape JSON body parsed", () => {
  it("when the server responds 4xx with ApiError JSON then code/message/correlation_id are surfaced", async () => {
    const { fetch } = makeFetch(() =>
      jsonResp(400, {
        error: {
          code: "BET_REJECTED_PRICE_CHANGED",
          message: "Odds moved",
          retriable: false,
          correlation_id: "server-corr-1",
        },
      }),
    );
    const c = client({
      fetch,
      generateCorrelationId: () => "client-corr-1",
    });
    const result = await c.post("/api/v1/bet-slip/place", { body: {} });
    expect(result.status).toBe("error");
    if (result.status === "error") {
      expect(result.error.code).toBe("BET_REJECTED_PRICE_CHANGED");
      expect(result.error.message).toBe("Odds moved");
      expect(result.error.retriable).toBe(false);
      // Server-provided correlation_id in body is preserved.
      expect(result.error.correlation_id).toBe("server-corr-1");
      // Result-level correlation_id remains the client-canonical one.
      expect(result.correlation_id).toBe("client-corr-1");
    }
  });
});

describe("RestClient error: non-JSON 5xx synthesised", () => {
  it("when the server responds 5xx with a non-JSON body then a synthesised ApiError is returned", async () => {
    const { fetch } = makeFetch(() => ({
      status: 503,
      headers: { get: () => null },
      text: () => Promise.resolve("<html>service down</html>"),
    }));
    const c = client({ fetch });
    const result = await c.get("/x");
    expect(result.status).toBe("error");
    if (result.status === "error") {
      expect(result.error.code).toBe("HTTP_503");
      expect(result.error.retriable).toBe(true);
    }
  });
});

describe("RestClient error: network failure synthesised", () => {
  it("when fetch throws then a NETWORK_ERROR ApiError is returned with retriable=true", async () => {
    const fetch: FetchLike = async () => {
      throw new Error("ECONNREFUSED");
    };
    const c = client({ fetch });
    const result = await c.get("/x");
    expect(result.status).toBe("error");
    if (result.status === "error") {
      expect(result.error.code).toBe("NETWORK_ERROR");
      expect(result.error.retriable).toBe(true);
      expect(result.error.message).toContain("ECONNREFUSED");
      expect(result.http_status).toBe(-1);
    }
  });
});

// =================== Telemetry ===================

describe("RestClient telemetry: emits request metric on completion", () => {
  it("when a request completes then a metric event is emitted with method/status/latency", async () => {
    const { fetch } = makeFetch(() => jsonResp(200, {}));
    let nowCalls = 0;
    const telemetry = {
      requestCompleted: vi.fn(),
      requestFailed: vi.fn(),
    };
    const c = client({
      fetch,
      telemetry,
      now: () => (nowCalls++ === 0 ? 1000 : 1150),
    });
    await c.get("/api/v1/foo");
    expect(telemetry.requestCompleted).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "GET",
        path: "/api/v1/foo",
        http_status: 200,
        outcome: "ok",
        duration_ms: 150,
      }),
    );
    expect(telemetry.requestFailed).not.toHaveBeenCalled();
  });
});

describe("RestClient telemetry: emits error event on failure", () => {
  it("when a request fails then an error event is emitted with correlation_id", async () => {
    const { fetch } = makeFetch(() =>
      jsonResp(400, {
        error: { code: "VALIDATION", message: "bad" },
      }),
    );
    const telemetry = {
      requestCompleted: vi.fn(),
      requestFailed: vi.fn(),
    };
    const c = client({
      fetch,
      telemetry,
      generateCorrelationId: () => "c-1",
    });
    await c.post("/x", { body: {} });
    expect(telemetry.requestCompleted).toHaveBeenCalledWith(
      expect.objectContaining({ outcome: "error", http_status: 400 }),
    );
    expect(telemetry.requestFailed).toHaveBeenCalledWith({
      method: "POST",
      path: "/x",
      message: "bad",
      code: "VALIDATION",
      correlation_id: "c-1",
    });
  });

  it("when fetch throws then telemetry sees NETWORK_ERROR", async () => {
    const fetch: FetchLike = async () => {
      throw new Error("boom");
    };
    const telemetry = {
      requestCompleted: vi.fn(),
      requestFailed: vi.fn(),
    };
    const c = client({
      fetch,
      telemetry,
      generateCorrelationId: () => "c-1",
    });
    await c.get("/x");
    expect(telemetry.requestCompleted).toHaveBeenCalledWith(
      expect.objectContaining({ outcome: "network_error", http_status: -1 }),
    );
    expect(telemetry.requestFailed).toHaveBeenCalledWith(
      expect.objectContaining({ code: "NETWORK_ERROR" }),
    );
  });
});
