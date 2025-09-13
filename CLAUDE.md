# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Wexler** is an AI Configuration Management Tool that unifies AI assistant configurations across multiple tools (Claude Code, Cursor). It maintains a single source of configuration truth that syncs across different AI tools, preventing configuration fragmentation and ensuring team consistency.

## Current Status

This project is in the **specification and planning phase**. All implementation artifacts are specifications located in the `specs/` directory. No Go code has been written yet.

## Core CLI Commands (To Be Implemented)

```bash
# Initialize Wexler in project directory
wexler init [--source=/path/to/configs]

# Import existing AI tool configurations to central storage
wexler import [--tool=claude|cursor] [--dry-run]

# Apply configurations from source to AI tools
wexler apply [--tool=claude|cursor] [--force] [--dry-run]

# List all managed configurations
wexler list [--mcp] [--tools] [--format=table|json|yaml]
```

## Architecture

### Technology Stack
- **Go 1.21+** with Cobra CLI framework
- **BoltDB** for MCP configurations and API keys (base64 encoded)
- **YAML** for project configuration files
- **File system** operations for source directory management

### Library-First Architecture
Every feature is implemented as a library with CLI wrapper:

- **config/**: Project configuration management (wexler.yaml)
- **source/**: Source directory operations (memory.mdc, subagent/*.mdc)
- **storage/**: BoltDB operations for MCP configurations
- **tools/**: AI tool adapters (Claude Code, Cursor)
- **apply/**: Configuration application with conflict resolution

### Project Structure (Planned)
```
src/
├── models/              # Data structures
├── config/              # Project configuration library
├── source/              # Source directory management library  
├── storage/             # BoltDB operations library
├── tools/               # AI tool adapter library
├── apply/               # Configuration application library
└── cli/                 # CLI commands

tests/
├── contract/            # API contract tests
├── integration/         # Cross-library integration tests
└── unit/                # Unit tests per library

cmd/
└── wexler/             # Main CLI entry point
```

### Data Flow

**Memory vs Subagent Configuration Handling:**
- **Memory configurations** (memory.mdc): Parsed by markdown sections, merged into existing tool configurations using section-based approach
- **Subagent configurations** (subagent/*.mdc): Applied as entire file replacements, no section parsing

**MCP Configuration Storage:**
- Stored as `map[string]string` where key = server name, value = base64 encoded JSON
- No need to parse mcpServers JSON structure - entire server blocks are base64 encoded
- Example: `mcpServers.context7` entire JSON object → base64 encode → store in `Servers["context7"]`

## Key Implementation Requirements

### Conflict Resolution (Critical UX Feature)
- **Three-option system**: Continue, Continue All, Stop
- **Partial apply behavior**: Changes made before "Stop" are preserved
- **Progressive conflict handling**: Display differences, get user choice, continue or halt

### File Generation Patterns
- **Claude Code**: CLAUDE.md (section-based), .claude/agents/*.wexler.md (full file), .mcp.json
- **Cursor**: .cursor/rules/general.wexler.mdc, .cursor/rules/{subagent}.wexler.mdc (full file), .cursor/mcp.json
- **Wexler-managed files**: Use .wexler extensions to identify managed content

### Security Requirements
- Base64 encoding for sensitive data (MVP approach)
- File permissions 0600 for sensitive storage
- BoltDB for secure MCP configuration storage

## Testing Strategy

### Test-Driven Development (Enforced)
1. **Contract tests** → **Integration tests** → **Unit tests** → **Implementation**
2. Use real dependencies (actual filesystem, real BoltDB files)
3. RED-GREEN-Refactor cycle strictly followed
4. No implementation before failing tests

### Validation Scenarios
- 5-minute setup: `wexler init` → `wexler apply` workflow
- Cross-tool consistency verification
- Conflict resolution workflows
- Team configuration sharing

## Development Commands

### Build and Test (When Implemented)
```bash
# Build CLI tool
go build -o bin/wexler cmd/wexler/main.go

# Run tests (TDD approach)
go test ./tests/contract/...    # Contract tests first
go test ./tests/integration/... # Integration tests  
go test ./tests/unit/...        # Unit tests last

# Run single test
go test -run TestConfigLibraryLoadProject ./tests/unit/config/

# Performance validation
go test -bench=. ./tests/performance/
```

## Code Style Notes

- 新文件命名: 若在 service 包下创建用户服务，文件名直接使用 `user.go`，而不是 `user_service.go`
- Direct framework usage (cobra directly, no wrapper patterns)
- Table-driven tests following Go conventions
- Structured logging with operation context

## Hackathon Timeline Constraints

**8-hour implementation window:**
- Phase 1: Project Setup (1.5h) - Go modules, Cobra setup, basic structure
- Phase 2: Core Libraries (3h) - config, source, storage libraries with tests
- Phase 3: Tool Adapters (2h) - Claude and Cursor configuration generation  
- Phase 4: CLI Integration (1h) - Command wiring and error handling
- Phase 5: Testing & Documentation (0.5h) - E2E validation

Priority: Focus on Claude Code + Cursor only, defer other tools for post-MVP.