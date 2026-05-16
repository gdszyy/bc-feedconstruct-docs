import type { ApiError } from "@/contract/rest";

// ---------------------------------------------------------------------------
// RestClient — central wrapper for /api/v1 endpoints.
//
// Locked decisions (PR thread):
//   1. Result discriminated union (`status: ok | not-modified | error`).
//      No throw on non-2xx; network errors are normalised to ApiError shape
//      with code=NETWORK_ERROR.
//   2. tokenProvider is async — supports refresh-token flows.
//   3. correlation_id is client-generated, canonical. The server cannot
//      override it; the same id flows through telemetry.
//   4. ETag cache is caller-managed: methods accept `ifNoneMatch`, return
//      `etag` on success. Adapters/stores cache the ETag themselves.
// ---------------------------------------------------------------------------

export type FetchLike = (
  input: string,
  init?: {
    method?: string;
    headers?: Record<string, string>;
    body?: string;
  },
) => Promise<FetchResponseLike>;

export interface FetchResponseLike {
  status: number;
  headers: { get(name: string): string | null };
  text(): Promise<string>;
}

export type ApiErrorBody = ApiError["error"];

export type ApiResult<T> =
  | {
      status: "ok";
      body: T;
      etag?: string;
      correlation_id: string;
      http_status: number;
    }
  | {
      status: "not-modified";
      etag?: string;
      correlation_id: string;
      http_status: 304;
    }
  | {
      status: "error";
      error: ApiErrorBody;
      http_status: number;
      correlation_id: string;
    };

export interface RequestMetric {
  method: string;
  path: string;
  http_status: number | -1;
  duration_ms: number;
  correlation_id: string;
  outcome: "ok" | "not-modified" | "error" | "network_error";
}

export interface RequestErrorRecord {
  method: string;
  path: string;
  message: string;
  code: string;
  correlation_id: string;
}

export interface RestClientTelemetry {
  requestCompleted(record: RequestMetric): void;
  requestFailed(record: RequestErrorRecord): void;
}

export interface RestClientOptions {
  baseUrl: string;
  fetch: FetchLike;
  tokenProvider?: () => Promise<string | undefined>;
  generateCorrelationId?: () => string;
  now?: () => number;
  telemetry?: RestClientTelemetry;
}

export interface RequestOptions {
  query?: Record<string, string | number | boolean | undefined | null>;
  body?: unknown;
  idempotencyKey?: string;
  ifNoneMatch?: string;
  correlationId?: string;
  headers?: Record<string, string>;
}

export class RestClient {
  private readonly baseUrl: string;
  private readonly fetchFn: FetchLike;
  private readonly tokenProvider?: () => Promise<string | undefined>;
  private readonly genCorrId: () => string;
  private readonly now: () => number;
  private readonly telemetry?: RestClientTelemetry;
  private autoCorrSeq = 0;

  constructor(opts: RestClientOptions) {
    this.baseUrl = opts.baseUrl.replace(/\/+$/, "");
    this.fetchFn = opts.fetch;
    this.tokenProvider = opts.tokenProvider;
    this.now = opts.now ?? (() => Date.now());
    this.telemetry = opts.telemetry;
    this.genCorrId =
      opts.generateCorrelationId ?? (() => `corr-${++this.autoCorrSeq}`);
  }

  get<T>(path: string, options: RequestOptions = {}): Promise<ApiResult<T>> {
    return this.request<T>("GET", path, options);
  }

  post<T>(path: string, options: RequestOptions = {}): Promise<ApiResult<T>> {
    return this.request<T>("POST", path, options);
  }

  delete<T>(
    path: string,
    options: RequestOptions = {},
  ): Promise<ApiResult<T>> {
    return this.request<T>("DELETE", path, options);
  }

  private async request<T>(
    method: string,
    path: string,
    options: RequestOptions,
  ): Promise<ApiResult<T>> {
    const correlationId = options.correlationId ?? this.genCorrId();
    const url = this.buildUrl(path, options.query);
    const headers: Record<string, string> = {
      Accept: "application/json",
      "X-Correlation-Id": correlationId,
      ...(options.headers ?? {}),
    };

    const token = await this.resolveToken();
    if (token) headers.Authorization = `Bearer ${token}`;
    if (options.idempotencyKey) {
      headers["Idempotency-Key"] = options.idempotencyKey;
    }
    if (options.ifNoneMatch) {
      headers["If-None-Match"] = options.ifNoneMatch;
    }

    let bodyText: string | undefined;
    if (options.body !== undefined) {
      bodyText = JSON.stringify(options.body);
      headers["Content-Type"] = "application/json";
    }

    const startedAt = this.now();

    let response: FetchResponseLike;
    try {
      response = await this.fetchFn(url, {
        method,
        headers,
        body: bodyText,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      const duration_ms = this.now() - startedAt;
      const result: ApiResult<T> = {
        status: "error",
        error: {
          code: "NETWORK_ERROR",
          message,
          retriable: true,
          correlation_id: correlationId,
        },
        http_status: -1,
        correlation_id: correlationId,
      };
      this.telemetry?.requestCompleted({
        method,
        path,
        http_status: -1,
        duration_ms,
        correlation_id: correlationId,
        outcome: "network_error",
      });
      this.telemetry?.requestFailed({
        method,
        path,
        message,
        code: "NETWORK_ERROR",
        correlation_id: correlationId,
      });
      return result;
    }

    const duration_ms = this.now() - startedAt;
    const etag = response.headers.get("ETag") ?? undefined;
    const httpStatus = response.status;

    if (httpStatus === 304) {
      this.telemetry?.requestCompleted({
        method,
        path,
        http_status: 304,
        duration_ms,
        correlation_id: correlationId,
        outcome: "not-modified",
      });
      return {
        status: "not-modified",
        etag,
        correlation_id: correlationId,
        http_status: 304,
      };
    }

    const rawText = await response.text();
    const parsed = rawText ? safeJsonParse(rawText) : undefined;

    if (httpStatus >= 200 && httpStatus < 300) {
      this.telemetry?.requestCompleted({
        method,
        path,
        http_status: httpStatus,
        duration_ms,
        correlation_id: correlationId,
        outcome: "ok",
      });
      return {
        status: "ok",
        body: parsed as T,
        etag,
        correlation_id: correlationId,
        http_status: httpStatus,
      };
    }

    // Non-success — normalise to ApiError shape.
    const errorBody = extractApiError(parsed, httpStatus, correlationId);
    this.telemetry?.requestCompleted({
      method,
      path,
      http_status: httpStatus,
      duration_ms,
      correlation_id: correlationId,
      outcome: "error",
    });
    this.telemetry?.requestFailed({
      method,
      path,
      message: errorBody.message,
      code: errorBody.code,
      correlation_id: correlationId,
    });
    return {
      status: "error",
      error: errorBody,
      http_status: httpStatus,
      correlation_id: correlationId,
    };
  }

  private async resolveToken(): Promise<string | undefined> {
    if (!this.tokenProvider) return undefined;
    return this.tokenProvider();
  }

  private buildUrl(
    path: string,
    query?: RequestOptions["query"],
  ): string {
    const prefix = path.startsWith("/") ? path : `/${path}`;
    let url = this.baseUrl + prefix;
    if (!query) return url;
    const params: string[] = [];
    for (const [k, v] of Object.entries(query)) {
      if (v === undefined || v === null) continue;
      params.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`);
    }
    if (params.length > 0) url += `?${params.join("&")}`;
    return url;
  }
}

function safeJsonParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

function extractApiError(
  parsed: unknown,
  httpStatus: number,
  correlationId: string,
): ApiErrorBody {
  if (parsed && typeof parsed === "object" && parsed !== null) {
    const maybe = parsed as { error?: Partial<ApiErrorBody> };
    if (maybe.error && typeof maybe.error.code === "string") {
      return {
        code: maybe.error.code,
        message: maybe.error.message ?? "",
        retriable: maybe.error.retriable,
        correlation_id: maybe.error.correlation_id ?? correlationId,
      };
    }
  }
  return {
    code: `HTTP_${httpStatus}`,
    message: `Request failed with status ${httpStatus}`,
    retriable: httpStatus >= 500,
    correlation_id: correlationId,
  };
}
