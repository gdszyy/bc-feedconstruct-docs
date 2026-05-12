# M06 — 盘口状态机

## 目的

把市场状态作为独立领域处理，区分 Active / Suspended / Deactivated / Settled / Cancelled / HandedOver，避免把所有非 Active 都当作「停盘」。

## 数据来源

- 实时：`market.status_changed`、间接来自 M07/M08/M09
- 参考：上传指引 §3.2、`docs/03_sports_model_reference/`

## 状态机

见 [状态机 §3](../04_state_machines.md#3-market-fsmm05--m06)。

## 关键组件

| 组件 | 职责 |
|---|---|
| `marketStatusReducer` | 单纯 FSM 转移，拒绝非法转移并上报 M16 |
| `bettableSelector` | 组合 status + betStop + connection 派生 `bettable` |

## 与其他模块依赖

- 输入：M02、M07、M08、M09、M01（连接降级）
- 输出：M05、M13

## 未决问题

- [ ] `HandedOver` 在 UI 是显示「移交至 X」还是直接禁用？
- [ ] 是否需要保留状态历史用于调试？

## 验收要点

- 6 个状态的所有合法转移均有测试覆盖
- 非法转移不修改状态，且发出遥测
