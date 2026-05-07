---
title: Partner API Security Checks
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_security_checks
current_loc: betGuard
location: partner_api_security_checks
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Partner API Security Checks

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_security_checks`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_security_checks` |

## 文档正文
Partner API Security Checks

To ensure secure communication and verify the identity of both sender and recipient, the main part of the Partner API requests and responses include two mandatory parameters:

- **Hash** – A checksum used to validate data integrity.
- **TS** – A timestamp used to prevent replay attacks and detect expired requests.
