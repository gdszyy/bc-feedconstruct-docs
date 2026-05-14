//go:build integration

package odds_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/odds"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	oddPool *storage.Pool
	oddOnce sync.Once
	oddErr  error
)

func setup(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping odds integration tests")
	}
	oddOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			oddErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			oddErr = err
			return
		}
		oddPool = p
	})
	require.NoError(t, oddErr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := oddPool.Exec(ctx, `
		TRUNCATE TABLE market_status_history, outcomes, markets,
			fixture_changes, matches, competitions, regions, sports,
			raw_messages, metrics_counters RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return oddPool
}

func env(body string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(body))
	if err != nil {
		e = feed.Envelope{Payload: []byte(body)}
	}
	e.Payload = []byte(body)
	return e
}

// seedMatch inserts a minimal sport + match so odds_change finds its FK target.
func seedMatch(t *testing.T, pool *storage.Pool, matchID int64) {
	t.Helper()
	ctx := context.Background()
	cat := catalog.New(pool)
	require.NoError(t, cat.Handle(ctx, feed.MsgCatalogSport,
		env(`{"id":1,"name":"Soccer"}`), [16]byte{}))
	require.NoError(t, cat.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":`+itoa(matchID)+`,"sportId":1,"status":"live"}`),
		[16]byte{}))
}

// 验收 5 — 赔率
//
// Given an odds_change for matchId=42 with marketTypeId=1, specifier="",
//       outcomes [{id:1,odds:1.85,active:true},{id:2,odds:2.10,active:true}]
// When the OddsHandler processes it
// Then markets row (42,1,"") is upserted with status=active
//      AND outcomes rows are upserted with the exact odds and is_active flags
func TestGiven_OddsChange_When_Handled_Then_MarketAndOutcomesUpserted(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	h := odds.New(pool)

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","outcomes":[
		{"id":1,"odds":1.85,"isActive":true},
		{"id":2,"odds":2.10,"isActive":true}
	]}`
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange, env(body), [16]byte{}))

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42 AND market_type_id=1 AND specifier=''`,
	).Scan(&status))
	require.Equal(t, "active", status)

	rows, err := pool.Query(ctx, `
		SELECT outcome_id, odds, is_active
		FROM outcomes WHERE match_id=42 AND market_type_id=1 AND specifier=''
		ORDER BY outcome_id`)
	require.NoError(t, err)
	defer rows.Close()
	found := map[int]struct {
		odds   float64
		active bool
	}{}
	for rows.Next() {
		var oid int
		var odds float64
		var active bool
		require.NoError(t, rows.Scan(&oid, &odds, &active))
		found[oid] = struct {
			odds   float64
			active bool
		}{odds: odds, active: active}
	}
	require.Equal(t, 1.85, found[1].odds)
	require.True(t, found[1].active)
	require.Equal(t, 2.10, found[2].odds)

	// market_status_history captured the initial transition NULL -> active.
	var historyCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM market_status_history WHERE match_id=42`,
	).Scan(&historyCount))
	require.Equal(t, 1, historyCount)
}

// Given an odds_change carrying the same payload twice
// When both deliveries are processed
// Then the outcomes row is updated to the same values (no duplicates)
//      AND no extra market_status_history rows beyond the first transition
func TestGiven_DuplicateOddsChange_When_Handled_Then_NoDuplicateRow(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	h := odds.New(pool)

	body := `{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange, env(body), [16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange, env(body), [16]byte{}))

	var outcomeCount, historyCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM outcomes WHERE match_id=42`,
	).Scan(&outcomeCount))
	require.Equal(t, 1, outcomeCount)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM market_status_history WHERE match_id=42`,
	).Scan(&historyCount))
	require.Equal(t, 1, historyCount, "no transition the second time -> no history row")
}

// 验收 12 — 防回退（盘口级）
//
// Given a markets row currently status=settled (terminal)
// When an odds_change arrives that would set it to active
// Then the transition is rejected; markets.status stays settled
//      AND no history row records the rejected transition
func TestGiven_SettledMarket_When_OddsChangeWouldActivate_Then_NoRegression(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	h := odds.New(pool)

	// Bootstrap the market in settled status directly.
	_, err := pool.Exec(ctx, `
		INSERT INTO markets (match_id, market_type_id, specifier, status, updated_at)
		VALUES (42, 1, '', 'settled', now())`)
	require.NoError(t, err)

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","status":"active","outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange, env(body), [16]byte{}))

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42 AND market_type_id=1 AND specifier=''`,
	).Scan(&status))
	require.Equal(t, "settled", status, "settled must not regress to active")

	var historyCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM market_status_history WHERE match_id=42`,
	).Scan(&historyCount))
	require.Equal(t, 0, historyCount, "rejected transition must NOT be logged in history")
}

// 验收 6 — 停投
//
// Given a bet_stop for matchId=42 with status=suspended
// When BetStopHandler processes it
// Then every active market of match 42 transitions to suspended
//      AND a market_status_history row is appended for each transition
func TestGiven_BetStopAllMarkets_When_Handled_Then_MarketsSuspendedAndHistoryAppended(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	h := odds.New(pool)

	// Two markets active for match 42.
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`),
		[16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":2,"outcomes":[{"id":1,"odds":2.10}]}`),
		[16]byte{}))

	// Now bet_stop suspends everything for the match.
	require.NoError(t, h.Handle(ctx, feed.MsgBetStop,
		env(`{"matchId":42,"status":"suspended"}`), [16]byte{}))

	rows, err := pool.Query(ctx,
		`SELECT market_type_id, status FROM markets WHERE match_id=42 ORDER BY market_type_id`)
	require.NoError(t, err)
	defer rows.Close()
	var statuses []string
	for rows.Next() {
		var mt int
		var s string
		require.NoError(t, rows.Scan(&mt, &s))
		statuses = append(statuses, s)
	}
	require.Equal(t, []string{"suspended", "suspended"}, statuses)

	var historyCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM market_status_history WHERE match_id=42 AND to_status='suspended'`,
	).Scan(&historyCount))
	require.Equal(t, 2, historyCount, "history must record one suspension per market")
}

// Given a bet_stop targeting a single market_type
// When BetStopHandler processes it
// Then only markets with that market_type_id transition; others stay active
func TestGiven_BetStopByMarketType_When_Handled_Then_OnlyTargetedTransitions(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	h := odds.New(pool)

	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`),
		[16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":2,"outcomes":[{"id":1,"odds":2.10}]}`),
		[16]byte{}))

	require.NoError(t, h.Handle(ctx, feed.MsgBetStop,
		env(`{"matchId":42,"marketTypeId":1,"status":"suspended"}`), [16]byte{}))

	rows, err := pool.Query(ctx,
		`SELECT market_type_id, status FROM markets WHERE match_id=42 ORDER BY market_type_id`)
	require.NoError(t, err)
	defer rows.Close()
	got := map[int]string{}
	for rows.Next() {
		var mt int
		var s string
		require.NoError(t, rows.Scan(&mt, &s))
		got[mt] = s
	}
	require.Equal(t, "suspended", got[1])
	require.Equal(t, "active", got[2], "untargeted market must remain active")
}

func itoa(i int64) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = digits[i%10]
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
