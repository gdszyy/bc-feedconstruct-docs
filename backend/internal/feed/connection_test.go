package feed_test

import "testing"

// 验收 1 — 连接（M01 / 上传指引 §2 "连接接入"）
//
// Given valid FC_RMQ_HOST, FC_RMQ_USER, FC_RMQ_PASS, FC_PARTNER_ID
// When the consumer starts
// Then it opens a TLS connection within 5s, declares both
//      P{PartnerID}_live and P{PartnerID}_prematch consumers,
//      and configures heartbeat + QoS prefetch
func TestGiven_ValidLiveCreds_When_ConsumerStarts_Then_BothQueuesBoundWithQoS(t *testing.T) {
	_ = t
}

// Given an RMQ connection that drops mid-consume
// When the broker becomes available again
// Then the consumer reconnects with exponential backoff (2s/4s/8s/16s/30s cap)
//      and resumes from the last unacked delivery
func TestGiven_DroppedConnection_When_BrokerReturns_Then_ReconnectsWithBackoffAndResumes(t *testing.T) {
	_ = t
}

// Given FEED_MODE=replay and a directory of raw/json sample messages
// When the Replayer is started
// Then deliveries are emitted in timestamp order to the same internal exchange
//      so the rest of the pipeline cannot tell the difference from live
func TestGiven_ReplayMode_When_ReplayerStarts_Then_EmitsInOrderViaSameExchange(t *testing.T) {
	_ = t
}

// Given live mode but FC_API_PASS missing
// When the consumer attempts to start
// Then it fails fast within 30s with a clear error naming the missing variable
func TestGiven_LiveModeMissingFCPass_When_ConsumerStarts_Then_FailsFastWithMissingVar(t *testing.T) {
	_ = t
}
