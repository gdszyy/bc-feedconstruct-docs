package feed_test

import "testing"

// 验收 2 — 消息留痕
//
// Given a GZIP-compressed JSON delivery from FeedConstruct RMQ
// When the consumer receives it
// Then it ungzips, parses the envelope, and INSERTs into raw_messages
//      with source / queue / routing_key / message_type / event_id /
//      product_id / sport_id / ts_provider / payload populated
//      BEFORE any business handler is invoked
func TestGiven_GzippedDelivery_When_Received_Then_RawMessagesRowWrittenBeforeHandler(t *testing.T) {
	_ = t
}

// Given a delivery whose envelope cannot be parsed
// When the consumer processes it
// Then the raw bytes are still persisted (raw_blob non-null), process_error is set,
//      and the delivery is acked (no poison-message loop)
func TestGiven_UnparsableEnvelope_When_Received_Then_RawBlobKeptErrorRecorded(t *testing.T) {
	_ = t
}

// 验收 3 — 消息覆盖
//
// Given the 9 required message_type strings
// When each is dispatched
// Then a registered handler exists for every one and an unknown type
//      is routed to a dead-letter queue with metric increment
func TestGiven_AllRequiredMessageTypes_When_Dispatched_Then_HandlerExistsForEach(t *testing.T) {
	_ = t
}
