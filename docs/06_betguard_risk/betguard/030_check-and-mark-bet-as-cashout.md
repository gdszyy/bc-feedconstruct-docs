---
title: CheckAndMarkBetAsCashout
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=check_and_mark_bet_as_cashout
current_loc: betGuard
location: check_and_mark_bet_as_cashout
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# CheckAndMarkBetAsCashout

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=check_and_mark_bet_as_cashout`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `check_and_mark_bet_as_cashout` |

## 文档正文
CheckAndMarkBetAsCashout

The **CheckAndMarkBetAsCashout** endpoint **validates the cashout availability** for a given bet, compares the odd sent by the partner in the request with the current odd, and compares with the Possible Win, and if validation passes, **marks it as cashed out**.

The validation includes **Live Delay**.

**Note:**  The platform does **not calculate the cashout amount**. This is expected to be handled by the Partner.

**Note:**  The Cashout amount cannot be higher than Possible Win. In case the provided Amount will be higher, it will result in Cashout rejection.

This endpoint provides a safer alternative to [MarkBetAsCashout](../documentation?currentLoc=betGuard&location=mark_bet_as_cashout), ensuring that the system only updates the bet status when cashout is actually permitted.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/PartnerAPI/CheckAndMarkBetAsCashout Method: POST .../api/en/290/Bet/CheckAndMarkBetAsCashout | CashoutRequestModel | Success---------ResponseWrapper { StatusCode=”0”, Data={} } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- CashoutRequestModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| BetId | long | Mandatory | Bet Id |
| RequestHash | string | Mandatory | required for partner verification |
| Price | decimal | Mandatory | Indicates Cashout Amount |
| Selections | Array of CashoutSelectionModel | Mandatory | Array of selections |

- CashoutSelectionModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| SelectionId | long | Mandatory | SelectionId |
| Odd | decimal | Mandatory | Current odd of the selection at the time of cashout |
