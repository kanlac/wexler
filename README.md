# Mindful (Formerly Wexler) - Make Your Coding Agents Mindful of Your Taste

## 问题定位

随着 agentic AI 编程工具（如 Claude Code, Cursor, Copilot）的普及，选择日益增多，开发者面临着配置碎片化和迁移成本高的问题。每个工具都有自己的配置格式和存储位置，导致：

- 团队成员间难以保持编码规范和项目知识的一致性
- 切换到其他 AI 编程工具，需要重复配置相同的内容
- API 密钥等敏感信息散落在各处，管理困难且存在安全隐患

## 项目概述

Mindful 通过统一的配置源和自动化同步机制，将您精心维护的 AI 长期记忆、Subagents (Roles) 以及 MCP 配置，与特定 AI 工具解耦，实现"一次配置，多处使用"。


## 核心价值

- **配置统一化**：解决配置碎片化问题，建立单一配置源，支持以下配置类型：
  1. AI 长期记忆
  2. Subagents (Roles)
  3. MCP
- **迁移零成本**：在不同 AI 工具间自由切换，配置自动同步
- **团队知识共享**：统一管理团队的编码规范和项目知识
- **密钥集中管理**：API 密钥等敏感信息安全存储，一处配置多处使用

## 使用场景

### 团队协作模式

1. **维护 Mindful 源仓库**：创建一个 Git 仓库作为 Mindful source，统一管理团队共享的知识和 AI 配置
2. **嵌入工作流**：将 `mindful apply` 命令加入项目 Makefile，确保团队成员自动获取最新配置

**Makefile 示例**：

```makefile
# 开发环境初始化
dev-init:
    go mod download
    mindful init --source=../team-mindful-configs
    mindful apply

# 更新 AI 配置
update-ai:
    cd ../team-mindful-configs && git pull
    mindful apply

```

### 个人使用模式

维护个人的 Mindful 配置目录，在多个项目间复用配置，轻松切换不同的 AI 工具。

## 命令设计

### 1. 初始化（init）

```bash
mindful init
mindful init --source=/path/to/mindful-configs

```

在项目目录初始化，创建 `mindful.yaml` 配置文件。

### 2. 导入（import）

```bash
mindful import
mindful import --tool=claude

```

扫描当前 AI 工具配置，提取 MCP 配置并存储到 BoltDB。

### 3. 应用（apply）

```bash
mindful apply
mindful apply --tool=cursor

```

将 Mindful 管理的配置应用到各 AI 工具：

- 生成/更新配置文件
- 注入 MCP 配置和 API 密钥

### 4. 列表（list）

```bash
mindful list
mindful list --mcp

```

## 如何安装

```bash
make install
```

## 目录结构

### Mindful 源目录

```
$MINDFUL_DIR/                    # 团队共享的配置源
├── memory.mdc                  # 通用记忆配置
├── subagent/                   # Subagent/Role 配置
│   ├── code-reviewer.mdc
│   ├── test-writer.mdc
│   └── architect.mdc
└── mindful.db                   # BoltDB（API密钥等敏感信息）

```

### 用户项目目录（apply 后）

```
project/
├── mindful.yaml                 # Mindful 项目配置
├── CLAUDE.md                   # Claude Code 配置
├── .claude/
│   └── agents/
│       └── *.mindful.md
├── .cursor/
│   └── rules/
│       └── *.mindful.mdc
└── .mcp.json                   # MCP 配置

```

### 项目结构
本仓库目录，省略。

## 配置文件映射

### 1. 通用记忆（General Memory）

| Mindful 源 | Claude Code | Cursor |
| --- | --- | --- |
| `memory.mdc` | `CLAUDE.md`（二级标题标识） | `.cursor/rules/general.mindful.mdc` |

### 2. Subagent/Role 配置

| Mindful 源 | Claude Code | Cursor |
| --- | --- | --- |
| `subagent/code-reviewer.mdc` | `.claude/agents/code-reviewer.mindful.md` | `.cursor/rules/code-reviewer.mindful.mdc` |

### 3. MCP 配置

| Mindful 源 | Claude Code | Cursor |
| --- | --- | --- |
| `mindful.db` | `.mcp.json`（upsert） | `.cursor/mcp.json` |

## 项目配置文件（mindful.yaml）

```yaml
version: 1.0
source: /Users/alice/team-mindful-configs    # 本地文件系统路径
tools:
  - claude
  - cursor

```

## MCP 配置管理

### 存储方案

考虑到黑客松时间限制，采用简化的安全方案：

- 使用 BoltDB 存储完整 MCP 配置
- API Key 等敏感信息使用 base64 编码
- 文件权限设为 0600

### Apply 时的 Upsert 逻辑

1. 读取现有 `.mcp.json`
2. 从 BoltDB 解码 Mindful 管理的配置
3. 合并配置（Mindful 优先）
4. 写回 `.mcp.json`

## 未来扩展计划

1. **Phase 2**（配置增强）
    - 远程 Git 源支持
    - 更方便智能地导入现有 AI 配置
2. **Phase 3**（工具扩展）
    - 支持更多 AI 编程工具（GitHub Copilot, Windsurf）
    - Web UI 管理界面
3. **Phase 4**（企业特性）
    - 针对 MCP API key 等敏感数据做加密存储（如 AES-256）

# A Proposal：我们应如何管理 AI 长期记忆？

## 核心理念/最佳实践

- 所有长期记忆/上下文都不应与具体工具耦合，记忆应该只按 scope 分类管理，实现一处定义，随处使用
- 具体 AI 配置文件都应加入 .gitignore，因为它们都将是 mindful 生成文件，人们只需要维护和迭代记忆本身

## Scope 分类体系

### 1. scope: team
跨项目共享的整体规范和 subagent，团队成员维护一个 git 仓库作为 mindful source。

- 长期记忆
  - 源：`mindful_source/memory.mdc`
  - 目标：
    - Claude Code: `project_dir/CLAUDE.md#'Mindful Memory (scope: team)'`
- subagent
  - 源：`mindful_source/subagent/*.mdc`
  - 目标：
    - Claude Code: `project_dir/.claude/agents/`
- MCP

### 2. scope: project
属于当前仓库的业务知识编码规范等，。

- 长期记忆
  - 源：开发者维护 `project_dir/memory.mdc`
  - 目标：
    - Claude Code: `project_dir/CLAUDE.md#'Mindful Memory (scope: project)'`
- subagent
- MCP

### 3. scope: user

- 长期记忆
  - 源：用户个人维护的 `~/.mindful/memory.mdc`
  - 目标：
    - Claude Code: `project_dir/CLAUDE.md#'Mindful Memory (scope: user)'`
- subagent
- MCP