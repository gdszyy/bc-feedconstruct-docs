---
title: .Net/C#
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=cSharp
current_loc: betGuard
location: cSharp
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# .Net/C#

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=cSharp`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `cSharp` |

## 文档正文
.Net/C#

#### Source

(Version 0.0.1)FeedConstruct.Betguard.SDK.src.zip

#### Binary

FeedConstruct.Betguard.SDK.bin.zip

#### Usage Sample

```
1. BetGuard API

Use BetguardClient to call FeedConstruct API endpoints.

using FeedConstruct.Betguard.SDK;

var options = new BetguardClientOptions {

BaseUrl = [API URL], // API URL

SecretKey = "xxxxxx", // Secret key

PartnerId = "xxxxxx", // PartnerId

LanguageId = "en",

TimeoutSeconds = 30

};

var httpClient = new HttpClient();

var betguardClient = new BetguardClient(httpClient, options);

// Use betguardClient

var result = await betguardClient.CreateBetAsync(request);

// Create a Bet

var request = new CreateBetRequest {

AuthToken = "client-auth-token",

Amount = 10.00m,

BetType = BetTypeEnum.Single,

Currency = "USD",

ExternalId = 123456,

SystemMinCount = 0,

AcceptTypeId = AcceptTypeEnum.Any,

Selections = new List<RequestBetSelectionModel> {

new() {

SelectionId = 55001,

Price = 1.85m

}

},

// Optional: include client details to skip GetClientDetails callback

ClientDetail = new ClientDetailModel {

Login = "player1",

CurrencyId = "USD",

ExternalId = "EXT-001"

}

};

// All API methods are available.

// Get Max Bet Amount

var result = await betguardClient.GetMaxBetAmountAsync(request);

// Mark Bet As Cashout

var result = await betguardClient.MarkBetAsCashoutAsync(request);

// ............

2. Partner API

FeedConstruct sends HTTP POST callbacks to your backend when certain events occur (bet placed,

bet resulted, rollback, etc.). The SDK provides:

IBetguardPartnerApi — an interface defining all callback methods

PartnerApiHashValidator — validates the MD5 hash and timestamp on incoming requests

using FeedConstruct.Betguard.SDK;

using FeedConstruct.Betguard.SDK.Security;

// Create the hash validator

var hashValidator = new PartnerApiHashValidator();

string sharedKey = "xxxxxx"; // Secret key

// Create your service implementation manually

var clientRepo = new ClientRepository(/* ... */);

var betRepo = new BetRepository(/* ... */);

var partnerService = new BetguardPartnerService(clientRepo, betRepo);

// BetPlaced callback

public async Task HandleBetPlaced(BetPlacedRequest request) {

if (!hashValidator.ValidateBetPlaced(request, sharedKey)) {

throw new InvalidOperationException("Invalid hash or expired timestamp");

}

var result = await partnerService.BetPlacedAsync(request);

}

// ............
```
