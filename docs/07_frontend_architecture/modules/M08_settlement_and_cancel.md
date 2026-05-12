# M08 — 结算与取消（Settlement / Cancel）

## 目的

把 `bet_settlement.applied` 与 `bet_cancel.applied` 转化为 outcome 级与 bet 级的可视化结果，支撑 P06 我的投注与 P03 已结算盘口提示。

## 数据来源

- 实时：`bet_settlement.applied` / `bet_cancel.applied`
- REST：`GET /api/v1/my-bets/{id}` 详情含完整结算链
- 字段语义：OddsFeed `VoidNotification`、`BetDivident`、Sportsbook 规则文档（`docs/04_sportsbook_rules/`）

## 领域模型

```ts
SettlementRecord {
  id, matchId, marketId, outcomeId,
  result: 'won'|'lost'|'void'|'half_won'|'half_lost'|'dead_heat',
  certainty: 'live_scouted'|'confirmed',
  voidFactor?: number,
  dead_heat_factor?: number,
  appliedAt, version
}
CancelRecord {
  id, scope: 'match'|'market'|'outcome',
  voidReason, supercededBy?, range?, appliedAt
}
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `settlementStore` | 按 outcome 索引结算记录链 |
| `cancelStore` | 按 scope 索引取消记录链 |
| `settlementReducer` / `cancelReducer` | 幂等合并、版本守卫 |

## 与其他模块依赖

- 输入：M02
- 输出：M05（outcome.result）、M06（Market → Settled/Cancelled）、M14、M09

## 未决问题

- [ ] certainty=live_scouted 是否给 UI 一个「待确认」徽标？
- [ ] dead_heat / void_factor 的本地化展示规则？

## 验收要点

- 同一 outcome 收到多次结算（含 certainty 升级）能正确替换
- cancel 的 `supercededBy` 链在 UI 可追溯
- 与 M09 联动，回滚后状态可逆
