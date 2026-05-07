---
title: BetPlaced Request Sample
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placed_request_sample
current_loc: betGuard
location: bet_placed_request_sample
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# BetPlaced Request Sample

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placed_request_sample`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_placed_request_sample` |

## 文档正文
BetPlaced Request Sample

```
{

"AuthToken":"your_client_security_token",

"TS":1461670530,

"TransactionId":34234324,

"BetId":123456,

"Amount":123.00,l

"BonusBetAmount":10.00,

"Created":"2016-04-26T11:35:30.0543787Z",

"BetType":1,

"SystemMinCount":null,

"TotalPrice":2.500,

"Selections":[{

"SelectionId":42343,

"SelectionName":"P1",

"MarketTypeId":333,

"MarketName":"Match Result",

"MatchId":55,

"MatchName":"Barcelona - Real Madrid",

"MatchStartDate":"2016-04-26T11:35:30.0543787Z",

"RegionId":22,

"RegionName":"Spain",

"CompetitionId":44,

"CompetitionName":"La Liga",

"SportId":1,

"SportName":"Football",

"Price":2.5

}],

"Hash":"hash_code_generated_from_request_fields"

}
```

Where hash source string is:

**AuthToken**your\_client\_security\_token**TS**1461670530**TransactionId**34234324**BetId**123456**Amount**123.00**Created**2016-04-26T11:35:30.0543787Z**BetType**1**TotalPrice**2.500**your\_shared\_key**
