//go:build integration

package settlement_test

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
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/settlement"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	stPool *storage.Pool
	stOnce sync.Once
	stErr  error
)

func setup(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping settlement integration tests")
	}
	stOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			stErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			stErr = err
			return
		}
		stPool = p
	})
	require.NoError(t, stErr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := stPool.Exec(ctx, `
		TRUNCATE TABLE rollbacks, cancels, settlements, market_status_history,
			outcomes, markets, fixture_changes, matches, competitions,
			regions, sports, raw_messages, metrics_counters
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return stPool
}

func env(body string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(body))
	if err != nil {
		e = feed.Envelope{Payload: []byte(body)}
	}
	e.Payload = []byte(body)
	return e
}

func seedActiveMarket(t *testing.T, pool *storage.Pool, matchID int64, marketTypeID int64) {
	t.Helper()
	ctx := context.Background()
	cat := catalog.New(pool)
	require.NoError(t, cat.Handle(ctx, feed.MsgCatalogSport,
		env(`{"id":1,"name":"Soccer"}`), [16]byte{}))
	require.NoError(t, cat.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":42,"sportId":1,"status":"live"}`), [16]byte{}))
	od := odds.New(pool)
	require.NoError(t, od.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":1,"outcomes":[
			{"id":1,"odds":1.85},{"id":2,"odds":2.10}
		]}`), [16]byte{}))
}

// 验收 7 — 结算
//
// Given a bet_settlement for outcome (42,1,"",1) with result=win, certainty=1
// When SettlementHandler processes it
// Then a settlements row is inserted with result=win, certainty=1
//      AND markets row (42,1,"") transitions to status=settled
func TestGiven_BetSettlementWin_When_Handled_Then_SettlementRowAndMarketSettled(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","outcomes":[
		{"id":1,"result":"win","certainty":1}
	]}`
	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement, env(body), [16]byte{}))

	var result string
	var certainty int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT result, certainty FROM settlements
		 WHERE match_id=42 AND market_type_id=1 AND specifier='' AND outcome_id=1
	`).Scan(&result, &certainty))
	require.Equal(t, "win", result)
	require.Equal(t, 1, certainty)

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42 AND market_type_id=1`,
	).Scan(&status))
	require.Equal(t, "settled", status)
}

// Given a bet_settlement carrying void_factor=0.5 and dead_heat_factor=0.25
// When SettlementHandler processes it
// Then settlements row stores both factors verbatim
func TestGiven_VoidAndDeadHeatFactors_When_Handled_Then_FactorsPersistedExactly(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	body := `{"matchId":42,"marketTypeId":1,"outcomes":[
		{"id":1,"result":"void","certainty":1,"voidFactor":0.5,"deadHeatFactor":0.25}
	]}`
	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement, env(body), [16]byte{}))

	var (
		vf, dhf *float64
	)
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT void_factor, dead_heat_factor FROM settlements
		 WHERE match_id=42 AND outcome_id=1
	`).Scan(&vf, &dhf))
	require.NotNil(t, vf)
	require.NotNil(t, dhf)
	require.InDelta(t, 0.5, *vf, 1e-9)
	require.InDelta(t, 0.25, *dhf, 1e-9)
}

// Given an uncertain settlement followed by a certain one for the same outcome
// When both are processed in order
// Then the certain settlement supersedes the uncertain one in place
//      AND no duplicate row exists for that outcome
func TestGiven_UncertainThenCertain_When_Handled_Then_CertainSupersedes(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	uncertain := `{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"result":"win","certainty":0}]}`
	certain := `{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"result":"win","certainty":1}]}`
	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement, env(uncertain), [16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement, env(certain), [16]byte{}))

	var count int
	var certainty int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*), MAX(certainty) FROM settlements WHERE match_id=42 AND outcome_id=1`,
	).Scan(&count, &certainty))
	require.Equal(t, 1, count, "must update in place, not append")
	require.Equal(t, 1, certainty)
}

// 验收 8 — 取消（VoidNotification, VoidAction=1）
//
// Given a VoidNotification with VoidAction=1, ObjectType=13 (market)
// When CancelHandler processes it
// Then a cancels row is inserted with void_reason / from_ts / to_ts
//      AND the targeted markets row transitions to status=cancelled
func TestGiven_VoidNotificationVoid_When_Handled_Then_CancelRowAndMarketCancelled(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	body := `{"objectType":13,"matchId":42,"marketTypeId":1,"specifier":"",
		"voidAction":1,"reason":"event_void",
		"fromDate":"2026-05-14T11:00:00Z","toDate":"2026-05-14T13:00:00Z"}`
	require.NoError(t, h.Handle(ctx, feed.MsgBetCancel, env(body), [16]byte{}))

	var (
		reason    string
		fromTs    time.Time
		toTs      time.Time
		voidAct   int
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT void_reason, from_ts, to_ts, void_action FROM cancels WHERE match_id=42`,
	).Scan(&reason, &fromTs, &toTs, &voidAct))
	require.Equal(t, "event_void", reason)
	require.WithinDuration(t, time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC), fromTs, time.Second)
	require.WithinDuration(t, time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC), toTs, time.Second)
	require.Equal(t, 1, voidAct)

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42`,
	).Scan(&status))
	require.Equal(t, "cancelled", status)
}

// Given a cancel for ObjectType=4 (match) covering all markets
// When processed
// Then every market of that match transitions to status=cancelled
func TestGiven_MatchLevelCancel_When_Handled_Then_AllMarketsCancelled(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	od := odds.New(pool)
	require.NoError(t, od.Handle(ctx, feed.MsgOddsChange,
		env(`{"matchId":42,"marketTypeId":2,"outcomes":[{"id":1,"odds":1.5}]}`), [16]byte{}))
	h := settlement.New(pool)

	body := `{"objectType":4,"objectId":42,"matchId":42,"voidAction":1,"reason":"match_void"}`
	require.NoError(t, h.Handle(ctx, feed.MsgBetCancel, env(body), [16]byte{}))

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
	require.Equal(t, []string{"cancelled", "cancelled"}, statuses)
}

// 验收 9 — 回滚
//
// Given an existing settlements row for outcome (42,1,"",1)
// When a rollback_bet_settlement message arrives for the same outcome
// Then a rollbacks row is inserted (target='settlement') AND
//      settlements.rolled_back_at is set AND markets reverts from settled to active
func TestGiven_ExistingSettlement_When_RollbackArrives_Then_RollbackRecordedAndMarketReverts(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement,
		env(`{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"result":"win","certainty":1}]}`),
		[16]byte{}))

	require.NoError(t, h.Handle(ctx, feed.MsgRollback,
		env(`{"matchId":42,"marketTypeId":1,"specifier":""}`), [16]byte{}))

	var rolledBack *time.Time
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT rolled_back_at FROM settlements WHERE match_id=42 AND outcome_id=1`,
	).Scan(&rolledBack))
	require.NotNil(t, rolledBack)

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42 AND market_type_id=1`,
	).Scan(&status))
	require.Equal(t, "active", status)

	var rb int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM rollbacks WHERE target='settlement'`).Scan(&rb))
	require.Equal(t, 1, rb)
}

// Given an existing cancels row
// When a rollback_cancel (VoidAction=2) arrives for the same target
// Then a rollbacks row is inserted (target='cancel')
//      AND cancels.rolled_back_at is set
//      AND markets exits cancelled to its prior status
func TestGiven_ExistingCancel_When_UnvoidArrives_Then_RollbackRecordedAndMarketRecovers(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	require.NoError(t, h.Handle(ctx, feed.MsgBetCancel,
		env(`{"objectType":13,"matchId":42,"marketTypeId":1,"voidAction":1,"reason":"void"}`),
		[16]byte{}))

	require.NoError(t, h.Handle(ctx, feed.MsgRollbackCancel,
		env(`{"matchId":42,"marketTypeId":1,"specifier":""}`), [16]byte{}))

	var rolledBack *time.Time
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT rolled_back_at FROM cancels WHERE match_id=42`,
	).Scan(&rolledBack))
	require.NotNil(t, rolledBack)

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=42`,
	).Scan(&status))
	require.NotEqual(t, "cancelled", status, "must exit cancelled status after unvoid")

	var rb int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM rollbacks WHERE target='cancel'`).Scan(&rb))
	require.Equal(t, 1, rb)
}

// Given a rollback delivered twice (same raw_message_id)
// When both deliveries are processed
// Then rollbacks contains exactly one row (idempotent)
func TestGiven_DuplicateRollback_When_Handled_Then_Idempotent(t *testing.T) {
	pool := setup(t)
	seedActiveMarket(t, pool, 42, 1)
	ctx := context.Background()
	h := settlement.New(pool)

	require.NoError(t, h.Handle(ctx, feed.MsgBetSettlement,
		env(`{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"result":"win","certainty":1}]}`),
		[16]byte{}))

	// Build a real raw_messages id so the FK is testable.
	var rawID [16]byte
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO raw_messages (source, message_type, event_id, payload)
		VALUES ('test','rollback','42','{}'::jsonb)
		RETURNING id`).Scan(&rawID))

	require.NoError(t, h.Handle(ctx, feed.MsgRollback,
		env(`{"matchId":42,"marketTypeId":1,"specifier":""}`), rawID))
	require.NoError(t, h.Handle(ctx, feed.MsgRollback,
		env(`{"matchId":42,"marketTypeId":1,"specifier":""}`), rawID))

	var rb int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM rollbacks WHERE target='settlement'`).Scan(&rb))
	require.Equal(t, 1, rb)
}
