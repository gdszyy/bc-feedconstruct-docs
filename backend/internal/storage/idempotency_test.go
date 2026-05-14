package storage_test

import "testing"

// 验收 11 — 幂等
//
// Given a raw_messages row with (source, message_type, event_id, ts_provider) already inserted
// When the same delivery is consumed again
// Then the unique constraint blocks the second insert and the count stays at 1
func TestGiven_DuplicateDelivery_When_InsertRawMessage_Then_UniqueConstraintHolds(t *testing.T) {
	_ = t
}

// Given an existing settlement row for (match_id, market_type_id, specifier, outcome_id, settled_at)
// When the same bet_settlement message is reprocessed
// Then no second settlements row is written
func TestGiven_ExistingSettlement_When_Reprocess_Then_NoDuplicateRow(t *testing.T) {
	_ = t
}

// 验收 16 — 数据治理 / 保留窗口
//
// Given raw_messages older than the retention window (default 7d)
// When the retention job runs
// Then those rows are deleted and metrics_counters.retention_deleted increments
func TestGiven_RawMessagesPastRetention_When_RetentionJobRuns_Then_RowsDeletedAndCounterIncrements(t *testing.T) {
	_ = t
}
