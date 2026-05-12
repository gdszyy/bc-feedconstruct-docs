# 分层架构

## 总体拓扑

```
┌────────────────────────────────────────────────────────────┐
│                       供应商数据源                          │
│        (FeedConstruct OddsFeed AMQP / Web API / TCP)        │
└───────────────▲────────────────────────────────────▲────────┘
                │                                    │
                │ AMQP / REST                        │ BetGuard API
                │                                    │
┌───────────────┴────────────────────────────────────┴────────┐
│                  Go BFF（不在本规划范围内详述）              │
│  连接层 / 接入层 / 处理层 / 补偿层 / 订阅层 / 描述层 / 监控 │
│                                                              │
│  对前端暴露：                                                │
│   - WebSocket /ws/v1/stream（订阅式实时事件）                │
│   - REST /api/v1/*（快照、目录、描述、投注、我的投注）       │
└───────────────▲────────────────────────────────────▲────────┘
                │ WebSocket                          │ HTTPS
                │                                    │
┌───────────────┴────────────────────────────────────┴────────┐
│                Next.js 前端（本规划核心）                    │
│                                                              │
│  ┌─────────── Transport Layer (M01) ─────────────────────┐  │
│  │ WS Client / Reconnect / Heartbeat / Backoff           │  │
│  │ HTTP Client / Auth / Retry / Idempotency-Key          │  │
│  └──────────────────────┬────────────────────────────────┘  │
│  ┌──────────── Dispatcher (M02) ─────────────────────────┐  │
│  │ envelope.type → handler；保留 event_id / correlation  │  │
│  └──────────────────────┬────────────────────────────────┘  │
│  ┌────── Domain Stores（领域状态，每个模块一个）─────────┐  │
│  │ M03 Catalog  M04 Match  M05 Markets  M06 MarketStatus │  │
│  │ M07 BetStop  M08 Settle M09 Rollback  M10 Recovery    │  │
│  │ M11 Subscr.  M12 i18n   M13 BetSlip   M14 MyBets      │  │
│  │ M15 Health   M16 Telemetry                            │  │
│  └──────────────────────┬────────────────────────────────┘  │
│  ┌──────────── Selectors / Derivations ──────────────────┐  │
│  │ 派生「可下注」「显示赔率」「停盘提示」「降级横幅」    │  │
│  └──────────────────────┬────────────────────────────────┘  │
│  ┌──────── React Components / Next.js Pages ─────────────┐  │
│  │ App Router + Server Components 渲染骨架                │  │
│  │ Client Components 订阅 store 渲染实时部分              │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## Next.js 项目目录建议（仅作规划，不强制）

```
apps/web/
├── app/
│   ├── (lobby)/page.tsx                  # P01 首页/大厅
│   ├── sport/[sportId]/page.tsx          # P02 体育/联赛
│   ├── match/[matchId]/page.tsx          # P03 赛事详情
│   ├── live/page.tsx                     # P04 滚球大厅
│   ├── my-bets/page.tsx                  # P06 我的投注
│   └── settings/page.tsx                 # P08 设置
├── components/
│   ├── bet-slip/                         # P05 投注单组件树
│   ├── market/                           # 市场表/赔率按钮
│   ├── match/                            # 比分/状态徽章
│   └── system/                           # 降级横幅、stale 标记
├── stores/                               # 每个模块一个 store（zustand / redux 后续定）
│   ├── catalog.ts                        # M03
│   ├── match.ts                          # M04
│   ├── markets.ts                        # M05
│   ├── marketStatus.ts                   # M06
│   ├── betStop.ts                        # M07
│   ├── settlement.ts                     # M08
│   ├── rollback.ts                       # M09
│   ├── recovery.ts                       # M10
│   ├── subscription.ts                   # M11
│   ├── descriptions.ts                   # M12
│   ├── betSlip.ts                        # M13
│   ├── myBets.ts                         # M14
│   ├── health.ts                         # M15
│   └── telemetry.ts                      # M16
├── services/
│   ├── transport/
│   │   ├── ws.ts                         # M01 WS Client
│   │   └── http.ts                       # M01 HTTP Client
│   ├── dispatcher.ts                     # M02
│   └── api/                              # REST 调用封装
├── domain/
│   ├── state-machines/                   # Match/Market/Bet/Subscription/Connection FSM
│   └── selectors/                        # 派生 selector
└── test/
    ├── unit/                             # 模块单测
    ├── integration/                      # store + dispatcher 集成测
    └── e2e/                              # Playwright E2E
```

> 真正的工程目录由前端仓库决定；本规划只给出**职责分层**约束，不绑定具体目录命名。

## Go BFF → 前端契约的最小切面

后端必须提供以下能力，前端才能落地：

1. WebSocket 单一接入端点，支持基于 token 的订阅授权
2. envelope 统一结构（见 [`03_backend_data_contract.md`](./03_backend_data_contract.md)）
3. 重连时返回 `replay_from` 游标，由后端按缺口回放（前端不重放原始消息）
4. 静态描述按 `version` 提供 ETag 缓存
5. 投注、查询通过 REST，前端不直接调用 BetGuard

## 渲染策略（Next.js）

| 内容类型 | 渲染策略 |
|---|---|
| 体育/联赛目录、玩法描述、规则页 | RSC + ISR（按 `version` revalidate） |
| 赛事列表（赛前为主） | RSC 首屏 + Client 增量订阅 |
| 赛事详情（赔率/状态/比分） | Client Component 全量订阅 |
| 投注单 / 我的投注 | Client Component（含本地状态机） |
| 设置 / 偏好 | RSC + Server Actions |

## 不在本规划范围

- Go BFF 内部分层（连接 / 接入 / 处理 / 补偿 / 订阅 / 描述 / 监控）由后端文档承载
- 风控 / 派彩 / KYC 由 BetGuard 文档承载，前端只通过 BFF 间接消费
- 部署、CI/CD、灰度策略
