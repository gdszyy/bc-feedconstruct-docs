import { describe, it } from "vitest";

// 页面 P03 — 赛事详情（消费 M04/M05/M06/M07/M12）
//
// Given match=42 has 2 markets and 4 outcomes returned by GET /api/v1/matches/42
// When the match-detail page renders for id=42
// Then markets are grouped by market_type and outcomes show formatted odds
describe("given match snapshot for id=42", () => {
  it("when page renders then markets grouped and odds formatted", () => {
    // BDD placeholder
  });
});

// Given the WebSocket pushes an odds_update for match=42 after initial render
// When the new payload arrives
// Then only the affected outcome cell re-renders (no full-page rerender)
describe("given subscribed match-detail page", () => {
  it("when ws odds_update arrives then only affected outcome rerenders", () => {
    // BDD placeholder
  });
});

// Given a bet_stop frame arrives for the whole match
// When the page handles it
// Then every outcome cell becomes non-clickable and shows a "suspended" overlay
describe("given subscribed match-detail page", () => {
  it("when bet_stop arrives then outcomes become non-clickable with suspended overlay", () => {
    // BDD placeholder
  });
});
