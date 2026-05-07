---
title: AuthToken; Binding with Currency
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=authToken_binding_with_currency
current_loc: betGuard
location: authToken_binding_with_currency
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# AuthToken; Binding with Currency

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=authToken_binding_with_currency`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `authToken_binding_with_currency` |

## 文档正文
AuthToken; Binding with Currency

Requests sent via the Partner API use the [AuthToken](../documentation?currentLoc=betGuard&location=authToken) from the original [CreateBet](../documentation?currentLoc=betGuard&location=createBet) call. In some scenarios - especially with [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted)  - the [AuthToken](../documentation?currentLoc=betGuard&location=authToken) may be significantly old:

- For live bets, this may be several **hours** after placement.
- For pre-match bets, this may be several **days or even weeks** later.

This delay occurs because some bet outcomes are determined long after the original bet was placed. The Partner must be able to **distinguish such delayed transactions** and ensure they are still **accepted and processed** properly.

Additionally, each [AuthToken](../documentation?currentLoc=betGuard&location=authToken) is permanently bound to the currency returned in the [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details) response.

**Important:**The Partner **must not change the currency** associated with an active [AuthToken](../documentation?currentLoc=betGuard&location=authToken). Doing so will corrupt financial transactions and may result in data inconsistencies or processing failures.
