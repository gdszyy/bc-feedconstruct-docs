import type { ApiResult, RestClient } from "@/api/client";
import type {
  PlaceBetRequest,
  PlaceBetResponse,
  ValidateBetSlipRequest,
  ValidateBetSlipResponse,
} from "@/contract/rest";

// M13 — Thin functions over RestClient for /api/v1/bet-slip/{validate,place}.

export function validateBetSlip(
  client: RestClient,
  req: ValidateBetSlipRequest,
  correlationId?: string,
): Promise<ApiResult<ValidateBetSlipResponse>> {
  return client.post<ValidateBetSlipResponse>("/api/v1/bet-slip/validate", {
    body: req,
    correlationId,
  });
}

export function placeBet(
  client: RestClient,
  req: PlaceBetRequest,
  idempotencyKey: string,
  correlationId?: string,
): Promise<ApiResult<PlaceBetResponse>> {
  return client.post<PlaceBetResponse>("/api/v1/bet-slip/place", {
    body: req,
    idempotencyKey,
    correlationId,
  });
}
