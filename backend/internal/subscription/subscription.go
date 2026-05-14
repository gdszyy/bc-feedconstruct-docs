// Package subscription owns the booking lifecycle of matches:
//   requested -> subscribed -> unsubscribed (or expired / failed)
// and the subscription_events audit trail.
//
// Maps to acceptance #13. The package exposes:
//   - Handler implementing feed.Handler for MsgSubscriptionBook / Unbook
//   - Manager.OnMatchTerminal as a catalog.MatchObserver implementation
//     that auto-unbooks when a match transitions to ended/closed/cancelled
//   - Manager.CleanupStuck which marks long-pending requested rows as
//     failed (background tick from cmd/bffd)
package subscription
