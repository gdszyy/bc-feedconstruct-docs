# M05 — 盘口与赔率（Markets & Odds）

## 目的

维护单场赛事的市场（market）与候选项（outcome / selection）的赔率与活跃位，按 specifier 区分变体。

## 数据来源

- REST 快照：`GET /api/v1/matches/{id}/markets`
- 实时增量：`odds.changed`
- 参考字段：OddsFeed `Market` / `Selection` / `MarketType` / `SelectionType`（`docs/01_data_feed/rmq-web-api/009_market.md`、`010_selection.md`）
- 描述映射：M12

## 领域模型

```ts
Market {
  id,                        // (matchId, marketTypeId, specifierKey) 唯一
  matchId,
  marketTypeId, specifiers,  // 如 total=2.5
  status: MarketStatus,      // 来源 M06
  group, tab,                // 来源 M12
  outcomes: Outcome[],
  lastUpdatedAt, version
}
Outcome {
  id, selectionTypeId, name, // name 来源 M12
  odds, active, openingOdds?,
  result?, voidFactor?, deadHeatFactor?  // 结算后由 M08 写入
}
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `marketsStore` | 按 matchId → marketId 双层索引 |
| `oddsReducer` | 合并 snapshot + 增量；版本守卫 |
| `marketSelector` | 派生 `displayOdds` / `bettable`（参考状态机 §3） |

## 与其他模块依赖

- 输入：M02、M06（状态）、M07（停盘）、M12（命名）
- 输出：P03/P04 市场表、M13 投注单

## 未决问题

- [ ] 大量 specifier（如 handicap 阶梯）是否分页/虚拟列表？
- [ ] outcome 是否携带 result 还是由 M08 单独 join？

## 验收要点

- 快照与增量合并幂等，版本回滚被丢弃
- 描述未加载时仅显示骨架，不显示原始 ID
- `bettable` 派生在 status / bet_stop / connection 任一变化时即时刷新
