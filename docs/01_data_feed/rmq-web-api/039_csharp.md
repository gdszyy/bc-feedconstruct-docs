---
title: .Net/C#
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=cSharp
current_loc: oddsFeedRmqAndWebApi
location: cSharp
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# .Net/C#

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=cSharp`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `cSharp` |

## 文档正文
.Net/C#

#### Source

(Version 0.0.5)FeedConstruct.OddsFeed.src.zip

#### Binary

FeedConstruct.OddsFeed.bin.zip

#### Usage Sample

```
private static FeedConstruct.OddsFeed.SDK.OddsFeedClient client;

private const int PartnerId = 444555444; // client PartnerId

static async Task Main(string[] args)

{

try

{

await RunIntegration();

}

catch (Exception ex)

{

Console.WriteLine(ex);

}

Console.ReadLine();

}

/// <summary>;

/// Run integration

/// </summary>

/// <returns></returns>;

private static async Task RunIntegration()

{

await StartClient();

// Get sports

var Sports = await client.WebApiClient.GetSportsAsync();

// Get regions

var Regions = await client.WebApiClient.GetRegionsAsync();

// Get MarketTypes

var MarketTypes = await client.WebApiClient.GetMarketTypesAsync();

// Get DataSnapshot for Live matches and request changes in last 5 minutes ONLY

var LiveDataSnapshot = await client.WebApiClient.GetDataSnapshotAsync(true, 5);

// Get Live bookings

var LiveBookings = await client.WebApiClient.GetBookingsAsync(true);

// ............

}

/// <summary>;

/// Start client

/// </summary>;

/// <returns></returns>;

private static async Task StartClient()

{

FeedConstruct.OddsFeed.SDK.Configuration.OddsFeedClientConfig config = new FeedConstruct.OddsFeed.SDK.Configuration.OddsFeedClientConfig()

{

ApiURL = "[WEB API URL]/api/DataService"  // WEB API URL,

ApiUsername = "xxxxxx", // WEB API userName

ApiPassword = "xxxxxx", // WEB API password

RmqHost = "[RmqHost]", //RmqHost

RmqPort = 5673,

RmqUsername = "xxxxxx", // RMQ userName

RmqPassword = "xxxxxx", // RMQ password

RmqQueueNames = new string[] { $"P{PartnerId}_live", $"P{PartnerId}_prematch" }

};

client = new FeedConstruct.OddsFeed.SDK.OddsFeedClient(config);

// will fire in case of any exceptions

client.OnException += Client_OnException;

// will fire in case of any error in response

client.OnResponseError += Client_OnResponseError;

// will notify that platform suggesting to get data snapshot for avoiding possible data loss

client.OnGetDataSnapshotCommandReceived += Client_OnGetDataSnapshotCommandReceived;

// Subscribe to this event if you want to get received update in json string

client.OnUpdateRaw += Client_OnUpdateRaw;

// Sport add or change

client.OnSportUpdate += Client_OnSportUpdate;

// Region add or change

client.OnRegionUpdate += Client_OnRegionUpdate;

// Competition add or change

client.OnCompetitionUpdate += Client_OnCompetitionUpdate;

// MarketType add or change

client.OnMarketTypeUpdate += Client_OnMarketTypeUpdate;

// SelectionType add or change

client.OnSelectionTypeUpdate += Client_OnSelectionTypeUpdate;

// Match add or change

client.OnMatchUpdate += Client_OnMatchUpdate;

// MatchStat add or change

client.OnStatUpdate += Client_OnStatUpdate;

// Market add or change

client.OnMarketUpdate += Client_OnMarketUpdate;

// Some type of data(Sport,Match,...) is unbooked

client.OnUnBookedObject += Client_OnUnBookedObject;

// Fire in case of VoidNotification

client.OnVoidNotification += Client_OnVoidNotification;

// Start receive updates

await client.StartReceivingAsync();

}
```
