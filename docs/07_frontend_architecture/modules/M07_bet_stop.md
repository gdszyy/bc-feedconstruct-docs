# M07 — 停投与赛事级停盘（Bet Stop）

## 目的

处理 `bet_stop` 与 `bet_stop.lifted`，按赛事 / 市场组维度冻结可下注状态，防止供应商停盘后本地继续售卖。

## 数据来源

- 实时：`bet_stop.applied` / `bet_stop.lifted`
- 参考：上传指引 §2、OddsFeed bet_stop 语义

## 领域模型

```ts
BetStopState {
  byMatch: Map<matchId, { groups: Set<group>, scope: 'match'|'group', appliedAt }>
}
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `betStopStore` | 当前生效的停投范围 |
| `betStopReducer` | 应用 applied/lifted |
| `betStopSelector` | 查询给定 (matchId, market.group) 是否被停 |

## 与其他模块依赖

- 输入：M02
- 输出：M05/M06 的 `bettable` 派生、M13 下注校验、P03 详情页提示

## 未决问题

- [ ] `group` 维度如何与 M12 的 group/tab 对齐？
- [ ] bet_stop 与个别 market `Suspended` 同时存在时的优先级？

## 验收要点

- 停投生效时所有受影响盘口的下注按钮立即禁用
- 解除停投后 UI 立即恢复（不需要刷新）
