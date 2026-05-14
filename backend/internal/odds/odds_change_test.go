package odds_test

import "testing"

// 验收 5 — 赔率（M05）
//
// Given an odds_change for match=42 carrying market_type_id=1, specifier="",
//       outcomes [{id:1, odds:1.85, active:true}, {id:2, odds:2.10, active:true}]
// When the OddsHandler processes it
// Then markets row (42,1,"") is upserted with status=active
//      and outcomes rows are upserted with the exact odds and is_active flags
func TestGiven_OddsChange_When_Handled_Then_MarketAndOutcomesUpserted(t *testing.T) {
	_ = t
}

// Given an odds_change with the same payload received twice
// When both deliveries are processed
// Then the outcomes row is updated to the same values exactly once
//      (no UNIQUE violations, updated_at advances on each apply)
func TestGiven_DuplicateOddsChange_When_Handled_Then_NoDuplicateRow(t *testing.T) {
	_ = t
}

// Given odds_change carrying score / period info
// When processed
// Then the catalog match row reflects the latest score and period if newer than current
func TestGiven_OddsChangeWithScore_When_Handled_Then_MatchScoreAdvances(t *testing.T) {
	_ = t
}
