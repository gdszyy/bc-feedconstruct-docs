# BFF (Go) — 体育博彩数据源接入服务

> 本目录是 [`docs/08_backend_railway/`](../docs/08_backend_railway/) 的实现骨架。当前阶段处于 **BDD "空测试"** 阶段：仅提供包结构、Given-When-Then 注释和测试函数名，**没有正式测试与生产逻辑**。任何代码补充必须先经 `AskUserQuestion` 确认。

## 当前阶段

| 阶段 | 状态 |
|---|---|
| 行为建模（Gherkin 注释） | ✅ |
| 空测试文件 | ✅ |
| 用户确认是否进入正式测试 | ⏳ 等待 |
| 正式测试代码 | ⏸ 暂缓 |
| 生产实现 | ⏸ 暂缓 |

## 目录

| 路径 | 说明 |
|---|---|
| `cmd/bffd` | 主进程入口（占位 main） |
| `cmd/migrate` | 迁移 runner（占位 main） |
| `internal/config` | 环境变量加载（占位） |
| `internal/storage` | pgx pool + repository（占位） |
| `internal/feed` | M01 接入：FeedConstruct RMQ + replay |
| `internal/catalog` | M03/M04 主数据 |
| `internal/odds` | M05/M06/M07 赔率 / 状态 / 停投 |
| `internal/settlement` | M08/M09 结算 / 取消 / 回滚 |
| `internal/recovery` | M10 恢复补偿 |
| `internal/subscription` | 订阅生命周期 |
| `internal/health` | M15 监控 |
| `internal/webapi` | FeedConstruct WebAPI Token + Snapshot 客户端 |
| `internal/bff` | 分发层：WebSocket + REST |
| `migrations` | Postgres 迁移（按 NNN_ 顺序） |

## 本地启动

```bash
cp ../.env.example ../.env.local
docker compose --env-file ../.env.local up --build bff postgres rabbitmq
```

`FEED_MODE=replay` 默认不连外部 FeedConstruct；切换为 `live` 必须在 Railway 注入 `FC_*` 全部变量（参见 [`../docs/08_backend_railway/01_railway_topology.md`](../docs/08_backend_railway/01_railway_topology.md)）。

## 测试

```bash
go test ./...
```

当前所有测试为 BDD placeholder，运行成功但断言为空；这是**预期状态**。
