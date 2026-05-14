package feed

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// FeedEventsExchange is the topic exchange the BFF publishes into after
// raw_messages persistence. Handlers (M02 dispatcher) bind their queues to
// this exchange with routing keys like "odds_change.#".
const FeedEventsExchange = "feed.events"

// Publisher pushes a parsed delivery into the internal exchange. The
// interface lets unit tests substitute a stub without an AMQP broker.
type Publisher interface {
	Publish(ctx context.Context, msgType MessageType, sportID int32, routingHint string, body []byte) error
	Close() error
}

// AMQPPublisher publishes to FeedEventsExchange over the internal
// RabbitMQ provided by Railway. Safe for concurrent use.
type AMQPPublisher struct {
	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewAMQPPublisher dials the internal RabbitMQ, declares the topic
// exchange and returns a ready publisher.
func NewAMQPPublisher(url string) (*AMQPPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("feed: dial internal amqp: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("feed: channel: %w", err)
	}
	if err := ch.ExchangeDeclare(
		FeedEventsExchange, amqp.ExchangeTopic,
		true /* durable */, false /* auto-delete */, false, false, nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("feed: declare exchange: %w", err)
	}
	return &AMQPPublisher{conn: conn, ch: ch}, nil
}

func (p *AMQPPublisher) Publish(ctx context.Context, msgType MessageType, sportID int32, routingHint string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := string(msgType)
	if sportID > 0 {
		key = fmt.Sprintf("%s.%d", key, sportID)
	} else if routingHint != "" {
		key = fmt.Sprintf("%s.%s", key, routingHint)
	}
	return p.ch.PublishWithContext(
		ctx, FeedEventsExchange, key, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (p *AMQPPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var first error
	if p.ch != nil {
		if err := p.ch.Close(); err != nil && first == nil {
			first = err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// NopPublisher discards everything. Useful when the internal exchange is
// not configured (e.g. early bring-up) and for unit tests.
type NopPublisher struct{}

func (NopPublisher) Publish(context.Context, MessageType, int32, string, []byte) error { return nil }
func (NopPublisher) Close() error                                                      { return nil }
