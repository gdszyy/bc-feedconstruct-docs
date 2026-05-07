---
title: Access method
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=access
current_loc: oddsFeedRmqAndWebApi
location: access
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Access method

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=access`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `access` |

## 文档正文
Access method

We transmit the data with our clients using JSON (<http://www.json.org/>) format. The communication between the FeedConstruct system and the client system is being proceeded through the Web API and RMQ. All retrieved data compressed as GZIP (in WebAPI and RMQ).

- RMQ

Host: odds-stream-rmq-stage.feedstream.org (port: 5673)

Queues: P[PartnerId]\_live and P[PartnerId]\_prematch

MaxConnectionCount: 4

MaxChannelCount: 64

- Web API:

odds-stream-api-stage.feedstream.org (port:8070)
For security reasons, access to our servers is only allowed for whitelisted IPs. The Client will be provided with a platform address, PartnerId, Username and Password for Web API and separate credentials for RMQ, immediately after the execution of the agreement between the parties /Client & FeedConstruct/.

**Note:** The provided endpoints are designated for connecting to the Stage environment. Endpoints for the Live environment will be shared with the Client upon mutual confirmation of readiness for transitioning to production.

**Important:** To use API it is required to have an actual token, which must be updated every 24 hours. You have to use your credentials for the special Token method, which is described in this document.
