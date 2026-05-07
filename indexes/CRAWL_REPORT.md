# BC / FeedConstruct OddsFeed 文档抓取报告

本报告记录本次文档镜像的抓取范围、处理方式与索引结果。抓取入口为 FeedConstruct OddsFeed Documentation 的公开访客入口：`https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary`。[1]

## 抓取范围

页面首次访问会进入登录页，但可以通过 **Sign In As Guest** 进入公开文档。当前访客权限下可见并可渲染的一级文档模块包括 RABBIT MQ、TCP SOCKET、TRANSLATIONS SOCKET API、TRANSLATIONS RMQ & WEB API、MATCH LIFECYCLE、EVENT TYPES、MARKET TYPES、SPORTS、SPORTSBOOK NOTES、SPORTS RULES、ODDS CONVERSION 与 BETGUARD。

| 范围项 | 结果 |
|---|---|
| 原始抓取页面数 | 261 |
| 按 `currentLoc + location` 去重后的 Markdown 文档数 | 249 |
| 一级文档模块数 | 12 |
| 业务域分类数 | 6 |
| 原始 JSON | `raw/json/bc_feedconstruct_docs_scrape.json` |
| 原始 HTML | `raw/html/` |

尝试直接访问未在访客菜单中展示的 `currentLoc=user_guide&location=mng_booking` 时，页面回落至默认 `oddsFeedRmqAndWebApi/summary`，因此本镜像未包含 User Guide 等非访客可访问文档。

## 业务域映射

| 业务域 | 覆盖模块 | Markdown 数量 |
|---|---|---:|
| 数据源服务 / OddsFeed | RABBIT MQ、TCP SOCKET | 76 |
| 翻译数据服务 | TRANSLATIONS SOCKET API、TRANSLATIONS RMQ & WEB API | 20 |
| 体育数据模型参考 | MATCH LIFECYCLE、EVENT TYPES、MARKET TYPES、SPORTS | 5 |
| 体育博彩业务规则 | SPORTSBOOK NOTES、SPORTS RULES | 82 |
| 赔率与计算 | ODDS CONVERSION | 1 |
| 投注风控服务 / BetGuard | BETGUARD | 65 |

## 生成物

本仓库以 `docs/` 保存规范化 Markdown，并以 `indexes/` 保存多维索引。`SEARCH_INDEX.json` 是机器可读索引，包含标题、业务域、模块、路径、关键词命中与正文预览。`AGENTS.md` 与 `.cursor/rules/` 用于 AI 代理按需加载最小上下文。

## References

[1]: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary "FeedConstruct OddsFeed Documentation"
