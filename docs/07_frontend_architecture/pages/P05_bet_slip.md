# P05 — 投注单

## 目的

承载 M13 的全部用户交互。可以是浮层、抽屉、独立页面三种形态，状态由 `betSlipStore` 统一管理。

## 数据来源

- 本地：M13 betSlipStore
- 实时：M05/M06/M07 → 选项可用性、价格变更
- REST：`/bet-slip/validate`、`/bet-slip/place`
- 限额/规则：来自 BFF（`docs/06_betguard_risk/betguard/`）

## 状态机交互

| FSM 状态 | UI 表现 |
|---|---|
| Empty | 占位 |
| Editing | 显示选项、stake 输入、错误提示 |
| Ready | 按钮 enable |
| NeedsReview | 高亮变更价格，强制确认 |
| Submitting | 按钮 loading，禁用编辑 |
| Submitted | 转入 P06 入口 |

## 关键组件

- `<BetSlipList/>`
- `<StakeInput/>`
- `<PriceChangeReview/>`
- `<SubmitButton/>`
- `<RejectExplanation/>`

## 验收要点

- 任一 selection 变为 Suspended/BetStop 立即禁用提交
- 价格变更后强制进入 NeedsReview
- 重试不重复创建单
- producer down / Degraded 时无法 Submit
