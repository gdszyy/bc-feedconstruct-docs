import type { ApiResult, RestClient } from "@/api/client";
import type {
  GetMarketDescriptionsResponse,
  GetOutcomeDescriptionsResponse,
} from "@/contract/rest";

// ---------------------------------------------------------------------------
// M12 — DescriptionsApi
//
// Caller-managed ETag: pass the last-known etag as `ifNoneMatch`; the
// adapter forwards it via If-None-Match. On 304 the result is
// `status: 'not-modified'` and the store should keep its current data.
// ---------------------------------------------------------------------------

export interface FetchDescriptionsArgs {
  lang?: string;
  ifNoneMatch?: string;
  correlationId?: string;
}

export class DescriptionsApi {
  constructor(private readonly client: RestClient) {}

  fetchMarketDescriptions(
    args: FetchDescriptionsArgs = {},
  ): Promise<ApiResult<GetMarketDescriptionsResponse>> {
    return this.client.get<GetMarketDescriptionsResponse>(
      "/api/v1/descriptions/markets",
      {
        query: args.lang ? { lang: args.lang } : undefined,
        ifNoneMatch: args.ifNoneMatch,
        correlationId: args.correlationId,
      },
    );
  }

  fetchOutcomeDescriptions(
    args: FetchDescriptionsArgs = {},
  ): Promise<ApiResult<GetOutcomeDescriptionsResponse>> {
    return this.client.get<GetOutcomeDescriptionsResponse>(
      "/api/v1/descriptions/outcomes",
      {
        query: args.lang ? { lang: args.lang } : undefined,
        ifNoneMatch: args.ifNoneMatch,
        correlationId: args.correlationId,
      },
    );
  }
}
