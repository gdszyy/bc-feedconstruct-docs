package bff_test

import "testing"

// S2 — WebSocket origin enforcement
//
// Given WS_ALLOWED_ORIGINS = "https://web.up.railway.app"
// When a client connects with Origin = "https://evil.example"
// Then the upgrade is rejected with HTTP 403
func TestGiven_DisallowedOrigin_When_Upgrade_Then_Rejected403(t *testing.T) {
	_ = t
	// BDD placeholder — Wave 10-F owns the real test (OriginGuard).
}

// S3 — REST rate limit
//
// Given the default rate limit of 60 req/min/ip
// When a single IP issues 61 requests inside 60 seconds
// Then request 61 returns 429 with Retry-After header
func TestGiven_RateLimit60_When_61stRequest_Then_Returns429(t *testing.T) {
	_ = t
	// BDD placeholder — Wave 10-F owns the real test (RateLimit).
}
