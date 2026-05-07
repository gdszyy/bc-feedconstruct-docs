---
title: BetResulted Request Sample
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_resulted_request_sample
current_loc: betGuard
location: bet_resulted_request_sample
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# BetResulted Request Sample

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_resulted_request_sample`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_resulted_request_sample` |

## 文档正文
BetResulted Request Sample

```
{

"TransactionId":34234325,

"Amount":307.50,

"BetId":123456,

"BetState":4,

"AuthToken":"your_client_security_token",

"TS":1461671373,

"Hash":"hash_code_generated_from_request_fields"

}
```

Where hash source string is:

**AuthToken**your\_client\_security\_token**TS**1461671373**TransactionId**34234325**BetId**123456**BetState**4**Amount**307.50**your\_shared\_key**
