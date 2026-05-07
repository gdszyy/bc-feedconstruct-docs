---
title: MarketType
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketType
current_loc: oddsFeedRmqAndWebApi
location: marketType
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# MarketType

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketType`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `marketType` |

## 文档正文
MarketType

Represents the Type of the market.

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | int | Unique Id |
| Name\* | string | Market name |
| NameId\* | int | Translation Id |
| Kind\* | string | Unique text Id |
| IsHandicap\* | bool | Handicap market |
| IsOverUnder\* | bool | Over/Under market |
| SportId\* | int | Sport Id |
| SelectionCount\* | int | Number of selections in market |
| IsDynamic\* | bool | Is Dynamic market |
| TypeFlag | int | • 1 - for Live and Prematch • 2 - only for Prematch • 3 - only for Live • null - not defined |
| LiveDelay | int | LiveDelay |
| DisplayOrder | int | Display order |
