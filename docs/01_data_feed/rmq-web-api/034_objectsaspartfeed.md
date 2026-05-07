---
title: Objects as part of the feed
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=objectsAsPartFeed
current_loc: oddsFeedRmqAndWebApi
location: objectsAsPartFeed
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Objects as part of the feed

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=objectsAsPartFeed`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `objectsAsPartFeed` |

## 文档正文
Objects as part of the feed

|  |  |
| --- | --- |
| **Sport Update** | ```  {  "Command": "SportUpdate" ,  "ResultCode": 0 ,  "Objects": [ < Sport objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **Region Update** | ```  {  "Command": "RegionUpdate" ,  "ResultCode": 0 ,  "Objects": [ < Region objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **Competition Update** | ```  {  "Command": "CompetitionUpdate" ,  "ResultCode": 0 ,  "Objects": [ < Competition objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **Match Update** | ```  {  "Command": "MatchUpdate" ,  "ResultCode": 0 ,  "Type": "Match" ,  "Objects": [ < Match objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **Market Update** | ```  {  "Command": "MatchUpdate" ,  "ResultCode": 0 ,  "Type": "Market" ,  "Objects": [ < Market objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **MatchStat Update** | ```  {  "Command": "MatchUpdate" ,  "ResultCode": 0 ,  "Type": "MatchStat" ,  "Objects": [ < Stat objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ```  OR  ```  {  "Command": "MatchStat" ,  "ResultCode": 0 ,  "Objects": [ < Stat objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **MarketType Update** | ```  {  "Command": "MarketTypeUpdate" ,  "ResultCode": 0 ,  "Objects": [ < MarketType objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **SelectionType Update** | ```  {  "Command": "SelectionTypeUpdate" ,  "ResultCode": 0 ,  "Objects": [ < SelectionType objects > ],  "SocketTime": "2016-11-22T14:34:18.8103475Z"  } ``` |
| **VoidNotification** | ```  {  "Command": "VoidNotification" ,  "ResultCode": 0 ,  "Objects": [< VoidNotification objects >],  "SocketTime": "2020-09-25T05:54:55.4081874Z"  } ``` |
| **GetDataSnapshot** | ```  {  "Command": "GetDataSnapshot" ,  "ResultCode": 0 ,  "Objects": [],  "SocketTime": "2021-11-22T07:31:40.0561525Z"  } ``` |
| **UnBookedObject** | ```  {  "Command": "UnBookedObject" ,  "ResultCode": 0 ,  "Objects": [< UnBookObject objects >],  "SocketTime": "2022-01-17T09:32:49.5004650Z"  } ``` |
| **BookObject** | ```  {  "Command": "BookedObject" ,  "ResultCode": 0 ,  "Objects": [< BookObject objects >],  "SocketTime": "2022-06-07T08:12:23.4084310Z"  } ``` |
