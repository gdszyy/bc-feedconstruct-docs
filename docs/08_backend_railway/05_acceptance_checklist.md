# 后端最小合格验收清单

完整对应上传指引《体育博彩数据源接入说明指引》§4 的 16 项。每一项都必须**有自动化测试覆盖**（BDD `Given/When/Then`）+ **运行期可观测**（日志/指标/DB 行）。

| 序号 | 验收项 | 合格标准 | 自动化测试 | 运行期证据 |
|---|---|---|---|---|
| 1 | 连接 | 启动 30s 内完成 RMQ + WebAPI 双连接；Token 拿到；queue 绑定成功 | `internal/feed/connection_test.go` | `bff` 启动日志：`rmq.connected`、`webapi.token.acquired` |
| 2 | 消息留痕 | 任意消息处理前先写 `raw_messages`；可按 `event_id`/`message_type`/`source` 查询 | `internal/feed/raw_message_test.go` | `SELECT count(*) FROM raw_messages` 持续增长；查询 API |
| 3 | 消息覆盖 | 9 类 message_type 都有专用 handler 注册；未知类型走 dead-letter | `internal/feed/dispatch_test.go` | `metrics_counters.handler_<type>` |
| 4 | 主数据 | sport/region/competition/match upsert 后字段齐全；fixture_change 改变 `start_at`/`status` 即落 history | `internal/catalog/match_test.go`、`internal/catalog/fixture_change_test.go` | `SELECT * FROM matches WHERE id=...` |
| 5 | 赔率 | odds_change 解析 market+specifier+outcome+odds+active+status；并把 outcomes upsert | `internal/odds/odds_change_test.go` | `SELECT * FROM outcomes WHERE match_id=...` |
| 6 | 停投 | bet_stop / market suspended 必须更新 `markets.status` 且写 `market_status_history` | `internal/odds/bet_stop_test.go` | history 表行数 |
| 7 | 结算 | bet_settlement 解析 outcome/result/certainty/void_factor/dead_heat_factor；幂等 | `internal/settlement/settlement_test.go` | `settlements` 行 |
| 8 | 取消 | bet_cancel 写 void_reason + 时间区间 + superceded_by；并把 markets.status 置 cancelled | `internal/settlement/cancel_test.go` | `cancels` 行 |
| 9 | 回滚 | rollback_settlement / VoidAction=2 触发 `rollbacks` 行；对应 settlement/cancel `rolled_back_at` 非空；market 状态恢复 | `internal/settlement/rollback_test.go` | `SELECT * FROM rollbacks` |
| 10 | 恢复 | 启动/单赛事/状态消息/fixture change 四种 scope 在 `recovery_jobs` 落 `success`；429 走指数退避 | `internal/recovery/coordinator_test.go` | `recovery_jobs` 表 |
| 11 | 幂等 | 同一消息重复投递不增 settlement/cancel/outcome 行；`raw_messages` 唯一约束命中 | `internal/storage/idempotency_test.go` | unique violation count = 0 业务行 |
| 12 | 防回退 | match `ended/closed/cancelled` 不被 `live` 覆盖；market `settled/cancelled` 不被 `active` 覆盖 | `internal/catalog/status_no_regress_test.go`、`internal/odds/market_status_no_regress_test.go` | 日志 `status.regress.blocked` |
| 13 | 订阅 | book/unbook 事件维护 `subscriptions` + `subscription_events`；过期 live 自动 unbook | `internal/subscription/manager_test.go` | `subscriptions.released_at` |
| 14 | 描述数据 | market_type / selection_type 描述可同步并通过 `GET /api/v1/market-types` 返回 | `internal/webapi/descriptions_test.go`、`internal/bff/descriptions_handler_test.go` | API 200 |
| 15 | 监控 | `producer_health` 30s 无消息置 down；`/metrics` 暴露消息量/卡赛/recovery 次数/队列 lag | `internal/health/reporter_test.go` | Prometheus 抓取 |
| 16 | 数据治理 | 后台 job 按保留窗口（默认 raw_messages 7d、history 30d、settlements 永久）滚动清理 | `internal/storage/retention_test.go` | `metrics_counters.retention_deleted` |

## BFF 接口验收

| 序号 | 接口 | 验收 |
|---|---|---|
| B1 | `GET /healthz` | 始终 200 |
| B2 | `GET /readyz` | 依赖未就绪返 503，全部就绪返 200 |
| B3 | `GET /api/v1/sports` | 返回所有 active sport |
| B4 | `GET /api/v1/matches/{id}` | 单赛事快照（match + markets + outcomes + 最近 settlement/cancel） |
| B5 | `WS /ws` | 客户端 `subscribe { match_id }` 后能收到 `odds_update`/`market_status`/`bet_stop`/`settlement`/`cancel`/`rollback`/`producer_status` 事件 |

## 安全验收

| 序号 | 项目 | 验收 |
|---|---|---|
| S1 | 仓库无任何真实 `FC_*` 凭证（grep 全仓库） | CI 在 PR 检查 |
| S2 | WebSocket 校验 origin（`WS_ALLOWED_ORIGINS`） | `internal/bff/ws_origin_test.go` |
| S3 | REST 加 rate limit（默认 60 req/min/ip） | `internal/bff/ratelimit_test.go` |
