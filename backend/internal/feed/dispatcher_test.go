package feed_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// 验收 3 — 消息覆盖
//
// Given the 9 required message_type strings registered
// When each is dispatched
// Then a registered handler exists for every one and an unknown type
//      is routed to a dead-letter handler with metric increment
func TestGiven_AllRequiredMessageTypes_When_Dispatched_Then_HandlerExistsForEach(t *testing.T) {
	required := []feed.MessageType{
		feed.MsgOddsChange,
		feed.MsgBetStop,
		feed.MsgBetSettlement,
		feed.MsgBetCancel,
		feed.MsgFixture,
		feed.MsgFixtureChange,
		feed.MsgRollback,
		feed.MsgAlive,
		feed.MsgSnapshotComplete,
	}
	var hits int64
	var deadHits int64
	dead := feed.HandlerFunc(func(context.Context, feed.MessageType, feed.Envelope, [16]byte) error {
		atomic.AddInt64(&deadHits, 1)
		return nil
	})
	d := feed.NewDispatcher(dead)
	for _, t := range required {
		d.Register(t, feed.HandlerFunc(func(context.Context, feed.MessageType, feed.Envelope, [16]byte) error {
			atomic.AddInt64(&hits, 1)
			return nil
		}))
	}

	for _, mt := range required {
		require.NoError(t, d.Dispatch(context.Background(), mt, feed.Envelope{}, [16]byte{}))
	}
	require.Equal(t, int64(len(required)), atomic.LoadInt64(&hits))

	// Unknown type routes to dead-letter.
	require.NoError(t, d.Dispatch(context.Background(), feed.MsgUnknown, feed.Envelope{}, [16]byte{}))
	require.Equal(t, int64(1), atomic.LoadInt64(&deadHits))
	require.Equal(t, int64(1), d.UnknownCount(feed.MsgUnknown))
}

func TestGiven_RegisteredHandler_When_HandlerReturnsError_Then_DispatcherSurfaces(t *testing.T) {
	d := feed.NewDispatcher(nil)
	d.Register(feed.MsgOddsChange, feed.HandlerFunc(func(context.Context, feed.MessageType, feed.Envelope, [16]byte) error {
		return context.Canceled
	}))
	err := d.Dispatch(context.Background(), feed.MsgOddsChange, feed.Envelope{}, [16]byte{})
	require.ErrorIs(t, err, context.Canceled)
}
