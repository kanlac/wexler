# Project Scope Memory 功能技术规格书

## 功能概述

Mindful 将支持双层 Memory 配置结构：团队级别（team scope）和项目级别（project scope）。`mindful apply` 命令将解析两个层级的 memory 文件并合并生成最终的工具配置文件。

## 数据结构变更

### 关键架构区别

**1. 用户项目目录结构** (mindful 管理的目标):
```
用户项目/
├── mindful.yaml
└── mindful/                 # Project scope configurations
    ├── memory.mdc          # Project-specific memory
    └── subagents/          # Project-specific subagents
```

**2. 外部 Source 目录** (团队共享配置):
```
/path/to/external/source/    # Team scope configurations (项目外部)
├── memory.mdc              # Team-wide memory
└── subagents/              # Team-wide subagents
```

### 配置文件更新
```yaml
# mindful.yaml 配置
source: "/path/to/external/source"  # 外部团队配置目录路径

# 项目内配置路径固定为:
# - mindful/memory.mdc (project scope)
# - mindful/subagents/ (project scope subagents)
```

### 内存数据模型
```go
// 原有单层结构
type MemoryConfig struct {
    Content string
}

// 新双层结构
type MemoryConfig struct {
    TeamContent    string  // 外部 source/memory.mdc 内容
    ProjectContent string  // 项目内 mindful/memory.mdc 内容
    HasTeam        bool    // 团队级别文件是否存在
    HasProject     bool    // 项目级别文件是否存在
    TeamSourcePath string  // 团队配置来源路径 (用于标记)
    ProjectSourcePath string // 项目配置来源路径 (用于标记)
}
```

## API 流程变更

### mindful apply 执行流程

**原有流程**:
```
mindful apply → 解析外部source/memory.mdc → 生成工具配置文件
```

**新流程**:
```
mindful apply →
  1. 从 mindful.yaml 读取外部source目录路径
  2. 检查并解析外部 source/memory.mdc (team scope)
  3. 检查并解析项目内 mindful/memory.mdc (project scope)
  4. 合并两层内容 (分离保存，不做内容合并)
  5. 删除现有工具配置文件
  6. 重新生成包含双层结构和来源标记的配置文件
```

### 生成文件格式 (以 Claude Code 为例)

**CLAUDE.md 结构**:
```markdown
# Mindful Memory (scope: team)
<!-- Source: /path/to/external/source/memory.mdc -->
[外部source/memory.mdc 的完整内容]

# Mindful Memory (scope: project)
<!-- Source: mindful/memory.mdc -->
[项目内mindful/memory.mdc 的完整内容]

[其他现有配置内容保持不变...]
```

**生成规则**:
- 总是先生成 team scope，再生成 project scope
- 如果某个 scope 的文件不存在，静默跳过对应段落
- 必须标记内容来源文件路径 (HTML 注释形式)
- 完整删除并重新创建目标文件，不做部分更新，保证原子性
- 强制验证所有 .mdc 文件为 UTF-8 编码

## 核心测试场景

### 1. 双文件存在场景
```go
// 伪代码测试
func TestDualMemoryExist() {
    setup := {
        "/external/source/memory.mdc": "团队通用配置",
        "mindful/memory.mdc": "项目特定配置"
    }

    result := mindfulApply()

    assertions := {
        "包含团队标题": "# Mindful Memory (scope: team)",
        "包含项目标题": "# Mindful Memory (scope: project)",
        "内容顺序正确": team_content_before_project_content,
        "文件被重新创建": file_recreated_not_updated
    }
}
```

### 2. 单文件存在场景
```go
func TestSingleMemoryFile() {
    // 场景2A: 仅 team scope
    setup := {"/external/source/memory.mdc": "团队配置"}
    result := mindfulApply()
    assert.Contains("scope: team")
    assert.NotContains("scope: project")

    // 场景2B: 仅 project scope
    setup := {"mindful/memory.mdc": "项目配置"}
    result := mindfulApply()
    assert.NotContains("scope: team")
    assert.Contains("scope: project")

    // 场景2C: 无任何 memory 文件
    setup := {} // 无文件
    result := mindfulApply()
    assert.NotContains("Mindful Memory") // 静默跳过
}
```

### 3. 配置文件重新生成验证
```go
func TestFileRegeneration() {
    // 现有文件包含其他内容
    existing_claude_md := "现有内容\n# 其他标题\n内容"

    mindfulApply()

    // 验证文件被完全重新创建，而非部分更新
    assert.FileWasDeleted("CLAUDE.md")
    assert.FileWasRecreated("CLAUDE.md")
    assert.NewContentFormat("双层 memory 结构")
}
```

### 4. 错误处理测试
```go
func TestErrorHandling() {
    // UTF-8 编码验证
    invalidEncodingFile := "非 UTF-8 编码内容"
    result := mindfulApply()
    assert.Error("UTF-8 encoding validation failed")

    // 文件权限错误
    unreadableFile := "/root/restricted.mdc"
    result := mindfulApply()
    assert.Error("无法读取文件")

    // 空文件处理
    emptyFiles := {"/external/source/memory.mdc": "", "mindful/memory.mdc": ""}
    result := mindfulApply()
    assert.NotContains("Mindful Memory") // 静默跳过空内容

    // 目录不存在 - 静默跳过
    missingDirectories := {"mindful/": not_exists, "source/": not_exists}
    result := mindfulApply()
    assert.Success() // 不报错，静默跳过
}
```

## 向后兼容性分析

### 受影响的现有功能
1. **Memory 解析器**: 需要从单文件解析扩展为双文件解析
2. **配置文件生成器**: 需要支持双层内容合并
3. **文件路径解析**: 需要检查新增的 `mindful/` 目录

### 兼容性保证策略
- **渐进式采用**: 现有项目无需强制迁移
- **自动降级**: 如果 `mindful/memory.mdc` 不存在，自动使用单层模式
- **配置可选**: `mindful.yaml` 中的新配置项为可选

### 迁移路径
```bash
# 现有项目 - 无需更改，继续正常工作
mindful apply  # 仅使用外部 team scope memory

# 初始化项目级配置
mindful init   # 生成 mindful/ 目录和模板

# 手动添加项目级配置
mkdir -p mindful
echo "项目特定配置" > mindful/memory.mdc
mindful apply  # 自动启用双层模式

# mindful.yaml 配置外部 source 路径
source: "/path/to/team/configs"
```

## 破坏性变更清单

### 1. 生成文件重新创建行为变更
- **变更**: 从部分更新改为完整删除后重新创建 (CLAUDE.md, Cursor 配置文件)
- **影响**: 手动修改的工具配置文件内容将丢失
- **缓解**: 建议将手动配置迁移到 memory.mdc 或 subagents 文件中

### 2. 项目目录结构依赖
- **变更**: 需要检查项目内 `mindful/` 目录和外部 source 目录
- **影响**: 文件系统扫描逻辑需要更新，需支持双路径检查
- **缓解**: 向后兼容，任何一个目录不存在时静默跳过

### 3. Memory 配置解析器接口
- **变更**: Memory 解析器需支持双文件读取和 UTF-8 验证
- **影响**: 现有单文件解析逻辑需要扩展
- **缓解**: 保持向后兼容的 API 设计，支持单双文件模式

## 需要澄清的业务需求

### 已澄清的高优先级问题

1. **内容冲突处理**: ✅ **不做冲突检测** - team 和 project memory 数据完全分离，分别显示为两个一级标题，无需考虑优先级

2. **文件不存在策略**: ✅ **静默跳过** - 两个 memory 文件都不存在时静默跳过，不报错

3. **目录自动创建**: ✅ **不自动创建** - `mindful apply` 只负责读取，`mindful init` 可生成目录和模板内容

### 已澄清的技术实现问题

4. **编码格式验证**: ✅ **强制 UTF-8** - 需要强制验证所有 .mdc 文件为 UTF-8 编码

5. **文件权限设置**: ✅ **无特殊权限** - `mindful/memory.mdc` 不需要特殊文件权限

6. **错误恢复机制**: ✅ **原子文件替换** - 每次替换都是整个文件替换，保证原子性，无需考虑复杂的错误恢复

### 已澄清的用户体验问题

7. **标题本地化**: ✅ **未来预留** - 多语言应该是 mindful 交互界面的多语言，与 scope 无关。可为未来预留，但现在不实现

8. **内容分隔符**: ✅ **不需要** - 两个 memory 块之间不需要特殊的视觉分隔线

9. **来源标记**: ✅ **需要** - 必须在生成文件中使用 HTML 注释标记内容来源文件路径

### 已澄清的扩展性问题

10. **Cursor 工具适配**: ✅ **需要** - Cursor 工具也需要相同的双层 memory 结构

11. **其他 Scope 扩展**: ✅ **不考虑** - 未来不计划支持更多 scope 级别（如 workspace, organization）

12. **版本兼容标记**: ✅ **不需要** - 无需在 `mindful.yaml` 中声明支持的功能版本号

---

## 实施建议

### 开发阶段优先级
1. **Phase 1**: 实现核心双文件解析逻辑和 UTF-8 验证
2. **Phase 2**: 更新 Claude Code 和 Cursor 工具适配器以支持双层结构
3. **Phase 3**: 实现文件来源标记和原子文件替换逻辑
4. **Phase 4**: 向后兼容性测试和错误处理完善

### 测试驱动开发重点
- **核心功能测试**: 双文件解析、UTF-8 验证、来源标记
- **边界情况覆盖**: 文件不存在、空文件、权限错误的静默处理
- **格式验证**: 生成文件的正确结构和 HTML 注释格式
- **兼容性保证**: 现有单层项目的继续工作和新双层项目的正确运行

---

## 总结

本功能将 Mindful 的 memory 配置从单一 team scope 扩展为双层架构：

- **Team Scope**: 外部共享的团队配置 (source/memory.mdc)
- **Project Scope**: 项目特定的本地配置 (mindful/memory.mdc)

所有技术实现问题已澄清，支持向后兼容、UTF-8 验证、来源标记和原子文件操作，可直接进入 TDD 开发阶段。