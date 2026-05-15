// WebSocket event contract between Go BFF and Next.js frontend.
// Source of truth: docs/07_frontend_architecture/03_backend_data_contract.md §1–§4.
//
// This file is FROZEN after Wave 0. New event types must be added in
// a sibling file (e.g. events.<scope>.ts) and re-exported via index.ts
// to keep parallel tracks merge-conflict-free.

export type ProductID = "live" | "prematch";

/** Common envelope wrapping every WS event and async REST notification. */
export interface Envelope<TPayload = unknown> {
  /** Dot-delimited event type, see EventType union below. */
  type: EventType;
  /** Envelope schema version (currently "1"). */
  schema_version: string;
  /** Server-generated ULID; used as idempotency key on the client. */
  event_id: string;
  /** Trace id propagated to telemetry (M16). */
  correlation_id: string;
  /** Source product. */
  product_id: ProductID;
  /** Business timestamp (ISO-8601). */
  occurred_at: string;
  /** Server ingest timestamp (ISO-8601). */
  received_at: string;
  /** Routing key — main entity ids. All optional; presence varies by type. */
  entity: EntityRef;
  /** Type-specific payload. */
  payload: TPayload;
}

export interface EntityRef {
  sport_id?: string;
  tournament_id?: string;
  match_id?: string;
  market_id?: string;
  outcome_id?: string;
}

// ---------------------------------------------------------------------------
// 3.1 — system lifecycle (M01 / M10 / M15)
// ---------------------------------------------------------------------------

export interface SystemHelloPayload {
  session_id: string;
  heartbeat_interval_ms: number;
}
export interface SystemHeartbeatPayload {
  server_time: string;
}
export interface SystemReplayStartedPayload {
  from_cursor: string;
}
export interface SystemReplayCompletedPayload {
  to_cursor: string;
}
export interface SystemProducerStatusPayload {
  product: ProductID;
  is_down: boolean;
  last_message_at: string;
  down_since?: string;
}

// ---------------------------------------------------------------------------
// 3.2 — master data (M03 / M04)
// ---------------------------------------------------------------------------

export interface SportUpsertedPayload {
  sport_id: string;
  name_translations: Record<string, string>;
  sort_order: number;
}
export interface SportRemovedPayload {
  sport_id: string;
}
export interface TournamentUpsertedPayload {
  tournament_id: string;
  sport_id: string;
  category_id: string;
  name_translations: Record<string, string>;
}
export interface TournamentRemovedPayload {
  tournament_id: string;
}
export interface MatchUpsertedPayload {
  match_id: string;
  tournament_id: string;
  home_team: string;
  away_team: string;
  scheduled_at: string;
  is_live: boolean;
  version: number;
}
export interface MatchStatusChangedPayload {
  match_id: string;
  status: MatchStatus;
  period?: string;
  home_score?: number;
  away_score?: number;
  version: number;
}
export type MatchStatus =
  | "not_started"
  | "live"
  | "suspended"
  | "ended"
  | "closed"
  | "cancelled"
  | "abandoned";

export interface FixtureChangedPayload {
  match_id: string;
  change_type: "schedule" | "teams" | "format" | "other";
  /** Frontend MUST refetch /api/v1/matches/{id} after this event. */
  refetch_required: true;
}

// ---------------------------------------------------------------------------
// 3.3 — odds & markets (M05 / M06 / M07)
// ---------------------------------------------------------------------------

export interface OddsChangedPayload {
  match_id: string;
  market_id: string;
  specifiers?: Record<string, string>;
  outcomes: Array<{
    outcome_id: string;
    odds: number;
    active: boolean;
  }>;
  version: number;
}
export interface MarketStatusChangedPayload {
  match_id: string;
  market_id: string;
  status: MarketStatus;
  version: number;
}
export type MarketStatus =
  | "active"
  | "suspended"
  | "deactivated"
  | "handed_over"
  | "settled"
  | "cancelled";

export interface BetStopAppliedPayload {
  match_id: string;
  /** Empty array means full-match stop; otherwise scoped to market groups. */
  market_groups: string[];
}
export interface BetStopLiftedPayload {
  match_id: string;
  market_groups: string[];
}

// ---------------------------------------------------------------------------
// 3.4 — settlement / cancel / rollback (M08 / M09)
// ---------------------------------------------------------------------------

export type Certainty = "certain" | "settled_after_confirmation";

export interface BetSettlementAppliedPayload {
  match_id: string;
  market_id: string;
  outcomes: Array<{
    outcome_id: string;
    result: "win" | "lose" | "void" | "half_win" | "half_lose";
    void_factor?: number;
    dead_heat_factor?: number;
  }>;
  certainty: Certainty;
  version: number;
}
export interface BetSettlementRolledBackPayload {
  match_id: string;
  market_id: string;
  version: number;
}
export interface BetCancelAppliedPayload {
  match_id: string;
  market_id?: string;
  void_reason: string;
  start_time?: string;
  end_time?: string;
  superceded_by?: string;
}
export interface BetCancelRolledBackPayload {
  match_id: string;
  market_id?: string;
}

// ---------------------------------------------------------------------------
// 3.5 — subscription (M11)
// ---------------------------------------------------------------------------

export interface SubscriptionChangedPayload {
  subscription_id: string;
  match_id: string;
  state: "active" | "released" | "cancelled";
}

// ---------------------------------------------------------------------------
// 3.6 — bet flow (M13 / M14)
// ---------------------------------------------------------------------------

export interface BetAcceptedPayload {
  bet_id: string;
  user_id: string;
  accepted_odds: number;
  accepted_at: string;
}
export interface BetRejectedPayload {
  bet_id: string;
  user_id: string;
  code: string;
  message: string;
}
export interface BetStateChangedPayload {
  bet_id: string;
  from: string;
  to: string;
  at: string;
  reason?: string;
}

// ---------------------------------------------------------------------------
// Discriminated union — exhaustive event registry.
// ---------------------------------------------------------------------------

export type EventType =
  | "system.hello"
  | "system.heartbeat"
  | "system.replay_started"
  | "system.replay_completed"
  | "system.producer_status"
  | "sport.upserted"
  | "sport.removed"
  | "tournament.upserted"
  | "tournament.removed"
  | "match.upserted"
  | "match.status_changed"
  | "fixture.changed"
  | "odds.changed"
  | "market.status_changed"
  | "bet_stop.applied"
  | "bet_stop.lifted"
  | "bet_settlement.applied"
  | "bet_settlement.rolled_back"
  | "bet_cancel.applied"
  | "bet_cancel.rolled_back"
  | "subscription.changed"
  | "bet.accepted"
  | "bet.rejected"
  | "bet.state_changed";

export type EventPayloadMap = {
  "system.hello": SystemHelloPayload;
  "system.heartbeat": SystemHeartbeatPayload;
  "system.replay_started": SystemReplayStartedPayload;
  "system.replay_completed": SystemReplayCompletedPayload;
  "system.producer_status": SystemProducerStatusPayload;
  "sport.upserted": SportUpsertedPayload;
  "sport.removed": SportRemovedPayload;
  "tournament.upserted": TournamentUpsertedPayload;
  "tournament.removed": TournamentRemovedPayload;
  "match.upserted": MatchUpsertedPayload;
  "match.status_changed": MatchStatusChangedPayload;
  "fixture.changed": FixtureChangedPayload;
  "odds.changed": OddsChangedPayload;
  "market.status_changed": MarketStatusChangedPayload;
  "bet_stop.applied": BetStopAppliedPayload;
  "bet_stop.lifted": BetStopLiftedPayload;
  "bet_settlement.applied": BetSettlementAppliedPayload;
  "bet_settlement.rolled_back": BetSettlementRolledBackPayload;
  "bet_cancel.applied": BetCancelAppliedPayload;
  "bet_cancel.rolled_back": BetCancelRolledBackPayload;
  "subscription.changed": SubscriptionChangedPayload;
  "bet.accepted": BetAcceptedPayload;
  "bet.rejected": BetRejectedPayload;
  "bet.state_changed": BetStateChangedPayload;
};

export type TypedEnvelope = {
  [K in EventType]: Envelope<EventPayloadMap[K]> & { type: K };
}[EventType];

// ---------------------------------------------------------------------------
// 4 — client → server control frames
// ---------------------------------------------------------------------------

export type ControlFrame =
  | { op: "subscribe"; scope: SubscribeScope }
  | { op: "unsubscribe"; scope: SubscribeScope }
  | { op: "replay_from"; cursor: string; session_id: string };

export interface SubscribeScope {
  match_ids?: string[];
  sport_ids?: string[];
  tournament_ids?: string[];
}
