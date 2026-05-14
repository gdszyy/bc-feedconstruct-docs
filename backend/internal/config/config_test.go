package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/config"
)

// helper: sets the full env required by Load and clears the rest.
func setLiveEnv(t *testing.T) {
	t.Helper()
	t.Setenv("FEED_MODE", "live")
	t.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db")
	t.Setenv("RABBITMQ_URL", "amqp://guest:guest@h:5672/")
	t.Setenv("FC_API_BASE", "https://api.test")
	t.Setenv("FC_API_USER", "user")
	t.Setenv("FC_API_PASS", "pass")
	t.Setenv("FC_RMQ_HOST", "rmq.test:5673")
	t.Setenv("FC_RMQ_USER", "ru")
	t.Setenv("FC_RMQ_PASS", "rp")
	t.Setenv("FC_PARTNER_ID", "123")
}

func setReplayEnv(t *testing.T) {
	t.Helper()
	t.Setenv("FEED_MODE", "replay")
	t.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db")
	t.Setenv("RABBITMQ_URL", "amqp://guest:guest@h:5672/")
	for _, k := range []string{
		"FC_API_BASE", "FC_API_USER", "FC_API_PASS",
		"FC_RMQ_HOST", "FC_RMQ_USER", "FC_RMQ_PASS", "FC_PARTNER_ID",
	} {
		t.Setenv(k, "")
	}
}

// Given FEED_MODE=live and all FC_* env vars present
// When config.Load() is called
// Then it returns a Config and reports no missing variables
func TestGiven_LiveModeAndAllFCEnv_When_Load_Then_NoError(t *testing.T) {
	setLiveEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, config.ModeLive, cfg.Mode)
	require.Equal(t, "8080", cfg.Port) // default
	require.Equal(t, "123", cfg.FCPartnerID)
	require.True(t, cfg.FCRMQTLS, "FC_RMQ_TLS defaults to true")
}

// Given FEED_MODE=live and FC_API_PASS missing
// When config.Load() is called
// Then it fails fast with an error naming every missing variable
func TestGiven_LiveModeMissingFCPass_When_Load_Then_FailsFastNamingMissing(t *testing.T) {
	setLiveEnv(t)
	t.Setenv("FC_API_PASS", "")
	t.Setenv("FC_RMQ_PASS", "") // also missing

	cfg, err := config.Load()
	require.Nil(t, cfg)
	require.Error(t, err)
	require.True(t, config.IsMissing(err, "FC_API_PASS"))
	require.True(t, config.IsMissing(err, "FC_RMQ_PASS"))
	require.False(t, config.IsMissing(err, "DATABASE_URL"))
	require.Contains(t, err.Error(), "FC_API_PASS")
	require.Contains(t, err.Error(), "FC_RMQ_PASS")
}

// Given FEED_MODE=replay and no FC_* vars
// When config.Load() is called
// Then it returns a Config without requiring FC_* (replay mode)
func TestGiven_ReplayMode_When_Load_Then_NoFCRequired(t *testing.T) {
	setReplayEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, config.ModeReplay, cfg.Mode)
	require.Empty(t, cfg.FCAPIPass)
	require.Empty(t, cfg.FCPartnerID)
}

// Given DATABASE_URL absent
// When config.Load() is called
// Then it fails fast regardless of mode
func TestGiven_NoDatabaseURL_When_Load_Then_FailsFast(t *testing.T) {
	setReplayEnv(t)
	t.Setenv("DATABASE_URL", "")

	cfg, err := config.Load()
	require.Nil(t, cfg)
	require.Error(t, err)
	require.True(t, config.IsMissing(err, "DATABASE_URL"))
}

// Defensive: PORT default + custom + WS origins parsing.
func TestGiven_PortAndOrigins_When_Load_Then_ParsedCorrectly(t *testing.T) {
	setReplayEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("WS_ALLOWED_ORIGINS", " http://a.test , https://b.test ,, ")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "9000", cfg.Port)
	require.Equal(t, []string{"http://a.test", "https://b.test"}, cfg.WSAllowedOrigins)
}

// Redacted() must not leak any password.
func TestGiven_PopulatedConfig_When_Redacted_Then_NoSecretsInString(t *testing.T) {
	setLiveEnv(t)
	t.Setenv("FC_API_PASS", "supersecret-api")
	t.Setenv("FC_RMQ_PASS", "supersecret-rmq")

	cfg, err := config.Load()
	require.NoError(t, err)

	r := cfg.Redacted()
	require.False(t, strings.Contains(r, "supersecret-api"), "api pass leaked: %s", r)
	require.False(t, strings.Contains(r, "supersecret-rmq"), "rmq pass leaked: %s", r)
}
