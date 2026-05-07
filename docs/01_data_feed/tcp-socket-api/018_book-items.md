---
title: BookItems
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=book-items
current_loc: feedSocketApi
location: book-items
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# BookItems

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=book-items`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `book-items` |

## 文档正文
BookItems

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| ObjectId\* | int | Book or unbook object Id |
| ObjectTypeId\* | int | Object type Id: • Sport = 1 • Region = 2 • Competition = 3 • Match = 4 |
| IsSubscribed\* | bool | Indicating booking status |
| IsLive\* | bool | Indicating booking/unbooking type (Live or Prematch) |
| SportId | int | \* This field is mandatory only when booking a region. |
