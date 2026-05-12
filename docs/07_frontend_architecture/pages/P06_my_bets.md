# P06 — 我的投注

## 目的

展示用户的全部单据与其完整状态链路，能解释「为什么这个单据是这个金额、这个状态」。

## 数据来源

- M14 myBetsStore
- 实时：bet_settlement / bet_cancel / *.rolled_back / bet.state_changed
- REST：`/api/v1/my-bets`

## 视图组织

| Tab | 过滤 |
|---|---|
| 未结算 | state ∈ {Accepted} |
| 已结算 | state ∈ {Settled} |
| 已取消 | state ∈ {Cancelled} |
| 全部 | 全部含历史链 |

## 关键组件

- `<BetCard/>` 显示当前状态 + 历史时间线
- `<RollbackBadge/>` 标识回滚链路
- `<PayoutBreakdown/>`

## 验收要点

- 回滚链「已结算 → 已回滚 → 重新结算」完整可见
- 状态时间线与 envelope 的 `occurred_at` 一致
- producer down 时未结算单据显示降级
- 切换 tab 不重新拉取已加载详情
