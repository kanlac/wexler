# Wexler - Share Your Handy Prompts and MCPs Across Projects

## 问题定位

随着 agentic AI 编程工具（如 Claude Code, Cursor, Copilot）的普及，选择日益增多，开发者面临着配置碎片化和迁移成本高的问题。每个工具都有自己的配置格式和存储位置，导致：

- 团队成员间难以保持编码规范和项目知识的一致性
- 切换到其他 AI 编程工具，需要重复配置相同的内容
- API 密钥等敏感信息散落在各处，管理困难且存在安全隐患

## 项目概述

Wexler 通过统一的配置源和自动化同步机制，将您精心维护的 AI 长期记忆、Subagents (Roles) 以及 MCP 配置，与特定 AI 工具解耦，实现"一次配置，多处使用"。

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

1. **维护 Wexler 源仓库**：创建一个 Git 仓库作为 Wexler source，统一管理团队共享的知识和 AI 配置
2. **嵌入工作流**：将 `wexler apply` 命令加入项目 Makefile，确保团队成员自动获取最新配置

**Makefile 示例**：

```makefile
# 开发环境初始化
dev-init:
    go mod download
    wexler init --source=../team-wexler-configs
    wexler apply

# 更新 AI 配置
update-ai:
    cd ../team-wexler-configs && git pull
    wexler apply

```

### 个人使用模式

维护个人的 Wexler 配置目录，在多个项目间复用配置，轻松切换不同的 AI 工具。

## 命令设计

### 1. 初始化（init）

```bash
wexler init
wexler init --source=/path/to/wexler-configs

```

在项目目录初始化 Wexler，创建 `wexler.yaml` 配置文件。

### 2. 导入（import）

```bash
wexler import
wexler import --tool=claude

```

扫描当前 AI 工具配置，提取 MCP 配置并存储到 BoltDB。

### 3. 应用（apply）

```bash
wexler apply
wexler apply --tool=cursor

```

将 Wexler 管理的配置应用到各 AI 工具：

- 生成/更新配置文件
- 注入 MCP 配置和 API 密钥

### 4. 列表（list）

```bash
wexler list
wexler list --mcp

```

## 如何安装

```bash
make install
```

## 存储架构

### Wexler 源目录

```
$WEXLER_DIR/                    # 团队共享的配置源
├── memory.mdc                  # 通用记忆配置
├── subagent/                   # Subagent/Role 配置
│   ├── code-reviewer.mdc
│   ├── test-writer.mdc
│   └── architect.mdc
└── wexler.db                   # BoltDB（API密钥等敏感信息）

```

### 项目目录（apply 后）

```
project/
├── wexler.yaml                 # Wexler 项目配置
├── CLAUDE.md                   # Claude Code 配置
├── .claude/
│   └── agents/
│       └── *.wexler.md
├── .cursor/
│   └── rules/
│       └── *.wexler.mdc
└── .mcp.json                   # MCP 配置

```

## 配置文件映射

### 1. 通用记忆（General Memory）

| Wexler 源 | Claude Code | Cursor |
| --- | --- | --- |
| `memory.mdc` | `CLAUDE.md`（二级标题标识） | `.cursor/rules/general.wexler.mdc` |

### 2. Subagent/Role 配置

| Wexler 源 | Claude Code | Cursor |
| --- | --- | --- |
| `subagent/code-reviewer.mdc` | `.claude/agents/code-reviewer.wexler.md` | `.cursor/rules/code-reviewer.wexler.mdc` |

### 3. MCP 配置

| Wexler 源 | Claude Code | Cursor |
| --- | --- | --- |
| `wexler.db` | `.mcp.json`（upsert） | `.cursor/mcp.json` |

## 项目配置文件（wexler.yaml）

```yaml
version: 1.0
source: /Users/alice/team-wexler-configs    # 本地文件系统路径
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
2. 从 BoltDB 解码 Wexler 管理的配置
3. 合并配置（Wexler 优先）
4. 写回 `.mcp.json`

## 源管理（Source）

### 当前版本（MVP）

仅支持本地文件系统源：

- 通过绝对路径或相对路径指定
- 支持符号链接
- 自动检查源目录有效性

### 后续计划

远程 Git 源支持（包括 clone、pull、分支管理等）将在后续版本实现，以确保 MVP 的快速交付。

## 未来扩展计划

1. **Phase 2**（配置增强）
    - 远程 Git 源支持（clone、pull、分支管理）
    - 配置继承和覆盖机制
    - 配置模板功能
2. **Phase 3**（工具扩展）
    - 支持更多 AI 工具（GitHub Copilot、Continue、Codeium）
    - IDE 插件配置同步
    - Web UI 管理界面
3. **Phase 4**（企业特性）
    - 真正的加密存储（AES-256）
    - 团队权限管理
    - 配置审计日志