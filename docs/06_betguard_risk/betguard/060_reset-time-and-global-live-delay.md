---
title: Reset Time and Global Live Delay
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=reset_time_and_global_live_delay
current_loc: betGuard
location: reset_time_and_global_live_delay
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Reset Time and Global Live Delay

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=reset_time_and_global_live_delay`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `reset_time_and_global_live_delay` |

## 文档正文
Reset Time and Global Live Delay

- Reset Player Live Limits After minutes and Reset Player PreMatch Limits After minutes

Reset Player Limits After minutes – an option, which allows resetting client’s limits (from selection liability) after the individual set time (minutes) for Live and Prematch.

Example:

If the match is in prematch, and the client has expired its limit, the player can repeat the bet and place a bet on a selection only after 10 minutes. If the match is live – after 2 minutes.

- Configuration reset time (e.g. live - 2min., prematch - 10min.) can be set up based on the Client`s Sportsbook Profile. For High Risk clients can be used 20 min. in live and 60 min. in prematch, for VIP clients 1 min. in live and 3 min. in prematch.

- Global Live Delay

Global live delay is an option, which supplements interval (seconds) after clicking “place bet” and bet acceptance actual time. This helps to prevent the cases, when a client can see the event earlier (e.g. from the stadium), than it is shown on the website.

There is an option, which allows us to set individual Live Delay time depending on the client`s sportsbook profile.

Example:

For VIP clients can be set default negative -2 seconds live delay, which will help them to place bets faster without additional loading time. Meanwhile, for Late bet clients can get additional +5 seconds live delay, which will reduce the risk of Late bets on the event.
