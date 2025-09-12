# Apply Library Contract

## Interface: ApplyManager

### Apply Configuration
**Operation**: ApplyConfig(projectConfig ProjectConfig, sourceConfig SourceConfig, mcpConfigs []MCPConfig) (ApplyProgress, error)
**Input**:
  - projectConfig: Project configuration specifying tools and source
  - sourceConfig: Source configuration with memory and subagents
  - mcpConfigs: List of MCP configurations to apply
**Output**:
  - ApplyProgress: Progress tracking with successful and failed operations
  - error: Critical errors that prevent continuation
**Behavior**:
  - MUST process each enabled tool in projectConfig.Tools
  - MUST detect conflicts before applying changes
  - MUST prompt user for conflict resolution
  - MUST track progress and allow partial completion

### Apply Single Tool
**Operation**: ApplyTool(toolName string, sourceConfig SourceConfig, mcpConfigs []MCPConfig) (ApplyResult, error)
**Input**:
  - toolName: AI tool to apply configuration to
  - sourceConfig: Source configuration to apply
  - mcpConfigs: MCP configurations for the tool
**Output**:
  - ApplyResult: Result with restart notification requirements
  - error: Tool-specific application errors
**Behavior**:
  - MUST create tool adapter for specified tool
  - MUST generate tool-specific configuration
  - MUST handle existing configuration conflicts
  - MUST write configuration files atomically
  - MUST detect MCP configuration changes and set restart notification flag

### Detect All Conflicts
**Operation**: DetectConflicts(projectConfig ProjectConfig, sourceConfig SourceConfig) ([]ConflictInfo, error)
**Input**:
  - projectConfig: Project configuration specifying tools
  - sourceConfig: Source configuration to check against
**Output**:
  - []ConflictInfo: List of all detected conflicts with details
  - error: Analysis errors
**Behavior**:
  - MUST check all enabled tools for conflicts
  - MUST identify specific files and sections that conflict
  - MUST provide human-readable conflict descriptions
  - MUST distinguish overwrites from additions

### Resolve Conflict
**Operation**: ResolveConflict(conflict ConflictInfo) (ConflictResolution, error)
**Input**:
  - conflict: Specific conflict to resolve with user
**Output**:
  - ConflictResolution: User's choice for handling this conflict
  - error: User interaction errors
**Behavior**:
  - MUST display conflict details to user
  - MUST show diff between existing and new configuration
  - MUST present options: Continue, Continue All, Stop
  - MUST validate user input

### Preview Changes
**Operation**: PreviewChanges(projectConfig ProjectConfig, sourceConfig SourceConfig) ([]ChangeInfo, error)
**Input**:
  - projectConfig: Project configuration specifying tools
  - sourceConfig: Source configuration to preview
**Output**:
  - []ChangeInfo: List of changes that would be made
  - error: Analysis errors
**Behavior**:
  - MUST show all files that would be created/modified
  - MUST show summary of changes for each file
  - MUST identify Wexler-managed vs user-managed content
  - MUST provide change size estimates

## Data Structures

### ApplyResult
**Fields**:
  - RequiresRestart: bool - Whether AI tools need restart for changes to take effect
  - ChangedMCPConfigs: []string - List of MCP configuration files that were modified
  - Message: string - Human-readable result message

### ConflictInfo
**Fields**:
  - ToolName: string - AI tool where conflict occurs
  - FilePath: string - Configuration file path with conflict
  - ConflictType: ConflictType - Type of conflict (overwrite, format, etc.)
  - ExistingContent: string - Current content that would be replaced
  - NewContent: string - New content from Wexler source
  - Description: string - Human-readable conflict description

### ConflictType
**Values**:
  - OverwriteFile: Wexler would overwrite user-managed file
  - OverwriteSection: Wexler would overwrite user-managed section
  - FormatMismatch: Existing file format incompatible with Wexler
  - PermissionDenied: Unable to write to configuration file

### ChangeInfo
**Fields**:
  - ToolName: string - AI tool being configured
  - FilePath: string - Configuration file path
  - ChangeType: ChangeType - Type of change being made
  - Size: int - Approximate size of change in bytes
  - Description: string - Human-readable change description

### ChangeType
**Values**:
  - CreateFile: New configuration file will be created
  - UpdateFile: Existing file will be modified
  - AddSection: New section added to existing file
  - ReplaceSection: Existing section will be replaced

## Error Conditions
- `ErrNoProjectConfig`: project configuration not provided or invalid
- `ErrNoSourceConfig`: source configuration not provided or invalid
- `ErrToolNotSupported`: specified tool not supported
- `ErrUserAborted`: user chose to stop during conflict resolution
- `ErrWritePermissionDenied`: unable to write configuration files
- `ErrConfigurationLocked`: configuration files locked by another process