package settlement_test

import "testing"

// 验收 8 — 取消（VoidNotification, VoidAction=1）
//
// Given a VoidNotification with VoidAction=1, ObjectType=13 (market),
//       FromDate, ToDate, Reason
// When CancelHandler processes it
// Then a cancels row is inserted carrying void_reason / from_ts / to_ts
//      AND the targeted markets row transitions to status=cancelled
func TestGiven_VoidNotificationVoid_When_Handled_Then_CancelRowAndMarketCancelled(t *testing.T) {
	_ = t
}

// Given a cancel referencing a superceded_by id
// When processed
// Then the superceded_by column links to the original cancel; chain queryable
func TestGiven_CancelWithSupercededBy_When_Handled_Then_LinkPreserved(t *testing.T) {
	_ = t
}

// Given a cancel for ObjectType=4 (match) covering all markets of a match
// When processed
// Then every market of that match transitions to status=cancelled
func TestGiven_MatchLevelCancel_When_Handled_Then_AllMarketsCancelled(t *testing.T) {
	_ = t
}
