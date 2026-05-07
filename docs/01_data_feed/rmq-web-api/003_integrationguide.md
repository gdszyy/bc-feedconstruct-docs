---
title: Integration Guide
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=integrationGuide
current_loc: oddsFeedRmqAndWebApi
location: integrationGuide
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Integration Guide

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=integrationGuide`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `integrationGuide` |

## 文档正文
Integration Guide

- Purpose

The following Guide is intended for Odds Feed Data integration Partners using Rabbit-MQ (RMQ) as the integration type.

- Data Integration Strategy

There are several steps for correct integration, which are as follows:

- Establishing the connection

The Client must connect to RMQ and Web-API in order to begin the integration process. To open a
connection, the Client must use the provided WebAPI credentials; if the username and/or password
are incorrect, the connection will be denied. Provided RMQ credentials are to connect to RMQ and
get the Data from there. While sending any command the Client should use a Web-API connection
with Web-API credentials. Also, the IP address attempting to connect should be whitelisted by
Odds Feed's network administrators; if the IP address is not whitelisted or is blacklisted, the
connection will be denied.

- Token

The “Token” should be requested via WebAPI connection, it should include Web-API credentials and should be updated once every 24 hours. The Token response is a key, which should be later used in every web method as a mandatory parameter, to identify the Client.

- Synchronisation of the general Data

After connecting to Odds Feed for the first time (the initial start), the Client must synchronise via Web-API connection, with the following commands:

- Sports
- Regions
- Competitions
- MarketTypes
- SelectionTypes
- EventTypes
- Periods

The above-mentioned commands will provide the Client with the necessary Data, to get started
with the integration. There is no need to use these commands frequently because all updates are
pushed by Odds Feed. For the integration, it is recommended to run the above-mentioned commands
once a day to be safe from missed/lost updates from the Feed.

- Subscription

The Client should subscribe to corresponding queues (P[PartnerID]\_live and
P[PartnerID]\_prematch) in Rabbit-MQ. After that, the Data will start to be received and the
Client is free to keep them or process them.

- DataSnapshot

Via Web-API the Client should take a snapshot for PreMatch and Live and process the Data. At the
initial start of the platform or in case of an outage of more than 1 hour the “DataSnapshot”
method should be called without the “getChangesFrom” parameter, in order to have a full data
snapshot. If the outage was less than 1 hour, the “getChangesFrom” parameter should be sent, to
get the snapshot for changes that were made after mentioned minutes.

**Note:** The DataSnapshot method provides the Live or Prematch data (depending on whether the isLive parameter is set to true or false) that is active at the moment of the request. These responses only include updates for matches booked within your account for the specific requested type—Live or Prematch; if a match is not part of your current bookings, it will not appear in the snapshot. Additionally, matches that are Completed, Cancelled, or otherwise not visible in the active feed are excluded from snapshot responses. To retrieve data for such events, please use the **GetMatchByID command**.

Following any releases or scheduled maintenance on our side, we recommend requesting the DataSnapshot command using both isLive: true and isLive: false parameters. If any events are missing from these responses, please use the GetMatchByID command to retrieve them. By reconciling the data from both commands, you can ensure your offer remains up to date. Please remember to remove any completed, canceled, or non-visible events to prevent "stuck" matches in your system.

- GetDataSnapshot

If the command "GetDataSnapshot" from the message queue is received, the client should take a
snapshot because it is how our feed notifies the client platform to retrieve any potentially
missed data. That can occur if there are any releases or problems with Feed's global platform.

- Implementation Strategy

The following are the most important aspects:

- Examine the Commands in detail, paying close attention to their responses and any response
  errors in the message syntax (list of the possible errors for an exact command).
- “Get” commands should only be used in emergencies, when the Feed hasn't pushed any updates, or
  when something unexpected has been received.
- Instruction to handle the unacknowledged (unacked) Data should be added.

##### Queue Management Restriction

Deleting queues is strictly prohibited.

##### Data Storage

We recommend that the Client keeps the logs for the received Data for further inspections if necessary.
