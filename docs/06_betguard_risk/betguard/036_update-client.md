---
title: UpdateClient
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=update_client
current_loc: betGuard
location: update_client
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# UpdateClient

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=update_client`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `update_client` |

## 文档正文
UpdateClient

The **UpdateClient** endpoint is called by the Partner when  **a client’s details have been updated on the Partner’s side** and those changes need to be **synchronized Feedconstruct**.

This ensures that the client information remains consistent across both systems.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId// Client/UpdateClient Method: POST .../api/en/290/Client/UpdateClient | ClientModel object | Success---------ResponseWrapper { StatusCode=”0”, Data=ClientModel } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- ClientModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| id | int | Mandatory | Client Id |
| ExternalId | string | Mandatory, in case the Client ID is not provided | Either id or ExternalId must be provided |
| RequestHash | string | Mandatory | required for partner verification |
| FirstName | string | Optional |  |
| LastName | string | Optional |  |
| MiddleName | string | Optional |  |
| Login | string | Optional | username of client in website |
| IBAN | string | Optional | international bank account number of client |
| RegionCode | string | Optional | ISO ALPHA-2 Code of country (FR, GB, RU) |
| Gender | int | Optional | Male = 1, Female = 2 |
| ProfileId | int | Optional |  |
| DocNumber | string | Optional | Passport Number of client |
| PersonalId | string | Optional | Unique identity number of client |
| Address | string | Optional |  |
| Email | string | Optional |  |
| Language | string | Optional | preferred language of client, ISO 639-1 codes |
| Phone | string | Optional |  |
| MobilePhone | string | Optional |  |
| BirthDateStamp | long | Optional | UNIX timestamp representation of registration date |
| City | string | Optional | City where client lives |
| PromoCode | string | Optional | Promotional code by which client was registered |
| TimeZone | decimal | Optional | Timezone of client (in hours) |
| IsLocked | bool | Optional | Indicates whether the account is locked (true = locked, false = active). |
| CreatedStamp | long | Optional |  |
| ModifiedStamp | long | Optional |  |
| DocIssuedBy | string | Optional |  |
| LastLoginIp | string | Optional |  |
| LastLoginTimeStamp | long | Optional |  |
| PreMatchSelectionLimit | decimal | Optional |  |
| LiveSelectionLimit | decimal | Optional |  |
| IsVerified | bool | Optional |  |
| SportsbookProfileId | int | Optional | Identifier for the user’s assigned sportsbook profile. |
| GlobalLiveDelay | int | Optional |  |
| IsTest | bool | Optional | Mark if the account is a test account |
| CanBet | bool | Optional | Determines if the user is allowed to place bets (true = allowed, false = restricted). |
