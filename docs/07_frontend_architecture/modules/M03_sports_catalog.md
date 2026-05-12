# M03 — 体育目录（Sport / Category / Tournament）

## 目的

维护赛事所属维度（运动、地区/分类、联赛/赛事）的目录树，供导航与筛选使用。

## 数据来源

- REST 快照：
  - `GET /api/v1/catalog/sports`
  - `GET /api/v1/catalog/tournaments?sport_id=...`
- 实时增量：`sport.upserted` / `sport.removed` / `tournament.upserted` / `tournament.removed`

底层字段语义参考 OddsFeed `Sport`、`Region`、`Competition`（见 `docs/01_data_feed/rmq-web-api/005_sport.md` 等）。

## 领域模型

```ts
Sport { id, name, order, liveCapable, defaultMarketTabs[] }
Category { id, sportId, name }
Tournament { id, sportId, categoryId, name, season, liveCapable }
```

## 关键组件

| 组件 | 职责 |
|---|---|
| `catalogStore` | 体育/分类/联赛树状缓存 |
| `catalogService` | REST 拉取 + ETag 缓存 |
| `catalogReducer` | 应用 upsert / remove 增量 |

## 与其他模块依赖

- 输入：M02
- 输出：M04（赛事过滤）、M12（i18n 取名）、P01/P02 导航

## 未决问题

- [ ] 多语言名称是否随 catalog 返回还是走 M12？
- [ ] 「热门联赛」「自定义排序」由 BFF 提供还是前端拼？

## 验收要点

- 树状结构在 i18n 切换时不重新拉取（M12 单独管文案）
- 增量与快照合并幂等
