---
title: PartnerBooking
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=partnerBooking
current_loc: feedSocketApi
location: partnerBooking
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# PartnerBooking

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=partnerBooking`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `partnerBooking` |

## 文档正文
PartnerBooking

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| ObjectId\* | int | Booked or unbooked object Id |
| ObjectTypeId\* | int | Object type Id • Sport = 1 • Region = 2 • Competition = 3 • Match = 4 |
| SportId\* | int | Sport Id |
| RegionId\* | int | Region Id |
| CompetitionId\* | int | Competition Id |
| IsLive\* | bool | Indicating booking/unbooking type (Live or Prematch) |
| IsSubscribed\* | bool | Indicating booking status |
