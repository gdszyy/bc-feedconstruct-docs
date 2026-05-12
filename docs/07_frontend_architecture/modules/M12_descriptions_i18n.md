# M12 — 静态描述与 i18n（Market / Outcome / Group / Tab）

## 目的

把市场描述、玩法名称、outcome 名称、分组、tab 与实时事件解耦，按 version 缓存，供 UI 渲染。**禁止把供应商 ID 直接展示给用户**。

## 数据来源

- REST：
  - `GET /api/v1/descriptions/markets?version=...`
  - `GET /api/v1/descriptions/outcomes?version=...`
- 翻译来源：仓库 `docs/02_translations/`
- 字段语义：`docs/03_sports_model_reference/market-types/001_market-types.md`

## 领域模型

```ts
MarketDescription {
  marketTypeId, name, groups: string[], tab?: string,
  outcomeTemplates: OutcomeTemplate[],
  specifierSchema?: {...}
}
OutcomeTemplate { id, nameTemplate /* 可含 {player}, {handicap} */ }
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `descriptionsStore` | 按 marketTypeId / locale 索引 |
| `descriptionService` | ETag 缓存、增量加载 |
| `renderName` | 把 (template, specifiers, players) 渲染成本地化名称 |

## 与其他模块依赖

- 输入：M03（运动维度）、用户语言偏好
- 输出：M05 渲染、P03 / P04 标签

## 未决问题

- [ ] 翻译是否随描述包返回，还是独立 i18n 通道？
- [ ] 渲染 `{player}` 等占位的 fallback 策略？

## 验收要点

- 切换语言不重拉描述结构，只换文案
- 描述未到位时显示骨架，不显示 ID
- 描述 version 升级时按 ETag 增量更新
