# M09 — 回滚纠错（Rollback）

## 目的

处理供应商纠错消息（`bet_settlement.rolled_back` / `bet_cancel.rolled_back`），把先前结算或取消撤销，并保留**完整变更链**以便用户与运营追溯。

## 数据来源

- 实时：`bet_settlement.rolled_back` / `bet_cancel.rolled_back`
- 参考：上传指引 §3.3

## 领域模型

```ts
RollbackRecord {
  id, targetType: 'settlement'|'cancel', targetId,
  rolledBackAt, reason?, version
}
```

回滚后的事件链以 append-only 形式存放在 `settlementStore` / `cancelStore`，UI 上显示「⟲ 已回滚 → 等待重新结算」。

## 关键组件

| 组件 | 职责 |
|---|---|
| `rollbackReducer` | 接收 rolled_back，在目标 store 写入「已回滚」状态 |
| `betFsmTransition` | 触发 Bet FSM `Settled → Accepted` 回滚（M13/M14） |

## 与其他模块依赖

- 输入：M02
- 输出：M05、M06、M08、M14（我的投注链路）

## 未决问题

- [ ] 回滚后是否会紧跟一条新的结算？需要 UI 自动合并显示还是分两条？
- [ ] 回滚的等待时限是否需要在 UI 显式提示？

## 验收要点

- 原结算/取消不被覆盖，链路可追溯
- Bet FSM 显式从 `Settled` / `Cancelled` 回退到 `Accepted` 或上一稳定态
- 我的投注页能展示「已结算 → 已回滚 → 重新结算」的时间线
