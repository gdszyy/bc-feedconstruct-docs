// Package telemetry tracks correlation IDs, audit events and PII-safe logs
// for the Go BFF. Maps to upload-guideline 业务域 "监控治理" (M16).
//
// This file declares the seam shared with Wave-10-C/D/E (BFF route tracks).
// The implementation lives in sibling files owned by Wave-10-B:
//
//   recorder.go         — concrete Recorder backed by zap / slog
//   middleware.go       — net/http middleware injecting correlation_id
//   correlation.go      — context helpers for correlation_id propagation
//
// Wave-10-B will add real BDD placeholders + tests; this file only
// publishes the seam types so other tracks can depend on the interface
// without importing concrete logic.
package telemetry

import "context"

// Recorder is the audit / metric sink. Implementations must be safe for
// concurrent use and must never block the calling goroutine on I/O —
// buffer or drop with a counter.
type Recorder interface {
	// Audit emits a structured business event (bet placed, settlement
	// applied, recovery completed, etc.). PII must be stripped by the
	// caller; the recorder enforces redaction as a defense-in-depth.
	Audit(ctx context.Context, eventType string, fields map[string]any)

	// Error reports a recoverable error with correlation_id from ctx.
	// Use Audit for business events; Error is reserved for failures.
	Error(ctx context.Context, err error, fields map[string]any)
}

// Nop is a Recorder that discards everything. Useful as a default for
// tests and bootstrap before Wave-10-B lands the real implementation.
type Nop struct{}

func (Nop) Audit(context.Context, string, map[string]any) {}
func (Nop) Error(context.Context, error, map[string]any)  {}
