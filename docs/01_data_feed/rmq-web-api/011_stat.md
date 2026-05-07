---
title: Stat
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=stat
current_loc: oddsFeedRmqAndWebApi
location: stat
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Stat

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=stat`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `stat` |

## 文档正文
Stat

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Id | int | The Id is not used in case of 'Command': 'MatchStat'. It can be used in order to identify which event was canceled with 'Command': 'MatchUpdate', 'Type': 'MatchStat' with 'IsCanceled': true. |
| EventId\* | int | Match Id |
| IsCanceled | bool | Indicates if current event was cancelled |
| IsTimeout | bool |  |
| EventTimeUtc | datetime | Time of the event |
| EventType | string | Event type Name |
| EventTypeId | int | Event type Id |
| Side | int | • 1 - Home Team • 2 - Away Team **\*Note:** For Period event types the value of this field isn't actual and should be ignored. |
| CurrentMinute | int | Current minute of the match ( only for time sports ) |
| MatchLength | int | Length of the match in minutes |
| Score | string | Match Score |
| CornerScore | string |  |
| YellowcardScore | string |  |
| RedcardScore | string |  |
| ShotOnTargetScore | string |  |
| ShotOffTargetScore | string |  |
| DangerousAttackScore | string |  |
| AcesScore | string |  |
| DoubleFaultScore | string |  |
| SportKind\* | int |  |
| PeriodScore | string |  |
| SetScore | string |  |
| Set1Score | string |  |
| Set2Score | string |  |
| Set3Score | string |  |
| Set4Score | string |  |
| Set5Score | string |  |
| Set6Score | string |  |
| Set7Score | string |  |
| Set8Score | string |  |
| Set9Score | string |  |
| Set10Score | string |  |
| GameScore | string | Game Score(Tennis) |
| Server | int | Current team’s serve |
| Info | string | Game short info |
| RemainingTime | TimeSpan |  |
| Period | int | Current period |
| PeriodCount\* | int | Period count |
| PeriodLength | int | Period Length |
| SetCount | int | Total set count |
| PenaltyScore | string | Score of Penalty Shootouts in Football, and in Ice Hockey |
| FreeKickScore | string |  |
| ExtraTimeScore | string |  |
| Set1YellowCardScore | string |  |
| Set2YellowCardScore | string |  |
| Set1CornerScore | string |  |
| Set2CornerScore | string |  |
| Set1RedCardScore | string |  |
| Set2RedCardScore | string |  |
| AdditionalMinutes | int | Additional minutes |
| TeamId | int | Goalscorer Id |
| HomeShirtColor | string |  |
| AwayShirtColor | string |  |
| HomeShortsColor | string |  |
| AwayShortsColor | string |  |
| TypeId | int | Type Id |
| CurrentShot | int | Current shot number (only for Basketball Shots sport) |
| CurrentOverNumber | int | Current over number (only for Cricket sport) |
| CurrentDeliveryNumber | int | Current delivery number (only for Cricket sport) |
| RunsPoint | int | Runs point (only for Cricket sport) |
| CurrentLeg | int | Current leg(only for Darts sport) |
