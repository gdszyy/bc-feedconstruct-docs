# 体育博彩数据源 Go BFF — Railway 部署骨架（规划）

> 本目录是配合根目录《体育博彩数据源接入说明指引》落地的 **后端骨架文档**，与 [`docs/07_frontend_architecture/`](../07_frontend_architecture/) 形成"前后端契约对"。
>
> 本目录只包含 **规划与契约骨架**，不包含实现代码。任何代码落地前必须按 `CLAUDE.md` 的 TDD/BDD 流程进行：先写 Given-When-Then 空测试，`AskUserQuestion` 确认，再补正式测试，再实现。

## 设计来源

| 输入 | 说明 |
|---|---|
| 上传指引 | 《体育博彩数据源接入说明指引》16 业务域、16 项验收 |
| OddsFeed 文档 | [`docs/01_data_feed/rmq-web-api/`](../01_data_feed/rmq-web-api/) 提供 RMQ/WebAPI 真实契约（JSON、GZIP、`P{PartnerID}_live/_prematch` 队列、Token 24h 刷新） |
| BetGuard 文档 | [`docs/06_betguard_risk/betguard/`](../06_betguard_risk/betguard/) 提供下注/派彩/回滚 API |
| 前端规划 | [`docs/07_frontend_architecture/`](../07_frontend_architecture/) 定义 Go BFF ↔ Next.js 契约 |

## 文档总览

| 文件 | 用途 |
|---|---|
| [`01_railway_topology.md`](./01_railway_topology.md) | Railway 4 服务拓扑：Go BFF / Next.js / Postgres / RabbitMQ；环境变量、健康检查、构建命令 |
| [`02_postgres_schema.md`](./02_postgres_schema.md) | Postgres 数据模型：raw_messages / sports / matches / markets / outcomes / settlements / cancels / rollbacks / subscriptions / producer_health |
| [`03_modules.md`](./03_modules.md) | Go 包划分：把指引 16 业务域映射到 `internal/feed`、`internal/catalog`、`internal/odds`、`internal/settlement`、`internal/recovery`、`internal/health`、`internal/bff` |
| [`04_amqp_integration.md`](./04_amqp_integration.md) | FeedConstruct RMQ + WebAPI 真实集成：Token、队列、GZIP、Snapshot、订阅 |
| [`05_acceptance_checklist.md`](./05_acceptance_checklist.md) | 后端最小合格验收清单（映射上传指引 §4 全部 16 项） |

## 服务拓扑（4 个 Railway service）

```
┌───────────────┐     WebSocket + REST     ┌────────────────┐
│  Next.js Web  │ ───────────────────────▶ │   Go BFF       │
│ (frontend/)   │ ◀─────────────────────── │ (backend/)     │
└───────────────┘     (snapshot + stream)  └────────┬───────┘
                                                    │ AMQP / HTTPS
                              ┌─────────────────────┼─────────────────────┐
                              ▼                     ▼                     ▼
                     ┌────────────────┐  ┌────────────────┐  ┌──────────────────────┐
                     │  RabbitMQ      │  │  Postgres      │  │ FeedConstruct WebAPI │
                     │ (Railway plug) │  │ (Railway plug) │  │  (外部, 凭证注入)     │
                     └────────────────┘  └────────────────┘  └──────────────────────┘
                       ▲ replay/local         ▲ raw_messages
                       │                      │ + 领域表
                  FeedConstruct RMQ
                  (外部, 真实数据源)
```

| Railway 服务 | 镜像/构建 | 端口 | 关键环境变量 |
|---|---|---|---|
| `bff` | `backend/Dockerfile` | `$PORT` (Railway) | `DATABASE_URL`, `RABBITMQ_URL`, `FC_PARTNER_ID`, `FC_RMQ_USER`, `FC_RMQ_PASS`, `FC_RMQ_HOST`, `FC_API_BASE`, `FC_API_USER`, `FC_API_PASS`, `FEED_MODE`, `LOG_LEVEL` |
| `web` | `frontend/Dockerfile` | `$PORT` | `NEXT_PUBLIC_BFF_HTTP`, `NEXT_PUBLIC_BFF_WS` |
| `postgres` | Railway Postgres plugin | 5432 | 自动注入 `DATABASE_URL` |
| `rabbitmq` | Railway RabbitMQ plugin（或自部署 image `rabbitmq:3.13-management`） | 5672 / 15672 | 自动注入 `RABBITMQ_URL`；用作本地 replay/snapshot replay 的中转，**不是** FeedConstruct 外部 RMQ |

> **注意**：FeedConstruct 的 RMQ（`odds-stream-rmq-stage.feedstream.org:5673`）是**外部数据源**，由 BFF 直接连接消费；Railway 自带的 RabbitMQ 服务是**内部消息总线**，用于 BFF 内部解耦消费者/处理器、本地 replay 与多副本扩展。两者不要混淆。

## 与 `docs/07_frontend_architecture/` 的对应

| 后端模块 | 上传指引业务域 | 前端模块 |
|---|---|---|
| `internal/feed` (M01) | 连接接入 + 原始消息 | M01 实时数据通道 + M02 事件分发器 |
| `internal/catalog` (M03/M04) | 赛事主数据 | M03 体育目录 + M04 赛事与赛程 |
| `internal/odds` (M05/M06/M07) | 赔率 + 停投状态 | M05 盘口与赔率 + M06 盘口状态机 + M07 停投 |
| `internal/settlement` (M08/M09) | 结算 + 取消 + 回滚 | M08 结算与取消 + M09 回滚 + M14 我的投注 |
| `internal/recovery` (M10) | 恢复补偿 | M10 恢复补偿 |
| `internal/health` (M15) | 监控治理 | M15 健康与降级提示 |
| `internal/bff` | 分发层 | M01 实时数据通道（WebSocket）、所有 REST 快照端点 |

## 落地顺序（强制）

1. 锁定 [`01_railway_topology.md`](./01_railway_topology.md) 的服务拓扑与环境变量
2. 锁定 [`02_postgres_schema.md`](./02_postgres_schema.md) 的领域表和迁移命名
3. 锁定 [`04_amqp_integration.md`](./04_amqp_integration.md) 的 FeedConstruct 契约引用
4. 按模块顺序 `feed → storage → recovery → catalog → odds → settlement → health → bff` 编写 BDD 空测试文件并请用户确认
5. 用户确认后补齐正式测试与最小实现，每步对照 [`05_acceptance_checklist.md`](./05_acceptance_checklist.md)

## 不做什么

- 不在 BFF 中做任何 UI 渲染；不在前端中调用 FeedConstruct
- 不复刻 16 项业务域的全部细节；只覆盖上传指引 §4 验收所需的最小路径
- 不引入额外消息中间件（Kafka 等）；不引入 ORM（直接 `pgx`）
- 不在本仓库提交任何真实供应商凭证；所有凭证通过 Railway 环境变量注入
