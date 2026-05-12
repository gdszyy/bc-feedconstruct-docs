# M04 — 赛事与赛程（Match / Fixture / Schedule）

## 目的

维护单场赛事的主数据、状态机、比分、阶段，并响应 fixture_change。

## 数据来源

- REST：`GET /api/v1/matches`、`GET /api/v1/matches/{id}`
- 实时：`match.upserted` / `match.status_changed` / `fixture.changed`
- 参考字段：OddsFeed `Match` / `Stat` / `MatchMember`（`docs/01_data_feed/rmq-web-api/008_match.md` 等）
- 状态参考：`docs/03_sports_model_reference/match-lifecycle/001_match-lifecycle.md`

## 领域模型

```ts
Match {
  id, sportId, tournamentId,
  home: MatchMember, away: MatchMember,
  scheduledAt, liveOdds: bool,
  status: MatchStatus,       // 见状态机 §2
  score: { home, away, period? },
  phase: { periodId, clock? },
  lastUpdatedAt, version
}
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `matchStore` | 按 id 索引 |
| `matchReducer` | 应用 upserted / status_changed；执行防回退（P6） |
| `fixtureChangeHandler` | 收到 fixture.changed 时拉 REST 快照覆盖 |

## 与其他模块依赖

- 输入：M02
- 输出：M05（赔率定位）、M11（订阅）、P02/P03/P04

## 未决问题

- [ ] 比分增量是否走单独事件还是嵌入 match.upserted？
- [ ] 滚球阶段（period/clock）频率是否与赔率事件解耦？

## 验收要点

- 高阶状态不被低阶覆盖（防回退）
- fixture_change 后能拉取并替换主数据
- 比分/状态在 stale 期间显示视觉降级（M15）
