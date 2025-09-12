# Config Library Contract

## Interface: ConfigManager

### Load Project Configuration
**Operation**: LoadProject(projectDir string) (ProjectConfig, error)
**Input**: 
  - projectDir: Absolute path to project directory containing wexler.yaml
**Output**: 
  - ProjectConfig: Parsed project configuration
  - error: Validation or file system errors
**Behavior**:
  - MUST validate wexler.yaml exists in projectDir
  - MUST validate source path accessibility
  - MUST validate tool names are supported
  - MUST return error if configuration invalid

### Save Project Configuration  
**Operation**: SaveProject(projectDir string, config ProjectConfig) error
**Input**:
  - projectDir: Absolute path to project directory
  - config: ProjectConfig to save
**Output**:
  - error: File system or validation errors
**Behavior**:
  - MUST write wexler.yaml with proper formatting
  - MUST validate configuration before writing
  - MUST use atomic file operations
  - MUST set appropriate file permissions

### Validate Project Configuration
**Operation**: ValidateProject(config ProjectConfig) error
**Input**:
  - config: ProjectConfig to validate
**Output**:
  - error: Validation errors or nil if valid
**Behavior**:
  - MUST check version compatibility
  - MUST validate source directory exists
  - MUST validate tools are supported
  - MUST check source directory accessibility

## Error Conditions
- `ErrProjectNotFound`: wexler.yaml not found in directory
- `ErrInvalidVersion`: unsupported configuration version
- `ErrSourceNotFound`: source directory does not exist  
- `ErrSourceNotAccessible`: source directory not readable
- `ErrInvalidTool`: unsupported tool name specified
- `ErrInvalidFormat`: malformed YAML configuration