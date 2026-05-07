---
title: Market
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=market
current_loc: oddsFeedRmqAndWebApi
location: market
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Market

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=market`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `market` |

## 文档正文
Market

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | long | Unique Id |
| Name\* | string | Market name Templates: • {s},{sw} - as Sequence • {p},{pw} - as PointSequence • {h} - as Handicap |
| NameId\* | int | Translation Id |
| Handicap\* | double | Handicap Value |
| Selections\* | object array | List of [Selection](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=selection) objects |
| Sequence\* | int | See [here](faq?page=received_data) |
| PointSequence\* | int | See [here](faq?page=received_data) |
| IsSuspended\* | bool | Market suspended |
| MatchId\* | int | Parent match Id |
| IsVisible\* | bool | Market visible |
| MarketTypeId\* | int | Market type Id |
| CashOutAvailable | bool | Is cash out available or not\* |
| HomeScore | int | Home score (note:{hv} in market name) |
| AwayScore | int | Away score (note:{av} in market name) |
| ObjectVersion | string | Version number |
| IsSelectionsOrderedByPrice | bool | Selections ordered by Price |
| StartingPriceEnabled | bool | Starting price is enabled (Racing Sports only) |
| ExtraInfo | object | [MarketExtraInfo](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketExtraInfo) object (Racing Sports only), |

**Note:** The Cashout flag in Odds Feed is supposed to be a trigger for the partners to make the cashout
available/unavailable in case they have implemented this feature. The Odds Feed doesn't provide
Cashout logic, and the feature must be implemented by a partner on his own platform.
