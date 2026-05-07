# AGENTS.md

本仓库是 BC / FeedConstruct OddsFeed 文档镜像仓库，面向 AI 代理与工程人员快速检索使用。进入仓库后应优先阅读本文件，然后根据任务目标选择对应索引。

## 入口策略

如果任务涉及接口接入、消息结构、对象字段或 SDK，请从 `indexes/NAVIGATION.md` 按文档站模块进入。如果任务涉及业务理解、系统集成分层或数据商能力梳理，请从 `indexes/BUSINESS_DOMAIN_INDEX.md` 进入。如果任务只提供关键词，如 BetGuard、RabbitMQ、Market、Settlement、SDK，请优先读取 `indexes/KEYWORD_INDEX.md` 或 `indexes/SEARCH_INDEX.json`。

## 模块路由表

| 任务类型 | 优先目录 |
|---|---|
| OddsFeed RMQ / Web API 对接 | `docs/01_data_feed/rmq-web-api/` |
| TCP Socket 数据源对接 | `docs/01_data_feed/tcp-socket-api/` |
| 翻译接口 | `docs/02_translations/` |
| 比赛、盘口、体育项目模型参考 | `docs/03_sports_model_reference/` |
| 结算、玩法规则、Sportsbook 规则 | `docs/04_sportsbook_rules/` |
| 赔率转换 | `docs/05_odds_math/` |
| 投注风控、下注、派彩、状态查询 | `docs/06_betguard_risk/betguard/` |

## 维护规则

本仓库是外部文档镜像。不要手工改写 `docs/` 下的外部文档正文，除非明确记录为本地注释或勘误。更新文档时应保留 `raw/json/` 的原始抓取文件，并重新生成 `indexes/` 下所有索引。
