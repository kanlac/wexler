# Data Model

## Core Entities

### ProjectConfig
**Purpose**: Represents wexler.yaml configuration file in project directory

**Fields**:
- `Version`: string - Configuration format version (e.g., "1.0")
- `Source`: string - Absolute or relative path to Wexler source directory
- `Tools`: []string - List of AI tools to manage ("claude", "cursor")

**Validation Rules**:
- Version must match supported format versions
- Source path must exist and be accessible
- Tools list must contain at least one supported tool
- Tools list cannot contain duplicates

**State Transitions**:
- Created: init command generates new config
- Loaded: apply/import commands read existing config
- Updated: Not applicable (config managed externally)

### MCPConfig
**Purpose**: Represents Model Context Protocol server configuration stored in BoltDB

**Fields**:
- `ID`: string - Unique identifier for MCP configuration
- `Servers`: map[string]string - Map of server name to base64 encoded JSON configuration
- `UpdatedAt`: time.Time - Timestamp of last modification

**Validation Rules**:
- ID must be non-empty and unique within database
- Servers map must contain at least one entry
- UpdatedAt must be valid timestamp
- Each server value must be valid base64 encoded JSON

**State Transitions**:
- Imported: import command extracts from AI tool configurations
- Updated: apply command modifies existing configuration
- Applied: apply command writes to AI tool configuration files

**Implementation Notes**:
- Server configurations are stored as base64 encoded JSON strings
- No need to parse mcpServers JSON structure - just encode/decode entire server blocks
- Example: mcpServers.context7 entire JSON object → base64 encode → store in Servers["context7"]

### SourceConfig
**Purpose**: Represents configuration files in Wexler source directory

**Fields**:
- `Memory`: MemoryConfig - General knowledge configuration
- `Subagents`: map[string]SubagentConfig - Role-specific configurations
- `Path`: string - Absolute path to source directory

**Validation Rules**:
- Path must exist and be readable
- Memory configuration must be present
- Subagent names must be valid identifiers

### MemoryConfig
**Purpose**: General AI knowledge and coding standards configuration

**Fields**:
- `Content`: string - Raw content from memory.mdc file
- `Sections`: map[string]string - Parsed sections by heading

**Validation Rules**:
- Content must be valid markdown
- Sections must have unique heading names

### SubagentConfig
**Purpose**: Role-specific AI configuration (code-reviewer, architect, etc.)

**Fields**:
- `Name`: string - Subagent identifier
- `Content`: string - Raw content from .mdc file (entire file content)

**Validation Rules**:
- Name must match filename without extension
- Content must be valid markdown
- Content must be non-empty

**Implementation Notes**:
- Subagent configurations are applied as entire file replacements
- Unlike MemoryConfig, no section-based parsing or merging
- Each subagent file generates a complete configuration file for the target tool

### ConflictResolution
**Purpose**: Tracks user decisions during configuration conflicts

**Fields**:
- `Mode`: ConflictMode - User's choice for handling conflicts
- `AppliedChanges`: []string - List of successfully applied configuration paths
- `PendingChanges`: []string - List of remaining configuration paths

**Validation Rules**:
- Mode must be valid enum value (Continue, ContinueAll, Stop)
- AppliedChanges must contain valid file paths
- PendingChanges must contain valid file paths

**State Transitions**:
- Continue: Process next conflict, prompt user
- ContinueAll: Apply all remaining changes without prompting
- Stop: Preserve applied changes, exit without remaining changes

### ConflictMode
**Purpose**: Enumeration of user choices during conflict resolution

**Values**:
- `Continue`: Apply current change and continue with next conflict
- `ContinueAll`: Apply current change and all subsequent changes without asking
- `Stop`: Apply current change but stop processing remaining conflicts

### ApplyProgress
**Purpose**: Tracks progress during configuration application

**Fields**:
- `TotalFiles`: int - Total number of configuration files to process
- `ProcessedFiles`: int - Number of files processed so far
- `SuccessfulFiles`: []string - List of successfully applied file paths
- `FailedFiles`: map[string]error - Map of file paths to error messages

**Validation Rules**:
- ProcessedFiles must not exceed TotalFiles
- SuccessfulFiles and FailedFiles combined should equal ProcessedFiles

### ToolConfig
**Purpose**: Abstract interface for tool-specific configuration generation

**Fields**:
- `ToolName`: string - Name of AI tool ("claude", "cursor")
- `ProjectRoot`: string - Project root directory path

**Methods** (Interface):
- `GenerateMemory(MemoryConfig) ([]ConfigFile, error)`: Generate memory configuration files
- `GenerateSubagents([]SubagentConfig) ([]ConfigFile, error)`: Generate subagent configuration files  
- `GenerateMCP(MCPConfig) ([]ConfigFile, error)`: Generate MCP configuration files
- `Validate(ConfigFile) error`: Validate generated configuration file
- `Merge(existing, new ConfigFile) (ConfigFile, error)`: Merge with existing configuration

### ConfigFile
**Purpose**: Represents a configuration file to be written

**Fields**:
- `Path`: string - Absolute path where file should be written
- `Content`: []byte - File content
- `Format`: ConfigFormat - File format type
- `IsSensitive`: bool - Whether file contains sensitive data (affects permissions)

### ConfigFormat
**Purpose**: Enumeration of configuration file formats

**Values**:
- `Markdown`: For CLAUDE.md and .wexler.md files
- `JSON`: For .mcp.json files  
- `YAML`: For .cursor-rules files

## Data Relationships

### Primary Relationships
- ProjectConfig → SourceConfig (1:1) - Project references single source
- SourceConfig → MemoryConfig (1:1) - Source contains one memory config
- SourceConfig → SubagentConfig (1:N) - Source contains multiple subagent configs
- ProjectConfig → MCPConfig (1:N) - Project can have multiple MCP configurations
- ApplyProgress → ConflictResolution (1:1) - Apply operation has conflict state

### Derived Relationships  
- ToolConfig ← MemoryConfig (N:1) - Tools generate from memory config
- ToolConfig ← SubagentConfig (N:N) - Tools generate from subagent configs
- ToolConfig ← MCPConfig (N:N) - Tools merge MCP configurations

## Storage Strategy

### File System Storage
- **ProjectConfig**: wexler.yaml in project directory
- **SourceConfig**: Files in source directory (memory.mdc, subagent/*.mdc)
- **ToolConfig**: Generated files in appropriate tool directories

### BoltDB Storage
- **Bucket: "mcp_configs"**: MCPConfig structs serialized as JSON
- **Bucket: "metadata"**: Timestamps and version information
- **Key Strategy**: Use config ID as key, JSON as value

### Memory Storage
- **ConflictResolution**: Ephemeral state during apply operation
- **ApplyProgress**: Progress tracking during operations
- **Parsed Configs**: Cached parsed configurations during operations

## Data Flow

### Import Operation
1. Load ProjectConfig from wexler.yaml
2. Scan AI tool configuration directories
3. Parse existing MCP configurations
4. Create MCPConfig entities with base64 encoded secrets
5. Store MCPConfig entities in BoltDB

### Apply Operation
1. Load ProjectConfig from wexler.yaml
2. Load SourceConfig from source directory
3. Load MCPConfig entities from BoltDB
4. For each enabled tool:
   - Generate ToolConfig using MemoryConfig, SubagentConfig, MCPConfig
   - Check for conflicts with existing configurations
   - Handle conflicts using ConflictResolution
   - Write configuration files
   - Update ApplyProgress

### List Operation
1. Load ProjectConfig from wexler.yaml
2. Load SourceConfig from source directory  
3. Load MCPConfig entities from BoltDB
4. Display summary of all managed configurations

This data model supports the hackathon timeline while providing clear structure for all configuration management operations.