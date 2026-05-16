import type { RestClient, ApiResult } from "@/api/client";
import type {
  PlaceBetRequest,
  PlaceBetResponse,
  ValidateBetSlipRequest,
  ValidateBetSlipResponse,
} from "@/contract/rest";

// ---------------------------------------------------------------------------
// M13 — BetSlipApi
// Thin adapter over RestClient for /api/v1/bet-slip/{validate,place}.
// ---------------------------------------------------------------------------

export class BetSlipApi {
  constructor(private readonly client: RestClient) {}

  validate(
    req: ValidateBetSlipRequest,
    correlationId?: string,
  ): Promise<ApiResult<ValidateBetSlipResponse>> {
    return this.client.post<ValidateBetSlipResponse>(
      "/api/v1/bet-slip/validate",
      { body: req, correlationId },
    );
  }

  place(
    req: PlaceBetRequest,
    idempotencyKey: string,
    correlationId?: string,
  ): Promise<ApiResult<PlaceBetResponse>> {
    return this.client.post<PlaceBetResponse>("/api/v1/bet-slip/place", {
      body: req,
      idempotencyKey,
      correlationId,
    });
  }
}
