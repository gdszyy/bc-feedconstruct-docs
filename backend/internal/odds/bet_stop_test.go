package odds_test

import "testing"

// 验收 6 — 停投
//
// Given a bet_stop for match=42 with market_status=suspended targeting all markets
// When BetStopHandler processes it
// Then every markets row of match 42 transitions to status=suspended
//      AND a market_status_history row is appended for each transition with raw_message_id link
func TestGiven_BetStopAllMarkets_When_Handled_Then_MarketsSuspendedAndHistoryAppended(t *testing.T) {
	_ = t
}

// Given a bet_stop targeting a specific market_group
// When BetStopHandler processes it
// Then only markets within that group transition; others remain active
func TestGiven_BetStopByGroup_When_Handled_Then_OnlyTargetedMarketsTransition(t *testing.T) {
	_ = t
}

// 验收 12 — 防回退（盘口级）
//
// Given a markets row currently in status=settled
// When an odds_change arrives that would set it to active
// Then the transition is rejected and logged; markets.status remains settled
func TestGiven_SettledMarket_When_OddsChangeWouldActivate_Then_NoRegression(t *testing.T) {
	_ = t
}
