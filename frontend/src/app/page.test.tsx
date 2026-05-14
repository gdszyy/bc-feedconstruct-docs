import { describe, it } from "vitest";

// 页面 P01 — 首页 / 大厅
//
// Given the BFF returns 3 sports and 5 live matches via REST snapshot
// When the home page is rendered
// Then it lists every sport name and shows a live match count badge of 5
describe("given a populated BFF snapshot", () => {
  it("when the home page renders then sports list and live count appear", () => {
    // BDD placeholder
  });
});

// Given the BFF /readyz returns 503
// When the home page renders
// Then a degradation banner is shown and live data is not requested
describe("given BFF not ready", () => {
  it("when home renders then degradation banner shown and no realtime subscribe issued", () => {
    // BDD placeholder
  });
});
