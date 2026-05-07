---
title: Integration Notes
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=integrationNotes
current_loc: feedSocketApi
location: integrationNotes
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Integration Notes

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=integrationNotes`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `integrationNotes` |

## 文档正文
Integration Notes

For Match and Market objects the field ‘ObjectVersion’ contains the numeric version number of
object(the type of field is STRING because of issues in some programming languages and platforms and
actually contains unsigned long numeric value). For correct integration the ‘version number’ check
of Match and Market object types must be implemented. All received updates where the received
version number for a concrete object is less than the current value, should be skipped. This logic
is actual for fast-changing markets (basketball, volleyball, etc.). If there is LiveResulting
subscription, there can't be subscription for live Matches and vice versa.

- Active item`s condition

The item will be considered active only if (IsSuspended = false and IsVisible = true) the following two conditions apply.
