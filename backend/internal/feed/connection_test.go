package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// 验收 1 — 连接（M01 / 上传指引 §2 "连接接入"）
//
// Given live mode but FC_API_PASS missing
// When the consumer attempts to start
// Then it fails fast with a clear error naming the missing variable
func TestGiven_LiveModeMissingFCPass_When_ConsumerStarts_Then_FailsFastWithMissingVar(t *testing.T) {
	c := &feed.LiveConsumer{
		Cfg: feed.LiveConsumerConfig{
			Host:      "rmq.test:5673",
			User:      "ru",
			Pass:      "", // missing
			PartnerID: "123",
		},
		// Processor unused because validation fails first; supplying a
		// minimal non-nil Processor avoids the "needs Processor" guard.
		Processor: &feed.Processor{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := c.Run(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "FC_RMQ_PASS")
}

// Given live mode with no Processor configured
// When Run is invoked
// Then it returns an immediate configuration error
func TestGiven_NoProcessor_When_LiveConsumerRun_Then_ImmediateError(t *testing.T) {
	c := &feed.LiveConsumer{
		Cfg: feed.LiveConsumerConfig{
			Host: "rmq.test:5673", User: "u", Pass: "p", PartnerID: "1",
		},
		Processor: nil,
	}
	err := c.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "Processor")
}

// 验收 10 适配 — exponential backoff bounds
//
// Given an unreachable FC host (loopback, closed port)
// When LiveConsumer.Run is invoked
// Then it retries with backoff until ctx is cancelled (never panics)
func TestGiven_UnreachableHost_When_Run_Then_RetriesUntilCancel(t *testing.T) {
	c := &feed.LiveConsumer{
		Cfg: feed.LiveConsumerConfig{
			Host:          "127.0.0.1:1", // closed
			User:          "u",
			Pass:          "p",
			PartnerID:     "1",
			ReconnectBase: 50 * time.Millisecond,
			ReconnectMax:  100 * time.Millisecond,
		},
		Processor: &feed.Processor{Repo: nil, Pub: feed.NopPublisher{}, Dispatcher: feed.NewDispatcher(nil)},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := c.Run(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
