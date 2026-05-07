---
title: Resulting Bets & Reporting Selection Outcomes
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=resulting_bets_reporting_selection_outcomes
current_loc: betGuard
location: resulting_bets_reporting_selection_outcomes
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Resulting Bets & Reporting Selection Outcomes

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=resulting_bets_reporting_selection_outcomes`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `resulting_bets_reporting_selection_outcomes` |

## 文档正文
Resulting Bets & Reporting Selection Outcomes

This section outlines how bets’ statuses are updated and how outcomes are reported by FeedConstruct (FC)

- Requesting Bet State Changes by Partner

Once the bet is in the **Accepted** state (placed but not yet settled), the Partner can request bet state updates via the following methods:

- [MarkBetAsCashout](../documentation?currentLoc=betGuard&location=mark_bet_as_cashout) – Instantly marks the bet as Cashed Out without validation.
- [CheckAndMarkBetAsCashout](../documentation?currentLoc=betGuard&location=check_and_mark_bet_as_cashout) – Validates cashout availability (including Live Delay) before marking the bet as **Cashed Out**. Note: Amount is not checked.
- [ReturnBet](../documentation?currentLoc=betGuard&location=return_bet) – Changes the bet status to Returned.

**Important:** Cashout logic must be implemented by the Partner.

- Reporting Bet States by FC

Once the  **final state** of a bet is determined (Won, Lost, Returned, or Cashed Out), or when the result is reverted (the bet is returned to the Accepted state), FeedConstruct notifies the Partner using the [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted) call.

- Reporting Selection Outcomes By FC

**For Multiple Bets only:**

If a selection is settled but does not change the overall bet state:

- FC calls the **/bet/event**  endpoint ([BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) method of Export API) to report the outcome of the individual selection.

**Example 1:** A multiple bet consists of two selections: Selection A and Selection B.

- Selection A is **lost**, so the whole bet becomes **Lost** → FC sends this via [BetResulted](../documentation?currentLoc=betGuard&location=bet_resulted).
- Later, Selection B is calculated (won/lost) → sent via [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) method:

**Example 2:**  A multiple bet consists of three selections:

- Selection 1 is **won** → sent via [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) method.
- Selection 2 is **lost** → this determines the overall bet as **Lost**  → sent via [BetResulted.](../documentation?currentLoc=betGuard&location=bet_resulted)
- Selection 3 is **won/lost** → does not impact overall state → sent via [BetSelection](../documentation?currentLoc=betGuard&location=partner_api_bet_selection) method.
