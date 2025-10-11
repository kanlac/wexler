# Mindful 软链接架构重设计方案

**文档版本**: v1.0
**创建日期**: 2025-09-26
**重大架构转型**: 从复杂的内容生成转向极简的软链接引用

## 🚀 架构革新的核心洞察

### 根本性发现
**传统方案的复杂度来源**: 为不同工具生成"本质相同但格式略异"的配置文件
**软链接方案的核心**: 操作系统层面的文件引用机制，一份内容，多个访问路径

### 架构转型对比

| 维度 | 传统内容生成方案 | 软链接引用方案 |
|------|------------------|----------------|
| **核心流程** | Source → Parse → Generate → Write | Source → Create Symlinks |
| **代码复杂度** | ~3000 行代码 | ~800 行代码 (-70%) |
| **运行时性能** | O(N*M) 解析转换 | O(N) 软链接创建 |
| **概念复杂度** | 适配器+生成器+提取器 | 软链接映射表 |
| **维护成本** | 每个工具需要独立适配器 | 配置映射表维护 |
| **扩展性** | 新工具需要编程实现 | 新工具只需配置映射 |

## 📁 全新的用户端项目文件结构

```
用户项目根目录/
├── mindful/
│   ├── mindful.yaml
│   ├── project-memory.mdc (可选，手动维护项目级配置)
│   ├── project-subagents/ (可选，团队共享子代理)
│   └── out/ (mindful 生成的最终配置，覆盖 project-scope 与 team-scope)
│       ├── memory.md
│       ├── subagents/
│       └── mcp.json
├── CLAUDE.md -> mindful/out/memory.md
├── AGENTS.md -> mindful/out/memory.md (Codex, Gemini CLI)
├── .mcp.json -> mindful/out/mcp.json (Claude Code MCP 配置)
├── .cursor/
│   ├── rules/
│   │   ├── general.mindful.mdc -> ../../mindful/out/memory.md
│   │   ├── researcher.mindful.mdc -> ../../mindful/out/subagents/researcher.mdc
│   │   └── reviewer.mindful.mdc -> ../../mindful/out/subagents/reviewer.mdc
│   └── mcp.json -> ../mindful/out/mcp.json
└── .claude/
    └── agents/
        ├── researcher.mindful.md -> ../../mindful/out/subagents/researcher.md
        └── reviewer.mindful.md -> ../../mindful/out/subagents/reviewer.md
```

## 🏗️ 极简化的代码架构

### 新的核心模块结构
```
src/
├── models/
│   ├── config.go          # 项目配置 + 软链接映射配置
│   ├── source.go          # 源文件配置 (简化)
│   └── symlink.go         # 软链接配置模型
├── config/
│   └── manager.go         # 项目配置管理
├── source/
│   ├── manager.go         # 源配置解析 (简化)
│   └── parser.go          # 标记解析 (保留)
├── storage/
│   └── manager.go         # MCP 配置存储 (保留)
├── symlink/               # 🆕 软链接管理核心模块
│   ├── manager.go         # 软链接创建/更新/清理
│   ├── config.go          # 工具软链接配置
│   └── resolver.go        # 路径解析和验证
└── cli/
    ├── init.go            # 初始化 (简化)
    ├── build.go           # 构建 mindful/out 产物
    ├── apply.go           # 🔄 构建产物并分发软链接
    ├── list.go            # 列出配置 (简化)
    └── import.go          # 导入配置 (简化)
```

### 删除的模块 (大幅简化)
```
🗑️ 完全删除的目录和文件:
├── src/tools/             # 整个工具适配器目录
│   ├── claude/
│   ├── cursor/
│   ├── adapter.go
│   └── types/
├── src/apply/
│   ├── content_extractor.go  # 内容提取器
│   └── manager.go (部分逻辑) # 复杂的应用管理逻辑
└── tests/
    └── 相关的适配器测试文件
```

## 🎯 软链接配置方案

### mindful.yaml 配置格式扩展
```yaml
# 基础项目配置
name: "my-project"
version: "1.0.0"
source: "/path/to/team-configs"

# 工具启用配置
enable-coding-agents:
  - claude
  - cursor
  - codex

# 🆕 软链接映射配置
symlinks:
  claude:
    memory: "CLAUDE.md"
    subagents: ".claude/agents/{name}.mindful.md"
    mcp: ".mcp.json"

  cursor:
    memory: ".cursor/rules/general.mindful.mdc"
    subagents: ".cursor/rules/{name}.mindful.mdc"
    mcp: ".cursor/mcp.json"

  codex:
    memory: "AGENTS.md"
    subagents: ""   # Codex 暂不支持 subagent
    mcp: ""         # Codex 通过 ~/.codex/config.toml 文件管理 MCP，不支持直接加载项目目录中的 mcp json 文件
```

### 软链接管理器接口设计
```go
// src/symlink/manager.go
type Manager struct {
    projectPath string
    config     *models.SymlinkConfig
}

func NewManager(projectPath string, config *models.SymlinkConfig) *Manager

func (m *Manager) CreateSymlinks(toolName string) error
func (m *Manager) UpdateSymlinks(toolName string) error
func (m *Manager) CleanupSymlinks(toolName string) error
func (m *Manager) ValidateSymlinks(toolName string) error
func (m *Manager) ListSymlinks(toolName string) ([]SymlinkInfo, error)

type SymlinkInfo struct {
    LinkPath   string
    TargetPath string
    IsValid    bool
    IsDirectory bool
}
```

## 🚀 mindful CLI 工作流程

### mindful build：生成 mindful/out
- 单一职责：读取源配置、渲染 memory/subagents/mcp，并写入 `mindful/out`
- 幂等设计：先清理旧产物，再写入全量结果，确保后续 `apply` 始终指向最新文件

```go
func runBuild(cmd *cobra.Command, args []string) error {
    ctx := loadBuildContext(cmd, args)                   // 1. 加载构建上下文
    sourceConfig := loadSourceConfiguration(ctx)         // 2. 读取团队/项目级配置

    artifacts, err := renderArtifacts(ctx, sourceConfig) // 3. 渲染 memory/subagents/mcp
    if err != nil {
        return err
    }

    return writeArtifacts(ctx.OutPath, artifacts)        // 4. 写入 mindful/out
}
```

### mindful apply：分发软链接
- 自动构建：默认先执行一次 `mindful build`
- 核心任务：将工具所需文件软链接到 `mindful/out` 中的对应产物
- 输出优化：统一展示实际创建/更新的链接，以及校验结果

```go
func runApply(cmd *cobra.Command, args []string) error {
    ctx := loadApplyContext(cmd, args)                    // 1. 加载软链接上下文

    if err := buildArtifactsIfNeeded(ctx); err != nil {   // 2. 构建或校验 mindful/out 产物
        return err
    }

    enabledTools := determineEnabledTools(ctx)            // 3. 确定需要分发的目标工具

    return executeSymlinkCreation(ctx, enabledTools)      // 4. 创建或更新软链接
}

func executeSymlinkCreation(ctx *ApplyContext, tools []string) error {
    symlinkManager := symlink.NewManager(ctx.ProjectPath, ctx.SymlinkConfig)
    results := make([]ToolResult, len(tools))

    for i, toolName := range tools {
        result := ToolResult{Tool: toolName}

        if ctx.DryRun {
            // 干运行: 仅规划指向 mindful/out 的软链接
            links, err := symlinkManager.PlanSymlinks(toolName)
            result.PlannedLinks = links
            result.Error = err
        } else {
            // 实际执行: 创建或更新软链接，指向 mindful/out
            err := symlinkManager.CreateSymlinks(toolName)
            result.Error = err
        }

        results[i] = result
    }

    return reportResults(ctx, results)
}
```

## 🔧 实施阶段规划

### 阶段 1: 软链接核心模块 (1.5小时)

#### 1.1 基础设施创建
- 创建 `src/symlink/` 包
- 实现 `Manager` 结构体和核心方法
- 实现软链接配置解析逻辑
- 添加路径验证和冲突检测

#### 1.2 配置格式扩展
- 扩展 `mindful.yaml` 支持软链接映射
- 更新 `src/models/config.go` 添加软链接配置结构
- 实现向后兼容性处理

### 阶段 2: CLI 命令重写 (1小时)

#### 2.1 build 命令新增
- 新增 `cli/build.go`，封装渲染与写入 mindful/out 的流程
- 抽离产物渲染逻辑，保证幂等写入与目录清理

#### 2.2 apply 命令重写
- 重写 `runApply()`，默认触发构建并分发指向 mindful/out 的软链接
- 实现软链接创建的错误处理 (跳过失败继续)
- 设计改进的用户输出格式，展示每个目标工具的链接状态

#### 2.3 其他命令简化
- 简化 `list` 命令显示软链接状态
- 更新 `init` 命令生成软链接配置模板
- 保持 `import` 命令基本不变

### 阶段 3: 大规模代码删除 (0.5小时)

#### 3.1 模块删除
- 删除整个 `src/tools/` 目录
- 删除 `src/apply/content_extractor.go`
- 清理 `src/apply/manager.go` 中的复杂逻辑

#### 3.2 依赖清理
- 更新所有 import 语句
- 删除无用的依赖和测试文件
- 简化模型结构体定义

### 阶段 4: 测试和优化 (1小时)

#### 4.1 软链接测试
- 编写软链接创建/更新/清理的测试
- 测试跨平台兼容性 (Windows/Linux/Mac)
- 验证相对路径和绝对路径处理

#### 4.2 用户体验优化
- 改进进度指示和成功/失败信息
- 添加软链接状态检查和修复功能
- 优化干运行模式的信息展示

## 📊 预期效果评估

### 性能提升
- **启动时间**: 从秒级降至毫秒级
- **内存使用**: 减少70% (无需加载多个适配器)
- **磁盘使用**: 减少重复文件存储

### 开发体验
- **代码行数**: 减少70% (从~3000行到~800行)
- **概念复杂度**: 显著降低 (无需理解适配器模式)
- **调试难度**: 大幅降低 (软链接状态直观可见)

### 用户体验
- **执行速度**: 毫秒级响应
- **配置一致性**: 真正的单一数据源
- **工具兼容**: 支持任意工具的自定义路径配置

### 扩展性
- **新工具支持**: 仅需配置软链接映射，无需编程
- **自定义路径**: 完全支持工具特定的文件路径需求
- **团队配置**: 软链接配置可版本控制和共享

## ⚠️ 实施风险和缓解策略

### 技术风险

#### R1: 跨平台兼容性 - **中等**
**风险**: Windows 对软链接的支持不同于 Unix 系统
**缓解**:
- 使用 Go 的 `os.Symlink()` API 处理跨平台差异
- Windows 下自动检测管理员权限，提供友好提示
- 提供 hardlink 备选方案用于不支持软链接的环境

#### R2: 现有项目迁移 - **低**
**风险**: 现有用户的项目迁移复杂
**缓解**:
- 实现自动迁移工具检测现有文件结构
- 提供向后兼容模式支持渐进迁移
- 详细的迁移指南和工具

### 用户体验风险

#### R3: 软链接概念理解 - **低**
**风险**: 用户不理解软链接概念
**缓解**:
- 提供清晰的软链接状态显示
- 实现 `mindful status` 命令显示链接健康度
- 文档中提供软链接基础知识说明

## 🎯 成功验收标准

### 功能验收
- ✅ 支持至少3个主流工具 (Claude Code, Cursor, Gemini)
- ✅ 软链接创建/更新/清理功能完整
- ✅ 跨平台兼容性验证通过
- ✅ 现有项目自动迁移成功

### 性能验收
- ✅ `mindful apply` 执行时间 < 100ms (vs 当前 1-2s)
- ✅ 代码行数减少 > 60%
- ✅ 内存占用减少 > 50%

### 用户体验验收
- ✅ 5分钟内完成新项目配置
- ✅ 编辑一次配置，所有工具同步更新
- ✅ 清晰的软链接状态反馈

---

## 总结

软链接策略代表了从"复杂的内容复制"到"简单的引用映射"的根本性架构转变。这不仅仅是重构，而是重新发明了一个更优雅、更高效、更易维护的解决方案。

通过利用操作系统层面的软链接机制，我们将一个复杂的多工具配置同步问题，转化为一个简单的文件系统引用问题。这是典型的"站在巨人肩膀上"的工程智慧体现。
