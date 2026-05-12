# 体育博彩前端数据来源整体方案（规划骨架）

> 本目录是面向 **Go 后端 + React/Next.js 前端** 的体育博彩前端数据来源规划文档。前端**不直连**博彩数据源（AMQP/REST），所有实时与静态数据均通过后端 Go BFF（WebSocket + REST）下发。
>
> 本目录只包含**规划与契约骨架**，不包含实现代码。任何代码落地前必须按 `CLAUDE.md` 的 TDD/BDD 流程进行。

## 设计来源

本方案以仓库根目录上传的《体育博彩数据源接入说明指引》为输入，把 16 项数据源接入业务域映射为前端可消费的模块、状态机与页面骨架，并将仓库已有的 OddsFeed / BetGuard 文档作为契约依据。

## 文档总览

| 文件 | 用途 |
|---|---|
| [`01_principles.md`](./01_principles.md) | 设计原则：接入层只做接入、状态领域化、纠错可恢复、订阅有生命周期 |
| [`02_layered_architecture.md`](./02_layered_architecture.md) | 分层架构：Go BFF / Next.js App Router / Store / Service / UI |
| [`03_backend_data_contract.md`](./03_backend_data_contract.md) | 后端 → 前端契约：WebSocket 事件、REST 端点、消息 envelope |
| [`04_state_machines.md`](./04_state_machines.md) | Match / Market / Bet / Subscription / Connection 状态机 |
| [`05_acceptance_checklist.md`](./05_acceptance_checklist.md) | 前端最小合格验收清单（映射上传指引 §4） |

## 模块骨架（`modules/`）

按上传指引 §2 业务域映射的 16 个前端模块。每个模块文件包含：目的 / 数据来源 / 领域状态 / 关键组件 / 与其他模块依赖 / 未决问题 / 验收要点。

| 编号 | 模块 | 上传指引对应业务域 | 文件 |
|---|---|---|---|
| M01 | 实时数据通道 | 连接接入 | [M01_realtime_transport.md](./modules/M01_realtime_transport.md) |
| M02 | 事件分发器 | 原始消息 / 接入层 | [M02_event_dispatcher.md](./modules/M02_event_dispatcher.md) |
| M03 | 体育目录（Sport / Category / Tournament） | 赛事主数据 | [M03_sports_catalog.md](./modules/M03_sports_catalog.md) |
| M04 | 赛事与赛程（Match / Fixture / Schedule） | 赛事主数据 + fixture_change | [M04_match_and_fixture.md](./modules/M04_match_and_fixture.md) |
| M05 | 盘口与赔率（Markets & Odds） | 赔率数据 | [M05_markets_and_odds.md](./modules/M05_markets_and_odds.md) |
| M06 | 盘口状态机 | 停投状态 / 状态领域化 | [M06_market_status.md](./modules/M06_market_status.md) |
| M07 | 停投与赛事级停盘（Bet Stop） | 停投状态 | [M07_bet_stop.md](./modules/M07_bet_stop.md) |
| M08 | 结算与取消（Settlement / Cancel） | 结算状态 + 取消状态 | [M08_settlement_and_cancel.md](./modules/M08_settlement_and_cancel.md) |
| M09 | 回滚纠错（Rollback） | 回滚纠错 | [M09_rollback.md](./modules/M09_rollback.md) |
| M10 | 恢复补偿（前端缺口检测/快照） | 恢复补偿 | [M10_recovery.md](./modules/M10_recovery.md) |
| M11 | 订阅生命周期（关注 / 订阅 / 释放） | 订阅生命周期 | [M11_subscription.md](./modules/M11_subscription.md) |
| M12 | 静态描述与 i18n | 静态描述 | [M12_descriptions_i18n.md](./modules/M12_descriptions_i18n.md) |
| M13 | 投注单（Bet Slip） | 综合（停投/赔率/状态） | [M13_bet_slip.md](./modules/M13_bet_slip.md) |
| M14 | 我的投注（My Bets） | 结算/取消/回滚 | [M14_my_bets.md](./modules/M14_my_bets.md) |
| M15 | 健康与降级提示 | 监控治理 | [M15_health_and_degradation.md](./modules/M15_health_and_degradation.md) |
| M16 | 前端遥测与审计 | 监控治理 / 可追溯 | [M16_telemetry.md](./modules/M16_telemetry.md) |

## 页面骨架（`pages/`）

| 编号 | 页面 | 主要消费模块 | 文件 |
|---|---|---|---|
| P01 | 首页 / 大厅 | M03, M04, M05, M15 | [P01_home_lobby.md](./pages/P01_home_lobby.md) |
| P02 | 体育 / 联赛页 | M03, M04 | [P02_sport_hub.md](./pages/P02_sport_hub.md) |
| P03 | 赛事详情 | M04, M05, M06, M07, M12 | [P03_match_detail.md](./pages/P03_match_detail.md) |
| P04 | 滚球大厅 | M04, M05, M06, M10, M15 | [P04_live_inplay.md](./pages/P04_live_inplay.md) |
| P05 | 投注单 | M05, M06, M07, M13 | [P05_bet_slip.md](./pages/P05_bet_slip.md) |
| P06 | 我的投注 | M08, M09, M14 | [P06_my_bets.md](./pages/P06_my_bets.md) |
| P07 | 搜索与收藏 | M03, M04, M11 | [P07_search_and_favorites.md](./pages/P07_search_and_favorites.md) |
| P08 | 设置（语言 / 赔率格式 / 通知） | M12, M15 | [P08_settings.md](./pages/P08_settings.md) |

## 与仓库其它文档的关系

| 关心点 | 应先读 |
|---|---|
| 字段含义、消息结构 | [`indexes/NAVIGATION.md`](../../indexes/NAVIGATION.md) |
| 业务概念、能力梳理 | [`indexes/BUSINESS_DOMAIN_INDEX.md`](../../indexes/BUSINESS_DOMAIN_INDEX.md) |
| 投注 / 派彩 / 回滚契约 | [`docs/06_betguard_risk/betguard/`](../06_betguard_risk/betguard/) |
| 玩法 / 结算规则 | [`docs/04_sportsbook_rules/`](../04_sportsbook_rules/) |

## 落地顺序

按以下顺序推进，每一步都必须先完成上一步的契约固化：

1. 锁定 `03_backend_data_contract.md` 中的 WebSocket envelope 与 REST 端点列表（依赖后端确认）
2. 锁定 `04_state_machines.md` 中的 Match / Market / Bet 状态枚举
3. 按模块顺序 M01 → M02 → M10 → M03~M07 → M08/M09 → M11 → M12 → M13/M14 → M15/M16 编写 BDD 空测试文件并请用户确认
4. 用户确认后补齐正式测试与最小实现
5. 页面装配阶段按 P01~P08 推进
