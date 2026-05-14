package feed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// RawInserter is the slice of storage.RawMessageRepo the Processor needs.
// Defining it here lets unit tests substitute an in-memory implementation.
type RawInserter interface {
	Insert(ctx context.Context, msg storage.RawMessage) (storage.InsertResult, error)
	MarkProcessError(ctx context.Context, id [16]byte, msg string) error
}

// DeliveryMeta is the broker-side context for a single delivery.
type DeliveryMeta struct {
	Source     string // "rmq.live" / "rmq.prematch" / "replay.<file>"
	Queue      string
	RoutingKey string
}

// ProcessResult reports what the processor did with a delivery.
type ProcessResult struct {
	MessageType MessageType
	RawMessage  storage.InsertResult
	Dispatched  bool
}

// Processor wires the ingest chain: gzip decode → envelope parse →
// raw_messages persist → internal exchange publish → dispatcher.
type Processor struct {
	Repo       RawInserter
	Pub        Publisher
	Dispatcher *Dispatcher
	Now        func() time.Time
}

// NewProcessor constructs a Processor with sensible defaults.
func NewProcessor(repo RawInserter, pub Publisher, dispatcher *Dispatcher) *Processor {
	if pub == nil {
		pub = NopPublisher{}
	}
	if dispatcher == nil {
		dispatcher = NewDispatcher(nil)
	}
	return &Processor{Repo: repo, Pub: pub, Dispatcher: dispatcher, Now: time.Now}
}

// Process runs the full ingest chain for one delivery. It tolerates an
// unparseable envelope: the body is still persisted with raw_blob set and
// process_error filled so operators can inspect later.
func (p *Processor) Process(ctx context.Context, body []byte, meta DeliveryMeta) (ProcessResult, error) {
	if p == nil || p.Repo == nil {
		return ProcessResult{}, errors.New("feed: processor not initialised")
	}

	decoded, decErr := DecodeBody(body)
	var jsonBytes []byte
	if decErr == nil {
		jsonBytes = decoded
	} else {
		jsonBytes = body // best effort: persist the raw payload as-is
	}

	env, parseErr := DecodeEnvelope(jsonBytes)
	msgType := MsgUnknown
	if parseErr == nil {
		msgType = Classify(env, meta.Queue)
	}

	raw := storage.RawMessage{
		Source:      meta.Source,
		RoutingKey:  meta.RoutingKey,
		Queue:       meta.Queue,
		MessageType: string(msgType),
		EventID:     env.EventKey(),
		ProductID:   env.ProductID,
		SportID:     env.SportID,
		TSProvider:  env.Timestamp,
		Payload:     jsonBytes,
	}
	if decErr != nil || parseErr != nil {
		// Original bytes aren't valid JSON, so we cannot use them for the
		// jsonb payload column. Persist them verbatim in raw_blob and
		// substitute a placeholder for payload.
		raw.RawBlob = body
		raw.Payload = []byte(`{"_unparseable":true}`)
	}

	ins, err := p.Repo.Insert(ctx, raw)
	if err != nil {
		return ProcessResult{MessageType: msgType}, fmt.Errorf("feed: persist raw_message: %w", err)
	}

	// Record process error after the row exists so it always has an audit trail.
	if decErr != nil || parseErr != nil {
		if err := p.markProcessError(ctx, ins.ID, joinErrors(decErr, parseErr)); err != nil {
			return ProcessResult{MessageType: msgType, RawMessage: ins}, err
		}
		// Don't fan out poison messages — Dispatcher would only see noise.
		return ProcessResult{MessageType: msgType, RawMessage: ins}, nil
	}

	// Fan out only on the first insert; duplicates short-circuit both
	// the internal exchange publish and the in-process dispatcher so the
	// rest of the system observes at-most-once handler invocation.
	if !ins.Inserted {
		return ProcessResult{MessageType: msgType, RawMessage: ins, Dispatched: false}, nil
	}

	sport := int32(0)
	if env.SportID != nil {
		sport = *env.SportID
	}
	if err := p.Pub.Publish(ctx, msgType, sport, env.EventKey(), jsonBytes); err != nil {
		// Publish failures don't roll back the audit row; we surface them
		// so the caller can NACK the source delivery.
		return ProcessResult{MessageType: msgType, RawMessage: ins}, fmt.Errorf("feed: publish: %w", err)
	}

	if err := p.Dispatcher.Dispatch(ctx, msgType, env, ins.ID); err != nil {
		return ProcessResult{MessageType: msgType, RawMessage: ins, Dispatched: false},
			fmt.Errorf("feed: dispatch %s: %w", msgType, err)
	}
	return ProcessResult{MessageType: msgType, RawMessage: ins, Dispatched: true}, nil
}

func (p *Processor) markProcessError(ctx context.Context, id [16]byte, err error) error {
	if err == nil {
		return nil
	}
	if mErr := p.Repo.MarkProcessError(ctx, id, err.Error()); mErr != nil {
		return fmt.Errorf("feed: mark process_error: %w", mErr)
	}
	return nil
}

func joinErrors(errs ...error) error {
	var msgs []string
	for _, e := range errs {
		if e != nil {
			msgs = append(msgs, e.Error())
		}
	}
	if len(msgs) == 0 {
		return nil
	}
	return errors.New(stringsJoin(msgs, "; "))
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}
