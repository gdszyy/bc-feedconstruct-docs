---
title: Command List
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=commandList
current_loc: translationSocketApi
location: commandList
top_category: TRANSLATIONS SOCKET API
product_line: 翻译数据服务
business_domain: 翻译数据服务
scraped_at: 2026-05-07T08:49:13.195Z
---

# Command List

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=translationSocketApi&location=commandList`。

| 字段 | 值 |
|---|---|
| 一级分类 | TRANSLATIONS SOCKET API |
| 产品线 | 翻译数据服务 |
| 业务域 | 翻译数据服务 |
| currentLoc | `translationSocketApi` |
| location | `commandList` |

## 文档正文
Command List

| Command Name | Command Sample | Response |  |
| --- | --- | --- | --- |
| **Login**  Login Partner | ```  {  "Command": "Login",  "Params": [  {  "UserName": "*****",  "Password": "*****"  }  ]  } ``` | ```  {  "Command": "Login",  "Error": null,  "Type": null,  "Objects": []  } ```   OR    ```  {  "Command": "Login",  "Error": {  "Key": "InvalidUsernamePassword",  "Message": "Invalid username or password"  },  "Objects": []  } ``` |
| **HeartBeat**  You should send this command for every 5 seconds, if server doesn’t receive it for 10 seconds it’ll close the connection | ```  {  "Command": "HeartBeat"  } ``` | There is no response for this message. |
| **GetLanguages**  Getting All Languages | ```  {  "Command": "GetLanguages"  } ``` | ```  {  "Command": "GetLanguages",  "Error": null,  "Type": null,  "Objects": [  {  "LangId": "en",  "Name": "English"  },  {  "LangId": "de",  "Name": "German"  }  ]  } ``` |
| **GetTranslations**  Getting translations by languageId | ```  {  "Command": "GetTranslations",  "CompressResponseToGZip": true,  "Params": [  {  "LangIds": [  "en",  "de"  ]  }  ]  } ``` | ```  {  "Command": "GetTranslations",  "Error": null,  "Type": null,  "Objects": [  [  {  "Translations": {  "en": [  {  "Id": 40,  "Text": "Tennis"  },  {  "Id": 218342,  "Text": "FH Hafnarfjordur(Wom)"  }  ],  "de": [  {  "Id": 40,  "Text": "Tennis"  },  {  "Id": 218342,  "Text": "FH Hafnarfjörður(Frauen)"  }  ]  }  }  ]  ]  } ``` |
| **SubscribeTranslations**  Subscribing for translations.  Note: This method should be called once when client starts. | ```  {  "Command": "SubscribeTranslations",  "Params": [  {  "LangIds": [  "en",  "de",  "zh"  ],  "GetChangesFrom": "2017-06-13T05:00:15.4303643Z"  }  ]  } ``` | ```  {  "Command": "SubscribeTranslations",  "Objects": [],  "SocketTime": "2017-06-13T13:59:04.449 9925Z"  } ``` |
| **UnSubscribeTranslations**  Unsubscribing from translations | ```  {  "Command": "UnSubscribeTranslations",  "Params": []  } ```     Unsubscribe from translations. | ```  {  "Command": "UnSubscribeTranslations",  "Error": null,  "Type": null,  "Objects": []  } ``` |
