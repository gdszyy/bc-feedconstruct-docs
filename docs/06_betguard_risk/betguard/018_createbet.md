---
title: CreateBet
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=createBet
current_loc: betGuard
location: createBet
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# CreateBet

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=createBet`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `createBet` |

## 文档正文
CreateBet

**Description:**

The call submits a bet placement request.

| Request URL Sample Request URL Sample | Request Body Request Body | Response Response |  |
| --- | --- | --- | --- |
| http://hostname/api/LangId/PartnerId/Bet/CreateBet Method: POST .../api/en/290/Bet/CreateBet | Request BetModel | Success---------ResponseWrapper { StatusCode=”0”, Data=Response BetModel } Error-----------ResponseWrapper { StatusCode=errorCode, Data=errorMessage } |

- Request BetModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| AuthToken | string | Mandatory | [AuthToken](/documentation?currentLoc=betGuard&location=authToken) |
| RequestHash | string | Mandatory | required for partner verification |
| Amount | decimal | Mandatory | Bet amount, Decimal(18, 2) |
| BetType | Int | Mandatory | Bet type (Single = 1, Multiple = 2, System = 3, Chain = 4, Trixie = 5, Yankee = 6, PermedYankee = 7, SuperYankee = 8, Heinz = 9, SuperHeinz = 10, Goliath = 11, Patent = 12, PermedPatent = 13, Lucky15 = 14, Lucky31 = 15, Lucky63 = 16, Alphabet = 17, StraightForecast = 40, ReverseForecast = 41, CombinationForecast = 42, StraightTricast = 43, CombinationTricast = 44, SystemCombination = 60 |
| SystemMinCount | Int | Mandatory in case of BetType=3 (System) | Multiple length in system bet, should **only** be used with BetType=3 (System) |
| Currency | string | Mandatory | [Currency ISO code](/documentation?currentLoc=betGuard&location=currency_codes) |
| Selections | List<Request BetSelectionModel> | Mandatory | [Bet’s selections](/documentation?currentLoc=betGuard&location=partner_api_bet_selection) |
| SystemCombinations | List<SystemCombinationModel> | Mandatory in case of BetTye=60 (SystemCombination) | This field should **only** be provided in case of BetTye=60 (SystemCombination) |
| ExternalId | long | Mandatory | Partner’s unique identifier of the bet. Will be returned with BetPlaced, Betresulted calls if provided. |
| AcceptTypeId | Int | Optional | Accept only if odd has not changed = 0 (is the default value if not sent) Accept only if odd has not changed or odd has been increased = 1 Accept with any odd changes = 2 |
| OddType | int | Optional | Odd type used to place the bet (Decimal = 0, Fractional = 1, American = 2, HongKong = 3, Malay = 4, Indo = 5) |
| IsEachWay | bool | Optional | Bet is an EachWay bet |
| ClientDetail | ClientDetailModel | Optional | Client details, which are necessary for identifying the client. If the mandatory fields are not sent, FC initiates GetClientDetails to obtain the details. |

- Request BetSelectionModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| SelectionId | long | Mandatory | Selection unique Id |
| Price | decimal | Mandatory | Selection’s price (odd) |
| IsBanker | bool | Optional | If the selection is a banker. Should be included **only** in case of BetTye=60 (SystemCombination) |

- ClientDetailModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| Login | String | Mandatory, max length - 255 | Client’s Login on the Partner’s side. The value of the ExternalID can be used instead |
| CurrencyId | String | Mandatory, max length - 3 | User’s currency code. |
| ExternalId | String | Mandatory, max length - 255 | Unique ID of the Client in the Partner’s Backend |
| LanguageId | String | Optional, max length - 2 | Two letter code of the Client’s language. For example “en” – for English |
| Email | String | Optional, max length - 255 | Email address of the Client. |
| BirthDate | Date | Optional | Birth date of the Client. |
| Gender | Int | Optional | 1 – for Male, 2 – for Female |
| CurrentIp | String | Optional, max length - 50 | IP address of the client currently logged in to the Partner’s website |
| Phone | String | Optional, max length - 255 | Primary phone number of the Client |
| CountryId | String | Optional, max length - 2 | ISO 3166-1 alpha-2 code of the registration country |
| PartnerFlag | String | Optional, max length - 50 | Optional flag field to differentiate some clients from others |
| SportsProfileID | Int | Optional | Identifier for the user’s assigned sportsbook profile. |
| RMTblocked | bool | Optional | Indicates if Risk Management Tool (RMT) is blocked for this user (true = blocked, false = allowed). |

- Request/Response SystemCombinationModel

| Field Name Field Name | Type Type | Requirement Requirement | Description Description |  |
| --- | --- | --- | --- | --- |
| BetAmount | decimal | Mandatory | Bet amount for a particular combination |
| SystemMinCount | int | Mandatory | Multiple length in system bet |

- Request/Response ErrorModel

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| StatusCode | string | [See Error codes](/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct) |
| Data | ErrorDataModel | [Error data object](/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct) |

- Request/Response ErrorDataModel

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| Key | string | [See Error codes](/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct) |
| Message | string | [Error message](/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct) |
| ErrorData | string | [See Error data](/documentation?currentLoc=betGuard&location=errors_returned_by_the_feed_construct) |

- Response BetModel

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| AuthToken | string | [AuthToken](/documentation?currentLoc=betGuard&location=authToken) |
| RequestHash | string | required for partner verification |
| TransactionId | long | Bet transaction unique Id |
| BetId | long | Bet unique Id |
| Amount | decimal | Bet amount, Decimal(18, 2) |
| Created | DateTime | Bet placement time |
| BetType | Int | Bet type (Single = 1, Multiple = 2, System = 3, Chain = 4, Trixie = 5, Yankee = 6, PermedYankee = 7, SuperYankee = 8, Heinz = 9, SuperHeinz = 10, Goliath = 11, Patent = 12, PermedPatent = 13, Lucky15 = 14, Lucky31 = 15, Lucky63 = 16, Alphabet = 17, StraightForecast = 40, ReverseForecast = 41, CombinationForecast = 42, StraightTricast = 43, CombinationTricast = 44, SystemCombination = 60 |
| AcceptTypeId | Int | Accept only if odd has not changed = 0 Accept only if odd has not changed or odd has been increased = 1 Accept with any odd changes = 2 |
| SystemMinCount | Int | Multiple length in system bet, is **only** used in case of BetType=3 (System) |
| TotalPrice | decimal | The total price (odd) of the bet |
| State | Int | Bet state (Accepted = 1, Returned = 2, Lost = 3, Won = 4, CashedOut = 5) |
| IsLive | bool | Bet is placed on Live selection |
| Currency | string | [Currency ISO code](/documentation?currentLoc=betGuard&location=currency_codes) |
| ClientId | Int | Player’s unique Id |
| PossibleWin | decimal | Bet possible win amount |
| Selections | List<BetSelectionModel>\* | Bet’s selections |
| SystemCombinations | List<SystemCombinationModel> | Used **only** for BetType=60 (SystemCombination) |
| OddType | Int | (optional) Odd type used to place the bet (Decimal = 0, Fractional = 1, American = 2, HongKong = 3, Malay = 4, Indo = 5) |
| IsEachWay | bool | Is used if the Bet is an EachWay bet |
| ExternalId | long | Partner’s unique identifier of the bet |

- Response BetSelectionModel

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| SelectionId | long | Selection unique Id |
| SelectionName | string | Selection name |
| MarketTypeId | int | Market template unique Id |
| MarketName | string | Market name |
| MatchId | int | Match unique Id |
| MatchName | string | Match name |
| MatchStartDate | DateTime | Match start date |
| RegionId | int | Region unique Id |
| RegionName | string | Region name |
| CompetitionId | int | Competition unique Id |
| CompetitionName | string | Competition Name |
| SportId | int | Sport unique Id |
| SportName | string | Sport name |
| Price | decimal | Selection’s price (odd) |
| IsLive | bool | Selection Live/Pre-match |
| Basis | decimal | Market’s handicap value |
| IsOutright | bool | Match is outright or not |
| IsBanker | bool | If the selection is a banker. Used **only** in case of BetTye=60 (SystemCombination) |
