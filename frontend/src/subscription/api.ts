import type { ApiResult, RestClient } from "@/api/client";
import type {
  CreateSubscriptionRequest,
  CreateSubscriptionResponse,
} from "@/contract/rest";

// M11 — Server-side subscription side effects. FavoritesStore is local-only
// and stays out of this adapter.

export function createSubscription(
  client: RestClient,
  req: CreateSubscriptionRequest,
  correlationId?: string,
): Promise<ApiResult<CreateSubscriptionResponse>> {
  return client.post<CreateSubscriptionResponse>("/api/v1/subscriptions", {
    body: req,
    correlationId,
  });
}

export function deleteSubscription(
  client: RestClient,
  subscriptionId: string,
  correlationId?: string,
): Promise<ApiResult<void>> {
  const path = `/api/v1/subscriptions/${encodeURIComponent(subscriptionId)}`;
  return client.delete<void>(path, { correlationId });
}
