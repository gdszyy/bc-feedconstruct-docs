import type { ApiResult, RestClient } from "@/api/client";
import type { GetSystemHealthResponse } from "@/contract/rest";

// M15 — Startup hydration; subsequent updates flow via system.producer_status WS events.

export function fetchSystemHealth(
  client: RestClient,
  correlationId?: string,
): Promise<ApiResult<GetSystemHealthResponse>> {
  return client.get<GetSystemHealthResponse>("/api/v1/system/health", {
    correlationId,
  });
}
