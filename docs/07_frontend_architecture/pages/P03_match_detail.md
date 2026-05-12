# P03 — 赛事详情

## 目的

单场赛事的完整信息：比分、阶段、市场分组（tab）、所有可下注盘口、停投提示、相关统计。是消费实时事件最密集的页面。

## 数据来源

| 区域 | 来源 |
|---|---|
| 顶部赛事头（队伍、比分、阶段、状态） | M04 |
| 市场 tabs / groups | M12 |
| 市场列表 + 赔率按钮 | M05、M06、M07 |
| 停投/停盘横幅 | M06、M07 |
| 投注单浮层 | M13 |
| 全页降级提示 | M15 |

## 渲染策略

- RSC：赛事头 + market 列表骨架（按 matchId 拉快照）
- Client：进入后订阅 `match_id={id}`，开始接收 odds / status / bet_stop
- 离开时取消订阅

## 关键组件

- `<MatchHeader/>`（M04）
- `<MarketTabs/>` / `<MarketGroupList/>`（M12）
- `<MarketRow/>`（M05 + M06 + M07 → bettable）
- `<OutcomeButton/>`（点击进入 M13）
- `<BetStopBanner/>`（M07）

## 验收要点

- market 状态变化即时反映到按钮可点击性
- 描述未到位时不显示原始 ID
- 重连期间显示 stale，replay_completed 后自动解除
- 离开页面后 WS 订阅被释放
