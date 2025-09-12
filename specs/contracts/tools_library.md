# Tools Library Contract

## Interface: ToolAdapter

### Generate Memory Configuration
**Operation**: GenerateMemory(memory MemoryConfig) ([]ConfigFile, error)
**Input**:
  - memory: General memory configuration from memory.mdc
**Output**:
  - []ConfigFile: Generated memory configuration files for the tool
  - error: Generation or validation errors
**Behavior**:
  - MUST generate memory configuration in tool's native format
  - MUST use section-based merge approach for memory content
  - MUST create appropriate file structures for the target tool

### Generate Subagent Configurations
**Operation**: GenerateSubagents(subagents []SubagentConfig) ([]ConfigFile, error)
**Input**:
  - subagents: List of role-specific configurations
**Output**:
  - []ConfigFile: Generated subagent configuration files
  - error: Generation or validation errors
**Behavior**:
  - MUST generate one complete file per subagent (entire file replacement)
  - MUST use tool's native format for subagent files
  - MUST preserve complete subagent content without section parsing

### Generate MCP Configuration
**Operation**: GenerateMCP(mcp MCPConfig) ([]ConfigFile, error)
**Input**:
  - mcp: MCP server configurations with decoded secrets
**Output**:
  - []ConfigFile: Generated MCP configuration files
  - error: Generation or validation errors
**Behavior**:
  - MUST decode base64 encoded server configurations
  - MUST merge MCP configurations using upsert logic
  - MUST generate tool-specific MCP configuration format

### Validate Tool Configuration
**Operation**: Validate(config []byte) error
**Input**:
  - config: Generated configuration content to validate
**Output**:
  - error: Validation errors or nil if valid
**Behavior**:
  - MUST validate configuration follows tool's format requirements
  - MUST check for required sections/fields
  - MUST validate syntax (markdown, JSON, etc.)
  - MUST check for conflicting configurations

### Merge Configurations
**Operation**: Merge(existing, new []byte) ([]byte, error)
**Input**:
  - existing: Current configuration file content
  - new: New configuration content to merge
**Output**:
  - []byte: Merged configuration content
  - error: Merge conflicts or format errors
**Behavior**:
  - MUST preserve non-Wexler managed content
  - MUST replace Wexler-managed sections/files
  - MUST use tool-specific merge strategies
  - MUST handle missing existing configuration

### Detect Configuration Conflicts
**Operation**: DetectConflicts(existing, new []byte) ([]string, error)
**Input**:
  - existing: Current configuration content
  - new: New configuration content to apply
**Output**:
  - []string: List of conflicting section/field descriptions
  - error: Analysis errors
**Behavior**:
  - MUST identify sections that would be overwritten
  - MUST report human-readable conflict descriptions
  - MUST distinguish Wexler-managed vs user-managed content
  - MUST return empty slice if no conflicts

## Interface: ClaudeAdapter (implements ToolAdapter)

### Generate Memory Configuration for Claude
**Behavior**:
  - MUST create CLAUDE.md with memory content in designated sections
  - MUST use second-level headings to separate memory sections in CLAUDE.md
  - MUST preserve existing non-Wexler content in CLAUDE.md
  - MUST use section-based merge approach

### Generate Subagent Configurations for Claude
**Behavior**:
  - MUST create .claude/agents/{subagent-name}.wexler.md for each subagent
  - MUST write complete subagent content without section parsing (entire file replacement)
  - MUST use .wexler.md extension to identify Wexler-managed files
  - MUST preserve existing .claude/agents files without .wexler extension

### Generate MCP Configuration for Claude
**Behavior**:
  - MUST generate/update .mcp.json with server configurations
  - MUST decode base64 encoded server configurations before writing
  - MUST use upsert logic to merge with existing MCP servers
  - MUST preserve existing non-Wexler MCP server configurations

## Interface: CursorAdapter (implements ToolAdapter)

### Generate Memory Configuration for Cursor
**Behavior**:
  - MUST create .cursor/rules/general.wexler.mdc with memory content
  - MUST use .wexler.mdc extension to identify Wexler-managed files
  - MUST preserve existing .cursor/rules files without .wexler extension
  - MUST use section-based merge approach if needed

### Generate Subagent Configurations for Cursor
**Behavior**:
  - MUST create .cursor/rules/{subagent-name}.wexler.mdc for each subagent
  - MUST write complete subagent content without section parsing (entire file replacement)
  - MUST use .wexler.mdc extension to identify Wexler-managed files
  - MUST preserve existing .cursor/rules files without .wexler extension

### Generate MCP Configuration for Cursor  
**Behavior**:
  - MUST generate/update .cursor/mcp.json with server configurations
  - MUST decode base64 encoded server configurations before writing
  - MUST use upsert logic to merge with existing MCP servers
  - MUST preserve existing non-Wexler MCP server configurations

## Factory Functions

### Create Tool Adapter
**Operation**: NewToolAdapter(toolName string) (ToolAdapter, error)
**Input**:
  - toolName: Name of AI tool ("claude", "cursor")
**Output**:
  - ToolAdapter: Tool-specific adapter implementation
  - error: Unsupported tool errors
**Behavior**:
  - MUST return appropriate adapter for supported tools
  - MUST return error for unsupported tool names
  - MUST validate tool name is non-empty

## Error Conditions
- `ErrUnsupportedTool`: tool name not supported
- `ErrInvalidFormat`: generated configuration has invalid format
- `ErrMergeConflict`: unable to merge configurations safely
- `ErrConfigurationCorrupted`: existing configuration file is malformed
- `ErrPermissionDenied`: unable to write to tool configuration directory