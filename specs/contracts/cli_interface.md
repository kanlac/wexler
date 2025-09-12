# CLI Interface Contract

## Root Command: wexler

### Global Flags
- `--verbose, -v`: Enable verbose logging output
- `--tool`: Specify specific AI tool for operation (optional, defaults to all configured tools)
- `--help, -h`: Display help information
- `--version`: Display version information

### Global Behavior
- MUST validate project directory contains wexler.yaml before operations (except init)
- MUST provide consistent error messages across all commands
- MUST support --help flag for all commands and subcommands
- MUST exit with appropriate exit codes (0 = success, 1 = error)

## Subcommand: init

### Usage
```bash
wexler init [flags]
```

### Flags
- `--source`: Specify custom source directory path (optional, defaults to ~/.wexler)

### Behavior
- MUST create wexler.yaml in current directory
- MUST validate source directory exists or create default if not specified
- MUST prompt for tool selection if not specified
- MUST fail if wexler.yaml already exists (unless --force flag added later)
- MUST create source directory with proper structure if using default

### Output Format
```
Initializing Wexler in current directory...
✓ Created wexler.yaml
✓ Source directory verified: /path/to/source
✓ Configuration ready for tools: claude, cursor
```

### Error Conditions
- Directory already contains wexler.yaml
- Source directory path invalid or inaccessible
- Insufficient permissions to create files

## Subcommand: import

### Usage
```bash
wexler import [flags]
```

### Flags
- `--tool`: Import from specific tool only (optional)
- `--dry-run`: Show what would be imported without making changes

### Behavior
- MUST scan current directory for AI tool configurations
- MUST extract MCP configurations from found tools
- MUST store configurations in BoltDB with base64 encoding
- MUST report what was imported and from which tools

### Output Format
```
Importing configurations...
✓ Found Claude configuration at .mcp.json
  - Imported 3 MCP servers
✓ Found Cursor configuration at .cursor/rules/
  - Imported 2 rule files
✓ Stored configurations in database
```

### Error Conditions
- No AI tool configurations found
- Invalid or corrupted configuration files
- Database write errors
- Insufficient permissions

## Subcommand: apply

### Usage
```bash
wexler apply [flags]
```

### Flags
- `--tool`: Apply to specific tool only (optional)
- `--dry-run`: Show what would be applied without making changes
- `--force`: Skip conflict resolution and overwrite all files

### Behavior
- MUST load project configuration and source configuration
- MUST detect conflicts with existing tool configurations
- MUST prompt user for conflict resolution (unless --force)
- MUST apply configurations to all enabled tools
- MUST provide progress feedback during application

### Output Format
```
Applying configurations...
⚠ Conflict detected in CLAUDE.md (lines 15-20)
  Would replace existing content with memory configuration
  Continue? [y/N/a/s] (y=yes, N=no, a=all, s=stop): y
✓ Applied memory configuration to Claude Code
✓ Applied subagent configurations: code-reviewer, architect  
✓ Updated MCP configuration with 3 servers
✓ All configurations applied successfully
```

### Error Conditions
- Conflicts detected and user chooses to stop
- Write permission denied for tool directories
- Source configuration invalid or missing
- Tool-specific configuration errors

## Subcommand: list

### Usage
```bash
wexler list [flags]
```

### Flags
- `--mcp`: Show only MCP configurations
- `--tools`: Show only tool configurations
- `--format`: Output format (table, json, yaml)

### Behavior
- MUST display all managed configurations
- MUST show source files and their status
- MUST show MCP configurations from database
- MUST show which tools are configured

### Output Format
```
Wexler Configuration Summary
Source: /Users/alice/team-wexler-configs
Tools: claude, cursor

Memory Configuration:
  ✓ memory.mdc (1.2kb)

Subagent Configurations:
  ✓ code-reviewer.mdc (0.8kb)
  ✓ test-writer.mdc (1.1kb) 
  ✓ architect.mdc (1.5kb)

MCP Configurations:
  ✓ context7 (http server)
  ✓ filesystem (command server)
  ✓ sqlite (command server)

Tool Configurations:
  ✓ Claude Code: CLAUDE.md, .claude/agents/, .mcp.json
  ✓ Cursor: .cursor/rules/
```

### Error Conditions
- Project not initialized (no wexler.yaml)
- Source directory not accessible
- Database read errors

## Global Error Handling

### Exit Codes
- 0: Success
- 1: General error
- 2: Misuse of command (invalid flags/arguments)  
- 3: Permission denied
- 4: File not found
- 5: Configuration invalid

### Error Message Format
```
Error: [operation] failed: [specific error]
Suggestion: [actionable advice]
```

### Examples
```
Error: apply failed: source directory not found
Suggestion: Run 'wexler init --source=/path/to/configs' to set up source

Error: import failed: no AI tool configurations found  
Suggestion: Ensure Claude or Cursor configurations exist in current directory
```