// Package odds processes odds_change and bet_stop deliveries, maintaining
// the markets, outcomes and market_status_history tables.
//
// Maps to acceptance #5 (赔率), #6 (停投) and #12 market-level (防回退):
// settled/cancelled markets must not be re-activated by a late odds_change.
package odds
