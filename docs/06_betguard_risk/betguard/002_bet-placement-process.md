---
title: Bet Placement process
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placement_process
current_loc: betGuard
location: bet_placement_process
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Bet Placement process

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_placement_process`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_placement_process` |

## 文档正文
Bet Placement process

This section outlines the end-to-end process of placing a bet. It describes how the Partner interacts with FeedConstruct (FC) backend during the bet placement and bet resulting lifecycle.

- Odds Feed Integration

- The Partner receives sports data via the **FeedConstruct Odds Feed** and renders the betting options in its Sportsbook interface.

- User (client) Interaction

- The end user (client) selects events from the Sportsbook and adds them to the betslip.
- **Important:** Only selections originating from **FeedConstruct`s Odds Feed** are valid.

- Bet Placement Request

- The Partner initiates a bet by calling [CreateBet](../documentation?currentLoc=betGuard&location=createBet) endpoint.

- User Identification

- Upon receiving the [CreateBet](../documentation?currentLoc=betGuard&location=createBet) request, FC needs to identify the client. For that
  - The Partner includes all the mandatory fields of the **ClientDetailModel**.
  - If not, FC calls the Partner’s [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details) endpoint.
  - In case of [GetClientDetails](../documentation?currentLoc=betGuard&location=partner_api_get_client_details), if the Partner:
    1. Returns an error response, or
    2. Fails to respond within 3 seconds (timeout), or
    3. Fails to send a valid response
  - → FC responds to the original [CreateBet](../documentation?currentLoc=betGuard&location=createBet) request with an error.

- Validation & Limitation Check

- Upon successful user identification, FC performs internal validations:

- The relevance of the price (odd).
- Market and event availability.
- Visibility rules.
- Bet limitations.
- Live delay (for in-play bets).

- If FC fails any of the internal checks it responds to the [CreateBet](../documentation?currentLoc=betGuard&location=createBet) call with a rejection without initiating [BetPlaced](../documentation?currentLoc=betGuard&location=partner_api_bet_placed).

- Most Common Rejection Reasons:
  1. Odds have changed (most common in fast-paced sports).
  2. Market is suspended (most common in fast-paced sports).
  3. Match is suspended (most common in fast-paced sports).
  4. The bet amount is less than the minimum or exceeds the maximum betslip values set for each currency.
  5. The bet stake exceeds the maximum allowed for a specific client on a particular selection.
  6. The client’s limit on that selection has expired (this limit is auto updated after the reset time).

- Bet Confirmation

- If validations pass, FC initiates the [BetPlaced](../documentation?currentLoc=betGuard&location=partner_api_bet_placed) call to the Partner for final confirmation and waits for the partner’s response:

- If the Partner:
  1. Responds with an error, or
  2. Does not respond within 5 seconds
- → FC initiates a [RollBack](../documentation?currentLoc=betGuard&location=partner_api_rollback) operation and responds to the original [CreateBet](../documentation?currentLoc=betGuard&location=createBet) request with a rejection.

- If the Partner responds to [BetPlaced](../documentation?currentLoc=betGuard&location=partner_api_bet_placed) with a success response within 5 seconds:

- FC:
  1. Accepts the bet with Accepted State.
  2. Responds to the Partner’s original [CreateBet](../documentation?currentLoc=betGuard&location=createBet) request with a success message.
