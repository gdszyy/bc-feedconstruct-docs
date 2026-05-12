# M14 — 我的投注（My Bets）

## 目的

聚合用户单据的完整生命周期：Pending / Accepted / Rejected / Settled / Cancelled / RolledBack。展示派彩、void、半赢半输、回滚链路。

## 数据来源

- REST：`GET /api/v1/my-bets?status=...`、`GET /api/v1/my-bets/{betId}`
- 实时：`bet.accepted` / `bet.rejected` / `bet.state_changed`、`bet_settlement.*`、`bet_cancel.*`、`*.rolled_back`
- 规则参考：`docs/04_sportsbook_rules/`、`docs/06_betguard_risk/betguard/004_resulting-bets-reporting-selection-outcomes.md`

## 领域模型

```ts
Bet {
  id, placedAt, stake, currency, betType,
  selections: BetSelection[],
  state: BetFsmState,              // 见状态机 §4
  history: BetTransition[],        // append-only 链路
  payout?: { gross, currency, breakdown },
  voidFactor?, deadHeatFactor?
}
BetTransition { at, from, to, reason?, eventId, correlationId }
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `myBetsStore` | 当前用户所有单据索引 |
| `myBetsReducer` | 合并 REST + 实时事件，维护历史链 |
| `payoutCalculator` | 仅做展示用派彩估算；权威值取自后端 |

## 与其他模块依赖

- 输入：M02、M08、M09、M13 提交
- 输出：P06 我的投注页

## 未决问题

- [ ] 是否需要本地缓存历史单据？多账号切换的清理？
- [ ] 派彩是否包含税费/汇率展示？

## 验收要点

- Bet 历史链 append-only，所有转移可追溯
- 回滚链路在 UI 显示完整时间线
- 实时与 REST 合并幂等
- 未结算单据在 producer down 时显示降级（M15）
