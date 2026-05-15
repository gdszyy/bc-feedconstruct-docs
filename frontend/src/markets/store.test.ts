import { describe, expect, it, vi } from "vitest";

import {
  MarketsStore,
  selectDisplayOdds,
  selectFrozen,
  type MarketRecord,
} from "./store";

function baseMarket(overrides: Partial<MarketRecord> = {}): MarketRecord {
  return {
    match_id: "sr:match:1",
    market_id: "sr:market:1",
    specifiers: {},
    status: "active",
    outcomes: [
      { outcome_id: "home", odds: 1.8, active: true },
      { outcome_id: "away", odds: 1.9, active: true },
    ],
    version: 1,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// REST snapshot hydration
// ---------------------------------------------------------------------------

// Given marketsStore 为空
// When 通过 hydrateMatchMarkets("sr:match:1", [2 个 market, 各含 2 个 outcome]) 注入
// Then getMarket / listMarkets("sr:match:1") 反映该快照，outcome 与 odds 字段一致
describe("given an empty markets store", () => {
  it("when a REST snapshot for sr:match:1 with two markets is hydrated then list/get queries return the snapshot data", () => {
    const store = new MarketsStore();
    store.hydrateMatchMarkets("sr:match:1", [
      baseMarket({ market_id: "sr:market:1", version: 7 }),
      baseMarket({
        market_id: "sr:market:2",
        specifiers: { total: "2.5" },
        outcomes: [
          { outcome_id: "over", odds: 1.95, active: true },
          { outcome_id: "under", odds: 1.85, active: true },
        ],
        version: 4,
      }),
    ]);

    const m1 = store.getMarket("sr:match:1", "sr:market:1");
    expect(m1?.version).toBe(7);
    expect(m1?.outcomes.map((o) => o.outcome_id)).toEqual(["home", "away"]);

    const m2 = store.getMarket("sr:match:1", "sr:market:2");
    expect(m2?.specifiers).toEqual({ total: "2.5" });
    expect(m2?.outcomes.find((o) => o.outcome_id === "over")?.odds).toBe(1.95);

    expect(
      store.listMarkets("sr:match:1").map((m) => m.market_id).sort(),
    ).toEqual(["sr:market:1", "sr:market:2"]);
  });
});

// Given marketsStore 中 sr:match:1 / sr:market:1 已被 odds.changed 写入（version=5）
// When 之后 hydrateMatchMarkets 再次包含同一 market 的旧字段
// Then 已存在条目不被覆盖（snapshot 不能回退 increment，M10 约束）
describe("given a market already populated by an odds increment", () => {
  it("when a later snapshot tries to re-hydrate the same market then the existing record is preserved", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      specifiers: { total: "2.5" },
      outcomes: [
        { outcome_id: "over", odds: 1.92, active: true },
        { outcome_id: "under", odds: 1.88, active: true },
      ],
      version: 5,
    });

    store.hydrateMatchMarkets("sr:match:1", [
      baseMarket({
        market_id: "sr:market:1",
        version: 1,
        outcomes: [
          { outcome_id: "over", odds: 9.99, active: false },
          { outcome_id: "under", odds: 9.99, active: false },
        ],
      }),
    ]);

    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.version).toBe(5);
    expect(m.specifiers).toEqual({ total: "2.5" });
    expect(m.outcomes.find((o) => o.outcome_id === "over")?.odds).toBe(1.92);
    expect(m.outcomes.find((o) => o.outcome_id === "over")?.active).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// odds.changed — first-seen path
// ---------------------------------------------------------------------------

// Given marketsStore 中无 (sr:match:1, sr:market:1)
// When 收到 odds.changed { match_id, market_id, specifiers, outcomes=[...], version=1 }
// Then 该 market 被创建，status 默认 "active"，outcomes / specifiers 与 payload 一致
describe("given a fresh markets store with no entry for sr:match:1/sr:market:1", () => {
  it("when odds.changed creates the market for the first time then status defaults to active and outcomes match the payload", () => {
    const store = new MarketsStore();
    const accepted = store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      specifiers: { hcp: "+0.5" },
      outcomes: [
        { outcome_id: "home", odds: 1.7, active: true },
        { outcome_id: "away", odds: 2.1, active: true },
      ],
      version: 1,
    });

    expect(accepted).toBe(true);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.status).toBe("active");
    expect(m.specifiers).toEqual({ hcp: "+0.5" });
    expect(m.outcomes).toEqual([
      { outcome_id: "home", odds: 1.7, active: true },
      { outcome_id: "away", odds: 2.1, active: true },
    ]);
    expect(m.version).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// odds.changed — version guard
// ---------------------------------------------------------------------------

// Given marketsStore 中 (sr:match:1, sr:market:1) 的 version=5
// When 收到 odds.changed（同 market, version=3, 新 odds）
// Then 该次增量被忽略，odds 与 version 保持 v=5 时的状态
describe("given a markets store where sr:match:1/sr:market:1 is at version 5", () => {
  it("when an older odds.changed at version 3 arrives then the older event is ignored and odds/version remain at 5", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.8, active: true }],
      version: 5,
    });

    const accepted = store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 9.99, active: false }],
      version: 3,
    });

    expect(accepted).toBe(false);
    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.version).toBe(5);
    expect(m.outcomes[0].odds).toBe(1.8);
    expect(m.outcomes[0].active).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// odds.changed — outcome merging
// ---------------------------------------------------------------------------

// Given (sr:match:1, sr:market:1) 已存在 [home@1.80, away@1.90]，specifiers={total: "2.5"}
// When 收到 odds.changed { outcomes=[home@1.75], version=higher }
// Then home 的 odds 被更新为 1.75，away 保留 1.90 不变，specifiers 不变
describe("given a market with [home@1.80, away@1.90] and specifiers total=2.5", () => {
  it("when odds.changed brings only the home outcome then only that outcome is updated and other fields are preserved", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      specifiers: { total: "2.5" },
      outcomes: [
        { outcome_id: "home", odds: 1.8, active: true },
        { outcome_id: "away", odds: 1.9, active: true },
      ],
      version: 1,
    });

    const accepted = store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.75, active: true }],
      version: 2,
    });
    expect(accepted).toBe(true);

    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.specifiers).toEqual({ total: "2.5" });
    expect(m.outcomes.find((o) => o.outcome_id === "home")?.odds).toBe(1.75);
    expect(m.outcomes.find((o) => o.outcome_id === "away")?.odds).toBe(1.9);
    expect(m.version).toBe(2);
  });
});

// Given (sr:match:1, sr:market:1) 已存在 [home, away]
// When 收到 odds.changed { outcomes=[home@1.75, away@2.00, draw@3.20] }
// Then draw 被新增到 outcomes，order 与 payload 顺序保持一致
describe("given a market with [home, away]", () => {
  it("when odds.changed brings an additional draw outcome then the new outcome is appended and ordering reflects the payload", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", odds: 1.8, active: true },
        { outcome_id: "away", odds: 1.9, active: true },
      ],
      version: 1,
    });

    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", odds: 1.75, active: true },
        { outcome_id: "away", odds: 2.0, active: true },
        { outcome_id: "draw", odds: 3.2, active: true },
      ],
      version: 2,
    });

    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.outcomes.map((o) => o.outcome_id)).toEqual([
      "home",
      "away",
      "draw",
    ]);
    expect(m.outcomes.find((o) => o.outcome_id === "draw")?.odds).toBe(3.2);
  });
});

// Given (sr:match:1, sr:market:1) 已存在 [home@1.80, away@1.90]
// When 收到 odds.changed { outcomes=[home@1.75, away@1.85] }，active 字段切换 true → false
// Then outcome.active 切换并被存储，selector 可识别失活
describe("given outcomes that are active", () => {
  it("when odds.changed flips active to false then the stored outcomes reflect the new active flags", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", odds: 1.8, active: true },
        { outcome_id: "away", odds: 1.9, active: true },
      ],
      version: 1,
    });

    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [
        { outcome_id: "home", odds: 1.75, active: false },
        { outcome_id: "away", odds: 1.85, active: false },
      ],
      version: 2,
    });

    const m = store.getMarket("sr:match:1", "sr:market:1")!;
    expect(m.outcomes.every((o) => o.active === false)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// Multi-match isolation
// ---------------------------------------------------------------------------

// Given marketsStore 已写入 (sr:match:1, sr:market:1)
// When 收到针对 (sr:match:2, sr:market:1) 的 odds.changed
// Then sr:match:2 下被创建独立 market，sr:match:1 下的同名 market_id 不受影响
describe("given two matches that reuse the same market_id namespace", () => {
  it("when odds.changed arrives for sr:match:2 then only sr:match:2's bucket is mutated and sr:match:1 stays intact", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.8, active: true }],
      version: 5,
    });

    store.applyOddsChanged({
      match_id: "sr:match:2",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 2.5, active: true }],
      version: 1,
    });

    const m1 = store.getMarket("sr:match:1", "sr:market:1")!;
    const m2 = store.getMarket("sr:match:2", "sr:market:1")!;
    expect(m1.outcomes[0].odds).toBe(1.8);
    expect(m1.version).toBe(5);
    expect(m2.outcomes[0].odds).toBe(2.5);
    expect(m2.version).toBe(1);
    expect(store.listMarkets("sr:match:1").length).toBe(1);
    expect(store.listMarkets("sr:match:2").length).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// Listener / subscribe
// ---------------------------------------------------------------------------

// Given 已注册 subscribe(listener)
// When odds.changed 实际更新了 outcomes
// Then listener 被回调一次
describe("given a registered store listener", () => {
  it("when odds.changed actually mutates the store then the listener is notified once", () => {
    const store = new MarketsStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.8, active: true }],
      version: 1,
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.75, active: true }],
      version: 2,
    });
    expect(listener).toHaveBeenCalledTimes(2);
  });
});

// Given 已注册 subscribe(listener)
// When 收到的 odds.changed 因 version guard 被丢弃
// Then listener 不被回调
describe("given a registered store listener observing dropped events", () => {
  it("when an odds.changed event is dropped by version-guard then no listener notification fires", () => {
    const store = new MarketsStore();
    store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 1.8, active: true }],
      version: 5,
    });

    const listener = vi.fn();
    store.subscribe(listener);

    const accepted = store.applyOddsChanged({
      match_id: "sr:match:1",
      market_id: "sr:market:1",
      outcomes: [{ outcome_id: "home", odds: 9.99, active: false }],
      version: 3,
    });
    expect(accepted).toBe(false);
    expect(listener).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// displayOdds / frozen selectors (status-only dimension)
// ---------------------------------------------------------------------------

// Given (sr:match:1, sr:market:1) status=active
// When 调用 selectDisplayOdds(market)
// Then 返回当前 outcomes 的 odds 快照
describe("given an active market", () => {
  it("when selectDisplayOdds is called then it returns the current outcome odds snapshot", () => {
    const m = baseMarket({
      status: "active",
      outcomes: [
        { outcome_id: "home", odds: 1.8, active: true },
        { outcome_id: "away", odds: 1.9, active: true },
      ],
    });
    const odds = selectDisplayOdds(m);
    expect(odds).toEqual([
      { outcome_id: "home", odds: 1.8, active: true },
      { outcome_id: "away", odds: 1.9, active: true },
    ]);
  });
});

// Given (sr:match:1, sr:market:1) status=deactivated
// When 调用 selectDisplayOdds(market)
// Then 返回 null（仅 Active/Suspended 暴露赔率）
describe("given a deactivated market", () => {
  it("when selectDisplayOdds is called then it returns null because only Active/Suspended expose odds", () => {
    const m = baseMarket({ status: "deactivated" });
    expect(selectDisplayOdds(m)).toBeNull();

    const settled = baseMarket({ status: "settled" });
    expect(selectDisplayOdds(settled)).toBeNull();

    const cancelled = baseMarket({ status: "cancelled" });
    expect(selectDisplayOdds(cancelled)).toBeNull();

    const suspended = baseMarket({ status: "suspended" });
    expect(selectDisplayOdds(suspended)).not.toBeNull();
  });
});

// Given (sr:match:1, sr:market:1) status=suspended
// When 调用 selectFrozen(market)
// Then 返回 true（Suspended 自身即视为 frozen，不含 bet_stop 维度）
describe("given a suspended market", () => {
  it("when selectFrozen is called then it returns true because Suspended is considered frozen on its own", () => {
    expect(selectFrozen(baseMarket({ status: "suspended" }))).toBe(true);
    expect(selectFrozen(baseMarket({ status: "active" }))).toBe(false);
    expect(selectFrozen(baseMarket({ status: "deactivated" }))).toBe(false);
  });
});
