---
title: Web Methods
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=webMethods
current_loc: oddsFeedRmqAndWebApi
location: webMethods
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Web Methods

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=webMethods`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `webMethods` |

## 文档正文
Web Methods

##### Token

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| POST  **\*Token should not be requested more than once per 24 hours.** | ```  /api/DataService/Token ```  ```  {  "Params": [  {  "UserName": "username",  "Password": "password"  }  ]  } ``` | ```  {  "Token": "token",  "ResultCode": 0  } ```   OR  CommandResponse   ```  {  "Error": {  "Key": "InvalidUsernamePassword",  "Message": "Invalid Username and/or password"  },  "Objects": [],  "ResultCode": 15  } ```       [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 15 | **Mandatory:** • UserName • Password |

##### Sport

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/Sport?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Sport](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=sport)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

##### Region

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/Region?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Region](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=region)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

##### Competition

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/Competition?token=[token]&SportId=[sportId]&RegionId=[regionId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Competition](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=competition)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token  **Optional:** • SportId • RegionId |

##### MarketType

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/MarketType?token=[token]&SportId=[sportId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[MarketType](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketType)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token  **Optional:** SportId |

##### SelectionType

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/SelectionType?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[SelectionType](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=selectionType)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

##### Calendar

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET  **\*Will be received up to 10 days data.This command should not be called earlier than 30 minutes.** | ```  /api/DataService/Calendar?token=[token]&SportId=[sportId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Calendar Match](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=calendar-match)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17, 105\* | **Mandatory:** token  **Optional:** SportId |

##### Booking

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET  **This command should not be called earlier than 30 minutes.** | ```  /api/DataService/Booking?token=[token]&IsLive=([true] or [false]) ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[PartnerBooking](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=partnerBooking+)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17, 105\* | **Mandatory:** • token • IsLive: true (Live) | false (Prematch) |

##### Book

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| POST  **\*Not more than 100 objects can be booked with one command.** | ```  /api/DataService/Book?token=[token] ```  ```  {  Params: [  {  "ObjectId": ObjectId,  "ObjectTypeId": ObjectTypeId,  "IsSubscribed": true or false,  "IsLive": true or false,  "SportId": Sport Id,  },  ],  } ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Book](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=book)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 12, 14, 17, 18, 106, 109\* | **Mandatory:** • token • ObjectId • ObjectTypeId • IsSubscribed • IsLive • SportId (only for region) |

##### CompetitionById

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/CompetitionById?token=[token]&competitionId=[competitionId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Competition](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=competition)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 12, 17\* | **Mandatory:** • token • competitionId |

##### MatchById

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET  **\*This command should not be called more than 5000 times in 24h.** | ```  /api/DataService/MatchById?token=[token]&MatchId=[matchId]&IncludeMatchStats=([true] or [false]) ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Match](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=match)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 12, 17, 18, 100, 101, 105\* | **Mandatory:** • MatchId • token  **Optional:** IncludeMatchStats - true | false |

##### MarketTypeById

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/MarketTypeById?token=[token]&marketTypeId=[marketTypeId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[MarketType](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketType)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 12, 17\* | **Mandatory:** • marketTypeId • token |

##### SelectionTypeById

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/SelectionTypeById?token=[token]&selectionTypeId=[selectionTypeId] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[SelectionType](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=selectionType)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 12, 17\* | **Mandatory:** • selectionTypeId • token |

##### DataSnapshot

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET  **\*This command should not be called more than 10 times in 24h both for Live and Prematch.** | ```  /api/DataService/DataSnapshot?token=[token]&isLive=([true] or [false])&getChangesFrom=[minutes] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Match](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=match)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17, 18, 105\* | **Mandatory:** • token • IsLive: true (Live) | false (Prematch)  **Optional:** getChangesFrom - will return all matches changed in last mentioned minutes |

##### EventTypes

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/EventType?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[EventTypes](/documentation?currentLoc=eventTypes&location=eventTypes)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

##### Periods

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/Period?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[Periods](/documentation?currentLoc=eventTypes&location=periods)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

##### MarketTypeBooking

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/MarketTypeBooking?token=[token]&IsLive=([true] or [false]) ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[MarketTypeBooking](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=marketTypeBooking)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** • token • IsLive: true (Live) | false (Prematch) |

##### SportOrder

| Access method | Address | Result | Params |
| --- | --- | --- | --- |
| GET | ```  /api/DataService/SportOrder?token=[token] ``` | CommandResponse   Where ‘Objects’ field contains list of ‘[SportOrder](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=sportOrder)’ models.  [Possible Result Codes:](/documentation?currentLoc=oddsFeedRmqAndWebApi&location=resultCodes) 0, 11, 17\* | **Mandatory:** token |

\* If you receive a response code of 17, it indicates that your token is invalid. You must acquire a new token and retry the request. Failing to do so may lead to the requester’s IP being blocked by the system.
