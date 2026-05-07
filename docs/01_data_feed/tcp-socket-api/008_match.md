---
title: Match
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=match
current_loc: feedSocketApi
location: match
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Match

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=match`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `match` |

## 文档正文
Match

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id\* | int | Unique Id |
| CompetitionId\* | int | Competition Id |
| Date\* | datetime | Match start time |
| SportId | int | [Sport](/documentation?currentLoc=feedSocketApi&location=sport) |
| RegionId | int | [Region](/documentation?currentLoc=feedSocketApi&location=region) |
| LiveStatus\* | int | • NotAvailable = 0 • Available = 1 • Completed = 2 • Cancelled = 3 |
| MatchStatus | int | • NotStarted = 0 • Started = 1 • Completed = 2 • Cancelled = 3 |
| IsVisible\* | bool | [Integration notes](/documentation?currentLoc=feedSocketApi&location=integrationNotes) |
| IsSuspended\* | bool | [Integration notes](/documentation?currentLoc=feedSocketApi&location=integrationNotes) |
| IsLive\* | bool | [Matchlifecycle](/documentation?currentLoc=match_lifecycle_for_live) |
| IsStarted\* | bool | [Matchlifecycle](/documentation?currentLoc=match_lifecycle_for_live) |
| MatchMembers\* | object array | List of [MatchMember](/documentation?currentLoc=feedSocketApi&location=matchMember) objects |
| MarketsList | object array | List of [Market](/documentation?currentLoc=feedSocketApi&location=market) objects |
| Stat | object | [Stat](/documentation?currentLoc=feedSocketApi&location=stat) object |
| CancelReason | string | Reason for match canceling(Live only) |
| IsBooked\* | bool | Flag which indicates if match was booked for the exact time of the update |
| Info | string | Additional information. Free text |
| MatchInfo | string | Match note |
| MatchResults | object array | List of [Stat](/documentation?currentLoc=feedSocketApi&location=stat) objects |
| MatchShortInfo | string | Match short info |
| IsOutright | bool | This field indicates that match is outright. And for these matches it will be better to have implemented the following logic for frontend: if StartTime of match is passed, remove it from frontend |
| IsStatAvailable | bool | This field indicates that we have statistics for this match. If partner has integrated Statistics API, he can have statistics on his frontend |
| IsNeutralVenue | bool | Is neutral venue |
| ObjectVersion | string | Version number  **Note:** The changes in the below mentioned fields DO NOT always result on the Object Version update.  • IsLive • IsStarted • IsVisible • IsSuspended |
| ParentId | int | Parent match Id Details in [Sportsbook notes](/documentation?currentLoc=sportsBookNotes) |
| Participants | object array | List of [MatchMember](/documentation?currentLoc=feedSocketApi&location=matchMember) objects (Matchday Statistics only) |
| LiveDelay | int | [LiveDelay](/documentation?currentLoc=sportsBookNotes) |
| InformationSource | int | • TV = 0 • Scout = 1 If the Information Source is missing, it means the source is 'Other' |
| AutobookRuleId | int | If a match has been delivered based on a Booking Rule By time, the Rule ID will be mentioned |
