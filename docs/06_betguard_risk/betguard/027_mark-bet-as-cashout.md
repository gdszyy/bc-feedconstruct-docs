---
title: MarkBetAsCashout
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=mark_bet_as_cashout
current_loc: betGuard
location: mark_bet_as_cashout
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# MarkBetAsCashout

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=mark_bet_as_cashout`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `mark_bet_as_cashout` |

## 文档正文
MarkBetAsCashout

The **MarkBetAsCashout** endpoint **marks a bet as cashed out without performing validation on odd or on cashout availability**.

This call is intended for use cases where the cashout is processed and fully checked by the Partner, the Partner checks the availability of the Cashout at the moment itself and needs to be reflected in the system for tracking or settlement purposes.

**Use with caution:** since no validation on the odd and cashout availability is performed, the caller is fully responsible for ensuring the correctness of the request.

**Note:** The Cashout amount cannot be higher than Possible Win. In case the provided Amount will be higher, it will result in Cashout rejection.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/PartnerAPI/MarkBetAsCashout Method: POST .../api/en/290/Bet/MarkBetAsCashout | CashoutModel | Success---------ResponseWrapper { StatusCode=”0”, Data={} } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- CashoutModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| BetId | long | Mandatory | Bet Id |
| Price | decimal | Mandatory | Indicates Cashout Amount |
| RequestHash | string | Mandatory | required for partner verification |
