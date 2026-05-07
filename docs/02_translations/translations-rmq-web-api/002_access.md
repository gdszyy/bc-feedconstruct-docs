---
title: Access Method
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=access
current_loc: translationWebApi
location: access
top_category: TRANSLATIONS RMQ & WEB API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Access Method

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=access`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS RMQ & WEB API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationWebApi` |
| location | `access` |

## 文档正文
Access Method

We transmit the data with our clients using JSON (<http://www.json.org/> ) format. The communication between the FeedConstruct system and the client system is being proceeded through the Web API and RMQ. All retrieved data compressed as GZIP (in WebAPI and RMQ).

**Note:** To ensure consistency and clarity in queue naming conventions,use the following name template:

If you want to receive translation updates through RMQ in addition to the Web API methods, contact your account manager.

- RMQ

Host: translations-rmq.feedstream.org: 5674

QueueName: T[PartnerId]\_name (e.g. translation)\*

Exchange: Translation

MaxConnectionCount: 4

MaxChannelCount: 64

**Note:** For Partners to receive updates, the queues they create must be bound to the Exchange. Updates will not be sent to these queues until the Partner manually configures this necessary binding.

- Web API

translations-api.feedstream.org
