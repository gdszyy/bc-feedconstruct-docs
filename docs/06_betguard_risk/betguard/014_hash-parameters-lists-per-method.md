---
title: Hash Parameters Lists per Method
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=hash_parameters_lists_per_method
current_loc: betGuard
location: hash_parameters_lists_per_method
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Hash Parameters Lists per Method

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=hash_parameters_lists_per_method`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `hash_parameters_lists_per_method` |

## 文档正文
Hash Parameters Lists per Method

| API Method API Method | Parameters Used in Hash Calculation (ordered) Parameters Used in Hash Calculation (ordered) |  |
| --- | --- | --- |
| [GetClientDetails](/documentation?currentLoc=betGuard&location=partner_api_get_client_details) | AuthToken, TS |
| [BetPlaced](/documentation?currentLoc=betGuard&location=partner_api_bet_placed) | AuthToken, TS, TransactionId, BetId, Amount, Created, BetType, SystemMinCount, TotalPrice |
| [BetResulted](/documentation?currentLoc=betGuard&location=bet_resulted) | AuthToken, TS, TransactionId, BetId, BetState, Amount |
| [Rollback](/documentation?currentLoc=betGuard&location=partner_api_rollback) | AuthToken, TS, TransactionId |



**No security checks** (e.g., hash or token validation) are required for the following methods of the Partner API:

- [Client](../documentation?currentLoc=betGuard&location=partner_api_client)
- [Bet Selection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection)

Push updates are sent by FeedConstruct to the endpoints, using the  **base URL provided by the Partner** during integration.

**Optional Security:**

If the Partner requires additional protection, FeedConstruct’s IP addresses can be provided upon request for  **IP whitelisting** on the Partner’s side.
