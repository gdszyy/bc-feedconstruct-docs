---
title: Samples
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=samples
current_loc: oddsFeedRmqAndWebApi
location: samples
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Samples

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=samples`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `samples` |

## 文档正文
Samples

##### WEB API

##### Token

##### Command:

[API\_URL] /api/DataService/Token

##### Body:

```
{

"Params": [

{

"UserName": "string",

"Password": "string"

}

]

}
```

##### Sport

##### Command:

[API\_URL] /api/DataService/Sport?token=[token]

##### Response:

```
{

"Command": "GetSports",

"ResultCode": 0,

"Objects": [

{

"Id": 1,

"Name": "Football",

"NameId": 21,

},

{

"Id": 2,

"Name": "Ice Hockey",

"NameId": 29,

},

{

"Id": 3,

"Name": "Basketball",

"NameId": 7,

},

{

"Id": 4,

"Name": "Tennis",

"NameId": 40,

},

{

"Id": 5,

"Name": "Volleyball",

"NameId": 41,

},

{

"Id": 6,

"Name": "American Football",

"NameId": 1,

},

{

"Id": 7,

"Name": "Athletics",

"NameId": 2,

},

{

"Id": 8,

"Name": "Aussie Rules",

"NameId": 3,

},

...

],

}
```

##### Region

##### Command:

[API\_URL] /api/DataService/Region?token=[token]

##### Response:

```
{

"Command": "GetRegions",

"ResultCode": 0,

"Objects": [

{

"Id": 1,

"Name": "World",

"NameId": 20860,

},

{

"Id": 2,

"Name": "Europe",

"NameId": 20861,

},

{

"Id": 3,

"Name": "Asia",

"NameId": 20862,

},

{

"Id": 4,

"Name": "Africa",

"NameId": 20863,

},

{

"Id": 5,

"Name": "North America",

"NameId": 21322,

},

{

"Id": 6,

"Name": "South America",

"NameId": 21323,

},

{

"Id": 7,

"Name": "Oceania",

"NameId": 21058,

},

{

"Id": 8,

"Name": "Afghanistan",

"NameId": 21059,

},

...

],

}
```

##### Competition

##### Command:

[API\_URL] /api/DataService/Competition?token=[token]&SportId=[sportId]&RegionId=[regionId]

##### Response:

```
{

"Command": "GetCompetitions",

"ResultCode": 0,

"Objects": [

{

"SportId": 1,

"RegionId": 254,

"IsTeamsReversed": false,

"Id": 35859,

"Name": "Cup",

"NameId": 2369996

},

{

"SportId": 1,

"RegionId": 111,

"IsTeamsReversed": false,

"Id": 35863,

"Name": "U19 Championship",

"NameId": 2370251

},

{

"SportId": 1,

"RegionId": 1,

"IsTeamsReversed": false,

"Id": 35903,

"Name": "Bangabandhu Cup",

"NameId": 2370656

},

{

"SportId": 1,

"RegionId": 2,

"IsTeamsReversed": false,

"Id": 35957,

"Name": "UEFA Champions League - Cards and Corners",

"NameId": 2371231

},

{

"SportId": 1,

"RegionId": 2,

"IsTeamsReversed": false,

"Id": 35958,

"Name": "UEFA Champions League - Fouls and Shots on Goal",

"NameId": 2371232

}

...

]

}
```

##### MarketTypes

##### Command:

[API\_URL] /api/DataService/MarketType?token=[token]&SportId=[sportId]

##### Response:

```
{

"Command": "GetMarketTypes",

"ResultCode": 0,

"Objects": [

{

"Kind": "CornerHandicap",

"IsHandicap": true,

"IsOverUnder": false,

"SportId": 1,

"SelectionCount": 2,

"IsDynamic": false,

"TypeFlag": 1,

"Id": 8956,

"Name": "Corners: Handicap",

"NameId": 415233

},

{

"Kind": "CornerOddEven",

"IsHandicap": false,

"IsOverUnder": false,

"SportId": 1,

"SelectionCount": 2,

"IsDynamic": false,

"TypeFlag": 1,

"Id": 8957,

"Name": "Corner: Even/Odd",

"NameId": 415237

}

...

]

}
```

##### SelectionType

##### Command:

[API\_URL] /api/DataService/SelectionType?token=[token]

##### Response:

```
{

"Command": "GetSelectionTypes",

"ResultCode": 0,

"Objects": [

{

"MarketTypeId": 369,

"Order": 2,

"Kind": "W2",

"Id": 101,

"Name": "W2",

"NameId": 143320

},

{

"MarketTypeId": 370,

"Order": 1,

"Kind": "Over",

"Id": 102,

"Name": "Over ({h})",

"NameId": 143321

},

{

"MarketTypeId": 370,

"Order": 2,

"Kind": "Under",

"Id": 103,

"Name": "Under ({h})",

"NameId": 143322

},

{

"MarketTypeId": 371,

"Order": 2,

"Kind": "Away",

"HandicapSign": 1,

"LiveDelay": 2,

"Id": 105,

"Name": "{t2} ({h})",

"NameId": 143324

},

{

"MarketTypeId": 372,

"Order": 1,

"Kind": "Over",

"LiveDelay": 1,

"Id": 106,

"Name": "Over ({h})",

"NameId": 143325

},

...

]

}
```

##### GetBookings

##### Command:

[API\_URL] /api/DataService/Booking?token=[token]&IsLive=([true] or [false])

##### Response:

```
{

"Command": "GetBookings",

"ResultCode": 0,

"Objects": [

{

"ObjectId": 67,

"ObjectTypeId": 2,

"SportId": 5,

"RegionId": 0,

"CompetitionId": 0,

"IsLive": false,

"IsSubscribed": true

},

{

"ObjectId": 37,

"ObjectTypeId": 1,

"SportId": 37,

"RegionId": 0,

"CompetitionId": 0,

"IsLive": false,

"IsSubscribed": false

}

...

]

}
```

##### GetCompetitionById

##### Command:

[API\_URL] /api/DataService/CompetitionById?token=[token]&competitionId=[competitionId]

##### Response:

```
{

"Command": "GetCompetitionById",

"ResultCode": 0,

"Objects": [

{

"SportId": 1,

"RegionId": 257,

"IsTeamsReversed": false,

"Id": 538,

"Name": "Premier League",

"NameId": 193154

}

]

}
```

##### MatchById

##### Command:

[API\_URL] /api/DataService/MatchById?token=[token]&MatchId=[matchId]&IncludeMatchStats=([true] or [false])

##### Response:

```
{

"Command": "GetMatchById",

"ResultCode": 0,

"Objects": [

{

"Date": "2018-10-31T13:00:00Z",

"CompetitionId": 9278,

"SportId": 1,

"RegionId": 90,

"LiveStatus": 1,

"MatchStatus": 1,

"IsVisible": true,

"IsSuspended": false,

"IsLive": true,

"IsStarted": true,

"MatchMembers": [

{

"TeamName": "HSV Barmbek-Uhlenhorst",

"TeamId": 101014,

"IsHome": true,

"NameId": 280818

},

{

"TeamName": "SC Condor",

"TeamId": 101022,

"IsHome": false,

"NameId": 280826

}

],

"IsBooked": false,

"IsStatAvailable": false,

"Id": 12862017

}

]

}
```

**Or in case of calling GetMatchById for unbooked/old match**

```
{

"Objects": [],

"Command": "GetMatchById",

"Error": {

"Key": "NotAllowed",

"Message": "Not allowed",

"Id": 18849956

},

"ResultCode": 100

}
```

##### GetMarketTypeById

##### Command:

[API\_URL] /api/DataService/MarketTypeById?token=[token]&marketTypeId=[marketTypeId]

##### Response:

```
{

"Command": "GetMarketTypeById",

"ResultCode": 0,

"Objects": [

{

"Kind": "CornerOddEven",

"IsHandicap": false,

"IsOverUnder": false,

"SportId": 1,

"SelectionCount": 2,

"IsDynamic": false,

"TypeFlag": 1,

"Id": 8957,

"Name": "Corner: Even/Odd",

"NameId": 415237

}

]

}
```

##### SelectionTypeById

##### Command:

[API\_URL] /api/DataService/SelectionTypeById?token=[token]&selectionTypeId=[selectionTypeId]

##### Response:

```
{

"Command": "GetSelectionTypeById",

"ResultCode": 0,

"Objects": [

{

"MarketTypeId": 371,

"Order": 1,

"Kind": "Home",

"HandicapSign": -1,

"Id": 104,

"Name": "{t1} ({-h})",

"NameId": 143323

}

]

}
```

##### Book

##### Command:

[API\_URL] /api/DataService/Book?token=[token]

##### Response:

```
{

"Command": "BookItems",

"ResultCode": 0,

"Object": []

}
```

##### Calendar Match

##### Command:

[API\_URL] /api/DataService/Calendar?token=[token]

##### Response:

```
{

"Command": "GetCalendar",

"ResultCode": 0,

"SocketTime": "2021-03-09T08:00:49.1411504Z",

"Objects": [

{

"Date": "2021-04-06T09:30:00Z",

"CompetitionId": 611,

"SportId": 11,

"RegionId": 126,

"MatchStatus": 1,

"IsStarted": false,

"MatchMembers": [

{

"TeamName": "SK Wyverns",

"TeamId": 9863,

"IsHome": false,

"NameId": 205178

},

{

"TeamName": "Hanwha Eagles",

"TeamId": 11857,

"IsHome": true,

"NameId": 207033

}

],

"PrematchBooked": false,

"LiveBooked": false,

"LiveStatus": 1,

"IsVisible": true,

"Id": 17791873

},

{

"Date": "2021-04-06T18:15:00Z",

"CompetitionId": 20011,

"SportId": 22,

"RegionId": 1,

"MatchStatus": 0,

"IsStarted": false,

"MatchMembers": [

{

"TeamName": "PDC Premier League 2021",

"TeamId": 633715,

"IsHome": true,

"NameId": 4471033

}

],

"IsOutright": true,

"PrematchBooked": false,

"LiveBooked": false,

"Id": 17398578

},

...

]

}
```

##### DataSnapshot

##### Command:

[API\_URL] /api/DataService/DataSnapshot?token=[token]&isLive=([true] or [false])

##### Response:

```
{

"Command": "LiveSnapshot",

"ResultCode": 0,

"Objects": [

{

"Matches": [...],

"Id": 0,

},

],

}
```

##### EventTypes

##### Command:

[API\_URL] /api/DataService/EventType?token=[token]

##### Response:

```
{

"Command": "GetEventTypes",

"ResultCode": 0,

"Objects": [

{

"SportId": 190,

"SportType": "3x3 Basketball",

"EventType": "period",

"EventTypeId": 0,

"TypeId": 2959,

"IsActive": true,

"Id": 0,

"Name": "Period"

},

{

"SportId": 190,

"SportType": "3x3 Basketball",

"EventType": "point",

"EventTypeId": 0,

"TypeId": 2960,

"IsActive": true,

"Name": "Point"

},

{

"SportId": 201,

"SportType": "Air Hockey",

"EventType": "point",

"EventTypeId": 0,

"TypeId": 3025,

"IsActive": true,

"Name": "Point"

}

...

]

}
```

##### Periods

##### Command:

[API\_URL] /api/DataService/Period?token=[token]

##### Response:

```
{

"Command": "GetPeriods",

"ResultCode": 0,

"Objects": [

{

"SportId": 6,

"Sport": "American Football",

"IsPeriod": true,

"Number": 0,

"Name": "Period 1 Start"

},

{

"SportId": 6,

"Sport": "American Football",

"IsPeriod": true,

"Number": 1,

"Name": "Period 1 End"

},

{

"SportId": 6,

"Sport": "American Football",

"IsPeriod": true,

"Number": 2,

"Name": "Period 2 Start"

}

...

]

}
```

##### MarketTypeBooking

##### Command:

[API\_URL] /api/DataService/MarketTypeBooking?token=[token]&IsLive=([true] or [false])

##### Response:

```
{

"Command": "GetMarketTypeBookings",

"ResultCode": 0,

"Objects": [

{

"SportId": 4,

"IsLive": true,

"IsUnSubscribed": true,

"Id": 12206

},

{

"SportId": 37,

"IsLive": true,

"IsUnSubscribed": true,

"Id": 12429

},

{

"SportId": 11,

"IsLive": true,

"IsUnSubscribed": false,

"Id": 369

},

{

"SportId": 11,

"IsLive": true,

"IsUnSubscribed": false,

"Id": 370

},

...

]

}
```

##### SportOrder

##### Command:

[API\_URL] api/DataService/SportOrder?token=[token]

##### Response:

```
{

"Command": "GetSportOrders",

"ResultCode": 0,

"Objects": [

{

"ObjectId": 1,

"ObjectTypeId": 1,

"Order": 2,

"Id": 0

},

{

"ObjectId": 8777,

"ObjectTypeId": 5,

"SportId": 18,

"Order": 14,

"Id": 0

},

{

"ObjectId": 190,

"ObjectTypeId": 1,

"Order": 4,

"Id": 0

},

...

]

}
```

##### RMQ updates:

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"Date": "2021-05-13T08:04:00Z",

"CompetitionId": 18273221,

"SportId": 1,

"RegionId": 1,

"LiveStatus": 2,

"MatchStatus": 2,

"IsVisible": false,

"IsSuspended": false,

"IsLive": false,

"IsStarted": true,

"MatchMembers": [

{

"TeamName": "Borussia Monchengladbach (DangerDim77)",

"TeamId": 614099,

"IsHome": false,

"NameId": 3565990

},

{

"TeamName": "Bayer 04 Leverkusen (Lion)",

"TeamId": 584285,

"IsHome": true,

"NameId": 3027432

}

],

"Stat": {

"EventId": 17998285,

"CurrentMinute": 90,

"MatchLength": 90,

"Score": "0:0",

"CornerScore": "0:0",

"YellowcardScore": "0:0",

"RedcardScore": "0:0",

"ShotOnTargetScore": "0:0",

"ShotOffTargetScore": "0:0",

"DangerousAttackScore": "0:0",

"SportKind": 1,

"Set1Score": "0:0",

"Set2Score": "0:0",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Server": 0,

"Info": "0 : 0, (0:0), (0:0); FT",

"Period": 6,

"PenaltyScore": "0:0",

"FreeKickScore": "0:0",

"Set1YellowCardScore": "0:0",

"Set2YellowCardScore": "0:0",

"Set1CornerScore": "0:0",

"Set2CornerScore": "0:0",

"Set1RedCardScore": "0:0",

"Set2RedCardScore": "0:0",

"AdditionalMinutes": 0,

"HomeShirtColor": "000000",

"AwayShirtColor": "000000",

"HomeShortsColor": "000000",

"AwayShortsColor": "000000",

"PeriodCount": 0,

"Id": 0

},

"IsBooked": true,

"IsStatAvailable": false,

"Id": 17998285,

"ObjectVersion": "153644086434"

}

],

"Type": "Match",

"SocketTime": "2021-05-13T08:36:18.0148449Z"

}
```

```
{

"Command": "MatchStat",

"ResultCode": 0,

"Objects": [

{

"EventId": 17995578,

"Score": "0:0",

"AcesScore": "0:0",

"DoubleFaultScore": "0:0",

"SportKind": 41,

"SetScore": "0:0",

"Set1Score": "0:0",

"Set2Score": "0:0",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Set6Score": "0:0",

"Set7Score": "0:0",

"GameScore": "0:0",

"Server": 1,

"Info": "",

"Period": 0,

"SetCount": 5,

"PeriodCount": 0,

"Id": 0

}

],

"SocketTime": "2021-05-13T08:36:10.2991076Z"

}
```

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"EventId": 17998285,

"EventTimeUtc": "2021-05-13T08:36:17.9193385",

"EventType": "period",

"EventTypeId": 10,

"Side": 1,

"CurrentMinute": 0,

"MatchLength": 90,

"Score": "0:0",

"CornerScore": "0:0",

"YellowcardScore": "0:0",

"RedcardScore": "0:0",

"ShotOnTargetScore": "0:0",

"ShotOffTargetScore": "0:0",

"DangerousAttackScore": "0:0",

"SportKind": 1,

"Set1Score": "0:0",

"Set2Score": "0:0",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Server": 0,

"Info": "",

"Period": 3,

"PenaltyScore": "0:0",

"FreeKickScore": "0:0",

"Set1YellowCardScore": "0:0",

"Set2YellowCardScore": "0:0",

"Set1CornerScore": "0:0",

"Set2CornerScore": "0:0",

"Set1RedCardScore": "0:0",

"Set2RedCardScore": "0:0",

"AdditionalMinutes": 0,

"PeriodCount": 0,

"Id": 453011407

}

],

"Type": "MatchStat",

"SocketTime": "2021-05-13T08:36:17.9351673Z"

}
```

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"Handicap": -2.5,

"Selections": [

{

"Handicap": -2.5,

"Order": 2,

"Price": 0,

"OriginalPrice": 5.5084,

"IsSuspended": false,

"IsVisible": false,

"Outcome": 0,

"Kind": "Away",

"SelectionTypeId": 6646,

"Id": 2122051696,

"Name": "{t2} ({h})",

"NameId": 143718,

"ObjectVersion": "153652722363"

},

{

"Handicap": 2.5,

"Order": 1,

"Price": 0,

"OriginalPrice": 1.2218,

"IsSuspended": false,

"IsVisible": false,

"Outcome": 0,

"Kind": "Home",

"SelectionTypeId": 6645,

"Id": 2122051729,

"Name": "{t1} ({-h})",

"NameId": 143717,

"ObjectVersion": "153652722619"

}

],

"Sequence": 1,

"PointSequence": 0,

"IsSuspended": true,

"MatchId": 17987907,

"IsVisible": false,

"MarketTypeId": 2449,

"CashOutAvailable": true,

"IsSelectionsOrderedByPrice": false,

"Id": 671010480,

"Name": "{sw} Set Games Handicap",

"NameId": 135942,

"ObjectVersion": "153652763253"

},

{

"Handicap": 2.5,

"Selections": [

{

"Order": 2,

"Price": 1.78,

"OriginalPrice": 1.92,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Under",

"SelectionTypeId": 16080,

"Id": 2120494803,

"Name": "Under ({h})",

"NameId": 2270005,

"ObjectVersion": 153652761839

},

{

"Order": 1,

"Price": 1.92,

"OriginalPrice": 2.087,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Over",

"SelectionTypeId": 16079,

"Id": 2120494804,

"Name": "Over ({h})",

"NameId": 2270004,

"ObjectVersion": "153652761843"

}

],

"Sequence": 0,

"PointSequence": 0,

"IsSuspended": false,

"MatchId": 17939626,

"IsVisible": true,

"MarketTypeId": 11630,

"CashOutAvailable": false,

"IsSelectionsOrderedByPrice": false,

"Id": 670478664,

"Name": "Corners: 2nd Half Team 1 Total",

"NameId": 2270000,

"ObjectVersion": "153652761790"

}

],

"Type": "Market",

"SocketTime": "2021-05-13T08:36:24.3228293Z"

}
```

```
{

"Command": "MarketTypeUpdate",

"ResultCode": 0,

"Objects": [

{

"Kind": "SoutheastDivisionWinner",

"IsHandicap": false,

"IsOverUnder": false,

"SportId": 3,

"SelectionCount": 0,

"IsDynamic": true,

"TypeFlag": 2,

"DisplayOrder": 180,

"Id": 12198,

"Name": "Southeast Division Winner",

"NameId": 2457847

}

],

"SocketTime": "2021-05-17T05:25:58.4238204Z"

}
```

```
{

"Command": "UnBookedObject",

"ResultCode": 0,

"Objects": [

{

"ObjectId": 19095744,

"ObjectTypeId": 4,

"IsLive": true,

"Id": 0

}

],

"SocketTime": "2022-01-17T09:32:49.5004650Z"

}
```

```
{

"Command": "BookedObject",

"ResultCode": 0,

"Objects": [

{

"ObjectId": 20005748,

"ObjectTypeId": 4,

"IsLive": true,

"Id": 0

}

],

"SocketTime": "2022-06-07T07:40:22.5887048Z"

}
```
