---
title: Calendar Match
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=calendar-match
current_loc: oddsFeedRmqAndWebApi
location: calendar-match
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Calendar Match

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=calendar-match`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `calendar-match` |

## 文档正文
Calendar Match

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | int | Match Id |
| Date\* | datetime | Match start time |
| CompetitionId\* | int | Competition Id |
| SportId | int | Sport Id |
| RegionId | int | Region Id |
| MatchStatus\* | int | • NotStarted = 0 • Started = 1 • Completed = 2 • Cancelled = 3 |
| IsStarted | bool |  |
| MatchMembers\* | object array | List of [MatchMember](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=matchMember) objects |
| PrematchBooked | bool | The flag indicates if the match is booked for Prematch (through Odds Feed portal, API or Backoffice). |
| LiveBooked | bool | The flag indicates if the match is booked for Live (through Odds Feed portal, API or Backoffice). |
| IsOutright | bool | This field indicates that match is outright. And for these matches it will be better to have implemented the following logic for frontend: if StartTime of match is passed, remove it from frontend |
| LiveStatus | int | • NotAvailable = 0 • Available = 1 • Completed = 2 • Cancelled = 3 |
| ParentId | int | Parent match Id |
| IsVisible | bool |  |
