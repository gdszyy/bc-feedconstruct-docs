import type { ApiResult, RestClient } from "@/api/client";
import type { GetSystemHealthResponse } from "@/contract/rest";

// ---------------------------------------------------------------------------
// M15 — HealthApi
// Hydrates HealthStore at startup; subsequent updates flow via
// `system.producer_status` over WS.
// ---------------------------------------------------------------------------

export class HealthApi {
  constructor(private readonly client: RestClient) {}

  fetchSystemHealth(
    correlationId?: string,
  ): Promise<ApiResult<GetSystemHealthResponse>> {
    return this.client.get<GetSystemHealthResponse>("/api/v1/system/health", {
      correlationId,
    });
  }
}
