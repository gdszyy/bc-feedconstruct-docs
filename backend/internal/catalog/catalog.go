// Package catalog handles FeedConstruct catalog objects:
// Sport (ObjectType=1) / Region (ObjectType=2) / Competition (ObjectType=3)
// / Match (ObjectType=4) / fixture_change.
//
// The Handler implements feed.Handler and is registered with the
// dispatcher for the matching MessageType values. Upserts are idempotent
// (ON CONFLICT) and match.status transitions enforce the no-regression
// rule from acceptance #12.
package catalog
