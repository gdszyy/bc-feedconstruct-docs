package config_test

import "testing"

// Given FEED_MODE=live and all FC_* env vars present
// When config.Load() is called
// Then it returns a Config and reports no missing variables
func TestGiven_LiveModeAndAllFCEnv_When_Load_Then_NoError(t *testing.T) {
	// BDD placeholder — formal test added after user confirmation
	_ = t
}

// Given FEED_MODE=live and FC_API_PASS missing
// When config.Load() is called
// Then it fails fast with an error naming every missing variable
func TestGiven_LiveModeMissingFCPass_When_Load_Then_FailsFastNamingMissing(t *testing.T) {
	_ = t
}

// Given FEED_MODE=replay and no FC_* vars
// When config.Load() is called
// Then it returns a Config without requiring FC_* (replay mode)
func TestGiven_ReplayMode_When_Load_Then_NoFCRequired(t *testing.T) {
	_ = t
}

// Given DATABASE_URL absent
// When config.Load() is called
// Then it fails fast regardless of mode
func TestGiven_NoDatabaseURL_When_Load_Then_FailsFast(t *testing.T) {
	_ = t
}
