# M16 — 前端遥测与审计

## 目的

为前端建立**端到端可追溯**的观测能力：连接事件、消息处理、用户操作、错误、降级都带 `correlation_id`，便于联合后端排障。

## 数据来源（内部）

- M01：连接生命周期事件
- M02：未知 type / 过期版本 / 重复事件计数
- M13：下注请求、结果、价格变更确认
- 通用：异常、未捕获 Promise、慢路径

## 关键能力

| 能力 | 说明 |
|---|---|
| Logger | 结构化日志，强制带 `correlation_id` / `session_id` / `user_id?` |
| Metrics | 计数器、直方图（连接断线次数、replay 时长、下单 RT 等） |
| ErrorReporter | Sentry 等 SaaS 接入，PII 过滤 |
| InteractionAudit | 用户关键操作（下单、确认价格变更、取消订阅）抽样记录 |

## 关键组件

| 组件 | 职责 |
|---|---|
| `telemetryStore` | 队列与批量上报 |
| `correlationContext` | React context 暴露当前 correlation_id |
| `redactor` | PII 过滤（金额、用户名等） |

## 与其他模块依赖

- 输入：所有模块
- 输出：监控平台（不在前端范围）

## 未决问题

- [ ] 采样率与隐私合规要求？
- [ ] 是否记录 envelope 原文（按抽样）？

## 验收要点

- 关键路径事件 100% 携带 `correlation_id`
- 错误日志不包含敏感信息
- 用户可见错误能反查到对应后端事件
