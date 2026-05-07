---
title: AuthToken
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=authToken
current_loc: betGuard
location: authToken
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# AuthToken

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=authToken`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `authToken` |

## 文档正文
AuthToken

The  **AuthToken** is the end user’s (client’s) security token and is included in all the requests.

**Important:** Different authentication tokens may be sent by the partner with separate bets from the same end user. When processing the BetResulted call, the system uses the most recently received auth token for that user, which may differ from the token originally sent with the specific bet being resulted.
