# P01 — 首页 / 大厅

## 目的

承载入口流量：今日推荐、热门联赛、滚球入口、即将开始、降级横幅。强调首屏快、二屏实时。

## 数据来源

| 区域 | 来源 |
|---|---|
| 顶部降级横幅 | M15 |
| 体育快捷入口 | M03 |
| 热门赛事卡片 | M04 + M05（按 sport 取若干） |
| 滚球大厅入口 | M04 筛选 `liveOdds=true` |
| 即将开始 | M04 按 scheduledAt |

## 渲染策略

- RSC 首屏渲染（缓存目录与赛事快照）
- Client Component 进入后订阅当前可见卡片范围（M11 范围订阅）

## 关键组件

- `<LobbyBanner/>`（M15）
- `<SportShortcuts/>`（M03）
- `<HotMatchesRail/>`（M04+M05，懒订阅）
- `<LiveEntry/>`

## 验收要点

- 首屏 LCP 不依赖 WS 建立
- 滚动到卡片可见时再订阅（节省 WS 流量）
- Degraded / Stale 时卡片显示降级视觉
