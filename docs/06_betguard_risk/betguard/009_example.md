---
title: Example
source_url: https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=example
current_loc: betGuard
location: example
top_category: BETGUARD
product_line: BetGuard 投注风控服务
business_domain: 投注风控服务 / BetGuard
scraped_at: 2026-05-07T08:49:13.195Z
---

# Example

> 来源：FeedConstruct OddsFeed Documentation；抓取入口为 `https://oddsfeed.feedconstruct.com/documentation?currentLoc=betGuard&location=example`。

| 字段 | 值 |
|---|---|
| 一级分类 | BETGUARD |
| 产品线 | BetGuard 投注风控服务 |
| 业务域 | 投注风控服务 / BetGuard |
| currentLoc | `betGuard` |
| location | `example` |

## 文档正文
Example

```
protected string CalculateRequestHash(string key)

{

var properties = this.GetType().GetProperties();

Dictionary<string, object> objectmap = new Dictionary<string, object>();

foreach (var property in properties)

{

var value = property.GetValue(this, null);

if (value != null && property.Name != "RequestHash" && property.CanWrite)

{

objectmap.Add(property.Name, value);

}

}

var jsonString = JsonConvert.SerializeObject(objectmap);

return ComputeHMACSHAT256(jsonString, key);

}

* key     // this is the partners secretKey

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
