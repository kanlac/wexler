# Source Library Contract

## Interface: SourceManager

### Load Source Configuration
**Operation**: LoadSource(sourcePath string) (SourceConfig, error)
**Input**:
  - sourcePath: Absolute path to Wexler source directory
**Output**:
  - SourceConfig: Parsed source configuration with memory and subagents
  - error: File system or parsing errors
**Behavior**:
  - MUST read memory.mdc from source directory
  - MUST scan subagent/ directory for .mdc files
  - MUST parse markdown content into sections
  - MUST validate all required files exist

### List Source Files
**Operation**: ListSourceFiles(sourcePath string) ([]string, error)
**Input**:
  - sourcePath: Absolute path to Wexler source directory
**Output**:
  - []string: List of configuration file paths
  - error: File system errors
**Behavior**:
  - MUST return memory.mdc path
  - MUST return all .mdc files in subagent/ directory
  - MUST return relative paths from source directory
  - MUST filter only .mdc files

### Parse Memory Configuration
**Operation**: ParseMemory(filePath string) (MemoryConfig, error)
**Input**:
  - filePath: Absolute path to memory.mdc file
**Output**:
  - MemoryConfig: Parsed memory configuration
  - error: File system or parsing errors
**Behavior**:
  - MUST read entire file content
  - MUST parse markdown sections by headings
  - MUST validate markdown format
  - MUST handle empty sections gracefully

### Parse Subagent Configuration
**Operation**: ParseSubagent(filePath string) (SubagentConfig, error)
**Input**:
  - filePath: Absolute path to subagent .mdc file
**Output**:
  - SubagentConfig: Parsed subagent configuration
  - error: File system or parsing errors
**Behavior**:
  - MUST extract subagent name from filename
  - MUST read entire file content as single unit
  - MUST validate markdown format (but no section parsing)
  - MUST validate content is non-empty

### Validate Source Structure
**Operation**: ValidateSource(sourcePath string) error
**Input**:
  - sourcePath: Absolute path to source directory
**Output**:
  - error: Validation errors or nil if valid
**Behavior**:
  - MUST check source directory exists and is readable
  - MUST check memory.mdc exists
  - MUST check subagent/ directory exists
  - MUST validate all .mdc files are readable

## Error Conditions
- `ErrSourceNotFound`: source directory does not exist
- `ErrSourceNotAccessible`: source directory not readable
- `ErrMemoryNotFound`: memory.mdc file not found
- `ErrSubagentDirNotFound`: subagent/ directory not found
- `ErrInvalidMarkdown`: malformed markdown content
- `ErrEmptyConfiguration`: configuration file has no content