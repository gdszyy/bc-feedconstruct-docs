---
title: Hash Calculation
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=hash_calculation
current_loc: betGuard
location: hash_calculation
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Hash Calculation

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=hash_calculation`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `hash_calculation` |

## 文档正文
Hash Calculation

The **RequestHash** is generated using the following steps:

- **Gather all request parameters**(properties of the request object) excluding:

- **RequestHash** itself.
- Properties with  **null** values.

- **Serialize** the remaining parameters into a JSON string.

- For numeric (decimal) values, at least two decimal points must be preserved.

- Valid: 1.10, 1.20, 1.00.
- Invalid: 1.1, 1.

- **Encrypt** the JSON string using the HMAC-SHA256 algorithm.

- Use the Partner’s **secret key** as the encryption key.

- Return the resulting lowercase hexadecimal string as the **RequestHash**.
