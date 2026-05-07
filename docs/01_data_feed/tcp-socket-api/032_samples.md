---
title: Samples
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=samples
current_loc: feedSocketApi
location: samples
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Samples

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=samples`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `samples` |

## 文档正文
Samples

After the subscription user will get all updates of subscribed games. The update format is

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"Date": "2021-05-11T20:28:00Z",

"CompetitionId": 18274437,

"SportId": 209,

"RegionId": 1,

"LiveStatus": 0,

"MatchStatus": 0,

"IsVisible": true,

"IsSuspended": false,

"IsLive": false,

"IsStarted": false,

"MatchMembers": [

{

"TeamName": "Bloodshot",

"TeamId": 586680,

"IsHome": true,

"NameId": 3037910

},

{

"TeamName": "Bruno",

"TeamId": 586682,

"IsHome": false,

"NameId": 3037912

}

],

"IsBooked": true,

"Id": 1502899561,

"ObjectVersion": "153296160718"

}

],

"Type": "Match",

"SocketTime": "2021-05-11T20:19:13.6787093Z"

}
```

OR

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"Date": "2021-05-11T09:00:00Z",

"CompetitionId": 4665,

"SportId": 1,

"RegionId": 21,

"LiveStatus": 1,

"MatchStatus": 1,

"IsVisible": true,

"IsSuspended": false,

"IsLive": true,

"IsStarted": true,

"MatchMembers": [

{

"TeamName": "North Pine SC",

"TeamId": 177735,

"IsHome": false,

"NameId": 362008

},

{

"TeamName": "North Brisbane FC",

"TeamId": 311951,

"IsHome": true,

"NameId": 1884625

}

],

"Stat": {

"EventId": 17984325,

"CurrentMinute": 64,

"MatchLength": 90,

"Score": "1:4",

"CornerScore": "4:3",

"YellowcardScore": "1:0",

"RedcardScore": "0:0",

"ShotOnTargetScore": "0:0",

"ShotOffTargetScore": "0:0",

"DangerousAttackScore": "11:10",

"SportKind": 1,

"Set1Score": "0:3",

"Set2Score": "1:1",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Server": 0,

"Info": "1 : 4, (0:3), (1:1) 64`",

"Period": 2,

"PenaltyScore": "0:0",

"FreeKickScore": "0:0",

"Set1YellowCardScore": "1:0",

"Set2YellowCardScore": "0:0",

"Set1CornerScore": "2:2",

"Set2CornerScore": "2:1",

"Set1RedCardScore": "0:0",

"Set2RedCardScore": "0:0",

"AdditionalMinutes": 0,

"HomeShirtColor": "000000",

"AwayShirtColor": "000000",

"HomeShortsColor": "000000",

"AwayShortsColor": "000000",

"PeriodCount": 0,

"CurrentShot": 0,

"Id": 0

},

"IsBooked": true,

"IsStatAvailable": false,

"IsNeutralVenue": false,

"Id": 17984325,

"ObjectVersion": "153132013563"

}

],

"Type": "Match",

"SocketTime": "2021-05-11T10:27:38.0842466Z"

}
```

OR

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"Handicap": 0.5,

"Selections": [

{

"Order": 2,

"Price": 1.44,

"OriginalPrice": 1.5273,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Under",

"SelectionTypeId": 6696,

"Id": 2117808710,

"Name": "Under ({h})",

"NameId": 143766,

"ObjectVersion": "153168938422"

},

{

"Order": 1,

"Price": 2.59,

"OriginalPrice": 2.8965,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Over",

"SelectionTypeId": 6695,

"Id": 2117808699,

"Name": "Over ({h})",

"NameId": 143765,

"ObjectVersion": "153168938402"

}

],

"Sequence": 0,

"PointSequence": 0,

"IsSuspended": false,

"MatchId": 17986173,

"IsVisible": true,

"MarketTypeId": 5502,

"CashOutAvailable": true,

"IsSelectionsOrderedByPrice": false,

"Id": 669565215,

"Name": "Team 2 Total Goals",

"NameId": 135983,

"ObjectVersion": "153168938356"

},

{

"Handicap": 1.5,

"Selections": [

{

"Handicap": -1.5,

"Order": 1,

"Price": 4.15,

"OriginalPrice": 5.064,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Home",

"SelectionTypeId": 5454,

"Id": 2117866711,

"Name": "{t1} ({-h})",

"NameId": 143481,

"ObjectVersion": "153168938446"

},

{

"Handicap": 1.5,

"Order": 2,

"Price": 1.19,

"OriginalPrice": 1.2461,

"IsSuspended": false,

"IsVisible": true,

"Outcome": 0,

"Kind": "Away",

"SelectionTypeId": 5455,

"Id": 2117866710,

"Name": "{t2} ({h})",

"NameId": 143482,

"ObjectVersion": "153168938440"

}

],

"Sequence": 0,

"PointSequence": 0,

"IsSuspended": false,

"MatchId": 17986173,

"IsVisible": true,

"MarketTypeId": 5503,

"CashOutAvailable": true,

"HomeScore": 1,

"AwayScore": 0,

"IsSelectionsOrderedByPrice": false,

"Id": 669585548,

"Name": "Goals Handicap",

"NameId": 135984,

"ObjectVersion": "153168938374"

}

],

"Type": "Market",

"SocketTime": "2021-05-11T10:27:52.6576213Z"

}
```

OR

```
{

"Command": "MatchUpdate",

"ResultCode": 0,

"Objects": [

{

"EventId": 17982665,

"EventTimeUtc": "2021-05-11T10:27:52.0736813",

"EventType": "ballSafe",

"EventTypeId": 20,

"Side": 0,

"CurrentMinute": 72,

"MatchLength": 90,

"Score": "2:2",

"CornerScore": "3:5",

"YellowcardScore": "0:3",

"RedcardScore": "0:0",

"ShotOnTargetScore": "2:0",

"ShotOffTargetScore": "4:7",

"DangerousAttackScore": "36:46",

"SportKind": 1,

"Set1Score": "1:1",

"Set2Score": "1:1",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Server": 0,

"Info": "2 : 2, (1:1), (1:1) 72`",

"Period": 2,

"PenaltyScore": "1:0",

"FreeKickScore": "10:9",

"Set1YellowCardScore": "0:1",

"Set2YellowCardScore": "0:2",

"Set1CornerScore": "2:2",

"Set2CornerScore": "1:3",

"Set1RedCardScore": "0:0",

"Set2RedCardScore": "0:0",

"AdditionalMinutes": 0,

"PeriodCount": 0,

"CurrentShot": 0,

"Id": 452312894

}

],

"Type": "MatchStat",

"SocketTime": "2021-05-11T10:27:52.1567667Z"

}
```

Match Statistics Information: After subscribing to Live Matches the following statistics and score information are sent.

```
{

"Command": "MatchStat",

"ResultCode": 0,

"Objects": [

{

"IsTimeout": true,

"EventId": 17976418,

"CurrentMinute": 41,

"MatchLength": 80,

"Score": "0:0",

"CornerScore": "3:0",

"YellowcardScore": "1:0",

"RedcardScore": "0:0",

"ShotOnTargetScore": "0:0",

"ShotOffTargetScore": "1:0",

"DangerousAttackScore": "19:9",

"SportKind": 1,

"Set1Score": "0:0",

"Set2Score": "0:0",

"Set3Score": "0:0",

"Set4Score": "0:0",

"Set5Score": "0:0",

"Server": 0,

"Info": "0 : 0, (0:0); HT",

"Period": 1,

"PenaltyScore": "0:0",

"FreeKickScore": "6:13",

"Set1YellowCardScore": "1:0",

"Set2YellowCardScore": "0:0",

"Set1CornerScore": "3:0",

"Set2CornerScore": "0:0",

"Set1RedCardScore": "0:0",

"Set2RedCardScore": "0:0",

"AdditionalMinutes": 0,

"PeriodCount": 0,

"CurrentShot": 0,

"Id": 0

}

],

"SocketTime": "2021-05-11T10:28:26.0654311Z"

}
```

CompetitionUpdate Information: After any change in competition following information is being sent

```
{

"Command": "CompetitionUpdate",

"ResultCode": 0,

"Objects": [

{

"SportId": 29,

"RegionId": 1,

"IsTeamsReversed": false,

"Id": 18277992,

"Name": "Spirit League",

"NameId": 5506699

},

{

"SportId": 44,

"RegionId": 213,

"IsTeamsReversed": false,

"Id": 18278001,

"Name": "EFC 85",

"NameId": 5510801

},

{

"SportId": 17,

"RegionId": 184,

"IsTeamsReversed": false,

"Id": 18277994,

"Name": "Gromda.Outright",

"NameId": 5506724

},

{

"SportId": 27,

"RegionId": 1,

"IsTeamsReversed": false,

"Id": 18277995,

"Name": "testt",

"NameId": 5506845

}

]

}
```

SelectionTypeUpdate information: After any change of selection type information following information is being sent

```
{

"Command": "SelectionTypeUpdate",

"ResultCode": 0,

"Objects": [

{

"MarketTypeId": 8703,

"Order": 2,

"Kind": "No",

"Id": 9196,

"Name": "No",

"NameId": 281982

}

],

"SocketTime": "2017-03-09T07:05:17.0995545Z"

}
```

MarketTypeUpdate information: After any change of marketType information following information is being sent

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

VoidNotification information: After any change of voidNotification information following information is being sent

```
{

"Command": "VoidNotification",

"ResultCode": 0,

"Objects": [

{

"FromDate": "2021-05-17T10:02:28+04:00",

"ToDate": "2021-05-17T10:02:30+04:00",

"ObjectType": 16,

"ObjectId": 2130183126,

"Reason": "Returned for a specified period(17/05/2021 10:02:28 +04:00 - 17/05/2021 10:02:30 +04:00) Selection (2130183126) Reason (Incorrect Odd (Correct odd 120.5 ))",

"VoidAction": 1,

"Id": 0

}

],

"SocketTime": "2021-05-17T06:10:02.4412382Z"

}
```

UnBookedObject information: After Unbook object following information is being sent

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

BookedObject information: After Bbook object following information is being sent

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
