---
title: MarketTypeBooking
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketTypeBooking
current_loc: oddsFeedRmqAndWebApi
location: marketTypeBooking
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# MarketTypeBooking

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketTypeBooking`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `marketTypeBooking` |

## 文档正文
MarketTypeBooking

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id \* | int | MarketType Id |
| SportId \* | int | Sport Id |
| IsLive \* | bool | Indicating booking/unbooking type (Live or Prematch) |
| IsUnSubscribed \* | bool | Indicating booking status |
| CompetitionId | int | Competition ID will be available only if an action is done on a competition level. |
