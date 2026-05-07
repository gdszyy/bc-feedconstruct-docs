---
title: Bet Limits - Global Limits
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_limits_global_limits
current_loc: betGuard
location: bet_limits_global_limits
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Bet Limits - Global Limits

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=bet_limits_global_limits`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `bet_limits_global_limits` |

## 文档正文
Bet Limits - Global Limits

![](/static/media/bet-liabilities.359143473cf364008db8.png)

- Maximum Single Selection Liability

Maximum Single Selection Liability is the maximum net winning amount which any player can win from any single selection. This is the most important indication in our limitation system as all other limits (such as game limit, client limit, market type limit) are percentages of this amount.

How the maximum bet amount of a single selection is calculated?

The Formula is:

**Max Single Selection Liability x (Competition or Sport Limit/100) x (Market Type limit/100) x (Client Limit/100) / (Odd -1)**

Example:

Selection liability is set at 15.000 EUR, Competition limit is 10 %, Market type limit is 100%, Client limit is 50%, the selection odd is 1.8.

15.000 x (10 %/100) x (100 % / 100) x (50 % / 100) / (1.8 – 1) =

The max bet of the selection of the example mentioned above will be 937,5 EUR

After reaching the maximum amount, the player cannot bet any more on the certain selection. The limit can be expired either by betting once the total amount or many times by smaller amounts.

Example:

Client has placed a bet on Liverpool vs Everton match on the selection W1. All the limits along with the partner selection liability allows to bet on this selection 10,000 EUR. Client can bet on this selection once with max bet (10,000 EUR) and he can place a bet with different amounts, but after reaching 10.000 EUR he can’t place bet on this selection any more until the delay time ends which has been set on global limits (“Reset Player limit after minutes”).

It is important to remember that each bet of the player should not be more and less than the minimum and maximum bet stakes amounts in each currency accordingly.

- Maximum Individual Bet Liability

The amount of pure winning (without bet stake) of a single player for one bet. (ex. multiple bet).

- Maximum Payout

Maximum Payout - The maximal amount, which is possible to be won by a single player for one coupon. Maximum Payout is the gross win that a client can receive on a coupon (exception – Super Bets).

**Note:**  MaxBet error **"ClientBetStakeLimitError"** may also happen, when there was a change with Maximum Payout.

**Important:** Limitation changes are not applied immediately, usually it’s needed some time (up to an hour) for the change to affect.

- Bet Stake per Currency

The Currency added in default is EUR. The Partner is allowed to request adding or removing currencies from their project․ Each currency should have the minimum and maximum allowed bet stake amount, which is also mentioned by the Partner.

![](/static/media/bet-stake-per-currency.b464e4b46f76f825df6b.png)

- Min Stake

Min Stake – the minimal stake for one bet for pre-match or live events. The client will not be able to place a bet with fewer amount than the value, which is set as Min Stake.

- Pre-Match Max Stake

PreMatch Max Bet Stake - maximal stake for one bet for pre-match event. This setting refers to one coupon. The client can place as many Max Stakes, as the Selection Liability (Limit) allows. After reaching the limit, the client cannot place bet on this selection any more until the delay time ends.

- Live Max Stake

Live Max Bet Stake - maximal stake for one bet for live event. This setting refers to one coupon. The client can place as many Max Stakes, as the Selection Liability (Limit) allows. After reaching the limit, the client cannot place bet on this selection any more until the delay time ends.

**Note:**  During the integration phase, Partners are subject to a pre-set Turnover Limit for their potential bets. If a Partner receives a **"PartnerLimitError"** rejection message while making testing bets, it indicates this limit has been reached. Partners cannot increase this Turnover Limit themselves and must contact the support team for assistance.
