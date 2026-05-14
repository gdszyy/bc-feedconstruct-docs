package feed

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// LiveConsumerConfig captures the FeedConstruct RMQ connection settings.
// Field semantics match docs/08_backend_railway/01_railway_topology.md.
type LiveConsumerConfig struct {
	Host      string // host:port, e.g. odds-stream-rmq-stage.feedstream.org:5673
	User      string
	Pass      string
	PartnerID string
	UseTLS    bool

	HeartbeatSec   int // default 30
	Prefetch       int // default 64
	ReconnectBase  time.Duration
	ReconnectMax   time.Duration
}

// LiveConsumer connects to FeedConstruct RMQ, consumes both partner
// queues and feeds each delivery through the Processor.
type LiveConsumer struct {
	Cfg       LiveConsumerConfig
	Processor *Processor
}

// Run blocks until ctx is done. It reconnects with exponential backoff
// on failures (acceptance #1 — connection robustness).
func (l *LiveConsumer) Run(ctx context.Context) error {
	if l.Processor == nil {
		return fmt.Errorf("feed: live consumer needs Processor")
	}
	if err := l.Cfg.validate(); err != nil {
		return err
	}

	base := l.Cfg.ReconnectBase
	if base == 0 {
		base = 2 * time.Second
	}
	max := l.Cfg.ReconnectMax
	if max == 0 {
		max = 30 * time.Second
	}

	delay := base
	for {
		err := l.runOnce(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err == nil {
			// runOnce only returns nil when the broker closes cleanly; resume.
			delay = base
			continue
		}
		// Wait then retry.
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
		delay *= 2
		if delay > max {
			delay = max
		}
	}
}

func (l *LiveConsumer) runOnce(ctx context.Context) error {
	url, err := l.amqpURL()
	if err != nil {
		return err
	}
	cfg := amqp.Config{
		Heartbeat: time.Duration(l.heartbeat()) * time.Second,
		Locale:    "en_US",
	}
	if l.Cfg.UseTLS {
		cfg.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	conn, err := amqp.DialConfig(url, cfg)
	if err != nil {
		return fmt.Errorf("feed: dial FC RMQ: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("feed: open channel: %w", err)
	}
	defer ch.Close()
	if err := ch.Qos(l.prefetch(), 0, false); err != nil {
		return fmt.Errorf("feed: qos: %w", err)
	}

	liveQ := fmt.Sprintf("P%s_live", l.Cfg.PartnerID)
	preQ := fmt.Sprintf("P%s_prematch", l.Cfg.PartnerID)

	liveCh, err := ch.Consume(liveQ, "bff-live", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("feed: consume %s: %w", liveQ, err)
	}
	preCh, err := ch.Consume(preQ, "bff-prematch", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("feed: consume %s: %w", preQ, err)
	}

	closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-closeCh:
			if e == nil {
				return nil
			}
			return fmt.Errorf("feed: FC RMQ connection closed: %w", e)
		case d, ok := <-liveCh:
			if !ok {
				return fmt.Errorf("feed: live delivery channel closed")
			}
			l.handle(ctx, d, "rmq.live", liveQ)
		case d, ok := <-preCh:
			if !ok {
				return fmt.Errorf("feed: prematch delivery channel closed")
			}
			l.handle(ctx, d, "rmq.prematch", preQ)
		}
	}
}

func (l *LiveConsumer) handle(ctx context.Context, d amqp.Delivery, src, queue string) {
	meta := DeliveryMeta{Source: src, Queue: queue, RoutingKey: d.RoutingKey}
	if _, err := l.Processor.Process(ctx, d.Body, meta); err != nil {
		// Persistence failures requeue; poison messages get acked because
		// processor already stored raw_blob + process_error.
		_ = d.Nack(false, isRetryable(err))
		return
	}
	_ = d.Ack(false)
}

func (l *LiveConsumer) amqpURL() (string, error) {
	scheme := "amqp"
	if l.Cfg.UseTLS {
		scheme = "amqps"
	}
	return fmt.Sprintf("%s://%s:%s@%s/", scheme, l.Cfg.User, l.Cfg.Pass, l.Cfg.Host), nil
}

func (l *LiveConsumer) heartbeat() int {
	if l.Cfg.HeartbeatSec > 0 {
		return l.Cfg.HeartbeatSec
	}
	return 30
}

func (l *LiveConsumer) prefetch() int {
	if l.Cfg.Prefetch > 0 {
		return l.Cfg.Prefetch
	}
	return 64
}

func (c *LiveConsumerConfig) validate() error {
	missing := make([]string, 0, 4)
	if c.Host == "" {
		missing = append(missing, "FC_RMQ_HOST")
	}
	if c.User == "" {
		missing = append(missing, "FC_RMQ_USER")
	}
	if c.Pass == "" {
		missing = append(missing, "FC_RMQ_PASS")
	}
	if c.PartnerID == "" {
		missing = append(missing, "FC_PARTNER_ID")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("feed: live consumer missing env: %s", strings.Join(missing, ", "))
}

// isRetryable reports whether an error from Processor warrants requeue.
// Persistence (DB) and publish errors are retryable; envelope-parse
// failures are not (already stored as poison row).
func isRetryable(err error) bool {
	s := err.Error()
	switch {
	case strings.Contains(s, "persist raw_message"),
		strings.Contains(s, "publish"):
		return true
	}
	return false
}
