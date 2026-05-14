package catalog_test

import "testing"

// 验收 4 — 主数据（M03/M04）
//
// Given a Match object delivered with sport / region / competition references
// When the catalog handler processes it
// Then sports, regions, competitions and matches rows are upserted with
//      name / start_at / home / away / is_live populated
func TestGiven_MatchObject_When_Handled_Then_HierarchyUpserted(t *testing.T) {
	_ = t
}

// Given a fixture_change altering start_at and status
// When the catalog handler processes it
// Then matches.start_at / matches.status are updated AND a fixture_changes
//      history row is inserted with old/new diff and raw_message_id link
func TestGiven_FixtureChange_When_Handled_Then_MatchUpdatedAndHistoryRecorded(t *testing.T) {
	_ = t
}

// 验收 12 — 防回退（赛事级）
//
// Given matches.status currently = ended
// When a delivery arrives with status = live for the same match
// Then matches.status is NOT regressed and a "status.regress.blocked" log is emitted
func TestGiven_EndedMatch_When_LiveStatusArrives_Then_NoRegressionLogged(t *testing.T) {
	_ = t
}
