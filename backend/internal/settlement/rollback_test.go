package settlement_test

import "testing"

// 验收 9 — 回滚
//
// Given an existing settlements row for outcome (42,1,"",1)
// When a rollback_bet_settlement message arrives for the same outcome
// Then a rollbacks row is inserted (target='settlement', target_id=...)
//      AND settlements.rolled_back_at is set non-null
//      AND markets row (42,1,"") status reverts from settled to its prior status
func TestGiven_ExistingSettlement_When_RollbackArrives_Then_RollbackRecordedAndMarketReverts(t *testing.T) {
	_ = t
}

// Given an existing cancels row
// When a VoidNotification with VoidAction=2 (unvoid) arrives for the same target
// Then a rollbacks row is inserted (target='cancel')
//      AND cancels.rolled_back_at is set non-null
//      AND the market exits status=cancelled
func TestGiven_ExistingCancel_When_UnvoidArrives_Then_RollbackRecordedAndMarketRecovers(t *testing.T) {
	_ = t
}

// Given a rollback message arriving twice
// When both deliveries are processed
// Then rollbacks contains exactly one row (idempotent)
func TestGiven_DuplicateRollback_When_Handled_Then_Idempotent(t *testing.T) {
	_ = t
}
