# M10 — 恢复补偿（前端缺口检测与快照对齐）

## 目的

实时流断线 / 重启 / 限流后，前端必须能：(1) 自动请求缺口回放；(2) 拉取关键视图的快照；(3) 标记并最终解除 stale。**前端不直接做 OddsFeed 级的 product/event/stateful 恢复**，那是后端职责。

## 数据来源

- 实时控制：`replay_from`（控制帧，M01 发起）
- 实时事件：`system.replay_started` / `system.replay_completed`
- REST 快照：`/api/v1/matches/{id}`、`/api/v1/matches/{id}/markets`、`/api/v1/my-bets`

## 触发点

| 触发 | 动作 |
|---|---|
| 重连成功（M01 Open） | 发送 `replay_from(lastEventId)` |
| `fixture.changed` | 对应 match 拉 REST 快照（覆盖式） |
| 长时间无事件（心跳正常但业务静默） | 抽样拉快照对账 |
| 页面 Hydration | 当前视图涉及的 match / my-bets 拉快照 |

## 关键组件

| 组件 | 职责 |
|---|---|
| `recoveryCoordinator` | 调度 replay / snapshot 拉取 |
| `staleTracker` | 维护 stale 标志（M15 显示） |
| `snapshotApi` | 封装 REST 快照调用 |

## 与其他模块依赖

- 输入：M01、M02、M04、M05、M14
- 输出：M15（stale 可视化）

## 未决问题

- [ ] 是否对每个 match 维护独立 stale，还是全局 stale？
- [ ] 快照 vs replay 的覆盖顺序？建议「先 replay 后 snapshot」并用 version 做最终对齐

## 验收要点

- 断线 → 重连 → replay_completed 期间，相关视图显示 stale
- replay 完成后能丢弃 stale 期间累积的过期消息
- 快照与增量到达顺序乱序时不引入回退
