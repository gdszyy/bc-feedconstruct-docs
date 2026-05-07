---
title: TS Parameter (Timestamp)
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=ts_parameter_timestamp
current_loc: betGuard
location: ts_parameter_timestamp
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# TS Parameter (Timestamp)

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=ts_parameter_timestamp`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `ts_parameter_timestamp` |

## 文档正文
TS Parameter (Timestamp)

The **TS** (timestamp) parameter protects against request/response replay attacks by validating the freshness of each message.

- The value of **TS** is the **UNIX timestamp** (seconds since January 1, 1970, 00:00:00 GMT).
- FC backend time is synchronized using **NTP;** the Partner must do the same.
- On every request/response, both sides must verify the **TS** value.

**Validation Rule:**

If the timestamp is **older than 20 seconds**, the message is considered **expired**and must be **rejected**.

No further processing should occur; instead, return an appropriate error code and message.
