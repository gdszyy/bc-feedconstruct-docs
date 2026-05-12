# M15 — 健康与降级提示

## 目的

把后端 producer 状态、连接状态、stale 标记转化为用户可理解的降级语义（横幅、徽标、禁用），避免在数据异常时让用户继续下注或误以为数据是最新的。

## 数据来源

- 实时：`system.producer_status`、`system.replay_*`
- REST：`GET /api/v1/system/health`（启动时拉取）
- 内部信号：M01 ConnectionFSM、M10 staleTracker

## 领域模型

```ts
HealthState {
  connection: ConnectionFsmState,
  producers: { [id]: 'up'|'down'|'degraded' },
  staleScope: Set<matchId> | 'global',
  banner?: { level: 'info'|'warn'|'error', message, since }
}
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `healthStore` | 汇聚信号 |
| `degradationPolicy` | 把状态映射成 UI 行为（横幅、禁用、提示） |

## 与其他模块依赖

- 输入：M01、M10、M02（producer 事件）
- 输出：所有页面顶部横幅、M13 下注按钮禁用

## 未决问题

- [ ] 降级文案与多语言对齐方案？
- [ ] producer down 时是否阻止下注？由谁决定？

## 验收要点

- producer down 时所有受影响盘口的下注禁用并显示原因
- 重连成功 + replay_completed 后横幅自动消失
- 不允许在 Degraded / Stale 状态下进入 BetSlip.Submitting
