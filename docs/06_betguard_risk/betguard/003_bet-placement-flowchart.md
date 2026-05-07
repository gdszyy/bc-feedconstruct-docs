---
title: Bet Placement Flowchart
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placement_flowchart
current_loc: betGuard
location: bet_placement_flowchart
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Bet Placement Flowchart

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placement_flowchart`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_placement_flowchart` |

## 文档正文
Bet Placement Flowchart

1Bet PlacementCreateBet → Validation Checks → Accepted / Rejected

2Bet Resulting & OutcomesEvent outcome → BetResulted → Retry Queue

3Cashout & ReturnsManual ReturnBet · User-Initiated Cashout · Finalizing

API Call Initiated By Partner/End User

**Note:** Partners have full autonomy over the UI design regarding the availability of ReturnBet and Cashout actions for the End User. The Partner`s system is responsible for triggering the corresponding API calls based on user actions. For Cashout functionality, the Partner`s backend must determine the appropriate method to invoke: MarkBetAsCashout or CheckAndMarkBetAsCashout.
