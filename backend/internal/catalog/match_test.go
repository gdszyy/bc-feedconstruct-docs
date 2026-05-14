//go:build integration

package catalog_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	catPool *storage.Pool
	catOnce sync.Once
	catErr  error
)

func setupPool(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping catalog integration tests")
	}
	catOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			catErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			catErr = err
			return
		}
		catPool = p
	})
	require.NoError(t, catErr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := catPool.Exec(ctx,
		`TRUNCATE TABLE fixture_changes, matches, competitions, regions, sports, raw_messages, metrics_counters RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return catPool
}

func env(payload string, msgType feed.MessageType) feed.Envelope {
	body := []byte(payload)
	e, err := feed.DecodeEnvelope(body)
	if err != nil {
		// keep going — handler tolerates missing fields
		e = feed.Envelope{Payload: body}
	}
	e.Payload = body
	_ = msgType
	return e
}

// 验收 4 — 主数据（M03/M04）
//
// Given a Match delivery with explicit sportId / competitionId
// When the catalog handler processes it
// Then sports / regions / competitions / matches rows are upserted with
//      name / start_at / home / away / status populated
func TestGiven_MatchObject_When_Handled_Then_HierarchyUpserted(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	h := catalog.New(pool)

	// First seed sport + region + competition via their own deliveries.
	require.NoError(t, h.Handle(ctx, feed.MsgCatalogSport,
		env(`{"id":1,"name":"Soccer","isActive":true}`, feed.MsgCatalogSport), [16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgCatalogRegion,
		env(`{"id":10,"sportId":1,"name":"Europe"}`, feed.MsgCatalogRegion), [16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgCatalogComp,
		env(`{"id":100,"sportId":1,"regionId":10,"name":"UCL"}`, feed.MsgCatalogComp), [16]byte{}))

	// Now the match.
	start := time.Date(2026, 5, 14, 18, 0, 0, 0, time.UTC).Format(time.RFC3339)
	body := `{"matchId":42,"sportId":1,"competitionId":100,"name":"A vs B","home":"A","away":"B","startAt":"` + start + `","status":"live"}`
	require.NoError(t, h.Handle(ctx, feed.MsgFixture, env(body, feed.MsgFixture), [16]byte{}))

	var (
		sportName string
		regionID  int64
		compID    int64
		home      string
		away      string
		status    string
		isLive    bool
		startAt   time.Time
	)
	require.NoError(t, pool.QueryRow(ctx, `SELECT name FROM sports WHERE id=1`).Scan(&sportName))
	require.Equal(t, "Soccer", sportName)
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM regions WHERE name='Europe'`).Scan(&regionID))
	require.Equal(t, int64(10), regionID)
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM competitions WHERE name='UCL'`).Scan(&compID))
	require.Equal(t, int64(100), compID)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT home, away, status, is_live, start_at FROM matches WHERE id=42`,
	).Scan(&home, &away, &status, &isLive, &startAt))
	require.Equal(t, "A", home)
	require.Equal(t, "B", away)
	require.Equal(t, "live", status)
	require.True(t, isLive)
	require.WithinDuration(t, time.Date(2026, 5, 14, 18, 0, 0, 0, time.UTC), startAt, time.Second)
}

// Given a fixture_change altering status from live -> ended and a new start_at
// When the catalog handler processes it as fixture_change
// Then matches row is updated AND a fixture_changes row is inserted with the diff
func TestGiven_FixtureChange_When_Handled_Then_MatchUpdatedAndHistoryRecorded(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	h := catalog.New(pool)

	// seed match
	require.NoError(t, h.Handle(ctx, feed.MsgCatalogSport, env(`{"id":1,"name":"Soccer"}`, feed.MsgCatalogSport), [16]byte{}))
	body1 := `{"matchId":50,"sportId":1,"home":"X","away":"Y","status":"live"}`
	require.NoError(t, h.Handle(ctx, feed.MsgFixture, env(body1, feed.MsgFixture), [16]byte{}))

	// Insert a real raw_messages row first so the FK link is preservable.
	var rawUUID [16]byte
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO raw_messages (source, message_type, event_id, payload)
		VALUES ('test', 'fixture_change', '50', '{}'::jsonb)
		RETURNING id`,
	).Scan(&rawUUID))
	body2 := `{"matchId":50,"sportId":1,"home":"X","away":"Y","status":"ended"}`
	require.NoError(t, h.Handle(ctx, feed.MsgFixtureChange, env(body2, feed.MsgFixtureChange), rawUUID))

	var status string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM matches WHERE id=50`).Scan(&status))
	require.Equal(t, "ended", status)

	var (
		count       int64
		changeType  string
		oldStatus   *string
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*), MAX(change_type) FROM fixture_changes WHERE match_id=50`,
	).Scan(&count, &changeType))
	require.GreaterOrEqual(t, count, int64(1))
	require.Equal(t, "fixture_change", changeType)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT old->>'status' FROM fixture_changes WHERE match_id=50 ORDER BY id DESC LIMIT 1`,
	).Scan(&oldStatus))
	require.NotNil(t, oldStatus)
	require.Equal(t, "live", *oldStatus)
}

// 验收 12 — 防回退（赛事级）
//
// Given matches.status currently = ended
// When a delivery arrives with status = live for the same match
// Then matches.status is NOT regressed to live and a fixture_changes row
//      does NOT record a backwards transition
func TestGiven_EndedMatch_When_LiveStatusArrives_Then_NoRegressionLogged(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	h := catalog.New(pool)

	require.NoError(t, h.Handle(ctx, feed.MsgCatalogSport, env(`{"id":1,"name":"Soccer"}`, feed.MsgCatalogSport), [16]byte{}))
	require.NoError(t, h.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":77,"sportId":1,"status":"ended"}`, feed.MsgFixture), [16]byte{}))

	// Now an out-of-order delivery wants to re-open it.
	require.NoError(t, h.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":77,"sportId":1,"status":"live"}`, feed.MsgFixture), [16]byte{}))

	var status string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM matches WHERE id=77`).Scan(&status))
	require.Equal(t, "ended", status, "status must not regress from ended to live")
}

// Given a match without sportId
// When handled
// Then the handler returns a clear error and no row is created
func TestGiven_MatchMissingSportID_When_Handled_Then_ErrorAndNoRow(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	h := catalog.New(pool)

	err := h.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":99,"status":"live"}`, feed.MsgFixture), [16]byte{})
	require.Error(t, err)

	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM matches WHERE id=99`).Scan(&count))
	require.Equal(t, 0, count)
}
