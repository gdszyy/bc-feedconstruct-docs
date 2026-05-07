---
title: BetGuard Notes
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=betGuard_notes
current_loc: betGuard
location: betGuard_notes
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# BetGuard Notes

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=betGuard_notes`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `betGuard_notes` |

## 文档正文
This section outlines important notes regarding BetGuard and describes how to handle specific scenarios to ensure error-free integration.

- UnbookedObject for BetGuard Partners

When a MarketType or any other level such as Sport, Region, Competition, Match is unbooked, OddsFeed sends an [UnbookedObject](../documentation?currentLoc=oddsFeedRmqAndWebApi&location=unBookObject) update for that specific level.

BetGuard Partners are expected to process this update immediately and close the corresponding offer on their side without delay.

No bet requests should be sent for the unbooked MarketType or any unbooked level, ensuring the closure is handled promptly at the exact level where the unbooking occurred.

This approach applies consistently across all levels, including Sport, Region, Competition, Match, and MarketType.

```
{

"Command": "UnBookedObject",

"Objects": [

{

"ObjectId": 10606,

"ObjectTypeId": 5,

"IsLive": false,

"CompetitionId": 18262443,

"Id": 4

}

]

}
```

- Implementation Logic for the mentioned example

- **Identification** – In the example above, ObjectId: 10606 refers to the MarketType (indicated by ObjectTypeId: 5).
- **Scope** – This specific MarketType is unsubscribed for CompetitionID: 18262443 for the Prematch offer (IsLive: false).
- **Partner Action** – Upon receipt, the Partner must close the offer for that MarketType for either Live or Prematch, depending on the IsLive flag.
- **Warning** – If the Partner fails to close the offer and submits a bet on an unbooked Market Type, the bet may **not** be rejected due to the unbook action of Market Type. It may be accepted unless other validation criteria triggers a rejection.
