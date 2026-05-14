# Postgres 数据模型

本文件锁定后端使用的领域表与迁移文件命名。所有表必须出现在 `backend/migrations/` 下，按 `NNN_<name>.sql` 升序运行。

## 1. 设计原则

| 原则 | 落实 |
|---|---|
| 原始消息先落库再处理 | `raw_messages` 是消息处理入口；任何业务表都通过 `raw_message_id` 反查 |
| 状态领域化 | 盘口、赛事、结算、取消独立枚举；不在代码中临时判断 |
| 可恢复 | snapshot/recovery 任务都写入 `recovery_jobs`，可幂等重放 |
| 可幂等 | 业务表对 `(raw_message_id, business_key)` 加唯一约束，重复消息不会写两次 |
| 可追溯 | 状态变更追加 `*_history`，不就地覆盖 |

## 2. 迁移文件计划

| 文件 | 内容 |
|---|---|
| `001_init.sql` | `raw_messages`、枚举类型、扩展 `pgcrypto` |
| `002_catalog.sql` | `sports`、`regions`、`competitions`、`matches`、`fixture_changes` |
| `003_markets.sql` | `markets`、`outcomes`、`market_status_history` |
| `004_settlement.sql` | `settlements`、`cancels`、`rollbacks` |
| `005_subscriptions.sql` | `subscriptions`、`subscription_events` |
| `006_recovery.sql` | `recovery_jobs`、`producer_health`、`metrics_counters` |

## 3. 表定义（要点）

### 3.1 `raw_messages`

```
id              uuid primary key default gen_random_uuid()
received_at     timestamptz not null default now()
source          text not null            -- 'rmq.live' / 'rmq.prematch' / 'webapi.snapshot'
routing_key     text                      -- RMQ routing key
queue           text                      -- 例 'P123_live'
message_type    text not null            -- 'odds_change' / 'bet_stop' / 'bet_settlement' / 'bet_cancel' / 'rollback' / 'fixture' / 'fixture_change' / 'alive' / 'snapshot_complete'
event_id        text                      -- match id 等
product_id      smallint
sport_id        int
ts_provider     timestamptz               -- 供应商发出时间
payload         jsonb not null            -- 解 GZIP 后的 JSON
raw_blob        bytea                     -- 选：原始 GZIP 字节，便于复盘
processed_at    timestamptz
process_error   text
unique (source, message_type, coalesce(event_id, ''), ts_provider)
```

> 唯一键覆盖**幂等**（指引验收 11）：相同消息重复投递不会写两条。

### 3.2 `sports` / `regions` / `competitions` / `matches`

| 表 | 关键字段 |
|---|---|
| `sports` | `id int pk`, `name text`, `is_active bool`, `updated_at timestamptz` |
| `regions` | `id int pk`, `sport_id int`, `name text`, `is_active bool` |
| `competitions` | `id int pk`, `region_id int`, `sport_id int`, `name text`, `is_active bool` |
| `matches` | `id bigint pk`, `sport_id int`, `competition_id int`, `name text`, `home text`, `away text`, `start_at timestamptz`, `is_live bool`, `status text`, `last_event_id text`, `updated_at timestamptz` |
| `fixture_changes` | `id bigserial pk`, `match_id bigint`, `change_type text`, `old jsonb`, `new jsonb`, `raw_message_id uuid`, `received_at timestamptz` |

`matches.status` 枚举：`not_started` / `live` / `ended` / `closed` / `cancelled` / `postponed`。**防回退**（验收 12）由代码层强制：`ended/closed/cancelled` 不被 `live` 覆盖。

### 3.3 `markets` / `outcomes` / `market_status_history`

| 表 | 关键字段 |
|---|---|
| `markets` | `match_id bigint`, `market_type_id int`, `specifier text`, `status text`, `group_id int`, `updated_at timestamptz`, pk = `(match_id, market_type_id, specifier)` |
| `outcomes` | `match_id bigint`, `market_type_id int`, `specifier text`, `outcome_id int`, `odds numeric(12,4)`, `is_active bool`, `updated_at timestamptz`, pk = `(match_id, market_type_id, specifier, outcome_id)` |
| `market_status_history` | `id bigserial pk`, `match_id`, `market_type_id`, `specifier`, `from_status`, `to_status`, `raw_message_id uuid`, `changed_at timestamptz` |

`markets.status` 枚举：`active` / `suspended` / `deactivated` / `settled` / `cancelled` / `handed_over`。

### 3.4 `settlements` / `cancels` / `rollbacks`

```
settlements (
  id bigserial pk,
  match_id bigint,
  market_type_id int,
  specifier text,
  outcome_id int,
  result text,                  -- win / lose / void / half_win / half_lose
  certainty smallint,
  void_factor numeric(5,4),
  dead_heat_factor numeric(5,4),
  raw_message_id uuid,
  settled_at timestamptz,
  rolled_back_at timestamptz,   -- 非空表示已回滚
  unique (match_id, market_type_id, specifier, outcome_id, settled_at)
)

cancels (
  id bigserial pk,
  match_id bigint,
  market_type_id int,
  specifier text,
  void_reason text,
  void_action smallint,         -- 1=void, 2=unvoid（对齐 VoidNotification）
  superceded_by uuid,
  from_ts timestamptz,
  to_ts timestamptz,
  raw_message_id uuid,
  cancelled_at timestamptz,
  rolled_back_at timestamptz
)

rollbacks (
  id bigserial pk,
  target text not null,         -- 'settlement' | 'cancel'
  target_id bigint not null,
  raw_message_id uuid,
  applied_at timestamptz default now()
)
```

### 3.5 `subscriptions`

```
subscriptions (
  match_id bigint pk,
  product text not null,         -- 'live' / 'prematch'
  status text not null,          -- 'requested'/'subscribed'/'unsubscribed'/'expired'/'failed'
  requested_at timestamptz,
  subscribed_at timestamptz,
  released_at timestamptz,
  last_event_id text,
  reason text
)

subscription_events (
  id bigserial pk,
  match_id bigint,
  from_status text,
  to_status text,
  reason text,
  occurred_at timestamptz default now()
)
```

### 3.6 `recovery_jobs` / `producer_health` / `metrics_counters`

```
recovery_jobs (
  id bigserial pk,
  scope text not null,           -- 'startup' / 'product' / 'event' / 'stateful' / 'fixture_change'
  product text,                  -- 'live' / 'prematch'
  match_id bigint,
  requested_at timestamptz default now(),
  started_at timestamptz,
  finished_at timestamptz,
  status text not null,          -- 'queued' / 'running' / 'success' / 'failed' / 'rate_limited'
  attempt smallint default 0,
  next_retry_at timestamptz,
  detail jsonb
)

producer_health (
  product text pk,               -- 'live' / 'prematch'
  last_alive_at timestamptz,
  last_message_at timestamptz,
  is_down bool,
  detail jsonb
)

metrics_counters (
  name text pk,
  value bigint default 0,
  updated_at timestamptz default now()
)
```

## 4. 与上传指引验收的对应

| 验收项 | 对应表/约束 |
|---|---|
| 1 连接 | `producer_health` 记录 alive 时间 |
| 2 消息留痕 | `raw_messages`，含 `payload` + 可选 `raw_blob` |
| 3 消息覆盖 | `raw_messages.message_type` 枚举字符串覆盖 9 种 |
| 4 主数据 | `sports`/`regions`/`competitions`/`matches` |
| 5 赔率 | `markets`/`outcomes` |
| 6 停投 | `markets.status` + `market_status_history` |
| 7 结算 | `settlements`（含 certainty/void_factor/dead_heat_factor） |
| 8 取消 | `cancels`（含 void_reason/from_ts/to_ts/superceded_by） |
| 9 回滚 | `rollbacks` + `settlements.rolled_back_at`/`cancels.rolled_back_at` |
| 10 恢复 | `recovery_jobs` |
| 11 幂等 | `raw_messages` 唯一约束 + `settlements`/`outcomes` 复合主键 |
| 12 防回退 | 代码层断言 + `matches.status` 枚举 |
| 13 订阅 | `subscriptions` + `subscription_events` |
| 14 描述数据 | `markets.market_type_id` → 描述层（独立字典；本期可放 `i18n_market_types` 表，可延后） |
| 15 监控 | `producer_health` + `metrics_counters` + `/metrics` |
| 16 数据治理 | 后台清理 job 按 `received_at` 滚动删 `raw_messages` 与历史表（保留窗口由配置决定） |
