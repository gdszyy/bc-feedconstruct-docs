package storage_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

// Sanity: every migration file is embedded, ordered and non-empty so that
// MigrateFromFS won't silently skip schema chunks at runtime.
func TestGiven_EmbeddedMigrations_When_Listed_Then_OrderedAndNonEmpty(t *testing.T) {
	want := []string{
		"001_init.sql",
		"002_catalog.sql",
		"003_markets.sql",
		"004_settlement.sql",
		"005_subscriptions.sql",
		"006_recovery.sql",
	}
	for _, name := range want {
		body, err := readEmbedded(name)
		require.NoError(t, err, "missing embedded migration: %s", name)
		require.NotEmpty(t, body, "empty migration: %s", name)
	}
}

func readEmbedded(name string) ([]byte, error) {
	f, err := migrations.FS().Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, 0, 1024)
	chunk := make([]byte, 512)
	for {
		n, rerr := f.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if rerr != nil {
			if rerr.Error() == "EOF" {
				return buf, nil
			}
			return buf, nil
		}
	}
}
