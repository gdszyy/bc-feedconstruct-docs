# 后端 → 前端数据契约（骨架）

本契约是 **Go BFF ↔ Next.js 前端**之间的最小切面。具体字段在落地阶段以后端确认为准；本文档只锁定结构、命名、语义，避免前端假设供应商原始字段。

## 1. 通用 Envelope

所有 WebSocket 事件与异步通知 REST 响应统一使用以下 envelope：

```jsonc
{
  "type": "odds.changed",          // 事件类型；点分命名，见 §3
  "schema_version": "1",           // envelope schema 版本
  "event_id": "01HX...ULID",       // 由后端生成，前端用于幂等
  "correlation_id": "...",         // 链路追踪，遥测必带（M16）
  "product_id": "live | prematch", // 数据源产品
  "occurred_at": "2026-05-12T...", // 业务时间
  "received_at":  "2026-05-12T...",// 后端落地时间
  "entity": {                      // 主体 ID（用于路由）
    "sport_id": "...",
    "tournament_id": "...",
    "match_id": "...",
    "market_id": "...",
    "outcome_id": "..."
  },
  "payload": { /* 类型特定 */ }
}
```

前端约束：

- 路由按 `type` 第一段（odds / match / market / bet / settlement / cancel / rollback / fixture / subscription / system）分发
- 幂等键：`event_id`；版本比较键：`payload.version` 或 `occurred_at`（具体由后端确认）

## 2. WebSocket 端点

| 端点 | 用途 | 备注 |
|---|---|---|
| `WS /ws/v1/stream` | 主事件流 | 鉴权 token 在 query 或 cookie |
| 客户端 → 服务端控制帧 | `subscribe` / `unsubscribe` / `replay_from` | 见 §4 |

## 3. 事件类型清单（前端必须处理）

> 命名前缀对应模块。具体 payload 由后端 OpenAPI / AsyncAPI 落定，本表只锁语义。

### 3.1 实时通道生命周期（M01 / M10）

| type | 说明 |
|---|---|
| `system.hello` | 连接建立后服务端首条消息，携带 `session_id` |
| `system.heartbeat` | 心跳；超时阈值由后端给定 |
| `system.replay_started` | 后端开始回放缺口 |
| `system.replay_completed` | 回放结束，前端可解除 `stale` |
| `system.producer_status` | producer alive / down，驱动降级（M15） |

### 3.2 主数据（M03 / M04）

| type | 说明 |
|---|---|
| `sport.upserted` / `sport.removed` | 体育目录变更 |
| `tournament.upserted` / `tournament.removed` | 赛事/联赛变更 |
| `match.upserted` | 赛事基础信息（队伍、开赛时间、live 标识） |
| `match.status_changed` | 赛事状态机变化（M04） |
| `fixture.changed` | fixture_change，前端需触发主数据刷新（拉 REST 快照） |

### 3.3 赔率与盘口（M05 / M06 / M07）

| type | 说明 |
|---|---|
| `odds.changed` | market + outcomes 的赔率/活跃位变化 |
| `market.status_changed` | Active / Suspended / Deactivated / HandedOver |
| `bet_stop.applied` | 赛事/市场组级停盘 |
| `bet_stop.lifted` | 停盘解除 |

### 3.4 结算 / 取消 / 回滚（M08 / M09）

| type | 说明 |
|---|---|
| `bet_settlement.applied` | outcome 级 result + certainty + void_factor + dead_heat_factor |
| `bet_settlement.rolled_back` | 结算回滚 |
| `bet_cancel.applied` | void_reason / 区间 / superceded_by |
| `bet_cancel.rolled_back` | 取消回滚 |

### 3.5 订阅（M11）

| type | 说明 |
|---|---|
| `subscription.changed` | 当前用户/会话的订阅状态机变化 |

### 3.6 投注流（M13 / M14）

| type | 说明 |
|---|---|
| `bet.accepted` / `bet.rejected` | 下注异步结果 |
| `bet.state_changed` | Pending → Accepted → Settled → ... 链路 |

## 4. 客户端控制帧

```jsonc
// 订阅范围（前端按页面/视图聚合发起）
{ "op": "subscribe",   "scope": { "match_ids": ["..."], "sport_ids": ["..."] } }
{ "op": "unsubscribe", "scope": { "match_ids": ["..."] } }

// 重连后的缺口回放
{ "op": "replay_from", "cursor": "<last event_id>", "session_id": "..." }
```

## 5. REST 端点（最小集）

| Method | Path | 用途 | 模块 |
|---|---|---|---|
| GET | `/api/v1/catalog/sports` | 体育目录（i18n、排序） | M03 |
| GET | `/api/v1/catalog/tournaments?sport_id=...` | 联赛列表 | M03 |
| GET | `/api/v1/matches?filter=...` | 赛事列表（赛前/滚球） | M04 |
| GET | `/api/v1/matches/{id}` | 赛事快照（含当前 markets） | M04/M05 |
| GET | `/api/v1/matches/{id}/markets` | 市场快照 | M05 |
| GET | `/api/v1/descriptions/markets?version=...` | 市场描述、玩法分组、tabs | M12 |
| GET | `/api/v1/descriptions/outcomes?version=...` | outcome 名称 | M12 |
| POST | `/api/v1/bet-slip/validate` | 价格/状态/限额预校验 | M13 |
| POST | `/api/v1/bet-slip/place` | 下注（幂等键必带） | M13 |
| GET | `/api/v1/my-bets?status=...` | 我的投注列表 | M14 |
| GET | `/api/v1/my-bets/{betId}` | 单注详情（含结算/取消/回滚链） | M14 |
| POST | `/api/v1/subscriptions` | 关注/订阅赛事 | M11 |
| DELETE | `/api/v1/subscriptions/{id}` | 取消订阅 | M11 |
| GET | `/api/v1/system/health` | producer 状态、降级窗口 | M15 |

## 6. 幂等与版本

- 下注：前端必须生成 `Idempotency-Key`（UUIDv7 或 ULID），重试同一 key 不产生新单
- 描述资源：`ETag` + `If-None-Match`，或 `version` 查询参数
- 事件版本：`payload.version`（单调递增）用于丢弃旧消息

## 7. 错误模型

```jsonc
{
  "error": {
    "code": "BET_REJECTED_PRICE_CHANGED",
    "message": "...",
    "retriable": false,
    "correlation_id": "..."
  }
}
```

错误码命名空间由后端枚举；前端 M02 必须把错误事件路由到 `M16 telemetry` 并按业务模块定位 UI 处理。

## 8. 未决问题（待后端确认）

- [ ] envelope 的版本字段是 `payload.version` 还是 `occurred_at` 单调？
- [ ] WebSocket 是否支持服务端推订阅列表回执？还是 fire-and-forget？
- [ ] `system.replay_completed` 是否分维度（按 sport / match）？
- [ ] 投注幂等键的 TTL？
- [ ] 描述资源是否支持增量？还是整包替换？
