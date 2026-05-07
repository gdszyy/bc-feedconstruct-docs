---
title: Web methods
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=webMethods
current_loc: translationWebApi
location: webMethods
top_category: TRANSLATIONS RMQ & WEB API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Web methods

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationWebApi&location=webMethods`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS RMQ & WEB API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationWebApi` |
| location | `webMethods` |

## 文档正文
Web methods

##### Languages \*

| Access method | Address | Result | Params |  |
| --- | --- | --- | --- | --- |
| **GET** | ```  /api/Translation/Languages ``` | Will return list of LanguageModel objects |  |

##### ByLanguage \*

| Access method | Address | Result | Params |  |
| --- | --- | --- | --- | --- |
| **GET** | ```  /api/Translation/ByLanguage/[languageId] ``` | Will return TranslationResponseModel object for given languageId | **Mandatory:** languageId |

##### ByID

| Access method | Address | Result | Params |  |
| --- | --- | --- | --- | --- |
| **GET** | ```  /api/Translation/ById/[languageId]?id=[id] ``` | Returns a TranslationResponseModel object containing the translation for the provided ID and LanguageID | **Mandatory:** • ID (required) – Identifier of the translation item • LanguageID (required) – ID of the desired language |

**\*Note**

The getByLanguage and Languages methods should not be called more frequently than once per hour for each language.
