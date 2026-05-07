---
title: UnBookObject
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=unBookObject
current_loc: oddsFeedRmqAndWebApi
location: unBookObject
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# UnBookObject

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=unBookObject`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `unBookObject` |

## 文档正文
UnBookObject

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| ObjectId \* | int | Unbook object Id |
| ObjectTypeId \* | int | Unbooked object type id: • Sport = 1, • Region = 2, • Competition = 3, • Match = 4 • MarketType = 5 |
| IsLive \* | bool | Indicating unbooking type (Live or Prematch) |
| Id | int | Indicates sportId when: • ObjectTypeId = 2 (Region), • ObjectTypeId = 5 (MarketType) |
| CompetitionId | int | Only for MarketType (ObjectTypeId = 5) |
