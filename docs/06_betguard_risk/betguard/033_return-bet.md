---
title: ReturnBet
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=return_bet
current_loc: betGuard
location: return_bet
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# ReturnBet

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=return_bet`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `return_bet` |

## 文档正文
ReturnBet

The **ReturnBet** endpoint **marks a bet in the Accepted state as returned**. This call is used when a Partner needs to return (void) a bet due to a specific reason.

**Important:** Only bets in the **Accepted** state can be returned using this endpoint. If the bet has already been resulted or its outcome has been determined, the call will be rejected.

This allows the Partner to handle exceptional cases where a placed bet must be voided and refunded before it is settled.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/PartnerAPI/ReturnBet Method: POST .../api/en/290/Bet/ReturnBet | FilterBetReturnModel | Success---------ResponseWrapper { StatusCode=”0”, Data={} } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- FilterBetReturnModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| BetId | long | Mandatory | Bet Id |
| RequestHash | string | Mandatory | required for partner verification |
| ExternalId | long | Optional | Partner’s unique identifier of the bet, **if specified, BetId should be 0.** |
