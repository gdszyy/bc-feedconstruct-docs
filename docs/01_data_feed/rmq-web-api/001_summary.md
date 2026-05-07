---
title: Summary
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary
current_loc: oddsFeedRmqAndWebApi
location: summary
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Summary

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=summary`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `summary` |

## 文档正文
## ODDS-FEED RMQ and WEB API

### Version 1.0.0

Summary

This document is intended for FeedConstruct Clients having own trading platform and willing to offer the end users (players) a comprehensive data on over 120 sports types. It describes how the client should implement the integration and the functionality that we provide via Web API and Rabbit MQ.

Web API specially designed for getting PreMatch and Live data snapshot and for static entities (Sport, Region, Competition, MarketType, SelectionType and etc.), also in Web API there was some methods for emergency and not frequent checking (ex. GetMatchById)

The API is updated on a regular basis, as we constantly add new features to it.

RabbitMQ is used for delivering asynchronous updates so that the clients don’t have to constantly send requests in order to receive the most recent data.

- Feed Usage and Security Notice

Partners must not engage in any activity that may disrupt, overload, or interfere with the performance and integrity of the feed. This includes, but is not limited to, DDoS attacks, unauthorized access, data breaches, malware distribution, or other malicious actions.

Any deviation from the valid integration process, protocols, or security measures outlined in the technical documentation may result in **immediate suspension or blocking of access** without prior notice.

FeedConstruct reserves the right to **conduct periodic audits and monitor partner systems** to ensure compliance and prevent malicious activity, without prior notice.
