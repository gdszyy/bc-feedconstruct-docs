//go:build integration

package feed_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

func gzipBytes(t *testing.T, in []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(in)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

var (
	feedPool *storage.Pool
	feedOnce sync.Once
	feedErr  error
)

func newProcessor(t *testing.T) (*feed.Processor, *storage.RawMessageRepo, *storage.Pool) {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping feed integration tests")
	}
	feedOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			feedErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			feedErr = err
			return
		}
		feedPool = p
	})
	require.NoError(t, feedErr)
	pool := feedPool

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE raw_messages, metrics_counters RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	repo := storage.NewRawMessageRepo(pool)
	proc := feed.NewProcessor(repo, feed.NopPublisher{}, feed.NewDispatcher(nil))
	return proc, repo, pool
}

// 验收 2 — 消息留痕
//
// Given a GZIP-compressed JSON delivery from FeedConstruct RMQ
// When the consumer receives it
// Then it ungzips, parses the envelope, and INSERTs into raw_messages
//      with source / queue / message_type / event_id / payload populated
//      BEFORE any business handler is invoked
func TestGiven_GzippedDelivery_When_Received_Then_RawMessagesRowWrittenBeforeHandler(t *testing.T) {
	proc, repo, pool := newProcessor(t)
	ctx := context.Background()

	body := gzipBytes(t, []byte(`{"objectType":13,"matchId":42,"sportId":1,"timestamp":"2026-05-14T12:00:00Z"}`))

	res, err := proc.Process(ctx, body, feed.DeliveryMeta{
		Source: "rmq.live", Queue: "P123_live", RoutingKey: "odds_change.42",
	})
	require.NoError(t, err)
	require.True(t, res.RawMessage.Inserted)
	require.Equal(t, feed.MsgOddsChange, res.MessageType)

	n, err := repo.CountSince(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	var (
		mt, src, queue, eventID string
		payload                 []byte
		procErr                 *string
	)
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT message_type, source, queue, COALESCE(event_id,''), payload, process_error
		FROM raw_messages LIMIT 1
	`).Scan(&mt, &src, &queue, &eventID, &payload, &procErr))
	require.Equal(t, string(feed.MsgOddsChange), mt)
	require.Equal(t, "rmq.live", src)
	require.Equal(t, "P123_live", queue)
	require.Equal(t, "42", eventID)
	// Postgres reserializes jsonb so we parse rather than substring-match.
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(payload, &parsed))
	require.EqualValues(t, 13, parsed["objectType"])
	require.EqualValues(t, 42, parsed["matchId"])
	require.Nil(t, procErr)
}

// Given a delivery whose envelope cannot be parsed
// When the consumer processes it
// Then the raw bytes are still persisted (raw_blob non-null), process_error is set
//      and no fan-out happens
func TestGiven_UnparsableEnvelope_When_Received_Then_RawBlobKeptErrorRecorded(t *testing.T) {
	proc, _, pool := newProcessor(t)
	ctx := context.Background()

	body := []byte("not json at all")
	res, err := proc.Process(ctx, body, feed.DeliveryMeta{
		Source: "rmq.live", Queue: "P123_live",
	})
	require.NoError(t, err)
	require.Equal(t, feed.MsgUnknown, res.MessageType)

	var (
		mt       string
		blob     []byte
		procErr  *string
	)
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT message_type, raw_blob, process_error FROM raw_messages LIMIT 1
	`).Scan(&mt, &blob, &procErr))
	require.Equal(t, string(feed.MsgUnknown), mt)
	require.Equal(t, body, blob, "raw_blob must hold the original bytes")
	require.NotNil(t, procErr)
	require.NotEmpty(t, *procErr)
}

// 验收 11 (feed level) — duplicate delivery collapses to one audit row
// AND only the first delivery fans out to handlers.
func TestGiven_DuplicateGzipDelivery_When_Processed_Then_OneRowOneFanout(t *testing.T) {
	proc, repo, _ := newProcessor(t)
	ctx := context.Background()

	dispatched := 0
	proc.Dispatcher.Register(feed.MsgOddsChange, feed.HandlerFunc(
		func(context.Context, feed.MessageType, feed.Envelope, [16]byte) error {
			dispatched++
			return nil
		}))

	body := gzipBytes(t, []byte(`{"objectType":13,"matchId":99,"timestamp":"2026-05-14T13:00:00Z"}`))
	meta := feed.DeliveryMeta{Source: "rmq.live", Queue: "P123_live", RoutingKey: "odds_change.99"}

	r1, err := proc.Process(ctx, body, meta)
	require.NoError(t, err)
	require.True(t, r1.RawMessage.Inserted)

	r2, err := proc.Process(ctx, body, meta)
	require.NoError(t, err)
	require.False(t, r2.RawMessage.Inserted, "duplicate must collapse to existing audit row")

	n, err := repo.CountSince(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.Equal(t, 1, dispatched, "duplicate must NOT trigger a second fan-out")
}
