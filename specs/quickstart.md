# Quickstart Guide

## Overview
This guide provides step-by-step instructions to validate the Wexler AI Configuration Management Tool from installation to successful configuration synchronization across multiple AI tools.

## Prerequisites
- Go 1.21 or later installed
- Claude Code or Cursor AI tool installed
- Terminal/command line access
- Write permissions to project directory

## Installation

### Build from Source
```bash
# Clone repository
git clone https://github.com/your-org/wexler
cd wexler

# Build CLI tool
go build -o bin/wexler cmd/wexler/main.go

# Add to PATH (optional)
export PATH=$PATH:$(pwd)/bin
```

### Verify Installation
```bash
wexler --version
# Expected output: wexler version 1.0.0

wexler --help  
# Expected output: Usage information with available commands
```

## Scenario 1: Individual Developer Setup (5 minutes)

### Step 1: Create Test Project
```bash
# Create new project directory
mkdir test-wexler-project
cd test-wexler-project

# Create sample AI configuration (simulates existing setup)
mkdir -p .claude/agents
echo '{"mcpServers": {"test": {"type": "http", "url": "http://localhost:3000"}}}' > .mcp.json
echo "# Test Agent\nYou are a testing assistant." > .claude/agents/test-agent.md
```

### Step 2: Initialize Wexler
```bash
# Initialize with default source
wexler init

# Verify wexler.yaml created
cat wexler.yaml
# Expected output: Configuration with version, source path, and tools list
```

### Step 3: Import Existing Configuration
```bash
# Import current AI tool configurations
wexler import

# Verify import success
wexler list
# Expected output: Summary showing imported MCP configuration
```

### Step 4: Create Source Configuration
```bash
# Navigate to source directory (created during init)
cd ~/.wexler

# Create memory configuration
cat > memory.mdc << 'EOF'
# General Coding Standards

## Code Style
- Use meaningful variable names
- Add comments for complex logic
- Follow language-specific conventions

## Testing Requirements
- Write tests for new features
- Maintain test coverage above 80%
- Use descriptive test names
EOF

# Create subagent directory and configuration
mkdir -p subagent
cat > subagent/code-reviewer.mdc << 'EOF'
# Code Reviewer Assistant

## Role
You are a thorough code reviewer focused on:
- Code quality and maintainability
- Security vulnerabilities
- Performance optimization opportunities
- Best practices adherence

## Review Checklist
- [ ] Code follows established patterns
- [ ] Error handling is comprehensive
- [ ] Tests cover edge cases
- [ ] Documentation is up to date
EOF

# Return to project directory
cd ../test-wexler-project
```

### Step 5: Apply Configuration
```bash
# Apply source configuration to AI tools
wexler apply

# Handle any conflicts (follow prompts)
# Expected: User prompts for conflict resolution if files would be overwritten

# Verify configurations applied
ls -la .claude/agents/
# Expected: code-reviewer.wexler.md file created

cat CLAUDE.md | head -10
# Expected: Memory configuration content in CLAUDE.md
```

### Step 6: Validate Integration
```bash
# List all managed configurations
wexler list

# Expected output should show:
# - Source path and tools
# - Memory configuration file
# - Subagent configurations
# - MCP configurations
# - Applied tool configurations

# Verify Wexler-managed files have correct naming
find . -name "*.wexler.*" | sort
# Expected: Files with .wexler suffix indicating Wexler management
```

## Scenario 2: Team Collaboration Setup (10 minutes)

### Step 1: Create Shared Configuration Source
```bash
# Create team configuration repository
mkdir ../team-wexler-configs
cd ../team-wexler-configs

# Create comprehensive team configuration
cat > memory.mdc << 'EOF'
# Team Coding Standards

## Architecture Principles
- Follow Domain-Driven Design patterns
- Use dependency injection for testability
- Implement proper error handling and logging

## Code Review Process
- All code requires peer review
- Tests must pass before merge
- Documentation must be updated with changes

## Security Guidelines
- Validate all input data
- Use parameterized queries for database operations
- Follow OWASP security practices
EOF

# Create multiple subagent roles
mkdir -p subagent

cat > subagent/architect.mdc << 'EOF'
# Software Architect Assistant

## Focus Areas
- System design and architecture decisions
- Technology stack recommendations
- Performance and scalability analysis
- Technical debt assessment

## Design Principles
- SOLID principles adherence
- Microservices vs monolith trade-offs
- Database design and optimization
- API design best practices
EOF

cat > subagent/test-writer.mdc << 'EOF'
# Test Writing Assistant

## Testing Philosophy
- Test-driven development (TDD)
- Comprehensive test coverage
- Integration and unit test balance
- Behavior-driven development (BDD)

## Test Types
- Unit tests for individual components
- Integration tests for system interactions
- End-to-end tests for user workflows  
- Performance tests for critical paths
EOF

# Return to project directory
cd ../test-wexler-project
```

### Step 2: Initialize with Team Source
```bash
# Initialize project with team configuration
wexler init --source=../team-wexler-configs

# Verify project points to team source
cat wexler.yaml
# Expected: source field points to team configuration directory
```

### Step 3: Import Team Configuration
```bash
# Import any existing local configuration first
wexler import

# Apply team configuration
wexler apply

# Handle conflicts (team configuration should take precedence)
# Expected: Prompts for overwriting local configuration with team standards
```

### Step 4: Validate Team Integration
```bash
# List all configurations
wexler list

# Expected output should show:
# - Team source path
# - Multiple subagent configurations (architect, test-writer)
# - Memory configuration with team standards

# Verify team configurations applied
ls -la .claude/agents/*.wexler.md
# Expected: architect.wexler.md and test-writer.wexler.md files

# Check CLAUDE.md contains team standards
grep -A 5 "Architecture Principles" CLAUDE.md
# Expected: Team architecture principles in CLAUDE.md
```

## Scenario 3: Tool Migration Test (5 minutes)

### Step 1: Add Cursor Support
```bash
# Update project configuration to include Cursor
# (This would normally be done by editing wexler.yaml)
wexler init --source=../team-wexler-configs
# When prompted, select both Claude and Cursor tools
```

### Step 2: Apply to Multiple Tools
```bash
# Apply configuration to all enabled tools
wexler apply

# Verify Cursor configurations created
ls -la .cursor/rules/
# Expected: general.wexler.mdc and subagent-specific .wexler.mdc files

# Verify content consistency between tools
diff <(head -5 CLAUDE.md | tail -3) <(head -5 .cursor/rules/general.wexler.mdc | tail -3)
# Expected: Similar content structure adapted to tool format
```

### Step 3: Selective Tool Application
```bash
# Apply to specific tool only
wexler apply --tool=cursor

# Verify only Cursor files updated
# Check timestamps to confirm selective update

# Test dry-run functionality
wexler apply --dry-run
# Expected: Preview of changes without applying them
```

## Validation Checklist

### Basic Functionality
- [ ] wexler --version displays correct version
- [ ] wexler init creates wexler.yaml with valid configuration
- [ ] wexler import successfully extracts MCP configurations
- [ ] wexler apply generates tool-specific configuration files
- [ ] wexler list displays comprehensive configuration summary

### File Management
- [ ] Wexler-managed files have .wexler suffix
- [ ] Original user files are preserved during apply
- [ ] Configuration files have appropriate permissions (0600 for sensitive data)
- [ ] Source directory validation works correctly

### Conflict Resolution
- [ ] Conflicts detected when applying over existing configurations
- [ ] User presented with clear conflict information and options
- [ ] "Continue" option applies change and continues
- [ ] "Continue All" option applies all remaining changes without prompts
- [ ] "Stop" option preserves applied changes and exits

### Integration
- [ ] Claude Code recognizes generated CLAUDE.md
- [ ] Cursor recognizes generated .cursor/rules/ files
- [ ] MCP configurations properly merged using upsert logic
- [ ] Team collaboration works with shared source directory

### Error Handling
- [ ] Meaningful error messages for common failures
- [ ] Graceful handling of missing files or directories
- [ ] Appropriate exit codes returned
- [ ] Help information accessible with --help flag

## Expected Results

### Successful Completion Indicators
1. **Configuration Consistency**: Same source generates equivalent configurations for different AI tools
2. **File Organization**: Clear separation between user-managed and Wexler-managed files
3. **Conflict Safety**: No data loss during configuration updates
4. **Team Integration**: Multiple developers can sync from shared source
5. **Tool Independence**: Easy switching between AI tools with consistent configuration

### Performance Metrics
- Commands complete in under 2 seconds for typical configurations
- Memory usage remains under 50MB during operations
- Database operations handle 100+ MCP server configurations efficiently

This quickstart guide validates all core functionality and ensures Wexler meets the success criteria of 5-minute setup and seamless team integration.