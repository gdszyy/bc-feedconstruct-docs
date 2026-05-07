---
title: Selection
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=selection
current_loc: oddsFeedRmqAndWebApi
location: selection
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Selection

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=selection`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `selection` |

## 文档正文
Selection

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | long | Unique Id |
| Name\* | string | Selection name Templates: • {t1},{t2} - as HomeTeam, AwayTeam • {h},{-h} - as Handicap (As market handicap)\* • {hv},{av} - as HomeValue/AwayValue  **\*Note:** To calculate the value of the Selection, take the Handicap field from the Market Object. For example, if the Handicap field in the Market Object is -0.5, and the Selections are represented as "{t1} ({-h})" and "{t2} ({h})", the value interpretation would be as follows: For "{t1} ({-h})", the value would be -h-0.5, which is --0.5, which is the same as 0.5 For "{t2} ({h})", the value would be -0.5.' |
| NameId\* | int | Translations Id |
| Kind | string | Sport level unique text |
| Order\* | int | Order |
| Price | double | Price with applied margin |
| OriginalPrice\* | double | 100% Price |
| HomeValue | int |  |
| SelectionTypeId | int |  |
| AwayValue | int |  |
| IsSuspended\* | bool |  |
| IsVisible\* | bool |  |
| Handicap | double |  |
| Outcome\* | int | Selection Result: • 0 - Not Resulted • 1 - Place\* • 2 - Return • 3 - Lost • 4 - Won • 5 - WinReturn • 6 - LoseReturn **\*Note:** For Golf sport kind, if the partner doesn’t use “Each Way” he should ignore Outcome “Place=1”, and it’s equivalent to “Lost=3”  **Important:** The field is not available with the BetGuard service. |
| NonRunner | int | Non runner (Racing Sports only) |
