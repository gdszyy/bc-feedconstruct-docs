package feed_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// memRepo is an in-memory implementation of feed.RawInserter for unit
// tests of the Replayer. It enforces the same idempotency contract as
// the real repository: identical (source, message_type, event_id) keys
// collapse to a single Inserted=true row.
type memRepo struct {
	mu       sync.Mutex
	rows     []storage.RawMessage
	keyIndex map[string][16]byte
	errors   map[[16]byte]string
}

func newMemRepo() *memRepo {
	return &memRepo{
		keyIndex: make(map[string][16]byte),
		errors:   make(map[[16]byte]string),
	}
}

func (m *memRepo) Insert(_ context.Context, msg storage.RawMessage) (storage.InsertResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := msg.Source + "|" + msg.MessageType + "|" + msg.EventID
	if id, ok := m.keyIndex[key]; ok {
		return storage.InsertResult{ID: id, Inserted: false, ReceivedAt: time.Now()}, nil
	}
	var id [16]byte
	copy(id[:], []byte(key)) // deterministic-ish for assertions
	m.keyIndex[key] = id
	m.rows = append(m.rows, msg)
	return storage.InsertResult{ID: id, Inserted: true, ReceivedAt: time.Now()}, nil
}

func (m *memRepo) MarkProcessError(_ context.Context, id [16]byte, msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[id] = msg
	return nil
}

type countingPublisher struct{ count int }

func (c *countingPublisher) Publish(context.Context, feed.MessageType, int32, string, []byte) error {
	c.count++
	return nil
}
func (c *countingPublisher) Close() error { return nil }

// Given the testdata/replay directory holds three filenames in timestamp order
// When the Replayer runs
// Then the Processor sees them in lexical (== timestamp) order, classifies
//      each correctly, and a duplicate pass causes no second fan-out
func TestGiven_TestdataDir_When_ReplayerRuns_Then_OrderedAndIdempotent(t *testing.T) {
	repo := newMemRepo()
	pub := &countingPublisher{}
	disp := feed.NewDispatcher(nil)

	var seenMu sync.Mutex
	var seen []feed.MessageType
	record := feed.HandlerFunc(func(_ context.Context, m feed.MessageType, _ feed.Envelope, _ [16]byte) error {
		seenMu.Lock()
		seen = append(seen, m)
		seenMu.Unlock()
		return nil
	})
	for _, mt := range []feed.MessageType{
		feed.MsgOddsChange,
		feed.MsgBetSettlement,
		feed.MessageType("bet_stop"),
	} {
		disp.Register(mt, record)
	}

	proc := feed.NewProcessor(repo, pub, disp)

	dir, err := filepath.Abs("testdata/replayer_unit")
	require.NoError(t, err)
	rep := &feed.Replayer{Dir: dir, Processor: proc, Source: "replay"}

	n, err := rep.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, n)
	require.Equal(t, []feed.MessageType{
		feed.MsgOddsChange,
		feed.MessageType("bet_stop"),
		feed.MsgBetSettlement,
	}, seen)
	require.Equal(t, 3, pub.count)
	require.Len(t, repo.rows, 3)

	// Second pass: every file collapses on the audit key. No new publishes
	// and no new dispatcher invocations.
	pubBefore := pub.count
	seenBefore := len(seen)
	n2, err := rep.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, n2)
	require.Equal(t, pubBefore, pub.count)
	require.Equal(t, seenBefore, len(seen))
	require.Len(t, repo.rows, 3)
}

// Given the Replayer is started with no Processor
// When Run is called
// Then it returns a configuration error
func TestGiven_NoProcessor_When_ReplayerRun_Then_ConfigError(t *testing.T) {
	r := &feed.Replayer{Dir: "/tmp"}
	_, err := r.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "Processor")
}

// Given the Replayer is pointed at a non-existent directory
// When Run is called
// Then it returns a read-dir error
func TestGiven_MissingDir_When_ReplayerRun_Then_ReadError(t *testing.T) {
	repo := newMemRepo()
	proc := feed.NewProcessor(repo, feed.NopPublisher{}, feed.NewDispatcher(nil))
	r := &feed.Replayer{Dir: "/nonexistent-dir-xyzzy-12345", Processor: proc}
	_, err := r.Run(context.Background())
	require.Error(t, err)
}
