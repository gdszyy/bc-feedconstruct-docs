# Railway 拓扑与服务清单

本文件锁定 4 个 Railway 服务的构建命令、启动命令、健康检查与环境变量。所有内容必须与仓库 `railway.json`、`backend/railway.json`、`frontend/railway.json` 保持一致；如有冲突，以仓库根目录的 `railway.json` 为准。

## 1. 服务清单

| 服务名 | 来源 | 构建器 | 启动命令 | 健康检查路径 |
|---|---|---|---|---|
| `bff` | `backend/` | `DOCKERFILE` (`backend/Dockerfile`) | `/app/bin/bffd` | `GET /healthz` |
| `web` | `frontend/` | `DOCKERFILE` (`frontend/Dockerfile`) | `node server.js` (Next.js standalone) | `GET /api/health` |
| `postgres` | Railway Postgres plugin | — | — | Railway 自带 |
| `rabbitmq` | Railway RabbitMQ plugin（或 `rabbitmq:3.13-management`） | — | — | Railway 自带；管理页 `/api/overview` |

> 部署顺序：`postgres` → `rabbitmq` → `bff` → `web`。Railway 环境变量引用：在 `bff` 的 `DATABASE_URL` 设为 `${{Postgres.DATABASE_URL}}`；在 `bff` 的 `RABBITMQ_URL` 设为 `${{RabbitMQ.RABBITMQ_URL}}`。

## 2. 环境变量

### 2.1 `bff` 服务

| 变量 | 必填 | 说明 |
|---|---|---|
| `PORT` | 自动 | Railway 注入；BFF 必须监听该端口 |
| `DATABASE_URL` | ✅ | Postgres 连接串，形如 `postgres://user:pass@host:5432/db?sslmode=require` |
| `RABBITMQ_URL` | ✅ | 内部 RabbitMQ；用于内部 fanout 与 replay |
| `FEED_MODE` | ✅ | `live`（连真实 FeedConstruct）或 `replay`（用 `raw/json` 重放）。生产 = `live` |
| `FC_API_BASE` | live 必填 | FeedConstruct WebAPI base，例 `https://odds-stream-api-stage.feedstream.org:8070` |
| `FC_API_USER` | live 必填 | WebAPI 用户名 |
| `FC_API_PASS` | live 必填 | WebAPI 密码 |
| `FC_RMQ_HOST` | live 必填 | FeedConstruct RMQ host:port，例 `odds-stream-rmq-stage.feedstream.org:5673` |
| `FC_RMQ_USER` | live 必填 | RMQ 用户名 |
| `FC_RMQ_PASS` | live 必填 | RMQ 密码 |
| `FC_PARTNER_ID` | live 必填 | PartnerID，用于绑定 `P{PartnerID}_live` 与 `P{PartnerID}_prematch` |
| `FC_RMQ_TLS` | 选 | 默认 `true`；非 5673 端口可关 |
| `RECOVERY_INITIAL` | 选 | 启动时是否触发全量 snapshot；默认 `true` |
| `LOG_LEVEL` | 选 | `debug`/`info`/`warn`/`error`，默认 `info` |
| `WS_ALLOWED_ORIGINS` | 选 | 逗号分隔，例 `https://web.up.railway.app,http://localhost:3000` |

### 2.2 `web` 服务

| 变量 | 必填 | 说明 |
|---|---|---|
| `PORT` | 自动 | Railway 注入 |
| `NEXT_PUBLIC_BFF_HTTP` | ✅ | BFF REST base，例 `https://bff.up.railway.app` |
| `NEXT_PUBLIC_BFF_WS` | ✅ | BFF WebSocket，例 `wss://bff.up.railway.app/ws` |

### 2.3 凭证管理原则

- **所有 `FC_*` 变量只在 Railway 环境注入**，禁止写入仓库（包括 `.env`、`docker-compose.yml`、文档）
- 仓库内 `.env.example` 只列变量名与占位说明，不写真实值
- 任何 PR 不得包含 `FC_API_PASS`、`FC_RMQ_PASS` 的真实值；CI 应配置 secret scanning

## 3. 健康检查与就绪

| 端点 | 含义 |
|---|---|
| `GET /healthz` | 进程存活；不依赖外部资源 |
| `GET /readyz` | Postgres 可连、内部 RabbitMQ 可连、（live 模式下）FeedConstruct RMQ 心跳近 30s 内、最近 alive 消息时间 |
| `GET /metrics` | Prometheus 文本格式（消息量、卡赛、producer 健康、recovery 次数） |

## 4. 构建与启动

### 4.1 `backend/Dockerfile`（要点）

```
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/bffd ./cmd/bffd
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/bffd /app/bin/bffd
COPY --from=build /out/migrate /app/bin/migrate
COPY migrations /app/migrations
ENTRYPOINT ["/app/bin/bffd"]
```

启动时 `bffd` 应先尝试运行迁移（或调用 `/app/bin/migrate up`）再监听端口。

### 4.2 `frontend/Dockerfile`（要点）

Next.js standalone 输出，多阶段构建；只暴露 `$PORT`，不暴露 `3000` 默认。

## 5. 与本地开发的对照

| 场景 | 本地 | Railway |
|---|---|---|
| 进程编排 | `docker-compose up`（4 服务齐备） | 4 个 Railway service |
| Postgres | `postgres:16-alpine` 容器 | Railway Postgres plugin |
| RabbitMQ | `rabbitmq:3.13-management` 容器 | Railway RabbitMQ plugin |
| FeedConstruct | `FEED_MODE=replay` + `raw/json` 样本 | `FEED_MODE=live` + 真实凭证 |

## 6. 验收

| 序号 | 验收项 |
|---|---|
| T1 | 在 Railway 创建 4 服务后，无人工拼接环境变量即可连通（用 `${{Postgres.DATABASE_URL}}` 引用） |
| T2 | `bff` 启动日志包含迁移成功、RabbitMQ 连接成功、（live 模式）FeedConstruct Token 获取成功 |
| T3 | `GET /readyz` 在所有依赖就绪前返回 503，全部就绪后返回 200 |
| T4 | `web` 启动后能从 `NEXT_PUBLIC_BFF_HTTP/WS` 拉取快照并接入 WebSocket |
| T5 | 任何 `FC_*` 变量未配置时，`FEED_MODE=live` 启动应在 30s 内 fail-fast 并打印缺失变量名 |
