---
title: Message Syntax
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=messageSyntax
current_loc: translationSocketApi
location: messageSyntax
top_category: TRANSLATIONS SOCKET API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Message Syntax

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=messageSyntax`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS SOCKET API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationSocketApi` |
| location | `messageSyntax` |

## 文档正文
Message Syntax

All messages are in JSON format. For all the messages that are sent or received, first 4 byte (binary integer) is the length of the message and the rest is the message in JSON format.

The format of the message is

```
{

"Command": "Command Name",

"Params": [

{

"param name": "value"

},

{

"param name": "value"

},

{

"param name": "value"

},

{

"param name": "value"

}

]

}
```

Where "Command Name" is one of our supported commands (See the list below) and "value" is a
parameter value for current command.
