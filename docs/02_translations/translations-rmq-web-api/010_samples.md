---
title: Samples
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=samples
current_loc: translationWebApi
location: samples
top_category: TRANSLATIONS RMQ & WEB API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Samples

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=samples`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS RMQ & WEB API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationWebApi` |
| location | `samples` |

## 文档正文
Samples

##### Languages

##### Command:

Request: [API\_URL] /api/Translation/Languages

##### Body:

```
[

{

"id": "it",

"name": "Italian",

},

{

"id": "en",

"name": "English",

},

{

"id": "bg",

"name": "Bulgarian",

},

{

"id": "cn",

"name": "Canadian English",

},

...

]
```

##### ByLanguage

##### Command:

Request: [API\_URL] /api/Translation/ByLanguage/{languageId}

##### Body:

```
[

{

"languageId": "fr",

"translations": [

{

"id": 511006,

"text": "Kolkheti Khobi 2",

},

{

"id": 2126263,

"text": "Etape d'élimination - Maroc",

},

...

],

},

]
```

##### ByID

##### Command:

Request: [API\_URL] /api/Translation/ById/{languageId}?id={id}

##### Body:

```
[

{

"languageId": "pl",

"translations": [

{

"id": 281451,

"text": "Na żywo"

}

],

},

]
```
