---
title: Errors returned by the FeedConstruct
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct
current_loc: betGuard
location: errors_returned_by_the_feed_construct
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Errors returned by the FeedConstruct

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `errors_returned_by_the_feed_construct` |

## 文档正文
Errors returned by the FeedConstruct

| Key Key | Message Message | Error data (sample) Error data (sample) |  |
| --- | --- | --- | --- |
| BetAmountError | Bet Amount Error (Cashout) |  |
| BetNotFoundError | The bet is not found |  |
| BetSelectionChanged | Odds have been changed | {"SelectionId": "4527713925", "OldPrice": "1.35", "ActualPrice": "1.290", "MarketId": "1500700331", "MatchId": "24433043", "CompetitionId": "4422", "RegionId": "73", "SportId": "4"} |
| BetSelectionsCanNotBeNullOrEmpty | Bet Selections Can Not Be Null Or Empty |  |
| BetSelectionsCombindedError | Please change Selections combination of your betslip | {SelectionId: 4822571750} |
| BetStateError | Bet State Error (Cashout) |  |
| BetTypeError | Bet Type Error |  |
| ClientBetMinStakeLimitError | The bet stake is lower than minimum allowed | {"MaxAllowedBetStake":"458.73","MinAllowedBetStake":"0.1"} |
| ClientBetStakeLimitError | The bet stake is greater than maximum allowed | {"MaxAllowedBetStake":"458.73","MinAllowedBetStake":"0.1"} |
| ClientBetStakeNoLimitError | ClientBetStakeNoLimitError | {"MaxAllowedBetStake":"180.17","MinAllowedBetStake":"192.66"} |
| ClientLocked | Client Locked |  |
| ClientRestrictedForAction | Client Restricted For Action |  |
| EachWayIsNotAvailable | Each Way Is Not Available |  |
| InternalError | Internal error |  |
| InvalidCategory | Invalid Category |  |
| InvalidParameters | Invalid Parameters |  |
| InvalidPartner | Invalid Partner |  |
| LinkedMatches | Linked (not combinable) matches |  |
| MarketNotFound | Market Not Found |  |
| MarketNotVisible | Market Not Visible |  |
| MarketSuspended | Market suspended |  |
| MatchNotBooked | Match Not Booked |  |
| MatchNotFound | Match not found |  |
| MatchNotVisible | Match Not Visible |  |
| MatchSuspended | Match Suspended |  |
| MaxSingleBetAmountError | Single bet amount error |  |
| NotAuthorized | Not authorised |  |
| NotSupportedCurrency | Not Supported Currency |  |
| OperationInProgress | Operation In Progress |  |
| PartnerApiClientBalanceError | Client balance is less |  |
| PartnerApiError | PartnerApiError | {"ApiErrorCode":"1005","ApiErrorMessage":"AuthToken Error"} |
| PartnerApiSecretKeyMissing | Partner Api Secret Key Missing in External Admin |  |
| PartnerApiWrongHash | Wrong hash |  |
| PartnerBlocked | Partner Blocked |  |
| PartnerLimitAmountExceed | Partner Limit Amount Exceed |  |
| PartnerMismatch | Partner Mismatch |  |
| PartnerNotFound | Partner Not Found |  |
| PriceWasChanged | Price Was Changed |  |
| RegionNotFound | Region was not found |  |
| RequiredFieldsMissing | Required Fields Missing |  |
| SelectionMultipleCount | Selection {0} must be combined with at least {1} other selections |  |
| SelectionNotFound | Selection Not Found |  |
| SelectionSinglesOnly | Selection is Singles only |  |
| SelectionSuspended | Selection suspended |  |
| SPMissing | SP Missing |  |
| TokenAlreadyExists | Token Already Exists |  |
| WrongCurrencyCode | Wrong Currency Code |  |
