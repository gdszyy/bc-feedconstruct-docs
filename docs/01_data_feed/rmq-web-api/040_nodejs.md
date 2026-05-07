---
title: Node.js
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=nodeJs
current_loc: oddsFeedRmqAndWebApi
location: nodeJs
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Node.js

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=nodeJs`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `nodeJs` |

## 文档正文
Node.js

#### Source

(Version 0.0.7)FeedConstruct.OddsFeed.src.zip

#### Usage Sample

```
import {config} from './config'

import Requestor from "./Requestor";

import RmqSubscribeClient from './RmqSubscribeClient';

const rmqSubscribeClient = new RmqSubscribeClient(config);

// Subscribe to RMQ

rmqSubscribeClient.subscribe((data: any) => {

console.log('data',data);

});

const requestor = new Requestor(config);

// Request data snapshot via web API

requestor.getDataSnapshot(true).then((res:any)=> {

console.log(res.data);

})
```
