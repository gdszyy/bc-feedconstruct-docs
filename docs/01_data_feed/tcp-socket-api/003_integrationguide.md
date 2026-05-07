---
title: Integration Guide
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=integrationGuide
current_loc: feedSocketApi
location: integrationGuide
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Integration Guide

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=integrationGuide`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `integrationGuide` |

## 文档正文
Integration Guide

- Purpose

The following Guide is intended for Odds Feed Data integrating Partners, who choose the TCP-Socket as the integration type.

- Data Integration Steps

There are several steps for correct integration, which are as follows:

- Establishing the connection

The Client must connect to the Odds Feed endpoint in order to begin the integration process.
To open a connection, the Client must use the provided WebAPI credentials; if the username
and/or password are incorrect, the connection will be denied. Also, the IP address
attempting to connect should be whitelisted by Odds Feed's network administrators; if the IP
address is not whitelisted or is blacklisted, the connection will be denied.

- Login

By sending the "Login" command, the Client should connect using an already whitelisted IP
address and the provided WebAPI credentials. The Client should not request any further
commands until the response to the "Login" command is received. The response for further
commands won’t be received, in case the Login command’s “Success” response is not yet
received. If everything is correct, the ResponseCode will be 0, which means "Success,"
allowing the Client to proceed to the next step. If the "Login" command returned an error
message, the Client should check the received ResponseCode to determine the cause of the
error. The following ResponseCodes are possible for the "Login" command:- 0, 10, 11, 13, 16,
18, 22, 23, 105.

- 0 - Success
- 10 - SystemUnavailable
- 11 - InternalError
- 13 - IncorrectRequest
- 16 - NotAuthorized
- 18 - NotAllowed
- 22 - NotAllowedUser
- 23 - NotAllowedIp
- 105 - LimitReached (when there are already 4 existing connections - max allowed amount)

- Heartbeat

Once the connection is established and the "Success" response from the Login command is
received, the Client should begin sending the "Heartbeat" command once every 5-10 seconds to
keep the opened connection alive. Odds Feed's system will automatically close the connection
if the "Heartbeat" is not received.

- Synchronisation of the general Data

After connecting to Odds Feed for the first time, the initial start, the Client must synchronise with the following commands:

- GetSports
- GetRegions
- GetCompetitions
- GetMarketTypes
- GetSelectionTypes
- GetEventTypes
- GetPeriods

The above-mentioned commands will provide the Client with the necessary Data, to get started with the integration. There is no need to use these commands frequently because all updates are pushed by Odds Feed, and in case of correct integration, there will be no data loss. For the integration, it is recommended to run the above-mentioned commands once a day to be safe from missed/lost updates from the feed.

- Subscription

Following the initial start and synchronisation of the Data, the Client should now send Subscribe commands to begin receiving the actual Data for the Booked events. SubscribeMatches for events available in Live, and SubscribePreMatches for events available in PreMatch.

Because the PreMatch is booked by default, the Client will receive the PreMatch data as
usual in the event of an initial start, but for the Live offer, the events must be booked
via Odds Feed Portal or via Commands. The Subscribe commands should be executed
with the parameters "IncludeMatches" and "IncludeMarkets" with the value "true" to obtain
full Data in the case of initial start or when the outage lasted longer than an hour
(snapshot). The "Subscribe" commands should be provided with the parameter "GetChangesFrom,"
which specifies the last disconnection time or the most recent update time in minutes, if it
is not the initial start or when there has been a reconnection or restart after a failure
lasting less than one hour. Only changes made after the given date and time are included in
the response.

**Note:** The SubscribeMatches and SubscribePrematches commands provide the last known updates for matches within their respective offers at the time of the request. SubscribeMatches returns only active Live matches, while SubscribePrematches returns only Prematch matches. These responses are limited to matches you have booked within your account for the requested type—Live or Prematch. Any matches that are Completed, Cancelled, or not currently visible will not be included in these subscription responses; for these cases, please use the **GetMatchByID command**.

Following any releases or scheduled maintenance on our side, we recommend requesting the **SubscribeMatches**, **SubscribePreMatches** commands. If any events are missing from these responses, please use the **GetMatchByID** command to retrieve them. By reconciling the data from both commands, you can ensure your offer remains up to date. Please remember to remove any completed, canceled, or non-visible events to prevent "stuck" matches in your system.

- Implementation Strategy

The following are the most important aspects:

- Examine the Commands in detail, paying close attention to their responses and any response errors in the message syntax (list of the possible errors for an exact command).
- Sending Heartbeat after a successful login.
- “Get” and “Subscribe” commands should only be used in emergencies, when the Feed hasn't
  pushed any updates, or when something unexpected has been received. Once the Subscribe
  command has been already sent, and the Client is requesting another “subscribe” command, the
  response will be an error message. If the Client has been connected but didn’t request
  “Subscribe” commands, the “Get” commands will not receive any response.

##### Message Syntax

The Client should calculate the length of the command, and send the (n) length value in the first 4 bytes, before the command JSON.

##### Data Storage

We recommend that the Client keeps the logs for the received Data for further inspections if necessary.
