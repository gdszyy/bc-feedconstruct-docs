// Package feed implements the M01 ingest layer: it connects to (or replays)
// the FeedConstruct source, decompresses GZIP, parses envelopes, persists
// every delivery into raw_messages BEFORE any business handler runs, and
// fans out into the internal RabbitMQ exchange "feed.events".
//
// Module split:
//   - envelope.go : Envelope type + classification rules
//   - decoder.go  : GZIP-aware body decoder
//   - dispatcher.go : handler registry (M02)
//   - publisher.go  : Publisher interface + AMQPPublisher implementation
//   - processor.go  : decode -> store raw_messages -> publish
//   - replayer.go   : FEED_MODE=replay file source
//   - live_consumer.go : FEED_MODE=live FC RMQ source
package feed
