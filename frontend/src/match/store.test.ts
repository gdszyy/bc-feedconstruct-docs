import { describe, expect, it, vi } from "vitest";

import {
  MatchStore,
  createFixtureChangeHandler,
  type MatchRecord,
} from "./store";

// 模块 M04 — 赛事与赛程（Match / Fixture / Schedule）
// 参考：docs/07_frontend_architecture/modules/M04_match_and_fixture.md
// 状态机：docs/07_frontend_architecture/04_state_machines.md §2
//
// 责任划分：
// - matchStore           : 按 match_id 索引的 Match 主数据
// - matchReducer         : 应用 match.upserted / match.status_changed；
//                          执行「高阶状态不被低阶覆盖」（FSM 防回退）
// - fixtureChangeHandler : 收到 fixture.changed 后拉 REST 快照覆盖（与 M10 协作）
//
// FSM 简化序：NotStarted < Live < Ended < Closed；
// Cancelled / Abandoned 为终态吸收态，不接受后续状态切换。

const baseUpserted = {
  match_id: "m1",
  tournament_id: "sr:t:1",
  home_team: "Alpha",
  away_team: "Beta",
  scheduled_at: "2026-05-15T18:00:00Z",
  is_live: false,
};

// ---------------------------------------------------------------------------
// 主数据写入
// ---------------------------------------------------------------------------

// Given matchStore 为空
// When 收到首条 match.upserted（match_id="m1", version=1）
// Then store 写入该 match，初始 status="not_started"
describe("given an empty match store", () => {
  it("when match.upserted arrives for the first time then the match is created with status not_started", () => {
    const store = new MatchStore();
    const result = store.applyUpserted({ ...baseUpserted, version: 1 });

    expect(result).toEqual({ applied: true });
    const record = store.get("m1");
    expect(record).toMatchObject({
      match_id: "m1",
      tournament_id: "sr:t:1",
      status: "not_started",
      version: 1,
    });
  });
});

// Given matchStore 已含 match m1 (version=5)
// When 再次收到 match.upserted (match_id="m1", version=6, scheduledAt 变更)
// Then 主数据按 id 合并更新；version 推进到 6；status 保持不变
describe("given an existing match in the store", () => {
  it("when a higher-version match.upserted arrives then the match is merged in place and version advances", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 5 });
    store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      home_score: 1,
      away_score: 0,
      version: 6,
    });

    const result = store.applyUpserted({
      ...baseUpserted,
      scheduled_at: "2026-05-15T20:00:00Z",
      version: 7,
    });

    expect(result.applied).toBe(true);
    const record = store.get("m1");
    expect(record).toMatchObject({
      scheduled_at: "2026-05-15T20:00:00Z",
      status: "live",
      home_score: 1,
      version: 7,
    });
  });
});

// Given matchStore 已含 match m1 (version=5)
// When 收到 match.upserted (match_id="m1", version=3)
// Then 旧版本被丢弃；store 内容保持 version=5
describe("given an existing match with a higher version", () => {
  it("when a stale-version match.upserted arrives then the store does not regress", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 5 });

    const result = store.applyUpserted({
      ...baseUpserted,
      home_team: "STALE",
      version: 3,
    });

    expect(result).toEqual({ applied: false, reason: "stale_version" });
    expect(store.get("m1")?.home_team).toBe("Alpha");
    expect(store.get("m1")?.version).toBe(5);
  });
});

// ---------------------------------------------------------------------------
// FSM 状态机 + 防回退
// ---------------------------------------------------------------------------

// Given match m1 status="not_started"
// When 收到 match.status_changed (status="live", home_score=0, away_score=0, version=2)
// Then status 推进到 live，score 写入
describe("given a not_started match", () => {
  it("when match.status_changed to live arrives then status advances to live and score is recorded", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 1 });

    const result = store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      home_score: 0,
      away_score: 0,
      version: 2,
    });

    expect(result).toEqual({ applied: true });
    const record = store.get("m1");
    expect(record?.status).toBe("live");
    expect(record?.home_score).toBe(0);
    expect(record?.away_score).toBe(0);
    expect(record?.version).toBe(2);
  });
});

// Given match m1 status="live"
// When 收到 match.status_changed (status="not_started", version=2)
// Then 状态保持 live（高阶状态不被低阶覆盖）；reducer 返回 rejected 标志
describe("given a live match", () => {
  it("when a status_changed event attempts to downgrade to not_started then the store keeps the live status", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 1 });
    store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      version: 2,
    });

    const result = store.applyStatusChanged({
      match_id: "m1",
      status: "not_started",
      version: 3,
    });

    expect(result).toEqual({ applied: false, reason: "fsm_downgrade" });
    expect(store.get("m1")?.status).toBe("live");
  });
});

// Given match m1 status="ended"
// When 收到 match.status_changed (status="live", version=...)
// Then 状态保持 ended；FSM 拒绝降级
describe("given an ended match", () => {
  it("when a status_changed event attempts to revert to live then the store keeps the ended status", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 1 });
    store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      version: 2,
    });
    store.applyStatusChanged({
      match_id: "m1",
      status: "ended",
      version: 3,
    });

    const result = store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      version: 4,
    });

    expect(result).toEqual({ applied: false, reason: "fsm_downgrade" });
    expect(store.get("m1")?.status).toBe("ended");
  });
});

// Given match m1 status="cancelled" (终态)
// When 收到任意 match.status_changed
// Then 状态保持 cancelled，终态吸收所有后续状态切换
describe("given a cancelled match in a terminal state", () => {
  it("when any subsequent status_changed arrives then the cancelled state is preserved", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 1 });
    store.applyStatusChanged({
      match_id: "m1",
      status: "cancelled",
      version: 2,
    });

    const result = store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      version: 3,
    });

    expect(result).toEqual({ applied: false, reason: "terminal_state" });
    expect(store.get("m1")?.status).toBe("cancelled");
  });
});

// ---------------------------------------------------------------------------
// fixture.changed → 快照覆盖
// ---------------------------------------------------------------------------

// Given matchStore 已含 match m1 (旧主数据)
// When fixtureChangeHandler 收到 fixture.changed (match_id="m1")，
//      并通过注入的 snapshot fetcher 返回新主数据
// Then store 用快照覆盖主数据（schedule / teams 等），
//      但不会降级已观测的高阶状态
describe("given an existing match and a fixture.changed event", () => {
  it("when the snapshot returns a refreshed match then the store applies the snapshot without regressing status or version", async () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, version: 5 });
    store.applyStatusChanged({
      match_id: "m1",
      status: "live",
      home_score: 2,
      away_score: 1,
      version: 6,
    });

    const refreshedSnapshot: MatchRecord = {
      match_id: "m1",
      tournament_id: "sr:t:1",
      home_team: "Alpha FC",
      away_team: "Beta United",
      scheduled_at: "2026-05-16T20:00:00Z",
      is_live: true,
      status: "not_started",
      version: 4,
    };
    const fetcher = vi.fn().mockResolvedValue(refreshedSnapshot);
    const handler = createFixtureChangeHandler(store, fetcher);

    await handler({
      match_id: "m1",
      change_type: "schedule",
      refetch_required: true,
    });

    expect(fetcher).toHaveBeenCalledWith("m1");
    const record = store.get("m1");
    expect(record?.scheduled_at).toBe("2026-05-16T20:00:00Z");
    expect(record?.home_team).toBe("Alpha FC");
    expect(record?.status).toBe("live");
    expect(record?.home_score).toBe(2);
    expect(record?.version).toBe(6);
  });
});

// ---------------------------------------------------------------------------
// selectors
// ---------------------------------------------------------------------------

// Given matchStore 已含 3 个 match，分别属于 sport sr:sport:1 / sr:sport:2
// When listByTournament("sr:t:1") 被调用
// Then 仅返回 tournament_id 与之匹配的 match
describe("given matches across multiple tournaments", () => {
  it("when listByTournament is queried then only matches belonging to that tournament are returned", () => {
    const store = new MatchStore();
    store.applyUpserted({ ...baseUpserted, match_id: "m1", tournament_id: "sr:t:1", version: 1 });
    store.applyUpserted({ ...baseUpserted, match_id: "m2", tournament_id: "sr:t:1", version: 1 });
    store.applyUpserted({ ...baseUpserted, match_id: "m3", tournament_id: "sr:t:9", version: 1 });

    const ids = store
      .listByTournament("sr:t:1")
      .map((m) => m.match_id)
      .sort();
    expect(ids).toEqual(["m1", "m2"]);
  });
});
