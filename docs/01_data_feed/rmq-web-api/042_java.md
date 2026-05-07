---
title: Java
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=java
current_loc: oddsFeedRmqAndWebApi
location: java
top_category: RABBIT MQ
product_line: OddsFeed 数据源服务
business_domain: 数据源服务 / OddsFeed
scraped_at: 2026-05-07T08:49:13.195Z
---

# Java

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=oddsFeedRmqAndWebApi&location=java`。

| 字段 | 值 |
|---|---|
| 一级分类 | RABBIT MQ |
| 产品线 | OddsFeed 数据源服务 |
| 业务域 | 数据源服务 / OddsFeed |
| currentLoc | `oddsFeedRmqAndWebApi` |
| location | `java` |

## 文档正文
Java

#### Source

(Version 0.0.2)FeedConstruct.OddsFeed.src.zip

#### Binary

FeedConstruct.OddsFeed.bin.zip

#### Usage Sample

```
public static void main(String[] args) throws Exception {

// define Live and PreMatch queues for partnerId=444555444

String[] queueNames = {"P444555444_live","P444555444_prematch"};

FeedRMQClient mqClient = new FeedRMQClient(

new FeedRMQConfig(

"[RmqHost]",

5673,

"xxxxxx", // RMQ userName

"xxxxxx", // RMQ password

queueNames), new ILogger() {

@Override

public void info(String msg) {

System.out.println(msg);

}

@Override

public void error(Exception e) {

e.printStackTrace();

}

}

);

mqClient.connect();

mqClient.subscribe();

RMQEvents rmqEvents = new RMQEvents();

// Susbscribe for market updates

rmqEvents.setMarketUpdate((marketCommandResponse, subscriptionType) -> {

System.out.println("market:subscriptionType: " + subscriptionType);

System.out.println("marketUpdateObject: " + marketCommandResponse);

});

// Susbscribe for match updates

rmqEvents.setMatchUpdate(((matchCommandResponse, subscriptionType) -> {

System.out.println("MatchUpdate: subscriptionType: " + subscriptionType);

System.out.println("matchCommandResponse: " + matchCommandResponse);

}));

// ............

mqClient.consumeAsync(rmqEvents);

FeedWebApiClient feedWebApiClient = new FeedWebApiClient(

new FeedApiConfig(

"[WEB API URL]/api/DataService",

"xxxxxx",// WEB API userName

"xxxxxx" // WEB API password

), new ILogger() {

@Override

public void info(String msg) {

System.out.println(msg);

}

@Override

public void error(Exception e) { e.printStackTrace(); }

});

System.out.println("**** getSports ****");

Callable> sportsAsync = feedWebApiClient.getSportsAsync();

CommandResponse call = sportsAsync.call();

System.out.println("Sports " + call);

System.out.println("**** getPreMatchDataSnapshot ****");

CommandResponse dataSnapShot = feedWebApiClient.getDataSnapshotAsync(false, 15).call();

System.out.println("PreMatchDataSnapshot" + dataSnapShot);

// ............

}
```
