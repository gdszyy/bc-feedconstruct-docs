---
title: Competition
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=competition
current_loc: feedSocketApi
location: competition
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Competition

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=competition`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `competition` |

## 文档正文
Competition

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | int | Unique Id |
| Name\* | string | Competition name |
| NameId\* | int | Translation Id |
| SportId\* | int | [Sport](/documentation?currentLoc=feedSocketApi&location=sport) |
| RegionId\* | int | [Region](/documentation?currentLoc=feedSocketApi&location=region) |
| IsTeamsReversed\* | bool | Indicates home team position (second - for american competitions) |
| LiveDelay | int | LiveDelay |
