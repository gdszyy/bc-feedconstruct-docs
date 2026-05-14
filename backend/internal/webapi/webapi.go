// Package webapi is the FeedConstruct WebAPI client.
//
// It owns Token acquisition (24h cache + singleflight refresh), DataSnapshot
// retrieval and GetMatchByID lookups. Descriptions (Sport / Region /
// MarketTypes / SelectionTypes / EventTypes / Periods) are added in a later
// wave; the current iteration covers what acceptance #1 and #10 need.
//
// FeedConstruct's wire protocol is JSON-over-HTTPS with GZIP responses.
// The exact endpoint paths and request envelope are configurable through
// MethodPaths so the client can be tuned against the real environment
// without touching call sites.
package webapi
