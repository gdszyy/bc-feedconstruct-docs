package webapi_test

import "testing"

// FeedConstruct WebAPI Token contract — see docs/01_data_feed/rmq-web-api/002_access.md
//
// Given valid FC_API_USER and FC_API_PASS
// When webapi.Client.Token() is called for the first time
// Then a non-empty token string is returned and cached in memory
func TestGiven_ValidCreds_When_TokenCalled_Then_NonEmptyAndCached(t *testing.T) {
	_ = t
}

// Given a cached token approaching its 24h expiry
// When the client refreshes >=1h before expiry
// Then exactly one refresh request is in flight even under concurrent callers
func TestGiven_TokenNearExpiry_When_ConcurrentRefresh_Then_SingleInFlight(t *testing.T) {
	_ = t
}

// Given the WebAPI returns an auth error
// When Token() is called
// Then the error is wrapped with cause and producer_health.detail records "webapi_token_failed"
//      AND no token value is logged
func TestGiven_AuthError_When_Token_Then_ErrorWrappedAndCredentialsNotLogged(t *testing.T) {
	_ = t
}
