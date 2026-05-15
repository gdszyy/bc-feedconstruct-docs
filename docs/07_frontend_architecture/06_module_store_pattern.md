# 模块 Store 落地约定（Wave 0 产物）

本文件锁定前端 16 个模块（M01–M16）与 8 个页面（P01–P08）在状态管理上的边界，确保多个并行 track 同时开发时不会争抢全局 store 或同一文件。

## 1. 原则

1. **每个模块自带状态域**：M0x 的 store 文件落在 `frontend/src/<module>/store.ts`，对外只暴露 `useXxxStore` hook 与一组纯函数 selector。
2. **禁止 mega-store**：不引入 `src/store/index.ts` 这种合并所有 slice 的全局 store。模块之间只通过 `frontend/src/contract/**` 的类型与显式 hook 组合。
3. **页面只组合，不存状态**：P0x 页面文件不得创建跨模块的 store；如需跨模块协调，使用 React context 包装在页面组件内或下放给负责的模块。
4. **store 实现自由**：每个 track 可选 Zustand / Redux Toolkit / React Context + useReducer，但要在该模块 README 中标注，并保证暴露的 hook API 形如 `useXxxStore(selector)`。

## 2. 文件骨架

每个模块目录建议如下结构：

```
frontend/src/<module>/
  index.ts             # public API barrel (hooks, selectors, types)
  store.ts             # 内部状态实现（不导出 raw store 引用）
  types.ts             # 模块私有类型；公共契约只从 ../contract 引入
  hooks.ts             # useXxx hooks，依赖 store + selector
  api.ts               # REST / WS 适配；只允许 import contract/
  __tests__/           # 单元 + 集成测试（与生产文件邻近）
```

`index.ts` 中可以 re-export 的内容：hooks、selector、公共类型。**禁止** re-export 内部 store 句柄或 mutator。

## 3. 跨模块通讯

| 场景 | 推荐做法 | 禁止 |
|---|---|---|
| Page 需要赛事 + 市场 + 描述 | 在 page 内分别 `useCatalogStore`、`useMarketsStore`、`useDescriptionsStore` | page 直接读 store 内部状态 |
| 模块需要监听事件 | 通过 M02 dispatcher 提供的 `subscribe(type, handler)` hook | 直接监听 WebSocket 实例 |
| 投注链需要赔率 | 12-E 投注 store 通过 11-C `useMarketSnapshot(matchId, marketId)` 选择器读取 | 12-E import `markets/store.ts` 内部 |

## 4. 命名空间

```
M01 realtime/       — useTransport, useConnectionState
M02 dispatcher/     — useEventSubscription
M03 catalog/        — useSportTree, useTournamentList
M04 match/          — useMatch, useMatchStatus
M05 markets/        — useMarketSnapshot, useOddsStream
M06 market-fsm/     — useMarketFSM
M07 bet-stop/       — useBetStop
M08 settlement/     — useSettlementChain
M09 rollback/       — useRollbackState
M10 recovery/       — useRecoveryStatus, useStaleFlag
M11 subscription/   — useSubscriptions
M12 i18n/           — useDescription, useLang
M13 bet-slip/       — useBetSlip, usePlaceBet
M14 my-bets/        — useMyBets, useBetTimeline
M15 health/         — useSystemHealth, useDegradedBanner
M16 telemetry/      — useTelemetry, withTrace
```

## 5. 验收要点

- 任一模块的 `__tests__/` 用例只 import 自己 + `../contract` + 测试工具，不 import 其他模块的具体路径。
- Page 测试可以 mock 多个模块 hook，不直接 mount 真实 store。
- `pnpm vitest run` 在 Wave 11/12 阶段每条 track 互不影响。

> Wave 0 之后，本文件追加内容时仍使用 PR 增量，禁止整体重写。
