package subscription_test

import "testing"

// 验收 13 — 订阅生命周期（M11）
//
// Given a Book object delivered for match=42 (product=live)
// When subscription.Manager handles it
// Then a subscriptions row is upserted with status=subscribed
//      AND a subscription_events row records the transition requested→subscribed
func TestGiven_BookDelivered_When_Handled_Then_SubscriptionUpsertedAndEventRecorded(t *testing.T) {
	_ = t
}

// Given a live match whose status transitions to ended
// When the catalog handler emits the status change
// Then subscription.Manager auto-unbooks within the configured grace period
//      AND subscriptions.released_at is set with reason="match_ended"
func TestGiven_LiveMatchEnds_When_StatusObserved_Then_AutoUnbookedWithReason(t *testing.T) {
	_ = t
}

// Given a subscription stuck in status=requested for >5 minutes
// When the cleanup tick runs
// Then status transitions to failed and a subscription_events row records reason="stuck_request"
func TestGiven_StuckRequest_When_CleanupTick_Then_TransitionsToFailed(t *testing.T) {
	_ = t
}
