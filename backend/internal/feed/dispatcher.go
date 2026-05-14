package feed

import (
	"context"
	"fmt"
	"sync"
)

// Handler reacts to a classified, persisted delivery. RawMessageID lets a
// handler look up the original payload from raw_messages without re-parsing.
type Handler interface {
	Handle(ctx context.Context, msgType MessageType, env Envelope, rawMessageID [16]byte) error
}

// HandlerFunc adapts a plain function into a Handler.
type HandlerFunc func(ctx context.Context, msgType MessageType, env Envelope, rawMessageID [16]byte) error

func (f HandlerFunc) Handle(ctx context.Context, t MessageType, env Envelope, id [16]byte) error {
	return f(ctx, t, env, id)
}

// Dispatcher maps message types to handlers and routes deliveries.
// At wave-2 every concrete handler is a Nop registered by cmd/bffd; later
// waves replace them with the real catalog/odds/settlement handlers.
type Dispatcher struct {
	mu            sync.RWMutex
	handlers      map[MessageType]Handler
	deadLetter    Handler
	unknownCounts map[MessageType]int64
}

// NewDispatcher returns an empty Dispatcher. deadLetter is invoked for
// MsgUnknown (or any unregistered MessageType) and must not be nil.
func NewDispatcher(deadLetter Handler) *Dispatcher {
	if deadLetter == nil {
		deadLetter = HandlerFunc(func(context.Context, MessageType, Envelope, [16]byte) error { return nil })
	}
	return &Dispatcher{
		handlers:      make(map[MessageType]Handler),
		deadLetter:    deadLetter,
		unknownCounts: make(map[MessageType]int64),
	}
}

// Register binds a handler to a message type. Re-registering overwrites.
func (d *Dispatcher) Register(t MessageType, h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[t] = h
}

// Dispatch invokes the handler bound to msgType or the dead-letter handler.
func (d *Dispatcher) Dispatch(ctx context.Context, msgType MessageType, env Envelope, rawID [16]byte) error {
	d.mu.RLock()
	h, ok := d.handlers[msgType]
	d.mu.RUnlock()
	if !ok {
		d.mu.Lock()
		d.unknownCounts[msgType]++
		d.mu.Unlock()
		if err := d.deadLetter.Handle(ctx, msgType, env, rawID); err != nil {
			return fmt.Errorf("dead-letter handler for %q: %w", msgType, err)
		}
		return nil
	}
	return h.Handle(ctx, msgType, env, rawID)
}

// Registered returns the set of currently registered MessageTypes.
func (d *Dispatcher) Registered() []MessageType {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]MessageType, 0, len(d.handlers))
	for k := range d.handlers {
		out = append(out, k)
	}
	return out
}

// UnknownCount returns how many times an unregistered type was routed to
// the dead-letter handler. Useful for the M15 metrics surface.
func (d *Dispatcher) UnknownCount(t MessageType) int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.unknownCounts[t]
}
