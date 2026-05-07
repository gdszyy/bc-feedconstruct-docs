---
title: Hash Calculation
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_hash_calculation
current_loc: betGuard
location: partner_api_hash_calculation
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Hash Calculation

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_hash_calculation`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_hash_calculation` |

## 文档正文
Hash Calculation

- Take the specified parameters for the API call (see table below).
- Exclude empty fields and the **Hash** itself.
- Concatenate **parameter names and values** in the defined order.
- Append the shared key to the end.
- Generate an **MD5 hash** of the resulting string.
- Add the resulting (lowercase) hash as the **Hash** parameter in the request/response.

The final hash must be in **lowercase**. Uppercase hashes will fail validation.
