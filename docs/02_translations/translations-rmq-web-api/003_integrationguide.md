---
title: Integration Guide
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=integrationGuide
current_loc: translationWebApi
location: integrationGuide
top_category: TRANSLATIONS RMQ & WEB API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Integration Guide

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=integrationGuide`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS RMQ & WEB API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationWebApi` |
| location | `integrationGuide` |

## 文档正文
Integration Guide

- Purpose

The following Guide is intended for Translation-Data integrating Partners, who choose the Translations RMQ and Web-API as the integration type.

- Data Integration Steps

There are several steps for correct integration, which are as follows:

- Establishing the connection

The Client must connect to the Translation’s endpoint in order to begin the integration
process. To open a connection, the Client must use the provided WebAPI credentials; if the
username and/or password are incorrect, the connection will be denied. Also, the IP address
attempting to connect should be whitelisted by Odds Feed's network administrators; if the IP
address is not whitelisted or is blacklisted, the connection will be denied.

- Synchronisation of the general Data

After connecting to Odds Feed for the first time, the initial start, the Client must synchronise with the following commands:

- GetLanguages
- GetTranslations

- Subscription

Client can use `languageid` as a binding key(s) for subscribing for the exact language(s) translation changes.

The above-mentioned commands will provide the Client with the necessary Data, to get started with the integration.

- Implementation Strategy

The following are the most important aspects:

- Examine the Commands in detail, paying close attention to their responses and any response errors in the message syntax.
- The getByLanguage and Languages methods should not be called more frequently than once per hour for each language. If you require immediate updates for newly translated data, please use either the Translations RMQ or Socket API, which are designed to provide real-time updates efficiently.
