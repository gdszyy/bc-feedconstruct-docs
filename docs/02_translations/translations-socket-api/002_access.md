---
title: Access Method
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=access
current_loc: translationSocketApi
location: access
top_category: TRANSLATIONS SOCKET API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Access Method

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=access`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS SOCKET API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationSocketApi` |
| location | `access` |

## 文档正文
Access Method

We transmit the data with our clients using JSON format (<http://www.json.org/>). The communication between the FeedConstruct’s system and the client system is proceeded through TCP sockets.

- Connection information:

TCP:  [translations-stream.feedstream.org](#) port 8088

For security reasons, accessing to our servers is only allowed for whitelisted IPs. The client should implement functionality of sending and receiving messages through TCP connection. If the connection is interrupted, the client should implement reconnection logic. The Client should also send instant messages for at least every 10 seconds; otherwise the server will close the connection.

In case of any global issue in platform, server will close all connections with clients. The Client will be provided a Username and Password to log in to our servers, immediately after the execution of the agreement between the parties /Client & FeedConstruct/.
