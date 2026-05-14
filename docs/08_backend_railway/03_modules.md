# Go 后端模块划分

把上传指引 16 业务域映射为 Go 包。所有包都遵守 `CLAUDE.md` 的 BDD 流程：先空测试（`*_test.go`，仅 Given-When-Then 注释 + 函数名）→ `AskUserQuestion` 确认 → 正式测试 → 实现。

## 1. 目录布局

```
backend/
├── cmd/
│   ├── bffd/        # 主进程：feed 消费 + 处理器 + WebSocket/REST
│   └── migrate/     # Postgres migration runner
├── internal/
│   ├── config/      # env 加载 + 校验
│   ├── storage/     # pgx pool + repository
│   ├── feed/        # M01 接入：FeedConstruct RMQ 客户端 + replay 模式
│   ├── catalog/     # M03/M04 主数据
│   ├── odds/        # M05/M06/M07 赔率/状态/停投
│   ├── settlement/  # M08/M09 结算/取消/回滚
│   ├── recovery/    # M10 恢复
│   ├── subscription/# 订阅生命周期
│   ├── health/      # M15 producer 健康 + metrics
│   ├── bff/         # 分发：WebSocket hub + REST handlers
│   └── webapi/      # FeedConstruct WebAPI 客户端（Token、snapshot、descriptions）
├── migrations/
│   └── 001_init.sql ...
├── test/
│   └── e2e/         # 跨包行为测试（godog 可选）
├── Dockerfile
├── railway.json
├── go.mod
└── go.sum
```

## 2. 包职责与对外接口

| 包 | 职责 | 主要类型/函数 |
|---|---|---|
| `config` | 解析 env，区分 `live`/`replay` 模式，缺失变量 fail-fast | `Load() (*Config, error)` |
| `storage` | pgx pool、迁移、repository 接口（不含 ORM） | `NewPool(ctx, dsn)`、`RawMessageRepo`、`MarketRepo` 等 |
| `feed` | RMQ 连接 + 队列绑定 + GZIP 解压 + 入库 + 内部 fanout | `Consumer`、`Replayer`（read `raw/json`） |
| `catalog` | 处理 sport/region/competition/match/fixture_change | `Handler.Handle(ctx, msg) error` |
| `odds` | odds_change → markets/outcomes；bet_stop → status 更新 | `OddsHandler`、`BetStopHandler` |
| `settlement` | bet_settlement / bet_cancel / rollback；维护 settled/cancelled/rolled_back 状态 | `SettlementHandler`、`CancelHandler`、`RollbackHandler` |
| `recovery` | snapshot 触发 + 限流退避；启动/产品/单事件/stateful 四种 scope | `Coordinator` |
| `subscription` | 跟踪 booking、释放过期 live | `Manager` |
| `health` | 维护 `producer_health`、消息量、卡赛检测；暴露 metrics | `Reporter` |
| `webapi` | Token 24h 刷新；DataSnapshot；GetMatchByID；descriptions | `Client` |
| `bff` | WebSocket hub（按 subscription 推送）+ REST（snapshot/match/bet 等） | `NewServer(deps Deps) *http.Server` |

## 3. 消息处理总流（M01 → M02 → 业务包）

```
RMQ delivery / replay file
       │
       ▼
[feed.Consumer]
   1. ungzip
   2. 解析 envelope（type/objectType/objectId/ts）
   3. INSERT raw_messages（幂等）
   4. publish 到内部 RabbitMQ exchange "feed.events"
       │
       ▼
[内部 RabbitMQ exchange "feed.events"]
   - routing_key = "<message_type>.<sport_id>"
   - bindings:
       catalog.#       → catalog handler
       odds_change.#   → odds.OddsHandler
       bet_stop.#      → odds.BetStopHandler
       settlement.#    → settlement.SettlementHandler
       cancel.#        → settlement.CancelHandler
       rollback.#      → settlement.RollbackHandler
       fixture.#       → catalog handler
       fixture_change.# → catalog handler + recovery（如缺数据）
       alive.#         → health.Reporter
       snapshot_complete.# → recovery.Coordinator
       │
       ▼
[handler 处理 → 写领域表 → 通过 bff.Hub 广播 WS 事件]
```

## 4. WebSocket / REST 端点（与 `docs/07_frontend_architecture/03_backend_data_contract.md` 对齐）

| 类型 | 路径 | 用途 |
|---|---|---|
| WS | `/ws` | 客户端 subscribe match/markets，BFF 推送 `odds_update`/`market_status`/`bet_stop`/`settlement`/`cancel`/`rollback`/`producer_status` |
| REST | `GET /api/v1/sports` | 体育目录 |
| REST | `GET /api/v1/competitions?sport_id=` | 联赛 |
| REST | `GET /api/v1/matches?live=true` | 赛事列表 |
| REST | `GET /api/v1/matches/{id}` | 赛事快照（含 markets/outcomes/status） |
| REST | `GET /api/v1/matches/{id}/markets` | 仅盘口 |
| REST | `GET /api/v1/matches/{id}/settlements` | 结算/取消/回滚历史 |
| REST | `POST /api/v1/recovery/event/{id}` | 手动单赛事恢复（运维） |
| REST | `GET /healthz` / `GET /readyz` / `GET /metrics` | 运维 |

## 5. 模块开发顺序

依次完成下列模块的 BDD 空测试 → 用户确认 → 正式测试 → 实现：

1. `config` + `storage`（基础）
2. `feed`（M01 + 幂等）
3. `recovery`（M10，先把骨架打通便于后续测试）
4. `catalog` + `webapi`（M03/M04）
5. `odds`（M05/M06/M07）
6. `settlement`（M08/M09）
7. `subscription`
8. `health`（M15）
9. `bff`（分发层；最后装配）

## 6. 测试策略

| 层级 | 工具 | 目录 |
|---|---|---|
| 单元 | `testing` + `testify/require` | 与生产同包 `*_test.go` |
| 集成（DB） | `testcontainers-go` 起 Postgres | `internal/storage/*_test.go` |
| 集成（RMQ） | `testcontainers-go` 起 RabbitMQ；`replay` 模式直接读 `raw/json` | `internal/feed/*_test.go` |
| 跨包行为 | `cucumber/godog`（可选） | `test/e2e/` |
| WebAPI 真实集成 | 仅在 `FEED_MODE=live` 且配置环境变量时跑；CI 默认跳过 | `internal/webapi/*_test.go`（带 `//go:build integration_live`） |
