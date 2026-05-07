---
title: BetResulted Retry Logic
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_resulted_retry_logic
current_loc: betGuard
location: bet_resulted_retry_logic
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# BetResulted Retry Logic

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_resulted_retry_logic`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_resulted_retry_logic` |

## 文档正文
BetResulted Retry Logic

When [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) call fails due to communication or network, Feedcosntruct tries to resend the message. Each retry is performed after **N minutes where N is the retry count**. After 5 retries the message remains in the queue with “No Answer” status.
