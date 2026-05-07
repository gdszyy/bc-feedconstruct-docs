---
title: Message Syntax
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=messageSyntax
current_loc: feedSocketApi
location: messageSyntax
top_category: TCP SOCKET
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Message Syntax

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=feedSocketApi&location=messageSyntax`。

| 字段 | 值 |
|---|---|
| 一级分类 | TCP SOCKET |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `feedSocketApi` |
| location | `messageSyntax` |

## 文档正文
Message Syntax

All messages are in JSON format. For all the messages that are sent or received, first 4 byte (binary integer) is the length of the message and the rest is the message in JSON format. The format of the message is

```
{

"Command": "Command Name",

"Params": [

{ "param name": "value" },

{ "param name": "value" },

{ "param name": "value" },

{ "param name": "value" },

],

}
```

Where "Command Name" is one of our supported commands (See the list below) and "value" is a
parameter value for current command(if the current command involves the transmission of parameters).

Response Models

- CommandResponse

| Field Name | Type | Description |  |
| --- | --- | --- | --- |
| Objects | Array | Collection of corresponding command response models |
| Command | string | Command name |
| Error | [ResponseError](/documentation?currentLoc=feedSocketApi&location=responseError) |  |
| Type | string | The type of command |
| ResultCode | int |  |

- ResponseError

| Field Name | Type | Description |  |
| --- | --- | --- | --- |
| Key | string | Error key |
| Message | string | Error message |
| Id | int | Match Id **\*Note:** Will be received only in case of GetMatchById command |
