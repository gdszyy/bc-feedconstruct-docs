---
title: ResendFailedTransfers
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=resend_failed_transfers
current_loc: betGuard
location: resend_failed_transfers
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# ResendFailedTransfers

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=resend_failed_transfers`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `resend_failed_transfers` |

## 文档正文
ResendFailedTransfers

Triggers the **resending of previously failed [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) transfers** within a specified time range (maximum range: 24 hours).

It is recommended to use this endpoint **no more than 1–2 times per day** to avoid performance impact and unnecessary load on the system.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/PartnerAPI/ResendFailedTransfers Method: POST .../api/en/290/PartnerAPI/ResendFailedTransfers | FilterTransferModel | Success---------ResponseWrapper { StatusCode=”0”, Data={} } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- FilterTransferModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| RequestHash | string | Mandatory | required for partner verification |
| StartDateStamp | long | Mandatory | from date in filter required field. |
| EndDateStamp | long | Mandatory | to date in filter required field, the maximum date range is 24 hours. |
| State | int | Optional | Optional State filter, possible values (-1 Error, 1 Processed, 2 No Answer, 4 Skipped) |
| BetIds | List<long> | Optional | Optional (List of bet Ids) |
| DocumentId | long | Optional | Optional (Unique transaction Id) |
