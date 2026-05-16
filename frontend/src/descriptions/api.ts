import type { ApiResult, RestClient } from "@/api/client";
import type {
  GetMarketDescriptionsResponse,
  GetOutcomeDescriptionsResponse,
} from "@/contract/rest";

// M12 — Caller-managed ETag: pass last-known etag as ifNoneMatch; on 304
// the result is status='not-modified' and the store keeps its current data.

export interface FetchDescriptionsArgs {
  lang?: string;
  ifNoneMatch?: string;
  correlationId?: string;
}

export function fetchMarketDescriptions(
  client: RestClient,
  args: FetchDescriptionsArgs = {},
): Promise<ApiResult<GetMarketDescriptionsResponse>> {
  return client.get<GetMarketDescriptionsResponse>(
    "/api/v1/descriptions/markets",
    {
      query: args.lang ? { lang: args.lang } : undefined,
      ifNoneMatch: args.ifNoneMatch,
      correlationId: args.correlationId,
    },
  );
}

export function fetchOutcomeDescriptions(
  client: RestClient,
  args: FetchDescriptionsArgs = {},
): Promise<ApiResult<GetOutcomeDescriptionsResponse>> {
  return client.get<GetOutcomeDescriptionsResponse>(
    "/api/v1/descriptions/outcomes",
    {
      query: args.lang ? { lang: args.lang } : undefined,
      ifNoneMatch: args.ifNoneMatch,
      correlationId: args.correlationId,
    },
  );
}
