---
title: Hash Parameter
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_hash_parameter
current_loc: betGuard
location: partner_api_hash_parameter
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Hash Parameter

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_hash_parameter`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_hash_parameter` |

## 文档正文
Hash Parameter

The **Hash** parameter validates the authenticity of the message and ensures that the data has not been modified during transmission.

- It is calculated from a predefined list of request fields (excluding **Hash** itself), in a strict order, and concatenated with the Partner’s **shared key**.
- Each Partner (Operator backend) has a unique shared key issued by FeedConstruct.
