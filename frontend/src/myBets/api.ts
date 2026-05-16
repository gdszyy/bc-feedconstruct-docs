import type { ApiResult, RestClient } from "@/api/client";
import type {
  GetMyBetsQuery,
  GetMyBetsResponse,
  MyBet,
} from "@/contract/rest";

// M14 — Multi-valued `status` is encoded as repeated query params:
//   ?status=Accepted&status=Settled
// This relies on RestClient.buildUrl accepting array values for query keys.

export function fetchMyBets(
  client: RestClient,
  query: GetMyBetsQuery = {},
  correlationId?: string,
): Promise<ApiResult<GetMyBetsResponse>> {
  return client.get<GetMyBetsResponse>("/api/v1/my-bets", {
    query: {
      user_id: query.user_id,
      limit: query.limit,
      status: query.status,
    },
    correlationId,
  });
}

export function fetchMyBetById(
  client: RestClient,
  betId: string,
  correlationId?: string,
): Promise<ApiResult<MyBet>> {
  const path = `/api/v1/my-bets/${encodeURIComponent(betId)}`;
  return client.get<MyBet>(path, { correlationId });
}
