---
title: Bet Selection
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_bet_selection
current_loc: betGuard
location: partner_api_bet_selection
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Bet Selection

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_bet_selection`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_bet_selection` |

## 文档正文
Bet Selection

This method pushes settlement outcomes for individual selections in a multiple bet. It is used **only for Multiple bet types and only when the settlement does not change the overall bet state**.

If the bet state changes, the [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) method from the Partner API is used instead.

**Important:** The **BetSelection** method requires subscription; therefore, if the Partner intends to receive these updates, they must inform their project or account manager accordingly.

URL: /bet/event

Example: **http://hostname:port/api/bet/event**

Method: POST

- Push Update Parameters

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| betId | long | Mandatory | Bet Id |
| oddId | long | Mandatory | Id of the bet selection (oddId is SelectionId) |
| settlementDate | long | Mandatory | UTC date of settlement |
| status | short | Mandatory | • State of bet event • NotResulted = 0, • Placed = 1, • Returned = 2, • Lost = 3, • Won = 4, • WinReturn = 5 , • LossReturn = 6 |
| externalId | long | Optional | External Id of the bet provided by the partner in the CreateBet call |
