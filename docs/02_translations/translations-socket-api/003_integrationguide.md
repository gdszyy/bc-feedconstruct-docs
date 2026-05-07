---
title: Integration Guide
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=integrationGuide
current_loc: translationSocketApi
location: integrationGuide
top_category: TRANSLATIONS SOCKET API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Integration Guide

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=integrationGuide`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS SOCKET API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationSocketApi` |
| location | `integrationGuide` |

## 文档正文
Integration Guide

- Purpose

The following Guide is intended for Translation-Data integrating Partners, who choose the TCP-Socket as the integration type.

- Data Integration Steps

There are several steps for correct integration, which are as follows:

- Establishing the connection

The Client must connect to the Translation’s endpoint in order to begin the integration
process. To open a connection, the Client must use the provided WebAPI credentials; if the
username and/or password are incorrect, the connection will be denied. Also, the IP address
attempting to connect should be whitelisted by Odds Feed's network administrators; if the IP
address is not whitelisted or is blacklisted, the connection will be denied.

- Login

By sending the "Login" command, the Client should connect using an already whitelisted IP
address and the provided WebAPI credentials. The Client should not request any further
commands until the response to the "Login" command is received. The response for further
commands won’t be received, in case the Login command’s “Success” response is not yet
received. If the "Login" command returned an error message, the Client should check the
criteria mentioned in the error message.

- Heartbeat

Once the connection is established and the "Success" response from the Login command is
received, the Client should begin sending the "Heartbeat" command once every 5-10 seconds to
keep the opened connection alive. Translation's system will automatically close the
connection if the "Heartbeat" is not received.

- Synchronisation of the general Data

After connecting to Odds Feed for the first time, the initial start, the Client must synchronise with the following commands:

- GetLanguages
- GetTranslations

The above-mentioned commands will provide the Client with the necessary Data, to get started with the integration. There is no need to use these commands frequently because all updates are pushed by Translation, and in case of correct integration, there will be no data loss. For the integration, it is recommended to run the above-mentioned commands once a day to be safe from missed/lost updates from the feed.

- Subscription

Following the initial start and synchronisation of the Data, the Client should now send the
“SubscribeTranslation” command to begin receiving the actual translated Data. The
“SubscribeTranslation” commands should be called when the Client starts the connection. The
“SubscribeTranslation” command can be called with the parameter "LangIDs," which specifies
the exact languages with LanguageIDs. In case the Client needs Translation Data for all the
supported languages, the command “SubscribeTranslation” should be called without mentioning
the “LangIDs” parameter, for full data.

- Implementation Strategy

The following are the most important aspects:

- Examine the Commands in detail, paying close attention to their responses and any response errors in the message syntax.
- Sending Heartbeat after a successful login.
- “Get” and “Subscribe” commands should only be used in emergencies, when the Translation
  hasn't pushed any updates, or when something unexpected has been received. Once the
  “SubscribeTranslations” command has been already sent, and the Client is requesting another
  “subscribe” command, the response will be an error message. If the Client has been connected
  but didn’t request “Subscribe” commands, the “Get” commands will not receive any response.

##### Message Syntax

The Client should calculate the length of the command, and send the (n) length value in the first 4 bytes, before the command JSON.
