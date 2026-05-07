---
title: CreateBet Request Sample
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=createBet_request_sample
current_loc: betGuard
location: createBet_request_sample
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# CreateBet Request Sample

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=createBet_request_sample`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `createBet_request_sample` |

## 文档正文
CreateBet Request Sample

```
{

"AcceptTypeId": 0,

"Amount": 2.00,

"AuthToken": "supermegaauthtoken",

"BetType": 1,

"ClientDetail": {

"CurrencyId": "EUR",

"ExternalId": "uniqueid",

"Login": "player_login"

},

"Currency": "EUR",

"ExternalId": 1234567890,

"IsEachWay": false,

"OddType": 0,

"RequestHash": "hash_code_generated_from_request_fields",

"Selections": [{

"IsBanker": false,

"Price": 1.26,

"SelectionId": 5907544813

}

]

}
```
