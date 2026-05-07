---
title: Introduction
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=introduction
current_loc: betGuard
location: introduction
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Introduction

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=introduction`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `introduction` |

## 文档正文
Introduction

The **Betguard system** provides a robust API framework that enables seamless communication between FeedConstruct and its Partners throughout the entire bet lifecycle - from placement and validation to resulting, outcome reporting, and client updates.

Betguard is composed of **two APIs**, each serving a distinct communication purpose:

- BetGuard API

Used for requests initiated by the Partner’s backend to FeedConstruct’s backend. This API is responsible for handling bet placement, client management, and other operational commands.

**Available methods include:**

- [GetMaxBetAmount](../documentation?currentLoc=betGuard&location=get_max_bet_amount)  – Retrieves the maximum allowed bet amount for a specific selection and client.
- [CreateBet](../documentation?currentLoc=betGuard&location=createBet) – Submits a bet placement request.
- [ResendFailedTransfers](../documentation?currentLoc=betGuard&location=resend_failed_transfers) – Triggers resending of previously failed BetResulted calls.
- [MarkBetAsCashout](../documentation?currentLoc=betGuard&location=mark_bet_as_cashout) – Marks a bet as cashed out without validation.
- [CheckAndMarkBetAsCashout](../documentation?currentLoc=betGuard&location=check_and_mark_bet_as_cashout) – Checks cashout availability and then marks the bet as cashed out.
- [ReturnBet](../documentation?currentLoc=betGuard&location=return_bet) – Marks a bet with Accepted state as returned.
- [UpdateClient](../documentation?currentLoc=betGuard&location=update_client) – Updates client profile information.

**Note:** Only [CreateBet](../documentation?currentLoc=betGuard&location=createBet) is mandatory to integrate. The remaining methods are optional and can be skipped if not required by the Partner.

- Partner API

Handles communication initiated by FeedConstruct’s backend toward the Partner’s backend. Used for sending requests, data, or status updates related to clients and bets.

**Available methods include:**

- [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details) – This method is called if the mandatory fields of **ClientDetailModel** fields are not sent with [CreateBet](../documentation?currentLoc=betGuard&location=createBet) and [GetMaxBetAmount](../documentation?currentLoc=betGuard&location=get_max_bet_amount). Requests user information from the Partner for validation and limit checks.
- [BetPlaced](../documentation?currentLoc=betGuard&location=partner_api_bet_placed) – Notifies the Partner that the bet is pending final confirmation.
- [Rollback](../documentation?currentLoc=betGuard&location=partner_api_rollback) – Informs the Partner of a failed bet.
- [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) – Sends the final state and outcome of a bet.
- [Client](../documentation?currentLoc=betGuard&location=partner_api_client) – Pushes updates related to the client profile.
- [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) – Pushes settlement outcomes for individual selections in a multiple bet. The method is used **only for Multiple bets** and **only when the settlement does not change the overall bet state**. If the bet state changes, the [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) method from the Partner API is used instead.

**Note:** [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details), [BetPlaced](../documentation?currentLoc=betGuard&location=partner_api_bet_placed), [Rollback](../documentation?currentLoc=betGuard&location=partner_api_rollback),[BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) methods in this API are mandatory to integrate. [Client](../documentation?currentLoc=betGuard&location=partner_api_client) and [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) methods are not mandatory: the integration of the methods is not required if the Partner accepts only single bets or does not wish to receive selection-level outcome and client related updates..

**Important:** The [Client](../documentation?currentLoc=betGuard&location=partner_api_client) and [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) methods require subscription; therefore, if the Partner intends to receive these updates, they must inform their project or account manager accordingly.
