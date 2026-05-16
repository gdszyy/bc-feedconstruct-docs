// Test-only helpers shared across adapter tests. Not exported from
// public surface (file lives under api/ but is not re-exported via
// any barrel — only test files import it).

import type {
  ApiResult,
  RequestOptions,
  RestClient,
} from "./client";

export interface RecordedCall {
  method: "GET" | "POST" | "DELETE";
  path: string;
  options: RequestOptions;
}

export interface StubRestClientOptions {
  /** Default response for any unmatched call. */
  defaultResponse?: ApiResult<unknown>;
  /** Per-(method,path) response queue, drained in order. */
  responses?: Array<{
    match: { method: "GET" | "POST" | "DELETE"; path?: string };
    response: ApiResult<unknown>;
  }>;
}

export class StubRestClient {
  public readonly calls: RecordedCall[] = [];
  private readonly defaultResponse: ApiResult<unknown>;
  private readonly queue: NonNullable<StubRestClientOptions["responses"]>;

  constructor(opts: StubRestClientOptions = {}) {
    this.defaultResponse =
      opts.defaultResponse ??
      ({
        status: "ok",
        body: undefined,
        correlation_id: "stub-corr",
        http_status: 200,
      } as ApiResult<unknown>);
    this.queue = opts.responses ? [...opts.responses] : [];
  }

  asClient(): RestClient {
    return this as unknown as RestClient;
  }

  get<T>(path: string, options: RequestOptions = {}): Promise<ApiResult<T>> {
    return this.respond("GET", path, options) as Promise<ApiResult<T>>;
  }
  post<T>(path: string, options: RequestOptions = {}): Promise<ApiResult<T>> {
    return this.respond("POST", path, options) as Promise<ApiResult<T>>;
  }
  delete<T>(
    path: string,
    options: RequestOptions = {},
  ): Promise<ApiResult<T>> {
    return this.respond("DELETE", path, options) as Promise<ApiResult<T>>;
  }

  private respond(
    method: "GET" | "POST" | "DELETE",
    path: string,
    options: RequestOptions,
  ): Promise<ApiResult<unknown>> {
    this.calls.push({ method, path, options });
    const idx = this.queue.findIndex(
      (q) =>
        q.match.method === method &&
        (q.match.path === undefined || q.match.path === path),
    );
    if (idx >= 0) {
      const { response } = this.queue.splice(idx, 1)[0]!;
      return Promise.resolve(response);
    }
    return Promise.resolve(this.defaultResponse);
  }
}
