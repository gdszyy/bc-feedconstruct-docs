// REST contract between Go BFF and Next.js frontend.
// Source of truth: docs/07_frontend_architecture/03_backend_data_contract.md §5–§7.
//
// FROZEN after Wave 0. Add new endpoints in sibling files
// (e.g. rest.<scope>.ts) and re-export via index.ts.

import type {
  Certainty,
  MarketStatus,
  MatchStatus,
} from "./events";

// ---------------------------------------------------------------------------
// Shared error model — §7
// ---------------------------------------------------------------------------

export interface ApiError {
  error: {
    code: string;
    message: string;
    retriable?: boolean;
    correlation_id?: string;
  };
}

// ---------------------------------------------------------------------------
// Catalog — M03
// ---------------------------------------------------------------------------

export interface CatalogSport {
  sport_id: string;
  name: string;
  sort_order: number;
}
export interface GetCatalogSportsResponse {
  sports: CatalogSport[];
}

export interface CatalogTournament {
  tournament_id: string;
  sport_id: string;
  category_id: string;
  name: string;
}
export interface GetCatalogTournamentsQuery {
  sport_id: string;
}
export interface GetCatalogTournamentsResponse {
  tournaments: CatalogTournament[];
}

// ---------------------------------------------------------------------------
// Matches — M04 / M05
// ---------------------------------------------------------------------------

export type MatchListFilter = "prematch" | "live" | "today" | "upcoming";

export interface GetMatchesQuery {
  filter?: MatchListFilter;
  sport_id?: string;
  tournament_id?: string;
  limit?: number;
  cursor?: string;
}
export interface MatchSummary {
  match_id: string;
  tournament_id: string;
  home_team: string;
  away_team: string;
  scheduled_at: string;
  status: MatchStatus;
  is_live: boolean;
  version: number;
}
export interface GetMatchesResponse {
  matches: MatchSummary[];
  next_cursor?: string;
}

export interface MarketSnapshot {
  market_id: string;
  market_type_id: string;
  specifiers?: Record<string, string>;
  status: MarketStatus;
  outcomes: Array<{
    outcome_id: string;
    odds: number;
    active: boolean;
  }>;
  version: number;
}

export interface GetMatchSnapshotResponse {
  match: MatchSummary & {
    period?: string;
    home_score?: number;
    away_score?: number;
  };
  markets: MarketSnapshot[];
  recent_settlement?: {
    market_id: string;
    certainty: Certainty;
    applied_at: string;
  };
  recent_cancel?: {
    market_id?: string;
    void_reason: string;
    applied_at: string;
  };
}

export interface GetMatchMarketsResponse {
  match_id: string;
  markets: MarketSnapshot[];
}

// ---------------------------------------------------------------------------
// Descriptions / i18n — M12
// ---------------------------------------------------------------------------

export interface MarketDescription {
  market_type_id: string;
  name: string;
  group: string;
  tab: string;
}
export interface GetMarketDescriptionsQuery {
  version?: string;
  lang?: string;
}
export interface GetMarketDescriptionsResponse {
  version: string;
  descriptions: MarketDescription[];
}

export interface OutcomeDescription {
  outcome_type_id: string;
  name: string;
}
export interface GetOutcomeDescriptionsQuery {
  version?: string;
  lang?: string;
}
export interface GetOutcomeDescriptionsResponse {
  version: string;
  descriptions: OutcomeDescription[];
}

// ---------------------------------------------------------------------------
// Bet slip — M13
// ---------------------------------------------------------------------------

export interface BetSelection {
  position: number;
  match_id: string;
  market_id: string;
  outcome_id: string;
  locked_odds: number;
}

export interface ValidateBetSlipRequest {
  selections: BetSelection[];
  stake: number;
  currency: string;
  bet_type: "single" | "multiple" | "system";
}
export interface ValidateBetSlipResponse {
  valid: boolean;
  reasons?: Array<{
    code: string;
    selection_position?: number;
    message: string;
  }>;
  current_odds?: Array<{
    position: number;
    odds: number;
    delta: number;
  }>;
}

export interface PlaceBetRequest extends ValidateBetSlipRequest {
  user_id?: string;
}
export interface PlaceBetResponse {
  bet_id: string;
  state: string;
  deduped: boolean;
}

// ---------------------------------------------------------------------------
// My bets — M14
// ---------------------------------------------------------------------------

export interface GetMyBetsQuery {
  user_id?: string;
  status?: string[];
  limit?: number;
}
export interface MyBetTransition {
  at: string;
  from: string;
  to: string;
  reason?: string;
  event_id?: string;
  correlation_id?: string;
}
export interface MyBet {
  id: string;
  user_id: string;
  placed_at: string;
  stake: number;
  currency: string;
  bet_type: string;
  state: string;
  selections: BetSelection[];
  history: MyBetTransition[];
  payout_gross?: number;
  payout_currency?: string;
  void_factor?: number;
  dead_heat_factor?: number;
}
export interface GetMyBetsResponse {
  bets: MyBet[];
  count: number;
}

// ---------------------------------------------------------------------------
// Subscriptions — M11
// ---------------------------------------------------------------------------

export interface CreateSubscriptionRequest {
  match_id: string;
}
export interface CreateSubscriptionResponse {
  subscription_id: string;
  match_id: string;
  state: "active";
}

// ---------------------------------------------------------------------------
// System health — M15
// ---------------------------------------------------------------------------

export interface GetSystemHealthResponse {
  producers: Array<{
    product: "live" | "prematch";
    is_down: boolean;
    last_message_at: string;
    down_since?: string;
  }>;
  degraded: boolean;
}

// ---------------------------------------------------------------------------
// Endpoint registry — keep aligned with §5 table in the doc.
// ---------------------------------------------------------------------------

export const REST_ENDPOINTS = {
  catalogSports:       { method: "GET",    path: "/api/v1/catalog/sports" },
  catalogTournaments:  { method: "GET",    path: "/api/v1/catalog/tournaments" },
  matches:             { method: "GET",    path: "/api/v1/matches" },
  matchSnapshot:       { method: "GET",    path: "/api/v1/matches/{id}" },
  matchMarkets:        { method: "GET",    path: "/api/v1/matches/{id}/markets" },
  marketDescriptions:  { method: "GET",    path: "/api/v1/descriptions/markets" },
  outcomeDescriptions: { method: "GET",    path: "/api/v1/descriptions/outcomes" },
  betSlipValidate:     { method: "POST",   path: "/api/v1/bet-slip/validate" },
  betSlipPlace:        { method: "POST",   path: "/api/v1/bet-slip/place" },
  myBets:              { method: "GET",    path: "/api/v1/my-bets" },
  myBetById:           { method: "GET",    path: "/api/v1/my-bets/{betId}" },
  createSubscription:  { method: "POST",   path: "/api/v1/subscriptions" },
  deleteSubscription:  { method: "DELETE", path: "/api/v1/subscriptions/{id}" },
  systemHealth:        { method: "GET",    path: "/api/v1/system/health" },
} as const;

export type RestEndpointKey = keyof typeof REST_ENDPOINTS;
