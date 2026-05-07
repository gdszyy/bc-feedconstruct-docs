---
title: Message Syntax
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=messageSyntax
current_loc: oddsFeedRmqAndWebApi
location: messageSyntax
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Message Syntax

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=messageSyntax`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `messageSyntax` |

## 文档正文
Message Syntax

All messages are in JSON format. All responses are compressed as GZIP.

"GetDataSnapshot" command is an indicator in case of which snapshots must be gotten. The reason of this
command may be the release of platform.
