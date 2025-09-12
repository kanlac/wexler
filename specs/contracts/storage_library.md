# Storage Library Contract

## Interface: StorageManager

### Store MCP Configuration
**Operation**: StoreMCP(config MCPConfig) error
**Input**:
  - config: MCPConfig to store with base64 encoded secrets
**Output**:
  - error: Database or validation errors
**Behavior**:
  - MUST validate config ID is non-empty and unique
  - MUST serialize config to JSON
  - MUST store in "mcp_configs" bucket with ID as key
  - MUST update timestamp in "metadata" bucket

### Retrieve MCP Configuration
**Operation**: RetrieveMCP(id string) (MCPConfig, error)
**Input**:
  - id: Unique identifier for MCP configuration
**Output**:
  - MCPConfig: Retrieved configuration with decoded secrets
  - error: Database or not found errors
**Behavior**:
  - MUST look up config in "mcp_configs" bucket
  - MUST deserialize JSON to MCPConfig struct
  - MUST decode base64 secrets to plain text
  - MUST return error if ID not found

### List MCP Configurations
**Operation**: ListMCP() ([]MCPConfig, error)
**Input**: None
**Output**:
  - []MCPConfig: All stored MCP configurations
  - error: Database errors
**Behavior**:
  - MUST iterate through "mcp_configs" bucket
  - MUST deserialize all stored configurations
  - MUST decode all base64 secrets
  - MUST return empty slice if no configs exist

### Update MCP Configuration
**Operation**: UpdateMCP(config MCPConfig) error
**Input**:
  - config: MCPConfig with updated values
**Output**:
  - error: Database or validation errors
**Behavior**:
  - MUST verify config ID exists in database
  - MUST encode secrets to base64
  - MUST serialize config to JSON
  - MUST update both config and timestamp

### Delete MCP Configuration
**Operation**: DeleteMCP(id string) error
**Input**:
  - id: Unique identifier for MCP configuration to delete
**Output**:
  - error: Database errors
**Behavior**:
  - MUST remove config from "mcp_configs" bucket
  - MUST remove associated metadata
  - MUST return error if ID not found
  - MUST be idempotent (no error if already deleted)

### Initialize Database
**Operation**: InitDB(dbPath string) error
**Input**:
  - dbPath: Absolute path where database file should be created
**Output**:
  - error: File system or database errors
**Behavior**:
  - MUST create BoltDB file with secure permissions (0600)
  - MUST create "mcp_configs" bucket
  - MUST create "metadata" bucket
  - MUST handle existing database gracefully

### Close Database
**Operation**: CloseDB() error
**Input**: None
**Output**:
  - error: Database errors
**Behavior**:
  - MUST close all database connections
  - MUST flush any pending writes
  - MUST be safe to call multiple times

## Error Conditions
- `ErrDatabaseNotFound`: database file does not exist
- `ErrDatabaseCorrupted`: database file is corrupted
- `ErrConfigNotFound`: MCP configuration ID not found
- `ErrDuplicateID`: attempting to store config with existing ID
- `ErrInvalidConfig`: configuration fails validation
- `ErrPermissionDenied`: insufficient permissions for database operations