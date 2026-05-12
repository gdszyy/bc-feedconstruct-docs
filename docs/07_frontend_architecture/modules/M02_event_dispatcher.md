# M02 — 事件分发器（Dispatcher）

## 目的

把 M01 输出的 envelope 按 `type` 路由到对应领域 store/handler，并执行幂等、版本比较、错误归并。**不做业务计算**。

## 数据来源

- 输入：M01 推送的 envelope（结构见 [契约 §1](../03_backend_data_contract.md#1-通用-envelope)）
- 输入：REST 响应 envelope（用于异步通知场景，可选）

## 路由表（按 type 前缀）

| 前缀 | 目标 |
|---|---|
| `system.*` | M01 transportStore / M10 / M15 |
| `sport.*` / `tournament.*` | M03 catalog store |
| `match.*` | M04 match store |
| `fixture.*` | M04 + 触发 M10 快照刷新 |
| `odds.*` | M05 markets store |
| `market.status_changed` | M06 marketStatus store |
| `bet_stop.*` | M07 betStop store |
| `bet_settlement.*` | M08 settlement store + M14 |
| `bet_cancel.*` | M08 cancel store + M14 |
| `*.rolled_back` | M09 rollback store + 原模块状态机 |
| `subscription.*` | M11 subscription store |
| `bet.*` | M13 / M14 |

## 关键组件

| 组件 | 职责 |
|---|---|
| `dispatcher` | type → handler 注册表 |
| `EventDedup` | 按 `event_id` 幂等 |
| `VersionGuard` | 按 `payload.version` 或 `occurred_at` 拒绝过期事件 |
| `ErrorRouter` | 把解析失败 / 未知 type 路由到 M16 telemetry |

## 与其他模块依赖

- 上游：M01
- 下游：M03~M09、M11、M13、M14

## 未决问题

- [ ] 未知 type 是丢弃还是缓存？
- [ ] 跨实体事件（match.status_changed 同时影响 M04 + M05）的处理顺序？

## 验收要点

- 重复 `event_id` 不二次触发
- 过期版本被丢弃且计数可观测
- 未知 type 不会让前端崩溃，仅记录
