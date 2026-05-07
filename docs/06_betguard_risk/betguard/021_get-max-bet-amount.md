---
title: GetMaxBetAmount
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=get_max_bet_amount
current_loc: betGuard
location: get_max_bet_amount
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# GetMaxBetAmount

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=get_max_bet_amount`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `get_max_bet_amount` |

## 文档正文
GetMaxBetAmount

The  [GetMaxBetAmount](../documentation?currentLoc=betGuard&location=get_max_bet_amount) endpoint returns the **maximum allowable stake** for a specific selection or bet combination. This call helps Partners determine the betting limits before placing a bet, ensuring that submitted stakes do not exceed the platform’s predefined thresholds.

Same as in case of [CreateBet](../documentation?currentLoc=betGuard&location=createBet), if the mandatory fields of **ClientDetailModel** are not sent, **FC** calls [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details) to identify the client and validate applicable stake limits. After getting the response, FC does

- Client identification.
- Risk assessment.
- Max stake calculation and response delivery.

The processes described in 4, 5, 6, 10 points of the Process Flow - Bet Placement via Betguard API are actual here

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/Bet/GetMaxBetAmount Method: POST .../api/en/290/Bet/GetMaxBetAmount | Request BetModel (see in CreateBet part) | Success---------ResponseWrapper { StatusCode=”0”, Data=MaxBetResponse } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- MaxBetResponse

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| MaxBetAmount | decimal | The max stake that can be used to place the bet |
