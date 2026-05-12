# M13 — 投注单（Bet Slip）

## 目的

承担用户构建投注的本地状态、价格变更确认、停投拦截、幂等下注、与后端 `bet-slip/validate` 与 `bet-slip/place` 联动。

## 数据来源

- REST：`POST /api/v1/bet-slip/validate`、`POST /api/v1/bet-slip/place`
- 实时：`market.status_changed`、`bet_stop.*`、`odds.changed`（影响选项可用性）
- 实时：`bet.accepted` / `bet.rejected` / `bet.state_changed`（返回投注结果）
- 投注规则：`docs/06_betguard_risk/betguard/` 与 `docs/04_sportsbook_rules/`

## 领域模型

```ts
BetSlip {
  selections: Array<{ matchId, marketId, outcomeId, lockedOdds }>,
  betType: 'single'|'combo'|'system',
  stake, currency,
  state: BetSlipFsmState,    // 见 §6
  idempotencyKey,            // 进入 Submitting 时生成
  priceChanges: Array<{ outcomeId, from, to }>
}
```

状态机：见 [§6](../04_state_machines.md#6-betslip-fsmm13)。

## 关键组件

| 组件 | 职责 |
|---|---|
| `betSlipStore` | 选项列表、stake、FSM |
| `betSlipValidator` | 调 BFF `validate` 做服务端校验 |
| `betSlipSubmitter` | 调 BFF `place`，幂等键防重 |
| `priceChangeWatcher` | 监听 odds.changed，触发 NeedsReview |
| `availabilityWatcher` | 监听 market/bet_stop，禁用不可下选项 |

## 与其他模块依赖

- 输入：M05/M06/M07（可用性）、M02、用户操作
- 输出：M14（提交后转入我的投注）、M16 遥测

## 未决问题

- [ ] 多选项组合与 system bet 的本地校验规则？
- [ ] 价格变更阈值（直接拒绝 / 仅提示）由谁决定？

## 验收要点

- 任一 selection 变为不可下注时按钮禁用且解释原因
- 价格变化必须进入 NeedsReview，不允许悄悄按新价格提交
- 重试同一 Idempotency-Key 不产生新单
- Submitting 失败有重试入口
