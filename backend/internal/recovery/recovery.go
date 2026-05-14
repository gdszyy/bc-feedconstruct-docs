// Package recovery coordinates FeedConstruct DataSnapshot calls across
// five scopes (startup, product, event, stateful, fixture_change) and
// drives the recovery_jobs table to completion with 429-aware backoff.
//
// Map to upload-guideline acceptance #10 "恢复".
package recovery
