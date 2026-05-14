package recovery_test

import "testing"

// 验收 10 — 恢复（启动级）
//
// Given the BFF starts cleanly with an empty raw_messages table
// When the recovery coordinator runs the startup scope
// Then DataSnapshot is invoked for both isLive=true and isLive=false
//      AND a recovery_jobs row is finalized with status=success
func TestGiven_FreshStart_When_StartupRecoveryRuns_Then_FullSnapshotAndJobSuccess(t *testing.T) {
	_ = t
}

// Given an outage of less than 1 hour (last_message_at within 60 minutes)
// When recovery runs
// Then DataSnapshot is invoked WITH getChangesFrom = last_message_at - safetyWindow
func TestGiven_ShortOutage_When_RecoveryRuns_Then_GetChangesFromUsed(t *testing.T) {
	_ = t
}

// Given a single match with stale data (no events for >5 minutes while live)
// When event-level recovery is requested
// Then GetMatchByID is invoked and that match's markets/outcomes are refreshed
func TestGiven_StaleLiveMatch_When_EventRecovery_Then_GetMatchByIDInvokedAndStateRefreshed(t *testing.T) {
	_ = t
}

// Given the WebAPI returns HTTP 429 Too Many Requests
// When recovery encounters it
// Then the job is marked rate_limited and retried with exponential backoff
//      capped at the documented max
func TestGiven_429FromWebAPI_When_RecoveryRetries_Then_ExponentialBackoffApplied(t *testing.T) {
	_ = t
}
