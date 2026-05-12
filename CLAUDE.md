# CLAUDE.md

本文件为 Claude / Claude Code 在本仓库中的工作入口。进入仓库后，必须先读取并遵循 [`AGENTS.md`](./AGENTS.md)，再根据任务目标选择对应索引与文档目录。`AGENTS.md` 是当前仓库面向 AI 代理的总索引与路由文件，包含入口策略、模块路由表与维护规则。

## 技术栈

本工程的实现栈如下，所有规划、契约与测试规范都围绕此栈展开：

| 层 | 技术 |
|---|---|
| 后端 | Go（建议 stdlib + chi/echo；AMQP 使用 amqp091-go；WebSocket 使用 gorilla/websocket 或 nhooyr/websocket） |
| 前端 | React + Next.js（App Router，TypeScript） |
| 前端 ↔ 后端通道 | WebSocket（实时事件流）+ REST（快照 / 描述 / 投注） |
| 边界约束 | 前端**不直连**博彩数据源 AMQP / 供应商 REST；这些只在 Go BFF 中使用 |

> 历史曾使用 Godot 进行可视化原型；当前主线**不再以 Godot 为目标栈**。若未来需要 Godot 模块，须独立标注并单独适配测试规范。

## AGENTS.md 索引

| 索引项 | 位置 | 使用场景 |
|---|---|---|
| 仓库总入口 | [`AGENTS.md`](./AGENTS.md) | 任何任务开始前必须先读，用于理解仓库定位、入口策略、模块路由和维护规则。 |
| 文档站模块导航 | [`indexes/NAVIGATION.md`](./indexes/NAVIGATION.md) | 涉及接口接入、消息结构、对象字段、SDK 或具体文档目录定位时优先读取。 |
| 业务领域索引 | [`indexes/BUSINESS_DOMAIN_INDEX.md`](./indexes/BUSINESS_DOMAIN_INDEX.md) | 涉及业务理解、系统集成分层、数据商能力梳理或跨模块概念分析时优先读取。 |
| 关键词索引 | [`indexes/KEYWORD_INDEX.md`](./indexes/KEYWORD_INDEX.md) | 任务只提供 BetGuard、RabbitMQ、Market、Settlement、SDK 等关键词时优先读取。 |
| 搜索索引 | [`indexes/SEARCH_INDEX.json`](./indexes/SEARCH_INDEX.json) | 需要程序化检索、批量搜索或关键词到文档路径映射时使用。 |
| 前端架构规划 | [`docs/07_frontend_architecture/`](./docs/07_frontend_architecture/) | 涉及前端模块、页面、Go BFF 与 Next.js 契约、状态机或验收清单时优先读取。 |

## 开发与测试工作流

本仓库的任何新增或修改逻辑都必须采用 **test-driven-development**。在定义测试之前，禁止编写任何业务逻辑、功能逻辑或修复逻辑代码。若任务需要改动代码，必须先完成测试意图建模、测试文件规划和用户确认，再进入正式测试与实现阶段。

## 强制 BDD 流程（适用于 Go 后端与 React/Next.js 前端）

| 阶段 | 必须完成的动作 | 禁止事项 |
|---|---|---|
| 行为建模 | 使用 Behaviour-Driven Development 实践，以 Gherkin schema 编写行为注释（Given-When-Then）。 | 禁止编写断言、测试逻辑、辅助逻辑或任何生产逻辑。 |
| 空测试文件 | 创建一个只包含 Given-When-Then 行为注释与测试函数名的空测试文件。 | 禁止在函数体中写入正式测试代码，函数体只能保持空实现或测试框架允许的占位语句。 |
| 用户确认 | 使用 `AskUserQuestion` 向用户确认是否继续进入正式测试代码编写阶段。 | 未获得用户确认前，禁止补充断言、测试夹具、mock、stub 或生产实现。 |
| 正式测试 | 用户确认后，才允许补充正式测试逻辑。 | 禁止跳过测试直接编写或修改逻辑代码。 |
| 实现逻辑 | 正式测试定义完成后，才允许编写最小必要生产逻辑以使测试通过。 | 禁止一次性实现超出测试覆盖范围的功能。 |

## 测试框架与目录约定

### Go 后端

| 项目 | 约定 |
|---|---|
| 单元测试框架 | Go 内建 `testing` + `testify/assert`（或 `testify/require`） |
| 行为测试框架（可选） | `cucumber/godog` 用于跨服务行为测试，与 BDD 注释保持一致 |
| 集成测试 | 使用 `testcontainers-go` 启 RabbitMQ / Postgres，或 docker compose 起 stack |
| 目录布局 | 生产代码与测试同包：`*_test.go`；跨包测试放 `internal/<module>/<module>_test.go`；E2E 放 `test/e2e/` |
| 命名 | 测试函数 `TestXxx`；子用例 `t.Run("given_<...>_when_<...>_then_<...>", ...)` 与 BDD 行为对齐 |

### React / Next.js 前端

| 项目 | 约定 |
|---|---|
| 单元测试框架 | Vitest 或 Jest + Testing Library |
| E2E 框架 | Playwright |
| 行为测试框架（可选） | `@cucumber/cucumber` 用于跨页面行为，与 BDD 注释保持一致 |
| 目录布局 | 单元/集成测试与生产代码邻近：`<module>/<file>.test.ts(x)`；E2E 集中放 `test/e2e/` |
| 命名 | `describe('given ...')` / `it('when ... then ...')` 与 BDD 行为对齐 |

## BDD 空测试文件要求

在编写正式测试代码之前，必须先创建空测试文件。该文件只能表达行为意图，不能包含真实测试逻辑。行为注释必须遵循 **Given-When-Then** 模式；测试函数名/`describe`/`it` 标题必须能准确概括行为场景。

### Go 示例

```go
package mymodule_test

// Given <前置条件>
// When <触发行为>
// Then <可观察结果>
func TestWhenXxx_ThenYyy(t *testing.T) {
    // BDD placeholder — 正式测试由用户确认后补充
    _ = t
}
```

### React/Next.js 示例

```ts
// Given <前置条件>
// When <触发行为>
// Then <可观察结果>
describe('given ...', () => {
  it('when ... then ...', () => {
    // BDD placeholder — 正式测试由用户确认后补充
  })
})
```

上述文件创建完成后，必须暂停后续实现，并通过 `AskUserQuestion` 询问用户是否继续。只有在用户明确确认后，才允许继续编写正式测试代码；只有正式测试代码定义完成后，才允许编写或修改生产逻辑代码。

## 前端模块落地顺序约束

凡是涉及 `docs/07_frontend_architecture/` 中模块的代码落地，必须遵守该目录 `README.md` 中给出的「落地顺序」，并以模块文件中的「验收要点」作为测试设计的最低基线。新增模块必须先在 `docs/07_frontend_architecture/` 内补齐骨架文档，再进入 BDD 流程。

## 强制执行原则

所有代理在本仓库中执行开发任务时，必须遵守以下顺序：先阅读 [`AGENTS.md`](./AGENTS.md)，再定位相关索引或文档；先写 BDD 行为注释和空测试函数，再向用户确认；先补充正式测试，再编写生产逻辑代码。任何与该顺序冲突的操作都应视为流程违规并立即停止。
