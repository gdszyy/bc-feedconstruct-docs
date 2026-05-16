import type { ApiResult, RestClient } from "@/api/client";
import type { GetMatchSnapshotResponse } from "@/contract/rest";

// M04/M05 — startup hydration for a single match. Subsequent updates flow
// via fixture / odds / status WS events handled by MatchStore + MarketsStore.

export function fetchMatchSnapshot(
  client: RestClient,
  matchId: string,
  correlationId?: string,
): Promise<ApiResult<GetMatchSnapshotResponse>> {
  return client.get<GetMatchSnapshotResponse>(
    `/api/v1/matches/${encodeURIComponent(matchId)}`,
    { correlationId },
  );
}
