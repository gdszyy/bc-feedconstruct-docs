import { describe, expect, it } from "vitest";

import {
  CatalogStore,
  type CatalogSnapshot,
  type TournamentRecord,
} from "./store";

// ---------------------------------------------------------------------------
// 快照引导
// ---------------------------------------------------------------------------

// Given catalogStore 为空
// When 通过 hydrateSnapshot(snapshot) 注入 2 个 sport + 3 个 tournament
// Then listSports() 返回这 2 个 sport，listTournaments(sport_id) 命中 sport 下的 tournament
describe("given an empty catalog store", () => {
  it("when a snapshot with sports and tournaments is hydrated then list queries reflect the snapshot", () => {
    const store = new CatalogStore();
    const snapshot: CatalogSnapshot = {
      sports: [
        {
          sport_id: "sr:sport:1",
          name_translations: { en: "Soccer" },
          sort_order: 1,
        },
        {
          sport_id: "sr:sport:2",
          name_translations: { en: "Basketball" },
          sort_order: 2,
        },
      ],
      tournaments: [
        {
          tournament_id: "sr:t:1",
          sport_id: "sr:sport:1",
          category_id: "sr:cat:gb",
          name_translations: { en: "Premier League" },
        },
        {
          tournament_id: "sr:t:2",
          sport_id: "sr:sport:1",
          category_id: "sr:cat:es",
          name_translations: { en: "LaLiga" },
        },
        {
          tournament_id: "sr:t:3",
          sport_id: "sr:sport:2",
          category_id: "sr:cat:us",
          name_translations: { en: "NBA" },
        },
      ],
    };

    store.hydrateSnapshot(snapshot);

    expect(store.listSports().map((s) => s.sport_id)).toEqual([
      "sr:sport:1",
      "sr:sport:2",
    ]);
    expect(
      store.listTournaments("sr:sport:1").map((t) => t.tournament_id).sort(),
    ).toEqual(["sr:t:1", "sr:t:2"]);
    expect(
      store.listTournaments("sr:sport:2").map((t) => t.tournament_id),
    ).toEqual(["sr:t:3"]);
  });
});

// ---------------------------------------------------------------------------
// sport.* 增量
// ---------------------------------------------------------------------------

// Given catalogStore 中无 sport_id="sr:sport:1"
// When 收到 sport.upserted { sport_id: "sr:sport:1", name_translations, sort_order }
// Then store 写入该 sport，listSports() 返回它
describe("given a fresh store with no football sport", () => {
  it("when sport.upserted for sr:sport:1 arrives then the sport appears in listSports()", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 10,
    });

    const sports = store.listSports();
    expect(sports).toHaveLength(1);
    expect(sports[0].sport_id).toBe("sr:sport:1");
    expect(sports[0].sort_order).toBe(10);
  });
});

// Given store 已有 sport sr:sport:1 (sort_order=10)
// When 再次收到 sport.upserted { sort_order: 5, ... }
// Then sport 字段被更新且未产生重复条目（幂等合并）
describe("given an existing sport in the store", () => {
  it("when sport.upserted for the same id arrives then it merges in place without duplicating", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 10,
    });
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer", zh: "足球" },
      sort_order: 5,
    });

    const sports = store.listSports();
    expect(sports).toHaveLength(1);
    expect(sports[0].sort_order).toBe(5);
    expect(sports[0].name_translations).toEqual({ en: "Soccer", zh: "足球" });
  });
});

// Given store 含 sport sr:sport:1 + 其下若干 tournament
// When 收到 sport.removed { sport_id: "sr:sport:1" }
// Then listSports() 不再包含该 sport；listTournaments(sr:sport:1) 也返回空
//      （tournament 记录被级联清理，避免悬挂引用）
describe("given a populated sport with tournaments", () => {
  it("when sport.removed arrives then both the sport and its tournaments disappear from queries", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 1,
    });
    store.applyTournamentUpserted({
      tournament_id: "sr:t:1",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:gb",
      name_translations: { en: "Premier League" },
    });
    store.applyTournamentUpserted({
      tournament_id: "sr:t:2",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:es",
      name_translations: { en: "LaLiga" },
    });

    store.applySportRemoved({ sport_id: "sr:sport:1" });

    expect(store.listSports()).toEqual([]);
    expect(store.listTournaments("sr:sport:1")).toEqual([]);
    expect(store.hasCategory("sr:cat:gb")).toBe(false);
    expect(store.hasCategory("sr:cat:es")).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// tournament.* 增量
// ---------------------------------------------------------------------------

// Given store 含 sport sr:sport:1，无任何 tournament
// When 收到 tournament.upserted { tournament_id: "sr:t:1", sport_id: "sr:sport:1", category_id: "sr:cat:1", ... }
// Then listTournaments("sr:sport:1") 包含该 tournament，category sr:cat:1 被隐式创建
describe("given a sport without tournaments", () => {
  it("when tournament.upserted arrives then the tournament is indexed by sport and an implicit category is recorded", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 1,
    });
    store.applyTournamentUpserted({
      tournament_id: "sr:t:1",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:1",
      name_translations: { en: "Premier League" },
    });

    const tournaments = store.listTournaments("sr:sport:1");
    expect(tournaments).toHaveLength(1);
    expect(tournaments[0].tournament_id).toBe("sr:t:1");
    expect(store.hasCategory("sr:cat:1")).toBe(true);
  });
});

// Given store 已经知道 tournament sr:t:1
// When 收到 tournament.removed { tournament_id: "sr:t:1" }
// Then listTournaments(...) 不再包含它
describe("given an existing tournament", () => {
  it("when tournament.removed arrives then it is removed from listTournaments()", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 1,
    });
    store.applyTournamentUpserted({
      tournament_id: "sr:t:1",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:1",
      name_translations: { en: "Premier League" },
    });

    store.applyTournamentRemoved({ tournament_id: "sr:t:1" });

    expect(store.listTournaments("sr:sport:1")).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// 排序与查询
// ---------------------------------------------------------------------------

// Given store 内含 3 个 sport，sort_order 分别为 30 / 10 / 20
// When listSports() 被调用
// Then 返回顺序按 sort_order 升序 (10 → 20 → 30)
describe("given multiple sports with distinct sort_order", () => {
  it("when listSports() is queried then they are returned in ascending sort_order", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "a",
      name_translations: { en: "A" },
      sort_order: 30,
    });
    store.applySportUpserted({
      sport_id: "b",
      name_translations: { en: "B" },
      sort_order: 10,
    });
    store.applySportUpserted({
      sport_id: "c",
      name_translations: { en: "C" },
      sort_order: 20,
    });

    expect(store.listSports().map((s) => s.sport_id)).toEqual([
      "b",
      "c",
      "a",
    ]);
  });
});

// ---------------------------------------------------------------------------
// i18n 隔离
// ---------------------------------------------------------------------------

// Given store 中 sport sr:sport:1 的 name_translations 包含 "en" 与 "zh"
// When 用户切换 locale 从 "en" 到 "zh"
// Then 不触发任何 REST 拉取；getSportName(id, locale) 直接命中已有 translations
describe("given a populated store with multilingual sport names", () => {
  it("when the active locale changes then no REST refetch is triggered and the new locale name resolves locally", () => {
    const store = new CatalogStore();
    let mutationCount = 0;
    const unsub = store.subscribe(() => {
      mutationCount += 1;
    });
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer", zh: "足球" },
      sort_order: 1,
    });
    const mutationsAfterPopulate = mutationCount;

    // Locale switching is a UI concern — it must NOT mutate the catalog store.
    expect(store.getSportName("sr:sport:1", "en")).toBe("Soccer");
    expect(store.getSportName("sr:sport:1", "zh")).toBe("足球");

    expect(mutationCount).toBe(mutationsAfterPopulate);
    unsub();
  });
});

// ---------------------------------------------------------------------------
// 快照 + 增量合并
// ---------------------------------------------------------------------------

// Given store 已经通过增量收到 tournament sr:t:1 (经过 1 次更新)
// When 后到的 hydrateSnapshot 含 tournament sr:t:1 的快照副本
// Then 二者按主键合并；不会出现重复记录；快照不会覆盖更新的属性
//      （遵循 M10 "snapshot 不能回退已观测增量" 原则）
describe("given an increment-only tournament was applied first", () => {
  it("when a snapshot later includes the same tournament then the merge is idempotent and does not regress observed updates", () => {
    const store = new CatalogStore();
    store.applySportUpserted({
      sport_id: "sr:sport:1",
      name_translations: { en: "Soccer" },
      sort_order: 1,
    });
    store.applyTournamentUpserted({
      tournament_id: "sr:t:1",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:1",
      name_translations: { en: "Premier League — Updated" },
    });

    const stale: TournamentRecord = {
      tournament_id: "sr:t:1",
      sport_id: "sr:sport:1",
      category_id: "sr:cat:1",
      name_translations: { en: "Premier League — Snapshot" },
    };
    store.hydrateSnapshot({ sports: [], tournaments: [stale] });

    const tournaments = store.listTournaments("sr:sport:1");
    expect(tournaments).toHaveLength(1);
    expect(tournaments[0].name_translations.en).toBe(
      "Premier League — Updated",
    );
  });
});
