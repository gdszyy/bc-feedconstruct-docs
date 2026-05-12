# 前端最小合格验收清单

把上传指引《§4 最小合格验收清单》16 项映射为**前端可验收项**。验收时不只看「赔率是否能显示」，必须覆盖状态、恢复、回滚、订阅、可追溯、降级。

| # | 验收项 | 前端合格标准 | 主要模块 |
|---|---|---|---|
| 1 | 实时通道 | 能稳定建立 WebSocket，心跳、断线指数退避、状态可见 | M01, M15 |
| 2 | 事件可追溯 | 每条事件保留 `event_id` / `correlation_id`，遥测可按 ID 反查 | M02, M16 |
| 3 | 事件覆盖 | UI 行为覆盖 odds / market.status / bet_stop / settlement / cancel / rollback / fixture / system | M02, M05~M09 |
| 4 | 主数据 | sport / tournament / match 主数据可缓存并按 i18n 渲染；fixture_change 触发刷新 | M03, M04, M12 |
| 5 | 赔率渲染 | market + outcome + 比分阶段在快照与增量下一致；旧版本消息丢弃 | M05 |
| 6 | 停盘行为 | Suspended / bet_stop 时禁止下注按钮可点击，UI 明确提示 | M06, M07, M13 |
| 7 | 结算可视化 | outcome 级 result / certainty / void_factor / dead_heat_factor 可展示 | M08, M14 |
| 8 | 取消可视化 | void_reason / 区间 / superceded_by 在我的投注页可见 | M08, M14 |
| 9 | 回滚 | 已结算 → 已回滚 → 重新结算的链路在 UI 可追溯，不覆盖原记录 | M09, M14 |
| 10 | 恢复 | 断线重连后 stale 标记可见，replay_completed 后才解除 | M10, M15 |
| 11 | 幂等 | 重复事件 / 重复下注请求不产生重复 UI 项或重复单 | M02, M13 |
| 12 | 防回退 | 高阶赛事状态不会被低阶覆盖；过期赔率版本不会回滚 UI | M04, M05 |
| 13 | 订阅 | 关注 / 订阅状态机持久化，比赛结束订阅自动释放 | M11 |
| 14 | 描述数据 | market / outcome 名称 / group / tab 可读，禁止直接显示 ID | M12 |
| 15 | 健康提示 | producer down / 卡赛 / 高延迟时显示降级横幅，并约束可交互操作 | M15 |
| 16 | 数据治理 | 长期不活跃数据释放内存；本地缓存按 version 失效 | M03, M12, M14 |

## 验收方法建议

- 单元测试覆盖每个 FSM 的转移矩阵（包括非法转移被拒绝）
- 集成测试模拟 WebSocket 断线 / replay / 乱序事件
- E2E（Playwright）覆盖 P03 详情、P05 投注单、P06 我的投注的关键路径
- 故障注入：producer down、replay 超时、价格变更确认
- 可追溯抽样：在生产/灰度抽取若干 `correlation_id`，验证 UI 行为可解释
