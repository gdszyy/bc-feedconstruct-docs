---
title: Client
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_client
current_loc: betGuard
location: partner_api_client
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Client

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_client`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_client` |

## 文档正文
Client

The method pushes updates related to a client profile.

**Important:** The **Client** method requires subscription; therefore, if the Partner intends to receive these updates, they must inform their project or account manager accordingly.

URL: /client/change

Example: **http://hostname:port/api/client/change**

Method: POST

- Push Update Parameters

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| clientId | int | Mandatory | Id of client |
| currencyId | string |  | currency |
| iban | string |  | Iban number of client |
| firstName | string |  | Name of client |
| lastName | string |  | Surname of client |
| middleName | string |  | Middle name of client |
| regionCode | string |  | Region code |
| gender | Int |  | Gender of client(1 male, 2 female) |
| address | string |  | Address of client |
| email | string |  | Email of client |
| language | string |  | Language |
| phone | string |  | Phone number of client |
| mobilePhone | string |  | Mobile phone of client |
| birthDate | long |  | Date of birth |
| isLocked | bool |  | Whether client is locked |
| isSubscribedToNewsletter | bool |  | Whether client is subscribed to news feed |
| preMatchSelectionLimit | decimal |  | Limit of event for pre match |
| liveSelectionLimit | decimal |  | Limit of event for live |
| isVerified | bool |  | Whether client is verified |
| globalLiveDelay | int |  | Delay for live matches |
| externalId | string |  | Unique ID of the Client in the Partner’s Backend |
| isTest | bool |  | Whether it is test client |
| sportsbookProfileId | int |  | Sportsbook profile Id |
