# 前端关键状态机

所有 reducer / selector 必须显式 switch over 以下状态枚举，禁止用 `!= active` 类的二分判断。

## 1. Connection FSM（M01 / M15）

```
                ┌──────────────┐
                │ Disconnected │◀──────────────┐
                └──────┬───────┘               │
                       │ connect()             │ close()
                       ▼                       │
                 ┌──────────┐                  │
                 │Connecting│──fail──▶┌────────┴────────┐
                 └────┬─────┘         │ Reconnecting    │
                      │open           │ (backoff)       │
                      ▼               └────────┬────────┘
                 ┌────────┐      heartbeat-loss│
                 │  Open  │────────────────────┘
                 └────┬───┘
                      │ producer.down / latency↑
                      ▼
                 ┌──────────┐
                 │ Degraded │ ──producer.up──▶ Open
                 └──────────┘
```

- `Degraded` 时 UI 显示降级横幅（M15），但仍允许只读消费
- `Reconnecting` 必须在重连成功后触发 `replay_from`（M10）

## 2. Match FSM（M04）

| 状态 | 进入事件 | 不可降级到 |
|---|---|---|
| NotStarted | `match.upserted`（首次） | — |
| Live | `match.status_changed → live` | NotStarted |
| Ended | `match.status_changed → ended` | Live, NotStarted |
| Closed | `match.status_changed → closed` | Ended, Live, NotStarted |
| Cancelled | `match.status_changed → cancelled` | — |
| Abandoned | `match.status_changed → abandoned` | — |

reducer 必须实现「高阶状态不被低阶覆盖」规则（P6）。

## 3. Market FSM（M05 / M06）

```
       ┌──────────┐    suspend    ┌───────────┐
       │  Active  │──────────────▶│ Suspended │
       └────┬─────┘               └─────┬─────┘
            │                            │ resume
            │                            ▼
            │   bet_stop          ┌──────────┐
            ├────────────────────▶│ Suspended│
            │                     └──────────┘
            │   deactivate
            ▼
       ┌─────────────┐  settle   ┌──────────┐
       │ Deactivated │──────────▶│ Settled  │
       └─────────────┘           └────┬─────┘
                                      │ rollback_settle
                                      ▼
                                 ┌─────────────┐
                                 │ Deactivated │
                                 └─────────────┘

       Any ──cancel──▶ Cancelled ──rollback_cancel──▶ <prev>
       Any ──handover──▶ HandedOver
```

派生属性：

| 派生 | 公式 |
|---|---|
| `bettable` | `status === Active && !inBetStop && connection !== Reconnecting` |
| `displayOdds` | `status ∈ {Active, Suspended} ? lastOdds : null` |
| `frozen` | `status === Suspended \|\| inBetStop` |

## 4. Bet FSM（M13 / M14）

```
   Draft (在投注单内) ──validate──▶ Validated
        │                                │ place()
        ▼                                ▼
     Discarded                       Submitting
                                         │ accepted/rejected
                              ┌──────────┴──────────┐
                              ▼                      ▼
                           Accepted               Rejected
                              │
                  bet_settlement.applied
                              ▼
                          Settled ──rollback──▶ Accepted
                              │
                  bet_cancel.applied
                              ▼
                         Cancelled ──rollback──▶ <prev>
```

变更链必须以 append-only 形式保存（M14 显示历史）。

## 5. Subscription FSM（M11）

```
 Idle ──book──▶ Booking ──ok──▶ Subscribed ──unbook──▶ Unbooking ──ok──▶ Released
                  │                                       │
                  └───fail──▶ Failed                       └───fail──▶ Failed
```

- 关注（Favorite）是本地纯偏好，不入此 FSM
- 「比赛结束自动 Release」由后端推 `subscription.changed`，前端只反映状态

## 6. BetSlip FSM（M13）

```
 Empty ──addSelection──▶ Editing ──validate──▶ Ready ──place──▶ Submitting
                            ▲                    │                   │
                            │                    │ priceChanged?     │
                            └───acceptChange─────┤                   │
                                                  ▼                   ▼
                                              NeedsReview         Submitted (转 Bet FSM)
```

价格变更必须显式让用户确认（NeedsReview），不允许悄悄按新价格提交。
