# FeedConstruct RMQ + WebAPI 集成契约

本文件**不重复**仓库已有的官方文档；只锁定本工程要消费的范围与字段映射。详细字段定义请引用：

- 接入方式：[`docs/01_data_feed/rmq-web-api/002_access.md`](../01_data_feed/rmq-web-api/002_access.md)
- 集成步骤：[`docs/01_data_feed/rmq-web-api/003_integrationguide.md`](../01_data_feed/rmq-web-api/003_integrationguide.md)
- 对象字段：[`docs/01_data_feed/rmq-web-api/004_objectsdescriptions.md`](../01_data_feed/rmq-web-api/004_objectsdescriptions.md)
- 体育/区域/赛事：`005_sport.md`、`006_region.md`、`007_competition.md`、`008_match.md`
- 盘口/赔率：`009_market.md`、`010_selection.md`、`013_markettype.md`、`014_selectiontype.md`
- 取消通知：`015_voidnotification.md`
- 订阅与 Booking：`016_partnerbooking.md`、`018_book.md`、`021_unbookobject.md`、`022_bookobject.md`
- 消息语法：`032_messagesyntax.md`
- WebAPI 方法：`033_webmethods.md`
- 对象作为 feed 一部分：`034_objectsaspartfeed.md`
- ObjectType ID 表：`035_objecttypes.md`
- 错误码：`037_resultcodes.md`、`031_responseerror.md`

## 1. 真实凭证与连接

| 资源 | Stage（文档示例） | 备注 |
|---|---|---|
| RMQ host | `odds-stream-rmq-stage.feedstream.org:5673` | 生产值由 FeedConstruct 提供后填入 `FC_RMQ_HOST` |
| WebAPI base | `https://odds-stream-api-stage.feedstream.org:8070` | 同上，填入 `FC_API_BASE` |
| 连接限制 | `MaxConnectionCount: 4`，`MaxChannelCount: 64` | BFF 默认 1 连接、2 channel（live + prematch）；预留扩展 |
| 队列 | `P{PartnerID}_live`、`P{PartnerID}_prematch` | `FC_PARTNER_ID` 注入 |
| IP 白名单 | 必须在 FeedConstruct 后台登记 Railway 出口 IP | 生产前通过 `curl https://ipinfo.io/ip` 在 BFF 启动日志打印出口 IP，便于报备 |
| TLS | RMQ 5673 = AMQPS；WebAPI 8070 = HTTPS | `FC_RMQ_TLS=true` 默认开启 |
| GZIP | RMQ delivery body 与 WebAPI response 都是 GZIP | `feed.Consumer` 必须先 `gzip.NewReader` 解压 |
| 编码 | JSON | 走 `encoding/json` 即可 |

## 2. WebAPI Token

来源：[`002_access.md`](../01_data_feed/rmq-web-api/002_access.md)、[`033_webmethods.md`](../01_data_feed/rmq-web-api/033_webmethods.md)。

| 行为 | 实现要点 |
|---|---|
| 获取 Token | 启动后立即调用 Token 方法（用户名/密码 → token） |
| 缓存 | `webapi.Client` 内部缓存，提前 1 小时刷新（防止 24h 边界） |
| 失败 | 指数退避（2s/4s/8s/16s/30s 上限），并在 `producer_health.detail` 标记 `webapi_token_failed` |
| 安全 | Token 与凭证只在内存；不写日志；不写 DB |

## 3. 启动同步（指引 §3 的"Synchronisation of the general Data"）

启动 BFF 后，依次拉取并 upsert：

1. `Sports` → `sports`
2. `Regions` → `regions`
3. `Competitions` → `competitions`
4. `MarketTypes` → `i18n_market_types`（描述层；可延后）
5. `SelectionTypes` → 同上
6. `EventTypes` → 同上
7. `Periods` → 同上

每天定时一次（cron 由 `recovery.Coordinator` 触发），保险起见。

## 4. Snapshot

来源：`033_webmethods.md` → `DataSnapshot`。

| 场景 | 调用方式 |
|---|---|
| 启动时 | `DataSnapshot(isLive=true)` + `DataSnapshot(isLive=false)`，**不带** `getChangesFrom` |
| 断流 < 1h | 带 `getChangesFrom = lastReceivedAt - safetyWindow` |
| 断流 ≥ 1h | 不带 `getChangesFrom`，全量重拉 |

`recovery.Coordinator` 通过 `recovery_jobs.scope = 'startup' | 'product'` 落库；执行结果写 `status`/`detail`。

## 5. 消息分发表

将 RMQ 投递的对象映射为内部消息类型（与 `docs/08_backend_railway/02_postgres_schema.md` 的 `raw_messages.message_type` 对齐）：

| FeedConstruct 对象 / 通知 | ObjectType ID（参考） | 内部 message_type |
|---|---|---|
| Sport | 1 | `catalog.sport` |
| Region | 2 | `catalog.region` |
| Competition | 3 | `catalog.competition` |
| Match | 4 | `fixture` 或 `fixture_change`（按是否新建判断） |
| MarketType | 5 | `catalog.market_type` |
| Market | 13 | `odds_change`（含 selection 时）/ `bet_stop`（status 变更） |
| Selection | 16 | 通常作为 Market 的子对象一起到来 |
| VoidNotification | — | `bet_cancel`（`VoidAction=1`）/ `rollback_cancel`（`VoidAction=2`） |
| Settlement / Resulting | — | `bet_settlement` |
| Book / Unbook | — | `subscription.book` / `subscription.unbook` |
| GetDataSnapshot 信号 | — | `recovery.snapshot_request` |

> **注意**：FeedConstruct 没有 Sportradar UOF 的 `producer alive`；存活通过"消息时间戳推进"判断（`health.Reporter` 维护 `last_message_at`）。如果 N 秒（默认 30s）无任何消息且 RMQ 心跳异常，则置 `producer_health.is_down=true`。

## 6. Replay 模式（`FEED_MODE=replay`）

用于本地与 PoC：

| 来源 | 用法 |
|---|---|
| `raw/json/bc_feedconstruct_docs_scrape.json` | 仅作字段示例 |
| 未来增加的 `raw/messages/*.json.gz` | `feed.Replayer` 按文件名时间戳顺序投递到内部 RabbitMQ exchange，复用同一处理链路 |

`replay` 模式下不连 FeedConstruct，但写入的 `raw_messages.source` 标记为 `replay.<file>`，确保审计可区分。

## 7. 不在范围

- 不实现 BetGuard 真实下注（`docs/06_betguard_risk/betguard/`）；本期仅在 BFF 侧暴露占位 REST `POST /api/v1/bets/preview`，调用 BetGuard 的能力延后阶段
- 不实现 i18n 描述同步（市场名/选项名翻译）；只保留表骨架
- 不实现 horse racing / racing-info 专门链路（`024_racinginfo.md`）

## 8. 验收

| 序号 | 验收项 |
|---|---|
| A1 | `webapi.Client` 在 23h 内自动刷新 Token，且并发刷新只发一次请求 |
| A2 | RMQ 消息从 GZIP 解压、入 `raw_messages`、fanout、被对应 handler 消费的延时 < 200ms（本地测） |
| A3 | 重复投递相同消息（同 `event_id` + `ts_provider`），`raw_messages` 只新增 0 行（唯一约束命中） |
| A4 | Snapshot 启动时把 live + prematch 全量数据 upsert 完成，并在 `recovery_jobs` 留 `success` |
| A5 | VoidNotification `VoidAction=2` 必须落 `rollbacks` 并把对应 `cancels.rolled_back_at` 写入 |
