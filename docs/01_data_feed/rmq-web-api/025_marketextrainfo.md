---
title: MarketExtraInfo
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketExtraInfo
current_loc: oddsFeedRmqAndWebApi
location: marketExtraInfo
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# MarketExtraInfo

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketExtraInfo`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `marketExtraInfo` |

## 文档正文
MarketExtraInfo

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| EachWayPlace | int | Each way place |
| EachWayK | int | Each way price |
| EarlyPrices | bool | Early prices |
| Created | datetime | Created date time |
| IsPartnerTermsEnabled | bool | Is partnerTerms enabled |
| ManualTerms | object array | List of [ManualEachWayTerms](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=manualEachWayTerms) |
| Dividents | object array | List of [BetDivident](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=betDivident) |
