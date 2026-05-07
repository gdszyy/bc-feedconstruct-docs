---
title: Summary
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=summary
current_loc: translationWebApi
location: summary
top_category: TRANSLATIONS RMQ & WEB API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Summary

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=summary`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS RMQ & WEB API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationWebApi` |
| location | `summary` |

## 文档正文
## TRANSLATIONS RMQ AND WEB API

### Version 0.01

Summary

This document is intended for FeedConstruct Clients already having our Odds Feed API. Only already integrated Clients can use this Web API. This document describes how the client should implement the integration and the functionality that we provide via Web API. Web API is designed especially for lookup commands for static entities (Translations, and Languages)
The API is updated on regular basis, as we constantly add new features to it.

RabbitMQ is used for delivering asynchronous updates so that the clients don’t have to constantly send requests in order to receive the most recent data.
