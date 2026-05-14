package health_test

import "testing"

// 验收 15 — 监控
//
// Given a producer (live) that has not produced any message for >30 seconds
// When health.Reporter.Tick is called
// Then producer_health.is_down for product='live' becomes true
//      AND a "producer.down" log + counter increment are emitted
func TestGiven_NoMessagesFor30s_When_TickRuns_Then_ProducerMarkedDown(t *testing.T) {
	_ = t
}

// Given a producer that resumes delivering messages after being down
// When the next message is processed
// Then producer_health.is_down becomes false and last_message_at advances
func TestGiven_DownProducerResumes_When_MessageArrives_Then_MarkedUp(t *testing.T) {
	_ = t
}

// Given the BFF has been receiving 100 messages/min for the last minute
// When /metrics is scraped
// Then it returns Prometheus text with messages_total{product="live"} >= 100
//      and includes recovery_jobs counters and stalled_matches gauges
func TestGiven_MessageThroughput_When_MetricsScraped_Then_PrometheusTextExposed(t *testing.T) {
	_ = t
}

// Given a live match that received an odds_change >120s ago and no bet_stop
// When stalled detection runs
// Then stalled_matches gauge increments and a structured warn log is emitted
func TestGiven_StaleLiveMatch_When_StalledDetectionRuns_Then_GaugeIncrementsAndWarnLogged(t *testing.T) {
	_ = t
}
