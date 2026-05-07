# BC / FeedConstruct OddsFeed 文档镜像与索引

本仓库保存从 FeedConstruct OddsFeed 文档站公开访客入口抓取并规范化的 BC 文档。BC 作为体育博彩数据商，其文档主要覆盖 **数据源服务** 与 **BetGuard 投注风控服务**，同时包含翻译接口、体育模型参考、Sportsbook 规则与赔率转换等配套资料。[1]

> 本仓库只镜像当前访客权限下可以访问和渲染的公开文档。尝试直接访问未在访客菜单中展示的 User Guide 时，页面回落到默认 OddsFeed RMQ & Web API Summary，因此未将其纳入公开镜像范围。

## 仓库内容概览

| 指标 | 数量 |
|---|---:|
| 原始抓取页面 | 261 |
| 去重后 Markdown 文档 | 249 |
| 一级文档模块 | 12 |
| 业务域分类 | 6 |

## 目录结构

| 路径 | 用途 |
|---|---|
| `docs/` | 按业务域与模块拆分后的 Markdown 文档 |
| `indexes/NAVIGATION.md` | 按文档站一级模块组织的导航索引 |
| `indexes/BUSINESS_DOMAIN_INDEX.md` | 按数据源、风控、翻译、规则等业务域组织的索引 |
| `indexes/KEYWORD_INDEX.md` | 按关键词命中的快速检索索引 |
| `indexes/CRAWL_REPORT.md` | 抓取范围、权限边界与统计报告 |
| `indexes/SEARCH_INDEX.json` | 机器可读全文检索元数据 |
| `raw/html/` | 每个页面的规范化原始 HTML 备份 |
| `raw/json/` | 浏览器抓取原始 JSON |
| `AGENTS.md` | AI 代理阅读本仓库时的入口与路由规范 |
| `docs/07_architecture/` | 基于公开文档综合整理的架构拆分、集成建议与专题设计 |

## 推荐阅读入口

一般应先阅读 [`indexes/BUSINESS_DOMAIN_INDEX.md`](indexes/BUSINESS_DOMAIN_INDEX.md)，然后根据目标能力进入对应目录。若要对接 BC 数据源服务，建议从 `docs/01_data_feed/rmq-web-api/` 和 `docs/01_data_feed/tcp-socket-api/` 开始。若要对接投注风控服务，建议从 `docs/06_betguard_risk/betguard/` 开始。若要先理解数据源与分控投注模块的系统边界，建议阅读 [`docs/07_architecture/001_bc_oddsfeed_betguard_split.md`](docs/07_architecture/001_bc_oddsfeed_betguard_split.md)。

| 业务域 | 目录 | 说明 |
|---|---|---|
| 数据源服务 / OddsFeed | `docs/01_data_feed/` | RabbitMQ、Web API 与 TCP Socket 数据接入 |
| 翻译数据服务 | `docs/02_translations/` | 翻译 Socket API、翻译 RMQ 与 Web API |
| 体育数据模型参考 | `docs/03_sports_model_reference/` | Match Lifecycle、Event Types、Market Types、Sports |
| 体育博彩业务规则 | `docs/04_sportsbook_rules/` | Sportsbook Notes 与 Sports Rules |
| 赔率与计算 | `docs/05_odds_math/` | Odds Conversion |
| 投注风控服务 / BetGuard | `docs/06_betguard_risk/` | BetGuard 流程、安全校验、下注、派彩、状态查询与报表接口 |
| 架构设计专题 | `docs/07_architecture/` | OddsFeed 数据源与 BetGuard 分控投注模块拆分等综合设计 |

## References

[1]: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary "FeedConstruct OddsFeed Documentation"
