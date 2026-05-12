# P07 — 搜索与收藏

## 目的

提供赛事/联赛/队伍搜索入口，并管理用户的本地收藏与服务端订阅。

## 数据来源

| 区域 | 来源 |
|---|---|
| 搜索结果 | BFF 搜索端点（待后端确认） + M03/M04 |
| 收藏列表 | M11 favoritesStore（本地） |
| 订阅状态 | M11 subscriptionStore |

## 关键组件

- `<SearchInput/>` + debounce
- `<SearchResults/>`
- `<FavoritesList/>` 展示本地收藏 + 订阅状态徽标
- `<SubscriptionActions/>` book / unbook

## 验收要点

- 收藏的赛事即使未订阅，也能显示开赛时间与状态
- 订阅失败可见且可重试（M11）
- 比赛结束后收藏可见但订阅自动 Released
