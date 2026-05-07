---
title: Example
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_example
current_loc: betGuard
location: partner_api_example
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Example

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=partner_api_example`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `partner_api_example` |

## 文档正文
Example

```
private readonly Dictionary<string, List<string>> hashPropertyMap = new Dictionary<string, List<string>>

{

{ "GetClientDetails" , new List<string>{ "AuthToken", "TS" } },

{ "BetPlaced" , new List<string>{ "AuthToken", "TS", "TransactionId", "BetId", "Amount", "Created", "BetType", "SystemMinCount", "TotalPrice" } },

{ "BetResulted" , new List<string>{ "AuthToken", "TS", "TransactionId", "BetId", "BetState", "Amount" } },

{ "Rollback" , new List<string>{ "AuthToken", "TS", "TransactionId" } }

};

protected string CalculateRequestHash(string key, string request)

{

var properties = this.GetType().GetProperties();

Dictionary<string, object> objectmap = new Dictionary<string, object>();

if (hashPropertyMap.TryGetValue(request, out var propertyNames))

{

foreach (var property in properties)

{

var value = property.GetValue(this, null);

if (value != null && propertyNames.Contains(property.Name) && property.CanWrite)

{

objectmap.Add(property.Name, value);

}

}

}

var jsonString = JsonConvert.SerializeObject(objectmap);

return ComputeHMACSHAT256(jsonString, key);

}

public static string ComputeHMACSHAT256(string data, string key)

{

var secretKey = Encoding.ASCII.GetBytes(key);

var message = Encoding.ASCII.GetBytes(data);

using (HMACSHA256 hash = new HMACSHA256(secretKey))

{

var hashArray = hash.ComputeHash(message);

return string.Concat(Array.ConvertAll(hashArray, b => b.ToString("x2")));

}

}
```
