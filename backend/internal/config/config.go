// Package config loads and validates environment variables for the BFF.
//
// The loader is fail-fast: any missing required variable for the active
// FEED_MODE is reported up-front by name. See
// docs/08_backend_railway/01_railway_topology.md for the authoritative list.
package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Mode is the data-source mode. Live consumes from FeedConstruct RMQ +
// WebAPI; replay drives the same pipeline from raw/json sample files.
type Mode string

const (
	ModeLive   Mode = "live"
	ModeReplay Mode = "replay"
)

// Config is the parsed and validated env state.
type Config struct {
	Mode Mode

	Port             string
	DatabaseURL      string
	RabbitMQURL      string
	WSAllowedOrigins []string

	// FeedConstruct WebAPI
	FCAPIBase string
	FCAPIUser string
	FCAPIPass string

	// FeedConstruct Translation Web API. Optional: when empty the BFF
	// skips the translation cache layer and serves raw IDs to the
	// frontend (which logs a warning per M12 "displays skeleton, not
	// ID" verification). Defaults to the same host as FCAPIBase.
	FCTranslationBase string

	// FCTranslationLanguages is the comma-separated language list the
	// translation manager refreshes on boot and on schedule (e.g.
	// "en,zh,ru"). Empty means no proactive refresh; cache misses are
	// still served on demand via ById.
	FCTranslationLanguages []string

	// FeedConstruct RMQ
	FCRMQHost   string
	FCRMQUser   string
	FCRMQPass   string
	FCRMQTLS    bool
	FCPartnerID string

	RecoveryInitial bool
	LogLevel        string

	// ReplayDir is the directory of JSON / JSON.gz fixtures the replay
	// mode iterates through. Defaults to backend/internal/feed/testdata/replay
	// when unset, so the binary boots usefully out of the box.
	ReplayDir string
}

// MissingEnvError lists the env var names whose absence prevented load.
type MissingEnvError struct{ Names []string }

func (e *MissingEnvError) Error() string {
	return "missing required env vars: " + strings.Join(e.Names, ", ")
}

// Load reads and validates env. It does not read any file; callers that
// want a .env should source it before invoking the binary.
func Load() (*Config, error) {
	c := &Config{
		Mode:             modeFromEnv(os.Getenv("FEED_MODE")),
		Port:             defaultStr(os.Getenv("PORT"), "8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		RabbitMQURL:      os.Getenv("RABBITMQ_URL"),
		WSAllowedOrigins: splitCSV(os.Getenv("WS_ALLOWED_ORIGINS")),
		FCAPIBase:              os.Getenv("FC_API_BASE"),
		FCAPIUser:              os.Getenv("FC_API_USER"),
		FCAPIPass:              os.Getenv("FC_API_PASS"),
		FCTranslationBase:      os.Getenv("FC_TRANSLATION_BASE"),
		FCTranslationLanguages: splitCSV(os.Getenv("FC_TRANSLATION_LANGUAGES")),
		FCRMQHost:        os.Getenv("FC_RMQ_HOST"),
		FCRMQUser:        os.Getenv("FC_RMQ_USER"),
		FCRMQPass:        os.Getenv("FC_RMQ_PASS"),
		FCRMQTLS:         parseBool(os.Getenv("FC_RMQ_TLS"), true),
		FCPartnerID:      os.Getenv("FC_PARTNER_ID"),
		RecoveryInitial:  parseBool(os.Getenv("RECOVERY_INITIAL"), true),
		LogLevel:         defaultStr(os.Getenv("LOG_LEVEL"), "info"),
		ReplayDir:        os.Getenv("REPLAY_DIR"),
	}

	missing := []string{}
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.RabbitMQURL == "" {
		missing = append(missing, "RABBITMQ_URL")
	}
	if c.Mode == ModeLive {
		live := map[string]string{
			"FC_API_BASE":   c.FCAPIBase,
			"FC_API_USER":   c.FCAPIUser,
			"FC_API_PASS":   c.FCAPIPass,
			"FC_RMQ_HOST":   c.FCRMQHost,
			"FC_RMQ_USER":   c.FCRMQUser,
			"FC_RMQ_PASS":   c.FCRMQPass,
			"FC_PARTNER_ID": c.FCPartnerID,
		}
		for k, v := range live {
			if v == "" {
				missing = append(missing, k)
			}
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return nil, &MissingEnvError{Names: missing}
	}
	return c, nil
}

func modeFromEnv(v string) Mode {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "live":
		return ModeLive
	case "", "replay":
		return ModeReplay
	default:
		return Mode(v)
	}
}

func defaultStr(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func parseBool(v string, d bool) bool {
	if v == "" {
		return d
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}
	return b
}

func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsMissing reports whether err is a MissingEnvError listing name.
func IsMissing(err error, name string) bool {
	var m *MissingEnvError
	if !errors.As(err, &m) {
		return false
	}
	for _, n := range m.Names {
		if n == name {
			return true
		}
	}
	return false
}

// Redacted returns a string safe to print in logs.
func (c *Config) Redacted() string {
	return fmt.Sprintf(
		"Config{Mode=%s Port=%s DBPresent=%t RMQPresent=%t FCAPIPresent=%t FCRMQPresent=%t Partner=%q TLS=%t LogLevel=%s RecoveryInitial=%t Origins=%v}",
		c.Mode, c.Port, c.DatabaseURL != "", c.RabbitMQURL != "", c.FCAPIPass != "", c.FCRMQPass != "", c.FCPartnerID, c.FCRMQTLS, c.LogLevel, c.RecoveryInitial, c.WSAllowedOrigins,
	)
}
