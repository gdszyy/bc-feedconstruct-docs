# CLAUDE.md

本文件为 Claude / Claude Code 在本仓库中的工作入口。进入仓库后，必须先读取并遵循 [`AGENTS.md`](./AGENTS.md)，再根据任务目标选择对应索引与文档目录。`AGENTS.md` 是当前仓库面向 AI 代理的总索引与路由文件，包含入口策略、模块路由表与维护规则。

## AGENTS.md 索引

| 索引项 | 位置 | 使用场景 |
|---|---|---|
| 仓库总入口 | [`AGENTS.md`](./AGENTS.md) | 任何任务开始前必须先读，用于理解仓库定位、入口策略、模块路由和维护规则。 |
| 文档站模块导航 | [`indexes/NAVIGATION.md`](./indexes/NAVIGATION.md) | 涉及接口接入、消息结构、对象字段、SDK 或具体文档目录定位时优先读取。 |
| 业务领域索引 | [`indexes/BUSINESS_DOMAIN_INDEX.md`](./indexes/BUSINESS_DOMAIN_INDEX.md) | 涉及业务理解、系统集成分层、数据商能力梳理或跨模块概念分析时优先读取。 |
| 关键词索引 | [`indexes/KEYWORD_INDEX.md`](./indexes/KEYWORD_INDEX.md) | 任务只提供 BetGuard、RabbitMQ、Market、Settlement、SDK 等关键词时优先读取。 |
| 搜索索引 | [`indexes/SEARCH_INDEX.json`](./indexes/SEARCH_INDEX.json) | 需要程序化检索、批量搜索或关键词到文档路径映射时使用。 |

## 开发与测试工作流

本仓库的任何新增或修改逻辑都必须采用 **test-driven-development**。在定义测试之前，禁止编写任何业务逻辑、功能逻辑或修复逻辑代码。若任务需要改动代码，必须先完成测试意图建模、测试文件规划和用户确认，再进入正式测试与实现阶段。

## Godot GUT 测试规范

如果任务涉及 Godot 项目代码，必须使用 **Godot GUT** 作为测试框架。所有单元测试代码必须放在 `test/unit/` 目录下，并按照被测模块或功能边界命名，确保测试文件与生产代码之间可以清晰追踪。

| 阶段 | 必须完成的动作 | 禁止事项 |
|---|---|---|
| 行为建模 | 使用 Behaviour-Driven Development 实践，以 Gherkin schema 编写行为注释。 | 禁止编写断言、测试逻辑、辅助逻辑或任何生产逻辑。 |
| 空测试文件 | 创建一个只包含 Given-When-Then 行为注释与测试函数名的空测试文件。 | 禁止在函数体中写入正式测试代码，函数体只能保持空实现或测试框架允许的占位语句。 |
| 用户确认 | 使用 `AskUserQuestion` 向用户确认是否继续进入正式测试代码编写阶段。 | 未获得用户确认前，禁止补充断言、测试夹具、mock、stub 或生产实现。 |
| 正式测试 | 用户确认后，才允许在 `test/unit/` 下补充 Godot GUT 测试逻辑。 | 禁止跳过测试直接编写或修改逻辑代码。 |
| 实现逻辑 | 正式测试定义完成后，才允许编写最小必要生产逻辑以使测试通过。 | 禁止一次性实现超出测试覆盖范围的功能。 |

## BDD 空测试文件要求

在编写正式测试代码之前，必须先创建空测试文件。该文件只能表达行为意图，不能包含真实测试逻辑。行为注释必须遵循 **Given-When-Then** 模式；测试函数名必须能准确概括行为场景。示例结构如下：

```gdscript
extends GutTest

# Given <前置条件>
# When <触发行为>
# Then <可观察结果>
func test_<behavior_name>():
    pass
```

上述文件创建完成后，必须暂停后续实现，并通过 `AskUserQuestion` 询问用户是否继续。只有在用户明确确认后，才允许继续编写正式测试代码；只有正式测试代码定义完成后，才允许编写或修改生产逻辑代码。

## 强制执行原则

所有代理在本仓库中执行开发任务时，必须遵守以下顺序：先阅读 [`AGENTS.md`](./AGENTS.md)，再定位相关索引或文档；先写 BDD 行为注释和空测试函数，再向用户确认；先补充正式 Godot GUT 测试，再编写生产逻辑。任何与该顺序冲突的操作都应视为流程违规并立即停止。
