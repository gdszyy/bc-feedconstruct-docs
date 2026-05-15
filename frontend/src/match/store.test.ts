import { describe, expect, it, vi } from "vitest";

import { MatchStore, type MatchRecord, type MatchStoreError } from "./store";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

function baseMatch(overrides: Partial<MatchRecord> = {}): MatchRecord {
  return {
    match_id: "sr:match:1",
    tournament_id: "sr:t:1",
    home_team: "Arsenal",
    away_team: "Chelsea",
    scheduled_at: "2026-05-15T12:00:00Z",
    is_live: false,
    status: "not_started",
    version: 1,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// REST snapshot hydration
// ---------------------------------------------------------------------------

// Given matchStore 为空
// When 通过 hydrateMatches(snapshot) 注入 3 个 match（不同 sport / tournament）
// Then getMatch(match_id) 命中每条，listByTournament(tournament_id) 命中对应集合
describe("given an empty match store", () => {
  it("when a REST snapshot with multiple matches is hydrated then per-id and per-tournament lookups reflect it", () => {
    const store = new MatchStore();
    store.hydrateMatches([
      baseMatch({ match_id: "sr:match:1", tournament_id: "sr:t:1" }),
      baseMatch({ match_id: "sr:match:2", tournament_id: "sr:t:1" }),
      baseMatch({ match_id: "sr:match:3", tournament_id: "sr:t:2" }),
    ]);

    expect(store.getMatch("sr:match:1")?.match_id).toBe("sr:match:1");
    expect(store.getMatch("sr:match:2")?.match_id).toBe("sr:match:2");
    expect(store.getMatch("sr:match:3")?.tournament_id).toBe("sr:t:2");

    expect(
      store.listByTournament("sr:t:1").map((m) => m.match_id).sort(),
    ).toEqual(["sr:match:1", "sr:match:2"]);
    expect(
      store.listByTournament("sr:t:2").map((m) => m.match_id),
    ).toEqual(["sr:match:3"]);
  });
});

// Given matchStore 已经收到 match.upserted（sr:match:1 version=5）
// When 之后 hydrateMatches 再次包含 sr:match:1 的旧字段
// Then 已存在条目不被覆盖（snapshot 不能回退 increment）
describe("given a match already populated by an increment", () => {
  it("when a later snapshot tries to re-hydrate the same id then the existing record is preserved", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: true,
      version: 5,
    });

    store.hydrateMatches([
      baseMatch({
        match_id: "sr:match:1",
        home_team: "STALE-HOME",
        away_team: "STALE-AWAY",
        is_live: false,
        version: 1,
      }),
    ]);

    const m = store.getMatch("sr:match:1")!;
    expect(m.home_team).toBe("Arsenal");
    expect(m.away_team).toBe("Chelsea");
    expect(m.is_live).toBe(true);
    expect(m.version).toBe(5);
  });
});

// ---------------------------------------------------------------------------
// match.upserted — first-seen path
// ---------------------------------------------------------------------------

// Given matchStore 中无 match_id="sr:match:1"
// When 收到 match.upserted { match_id: "sr:match:1", ..., version=1 }
// Then store 写入该 match，status 默认 NotStarted，version=1
describe("given a fresh match store with no entry for sr:match:1", () => {
  it("when match.upserted for sr:match:1 arrives then the match is created with status NotStarted and version 1", () => {
    const store = new MatchStore();
    const accepted = store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });

    expect(accepted).toBe(true);
    const m = store.getMatch("sr:match:1")!;
    expect(m.status).toBe("not_started");
    expect(m.version).toBe(1);
    expect(m.home_team).toBe("Arsenal");
    expect(store.listByTournament("sr:t:1").map((x) => x.match_id)).toEqual([
      "sr:match:1",
    ]);
  });
});

// ---------------------------------------------------------------------------
// match.upserted — version guard
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 的 version=5
// When 收到 match.upserted（同 match_id, version=3）
// Then 该次 upsert 被忽略，store 仍保留 version=5 的字段
describe("given a match store where sr:match:1 is at version 5", () => {
  it("when an older match.upserted at version 3 arrives then the older event is ignored and version stays 5", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: true,
      version: 5,
    });

    const accepted = store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "STALE-HOME",
      away_team: "STALE-AWAY",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 3,
    });

    expect(accepted).toBe(false);
    const m = store.getMatch("sr:match:1")!;
    expect(m.version).toBe(5);
    expect(m.home_team).toBe("Arsenal");
    expect(m.is_live).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// match.status_changed — forward transitions (NotStarted → Live → Ended → Closed)
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 status=NotStarted
// When 依次收到 status_changed: Live(v=2), Ended(v=3), Closed(v=4)
// Then store 的 status 依次推进到 Live、Ended、Closed
describe("given match sr:match:1 in NotStarted", () => {
  it("when status_changed events advance through Live, Ended, Closed in order then status follows each transition", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });

    expect(
      store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: "live",
        version: 2,
      }),
    ).toBe(true);
    expect(store.getMatch("sr:match:1")?.status).toBe("live");

    expect(
      store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: "ended",
        version: 3,
      }),
    ).toBe(true);
    expect(store.getMatch("sr:match:1")?.status).toBe("ended");

    expect(
      store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: "closed",
        version: 4,
      }),
    ).toBe(true);
    expect(store.getMatch("sr:match:1")?.status).toBe("closed");
  });
});

// ---------------------------------------------------------------------------
// match.status_changed — anti-regression
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 status=Ended (version=3)
// When 收到 status_changed { status: Live, version=4 }
// Then 该事件被拒绝（不可降级），status 仍为 Ended
describe("given match sr:match:1 already in Ended", () => {
  it("when a later status_changed tries to downgrade to Live then the regression is rejected and status stays Ended", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "ended",
      version: 3,
    });

    const accepted = store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "live",
      version: 4,
    });
    expect(accepted).toBe(false);
    expect(store.getMatch("sr:match:1")?.status).toBe("ended");
    expect(store.getMatch("sr:match:1")?.version).toBe(3);
  });
});

// Given matchStore 中 sr:match:1 status=Live
// When 收到 status_changed { status: NotStarted, version=higher }
// Then 该事件被拒绝（Live 不可回退到 NotStarted）
describe("given match sr:match:1 already in Live", () => {
  it("when a later status_changed tries to downgrade to NotStarted then the regression is rejected and status stays Live", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: true,
      version: 1,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "live",
      version: 2,
    });

    const accepted = store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "not_started",
      version: 10,
    });
    expect(accepted).toBe(false);
    expect(store.getMatch("sr:match:1")?.status).toBe("live");
    expect(store.getMatch("sr:match:1")?.version).toBe(2);
  });
});

// ---------------------------------------------------------------------------
// match.status_changed — absorbing terminal states
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 status=Cancelled
// When 收到任意 status_changed（Live / Ended / Closed / Abandoned）
// Then 状态保持 Cancelled（终态吸收）
describe("given match sr:match:1 already in Cancelled", () => {
  it("when any subsequent status_changed arrives then the match remains Cancelled", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "cancelled",
      version: 2,
    });

    for (const next of ["live", "ended", "closed", "abandoned"] as const) {
      const accepted = store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: next,
        version: 99,
      });
      expect(accepted).toBe(false);
      expect(store.getMatch("sr:match:1")?.status).toBe("cancelled");
    }
  });
});

// Given matchStore 中 sr:match:1 status=Closed
// When 收到 status_changed { status: Ended | Live | NotStarted }
// Then 状态保持 Closed
describe("given match sr:match:1 already in Closed", () => {
  it("when status_changed attempts to move back from Closed to Ended/Live/NotStarted then status stays Closed", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "closed",
      version: 4,
    });

    for (const next of ["ended", "live", "not_started"] as const) {
      const accepted = store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: next,
        version: 50,
      });
      expect(accepted).toBe(false);
      expect(store.getMatch("sr:match:1")?.status).toBe("closed");
    }
  });
});

// ---------------------------------------------------------------------------
// match.status_changed — score & period propagation
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 status=Live, score=0:0
// When 收到 status_changed { status: Live, home_score: 1, away_score: 0, period: "1H", version=higher }
// Then 比分与阶段被更新，但 status 仍为 Live（同态更新允许携带 score）
describe("given match sr:match:1 in Live with score 0:0", () => {
  it("when status_changed brings updated score and period then score/period are applied without status regression", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: true,
      version: 1,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "live",
      home_score: 0,
      away_score: 0,
      version: 2,
    });

    const accepted = store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "live",
      home_score: 1,
      away_score: 0,
      period: "1H",
      version: 3,
    });

    expect(accepted).toBe(true);
    const m = store.getMatch("sr:match:1")!;
    expect(m.status).toBe("live");
    expect(m.home_score).toBe(1);
    expect(m.away_score).toBe(0);
    expect(m.period).toBe("1H");
    expect(m.version).toBe(3);
  });
});

// ---------------------------------------------------------------------------
// fixture.changed — REST refetch + replace
// ---------------------------------------------------------------------------

// Given matchStore 中 sr:match:1 的字段为旧值，且 fixtureRefetcher 已配置
// When 收到 fixture.changed { match_id: "sr:match:1", refetch_required: true }
// Then fixtureChangeHandler 调用 fixtureRefetcher("sr:match:1") 并以返回值整体替换主数据
describe("given match sr:match:1 with stale snapshot fields and a fixture refetcher", () => {
  it("when fixture.changed arrives then the handler refetches the match via REST and replaces the stored entry", async () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });

    const fresh: MatchRecord = {
      match_id: "sr:match:1",
      tournament_id: "sr:t:9",
      home_team: "Arsenal-Fresh",
      away_team: "Chelsea-Fresh",
      scheduled_at: "2026-05-15T14:30:00Z",
      is_live: true,
      status: "live",
      home_score: 2,
      away_score: 1,
      period: "2H",
      version: 42,
    };
    const refetcher = vi.fn(async (_id: string) => fresh);
    store.setFixtureRefetcher(refetcher);

    const accepted = await store.applyFixtureChanged({
      match_id: "sr:match:1",
      change_type: "schedule",
      refetch_required: true,
    });

    expect(accepted).toBe(true);
    expect(refetcher).toHaveBeenCalledWith("sr:match:1");
    const m = store.getMatch("sr:match:1")!;
    expect(m).toEqual(fresh);
    expect(
      store.listByTournament("sr:t:9").map((x) => x.match_id),
    ).toEqual(["sr:match:1"]);
    // Re-indexed off the old tournament bucket.
    expect(store.listByTournament("sr:t:1")).toEqual([]);
  });
});

// Given fixtureRefetcher 抛错或返回 null
// When 收到 fixture.changed
// Then 旧主数据保留不变，且错误被上报到 errorSink（不污染 store）
describe("given a fixture refetcher that fails", () => {
  it("when fixture.changed arrives and the refetch fails then the existing match data is preserved and the error is surfaced", async () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });
    const before = store.getMatch("sr:match:1");

    const boom = new Error("network down");
    const refetcher = vi.fn(async (_id: string): Promise<MatchRecord> => {
      throw boom;
    });
    const errors: MatchStoreError[] = [];
    store.setFixtureRefetcher(refetcher);
    store.setErrorSink((e) => errors.push(e));

    const accepted = await store.applyFixtureChanged({
      match_id: "sr:match:1",
      change_type: "other",
      refetch_required: true,
    });
    expect(accepted).toBe(false);
    expect(store.getMatch("sr:match:1")).toEqual(before);
    expect(errors).toHaveLength(1);
    expect(errors[0].kind).toBe("fixture_refetch_failed");
    expect(errors[0].match_id).toBe("sr:match:1");
    expect(errors[0].cause).toBe(boom);

    // Now repeat with a refetcher that returns null.
    store.setFixtureRefetcher(async () => null);
    const accepted2 = await store.applyFixtureChanged({
      match_id: "sr:match:1",
      change_type: "other",
      refetch_required: true,
    });
    expect(accepted2).toBe(false);
    expect(store.getMatch("sr:match:1")).toEqual(before);
    expect(errors).toHaveLength(2);
    expect(errors[1].kind).toBe("fixture_refetch_empty");
  });
});

// ---------------------------------------------------------------------------
// Subscription / notify
// ---------------------------------------------------------------------------

// Given 已注册 subscribe(listener)
// When 任一 reducer 实际改变了 store
// Then listener 被回调一次
describe("given a registered store listener", () => {
  it("when a reducer mutation actually changes the store then the listener is notified once", () => {
    const store = new MatchStore();
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: false,
      version: 1,
    });
    expect(listener).toHaveBeenCalledTimes(1);

    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "live",
      version: 2,
    });
    expect(listener).toHaveBeenCalledTimes(2);
  });
});

// Given 已注册 subscribe(listener)
// When 收到的事件因 version guard / anti-regression 被丢弃
// Then listener 不被回调
describe("given a registered store listener observing dropped events", () => {
  it("when an event is dropped by version-guard or anti-regression then no listener notification fires", () => {
    const store = new MatchStore();
    store.applyMatchUpserted({
      match_id: "sr:match:1",
      tournament_id: "sr:t:1",
      home_team: "Arsenal",
      away_team: "Chelsea",
      scheduled_at: "2026-05-15T12:00:00Z",
      is_live: true,
      version: 5,
    });
    store.applyMatchStatusChanged({
      match_id: "sr:match:1",
      status: "ended",
      version: 6,
    });

    const listener = vi.fn();
    store.subscribe(listener);

    // Older version of upserted is dropped.
    expect(
      store.applyMatchUpserted({
        match_id: "sr:match:1",
        tournament_id: "sr:t:1",
        home_team: "STALE",
        away_team: "STALE",
        scheduled_at: "2026-05-15T12:00:00Z",
        is_live: false,
        version: 3,
      }),
    ).toBe(false);

    // Anti-regression drops ended → live.
    expect(
      store.applyMatchStatusChanged({
        match_id: "sr:match:1",
        status: "live",
        version: 7,
      }),
    ).toBe(false);

    // status_changed for an unknown match is also dropped silently.
    expect(
      store.applyMatchStatusChanged({
        match_id: "sr:match:unknown",
        status: "live",
        version: 1,
      }),
    ).toBe(false);

    expect(listener).not.toHaveBeenCalled();
  });
});
