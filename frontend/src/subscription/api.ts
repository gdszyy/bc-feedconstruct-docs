import type { ApiResult, RestClient } from "@/api/client";
import type {
  CreateSubscriptionRequest,
  CreateSubscriptionResponse,
} from "@/contract/rest";

// ---------------------------------------------------------------------------
// M11 — SubscriptionApi
// Server-side subscription side effects only; FavoritesStore is local-only
// and stays out of this adapter.
// ---------------------------------------------------------------------------

export class SubscriptionApi {
  constructor(private readonly client: RestClient) {}

  createSubscription(
    req: CreateSubscriptionRequest,
    correlationId?: string,
  ): Promise<ApiResult<CreateSubscriptionResponse>> {
    return this.client.post<CreateSubscriptionResponse>(
      "/api/v1/subscriptions",
      { body: req, correlationId },
    );
  }

  deleteSubscription(
    subscriptionId: string,
    correlationId?: string,
  ): Promise<ApiResult<void>> {
    const path = `/api/v1/subscriptions/${encodeURIComponent(subscriptionId)}`;
    return this.client.delete<void>(path, { correlationId });
  }
}
