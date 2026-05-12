# 设计原则

源自上传指引《体育博彩数据源接入说明指引》§3，针对 **Go BFF + Next.js 前端**的具体化。

## P1 接入层只做接入，不堆业务

| 后端（Go BFF） | 前端（Next.js） |
|---|---|
| 消费 AMQP / REST 原始消息，落库原文，分发到业务处理器 | 仅消费 BFF 暴露的 WebSocket 事件与 REST 端点；不解析供应商 XML |
| 输出统一 envelope（type / event_id / product_id / sport_id / occurred_at / payload） | 按 envelope.type 路由到对应 store/handler，禁止在 UI 层解析 envelope |

**前端推论**：

- 实时通道（M01）只负责连接、心跳、断线重连、原始事件分发，不更新业务 store
- 事件分发器（M02）按 type 路由到领域 store，不在分发层做业务计算
- UI 组件只订阅领域 store 的派生 selector，不直接订阅 WebSocket

## P2 状态必须领域化，不能靠临时判断

盘口、赛事、结算、取消必须使用**前后端一致的枚举**。

| 域 | 必须枚举 | 模块 |
|---|---|---|
| Market | Active / Suspended / Deactivated / Settled / Cancelled / HandedOver | M05 / M06 |
| Match | NotStarted / Live / Ended / Closed / Cancelled / Abandoned | M04 |
| Bet（前端视图） | Pending / Accepted / Rejected / Settled / Cancelled / RolledBack | M13 / M14 |
| Connection | Connecting / Open / Degraded / Reconnecting / Closed | M01 / M15 |
| Subscription | Idle / Booking / Subscribed / Unbooking / Released / Failed | M11 |

**前端推论**：禁止 UI 用 `status !== 'active'` 之类的二分判断决定是否允许下注或展示，必须显式 switch over 枚举，落到 selector / 状态机文件，由测试覆盖。

## P3 必须处理纠错消息

回滚（rollback_bet_settlement / rollback_bet_cancel）和替代取消（superceded_by）必须在前端可视化。

**前端推论**：

- M09（回滚）：单独事件类型 `bet_settlement.rolled_back` / `bet_cancel.rolled_back`，落到投注单/我的投注的状态机
- M14（我的投注）：必须能展示「已结算 → 已回滚 → 重新结算」链路，保留时间戳
- 不允许通过覆盖原记录的方式实现回滚展示，必须保留变更链

## P4 恢复机制是必需能力，不是兜底脚本

后端负责消息缺口恢复，前端负责**会话/视图缺口**恢复。

| 后端职责（不在本规划范围） | 前端职责（M10） |
|---|---|
| product/event/stateful/fixture_change 恢复 | 断线重连后拉取赛事快照、市场快照、未结投注快照 |
| 处理供应商限流 | 退避重连 + 标记降级（M15） |
| recovery 完成发送 snapshot_complete 等价事件 | 等到 `snapshot.complete` 事件才把页面从 stale 切回 fresh |

**前端推论**：所有受实时更新的视图必须有 `stale` 标志（M15）和 `lastUpdatedAt`，并且在 WebSocket 重连完成前禁止用户对受影响盘口下单。

## P5 订阅要有生命周期

前端「关注」「订阅」「自动取关」必须有明确状态机（M11），并通过后端 REST 接口落地，不允许由 UI 直接管控订阅。

**前端推论**：

- 关注（favorite）= 本地偏好；订阅（subscription）= 服务端订阅状态
- 收藏页 / 详情页只能读 `Subscribed` 状态，不能假设订阅一定成功
- 比赛结束后由后端发起 unbook，前端只负责清理本地视图

## P6 幂等与防回退

| 约束 | 实施位置 |
|---|---|
| 重复事件不应触发重复 UI 更新 | 领域 store 用 `(entity_id, version|timestamp)` 比较，旧消息丢弃 |
| 高阶赛事状态（ended/closed）不被低阶覆盖 | M04 状态机 reducer 显式拒绝降级转移 |
| 重复 bet_settlement 不重复弹通知 | M08/M14 用 settlement_id 去重 |

## P7 静态描述必须独立加载

market description / outcome name / group / tab 不应跟随实时事件下发，应通过 REST（M12）按版本号缓存。前端 UI 渲染必须用本地描述映射，不允许把 ID 直接显示给用户。

## P8 可追溯

所有前端错误、异常状态、用户操作都需要带上后端 envelope 的 `event_id` / `correlation_id`，写入遥测（M16），以便排障。
