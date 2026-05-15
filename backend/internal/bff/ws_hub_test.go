package bff_test

import "testing"

// B5 — WebSocket subscription
//
// Given a client connected to /ws/v1/stream and sent {"op":"subscribe","scope":{"match_ids":["42"]}}
// When odds.changed for match_id=42 is processed by the BFF
// Then the client receives the same envelope frame within 200ms
func TestGiven_ClientSubscribed_When_OddsChange_Then_PushedWithin200ms(t *testing.T) {
	_ = t
	// BDD placeholder — Wave 10-E owns the real test (RegisterWebSocketRoutes).
}
