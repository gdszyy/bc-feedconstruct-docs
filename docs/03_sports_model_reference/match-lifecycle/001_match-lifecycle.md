---
title: Match Lifecycle
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=match_lifecycle_for_live
current_loc: match_lifecycle_for_live
location: root
top_category: MATCH LIFECYCLE
product_line: 体育数据模型参考
business_domain: 体育数据模型参考
scraped_at: 2026-05-07T08:49:13.195Z
---

# Match Lifecycle

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=match_lifecycle_for_live`。

| 字段 | 值 |
|---|---|
| 一级分类 | MATCH LIFECYCLE |
| 产品线 | 体育数据模型参考 |
| 业务域 | 体育数据模型参考 |
| currentLoc | `match_lifecycle_for_live` |
| location | `root` |

## 文档正文
Match Lifecycle

- Match lifecycle for Live games

- - Match Creation
  - - Once match is created in the system, it will appear in the feed.
    - MatchStatus = 0
    - LiveStatus = 0
- - Mark As Available for Live
  - - Match can be directly created or marked later as available for Live.
    - MatchStatus = 0
    - LiveStatus = 1
- - Mark As Ready to Start
  - - Indicates that match is ready to go to Live, but is not started yet.
    - MatchStatus = 0
    - LiveStatus = 1
    - IsLive = true
- - Match Start
  - - Actual Match Start
    - MatchStatus = 1
    - LiveStatus = 1
    - IsLive = true
    - IsStarted = true
- - Match End
  - - MatchStatus = 2
    - LiveStatus = 2

- Match lifecycle for Prematch games

- - Match Creation
  - - Once match is created in the system, it will appear in the feed
    - MatchStatus = 0
    - LiveStatus = 0
- - Match Start
  - - Actual Match Start
    - MatchStatus = 1
    - LiveStatus = 1
- - Match End
  - - MatchStatus = 2
    - LiveStatus = 2

- Match end can be identified through two primary criteria:

- A "Command" : "MatchUpdate", "Type" : "Match" update where "LiveStatus" is 2 and "MatchStatus" is 2 (completed), or "LiveStatus" is 3 and "MatchStatus" is 3 (cancelled).
- A "Command" : "MatchUpdate", "Type" : "MatchStat" update with "EventType" as "Finished."

Both criteria are equally important. In cases where a "Completed" message (from "Command" : "MatchUpdate",
"Type" : "Match" with "LiveStatus" : 2, "MatchStatus" : 2) is received but a "Finished" event is not, the
"Completed" message should be processed as the match end.

- Canceled matches

Match cancelation is indicated by the MatchStatus = 3 and LiveStatus = 3.

- Removal of Events from Offer

## Removal of Events from PreMatch Offer:

Remove all Events (Match, Market Type) from the PreMatch Offer when the planned StartDate of the event is reached, and no updates for MatchStart have been received. Remove Events (Match, Market Type) from the PreMatch Offer when PreMatch Booking has been changed to Unbooked during the PreMatch offer.

## Removal of Events from Live Offer:

Remove all Events (Match, Market Type) from the Live Offer when no Live Booking has been made on them, or if a booking has been changed to Unbooked during the Live Offer.
