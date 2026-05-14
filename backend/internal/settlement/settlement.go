// Package settlement processes bet_settlement, bet_cancel and rollback
// deliveries. It maintains the settlements / cancels / rollbacks tables
// and propagates the resulting market.status transitions.
//
// Maps to acceptance #7 (结算), #8 (取消), #9 (回滚) and the idempotency
// requirements of #11 at the settlement/cancel boundary.
package settlement
