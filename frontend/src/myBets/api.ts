import type { ApiResult, RestClient } from "@/api/client";
import type {
  GetMyBetsQuery,
  GetMyBetsResponse,
  MyBet,
} from "@/contract/rest";

// ---------------------------------------------------------------------------
// M14 — MyBetsApi
// Multi-valued `status` is encoded as repeated query params:
//   ?status=Accepted&status=Settled
// ---------------------------------------------------------------------------

export class MyBetsApi {
  constructor(private readonly client: RestClient) {}

  fetchMyBets(
    query: GetMyBetsQuery = {},
    correlationId?: string,
  ): Promise<ApiResult<GetMyBetsResponse>> {
    const queryString = encodeMyBetsQuery(query);
    const path = queryString
      ? `/api/v1/my-bets?${queryString}`
      : "/api/v1/my-bets";
    return this.client.get<GetMyBetsResponse>(path, { correlationId });
  }

  fetchMyBetById(
    betId: string,
    correlationId?: string,
  ): Promise<ApiResult<MyBet>> {
    const path = `/api/v1/my-bets/${encodeURIComponent(betId)}`;
    return this.client.get<MyBet>(path, { correlationId });
  }
}

function encodeMyBetsQuery(q: GetMyBetsQuery): string {
  const params: string[] = [];
  if (q.user_id) params.push(`user_id=${encodeURIComponent(q.user_id)}`);
  if (q.limit !== undefined) params.push(`limit=${q.limit}`);
  if (q.status) {
    for (const s of q.status) params.push(`status=${encodeURIComponent(s)}`);
  }
  return params.join("&");
}
