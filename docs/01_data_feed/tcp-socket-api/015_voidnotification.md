---
title: VoidNotification
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=voidNotification
current_loc: feedSocketApi
location: voidNotification
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# VoidNotification

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=voidNotification`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `voidNotification` |

## 文档正文
VoidNotification

| Field Name Field Name | Type Type | Description Description |  |
| --- | --- | --- | --- |
| FromDate\* | datetime |  |
| ToDate\* | datetime |  |
| ObjectType\* | int | • 4 - match • 13 - market • 16 - selection |
| ObjectId\* | int |  |
| Reason | string | VoidReason |
| VoidAction\* | short | • 1 - void • 2 - unvoid |
