---
title: Access Method
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=access
current_loc: feedSocketApi
location: access
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Access Method

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=access`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `access` |

## 文档正文
Access Method

We transmit the data with our clients using JSON (<http://www.json.org/>) format. The communication between the FeedConstruct system and the client system is done through TCP sockets.

- Connection information for development platform:

TCP:[odds-stream-test.feedstream.org](#) port 8077

MaxConnectionCount: 4

For security reasons, access to our servers is only allowed for whitelisted IPs. The client should implement functionality of sending and receiving messages through TCP connection. If the connection is interrupted, the client should implement reconnection logic. The Client should also send instant messages for at least every 10 seconds; otherwise the server will close the connection.

In case of any global issue in platform, server will close all connections with clients, until the issue is fixed. As a response of Login command clients will receive error with message: “System temporarily unavailable”.

**Note:** The provided endpoint is designated for connecting to the Stage environment. The endpoint for the Live environment will be shared with the Client upon mutual confirmation of readiness for transitioning to production.

- Connection Handling:

Client should open a TCP connection and keep it opened, in order to get updates from feed. Client should also implement a reconnection logic, which will automatically reconnect to the feed, in case of broken connections. Client should detect immediately the disconnection of the feed and reconnect, which is normal from the network point of view. One of the reasons can be continuous redistribution of client connections in case of heavy load on platform. Also for reconnection client should handle the following case also: when there is no update command and heartbeat from feed in 15 seconds.

**Note:**  In case of each reconnection client should send subscribe commands to get updates from the feed and also send (GetChangesFrom) parameter for getting only changes from the mentioned period.
