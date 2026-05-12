# M11 — 订阅生命周期（关注 / 订阅 / 释放）

## 目的

把「关注（本地偏好）」与「订阅（服务端 booking）」分离，让前端有明确的订阅状态机，避免长期占用配额和成本。

## 数据来源

- REST：`POST /api/v1/subscriptions`、`DELETE /api/v1/subscriptions/{id}`、`GET /api/v1/subscriptions`
- 实时：`subscription.changed`
- 参考：OddsFeed `Book` / `BookObject` / `PartnerBooking` / `MarketTypeBooking`、上传指引 §3.5

## 领域模型

```ts
Favorite { matchId, addedAt }            // 本地，纯偏好
Subscription { id, scope: 'match'|'tournament', refId, status: SubFSM, lastTransitionAt }
```

状态机：见 [§5](../04_state_machines.md#5-subscription-fsmm11)。

## 关键组件

| 组件 | 职责 |
|---|---|
| `favoritesStore` | localStorage 持久化的本地偏好 |
| `subscriptionStore` | 服务端订阅状态 |
| `subscriptionService` | REST 调用封装，重试 + 幂等 |
| `lifecycleReducer` | 应用 `subscription.changed` |

## 与其他模块依赖

- 输入：M02、用户操作
- 输出：M03/M04 列表中的「已订阅」标记、P07 收藏页

## 未决问题

- [ ] 未登录用户的 Favorite 是否允许？
- [ ] 比赛结束的自动 unbook 由 BFF 还是前端发起？建议 BFF
- [ ] 订阅失败是否需要用户可见错误？

## 验收要点

- 收藏与订阅状态独立可见，不互相覆盖
- 比赛结束后 5 分钟内 UI 显示 Released
- 订阅失败有可重试入口并被遥测
