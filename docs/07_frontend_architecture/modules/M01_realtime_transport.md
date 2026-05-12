# M01 — 实时数据通道（WebSocket Transport）

## 目的

为前端提供一条**可靠、可观测、可恢复**的事件流通道。负责连接、心跳、退避重连、订阅控制帧、缺口回放请求；不解析业务 payload。

## 数据来源（后端契约）

- WebSocket：`WS /ws/v1/stream`（见 [契约 §2](../03_backend_data_contract.md#2-websocket-端点)）
- 控制帧：`subscribe` / `unsubscribe` / `replay_from`
- 生命周期事件：`system.hello` / `system.heartbeat` / `system.replay_*` / `system.producer_status`

## 领域状态

`Connection FSM`：Disconnected / Connecting / Open / Degraded / Reconnecting / Closed（见 [状态机 §1](../04_state_machines.md#1-connection-fsmm01--m15)）

存储：
- `sessionId`、`lastEventId`、`backoffAttempt`、`heartbeatMissed`

## 关键组件（待 TDD 实现）

| 组件 | 类型 | 职责 |
|---|---|---|
| `WsClient` | service | 建链、收发、关闭、错误归类 |
| `Reconnector` | service | 指数退避 + 抖动 + 上限 |
| `Heartbeat` | service | 心跳超时判定 |
| `transportStore` | store | 暴露 Connection FSM 给 UI |

## 与其他模块依赖

- 输入：用户登录 token、订阅意图（来自 M11、各页面）
- 输出：原始 envelope → M02；连接状态 → M15

## 未决问题

- [ ] token 续期与重连耦合策略？
- [ ] 同源多 tab 是否共享一个 WS（SharedWorker）？

## 验收要点

- 断网 30s 内能自动重连
- 重连后必须发送 `replay_from`（依赖 M10）
- 心跳丢失阈值与后端一致
- 所有连接事件可遥测（M16）
