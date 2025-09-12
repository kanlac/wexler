# Implementation Plan

**Input**: Feature specification from `/specs/spec.md`

## Execution Flow
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (single project)
   → Set Structure Decision based on project type
3. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
4. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
5. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
6. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
7. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
8. STOP - Ready for tasks subagent
```

**IMPORTANT**: This planner subagent STOPS at step 7. Phases 2-4 are executed by other subagents:
- Phase 2: tasks subagent creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Wexler is a Go CLI tool for unified AI assistant configuration management. It provides commands (init, import, apply, list) to maintain a single source of configuration that syncs across multiple AI tools (Claude Code, Cursor), enabling seamless tool switching and consistent team coding standards. Core technical approach uses Go 1.21+, Cobra CLI framework, YAML configuration, and BoltDB for secure storage with base64 encoding.

## Technical Context
**Language/Version**: Go 1.21+  
**Primary Dependencies**: cobra (CLI), gopkg.in/yaml.v3 (YAML), go.etcd.io/bbolt (BoltDB), os (file operations)  
**Storage**: BoltDB for MCP configurations and API keys, YAML for project config, filesystem for source management  
**Testing**: Go testing package with table-driven tests, integration tests with temp directories  
**Target Platform**: Cross-platform (Linux, macOS, Windows)  
**Project Type**: single - CLI tool with libraries  
**Performance Goals**: <500ms command execution, <10MB memory usage, handle 100+ MCP configs  
**Constraints**: 8-hour hackathon timeline, base64 encoding (not AES), local filesystem only  
**Scale/Scope**: Individual/team usage, 5-50 users per config source, 10-100 MCP servers per project

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Simplicity**:
- Projects: 1 (cli tool)
- Using framework directly? Yes (cobra directly, no wrapper)
- Single data model? Yes (shared structs for MCP, project config)
- Avoiding patterns? Yes (no Repository/UoW, direct file ops)

**Architecture**:
- EVERY feature as library? Yes
- Libraries listed: 
  - config (project config management)
  - source (source directory operations) 
  - storage (BoltDB operations)
  - tools (AI tool adapters)
  - apply (configuration application logic)
- CLI per library: wexler with subcommands (init --help, apply --help, import --help, list --help)
- Library docs: llms.txt format planned? Yes

**Testing (NON-NEGOTIABLE)**:
- RED-GREEN-Refactor cycle enforced? Yes (tests written before implementation)
- Git commits show tests before implementation? Yes
- Order: Contract→Integration→E2E→Unit strictly followed? Yes
- Real dependencies used? Yes (actual filesystem, real BoltDB files)
- Integration tests for: new libraries, contract changes, shared schemas? Yes
- FORBIDDEN: Implementation before test, skipping RED phase

**Observability**:
- Structured logging included? Yes (log levels, operation context)
- Frontend logs → backend? N/A (CLI tool)
- Error context sufficient? Yes (file paths, operation details)

**Versioning**:
- Version number assigned? 1.0.0 (MAJOR.MINOR.BUILD)
- BUILD increments on every change? Yes
- Breaking changes handled? Yes (config format validation)

## Project Structure

### Documentation
```
specs/
├── plan.md              # This file (planner subagent output)
├── research.md          # Phase 0 output (planner subagent)
├── data-model.md        # Phase 1 output (planner subagent)
├── quickstart.md        # Phase 1 output (planner subagent)
├── contracts/           # Phase 1 output (planner subagent)
└── tasks.md             # Phase 2 output (tasks subagent - NOT created by planner)
```

### Source Code (repository root)
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

internal/               # Internal utilities
└── testutil/          # Test helpers
```

**Structure Decision**: Option 1 (single project) - CLI tool with supporting libraries

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - BoltDB best practices for CLI tools
   - Cobra CLI patterns for file operations
   - YAML validation strategies
   - Cross-platform file permission handling
   - Base64 encoding security considerations

2. **Generate and dispatch research agents**:
   ```
   For BoltDB usage:
     Task: "Research BoltDB best practices for CLI configuration storage"
   For Cobra patterns:
     Task: "Find Cobra CLI best practices for file system operations"
   For YAML handling:
     Task: "Research YAML validation and parsing strategies in Go"
   For cross-platform concerns:
     Task: "Research cross-platform file permission handling in Go"
   For security approach:
     Task: "Research base64 encoding limitations and alternatives for MVP"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all technical decisions documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - ProjectConfig (wexler.yaml structure) - NOTE: MVP supports single source, future versions will support multiple sources
   - MCPConfig (server configurations, API keys)
   - SourceConfig (memory.mdc, subagent definitions)
   - ConflictResolution (user choices during apply)
   - ApplyProgress (tracking applied changes)
   - ApplyResult (restart notifications for MCP changes)

2. **Generate API contracts** from functional requirements:
   - Config Library: Load/Save/Validate project config
   - Source Library: Read/Parse/List source configurations
   - Storage Library: Store/Retrieve/Update MCP configurations
   - Tools Library: Generate tool-specific config files
   - Apply Library: Merge configurations with conflict resolution
   - Output OpenAPI-style specs to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per library interface
   - Assert input validation and output formats
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each acceptance scenario → integration test
   - End-to-end workflow tests
   - Quickstart validation scenarios

5. **Update CLAUDE.md incrementally**:
   - Add Go CLI development context
   - Add BoltDB and YAML handling patterns
   - Add testing strategies for file operations
   - Preserve existing manual additions
   - Keep under 150 lines

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, CLAUDE.md

## Phase 2: Task Planning Approach
*This section describes what the tasks subagent will do - DO NOT execute during planner*

**Task Generation Strategy**:
- Load `/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each library contract → contract test task [P]
- Each entity → model creation task [P] 
- Each CLI command → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:
- TDD order: Contract tests → Integration tests → Unit tests → Implementation
- Dependency order: Models → Storage → Config → Source → Tools → Apply → CLI
- Mark [P] for parallel execution (independent libraries)

**Estimated Output**: 35-40 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the tasks subagent, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the planner subagent*

**Phase 3**: Task execution (tasks subagent creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Future Enhancement: Multi-Source Support
*Not implemented in MVP version*

The current MVP design supports a single Wexler source per project (specified in `wexler.yaml`). Future versions will support configuring multiple Wexler sources with priority ordering:

```yaml
# Future wexler.yaml structure
version: 2.0
sources:
  - path: /team/shared-configs
    priority: 1
  - path: ~/.wexler/personal-configs  
    priority: 2
tools:
  - claude
  - cursor
```

**Design Considerations**:
- Current `ProjectConfig.Source` field is easily extensible to `ProjectConfig.Sources []SourceConfig`
- Source library contracts already support multiple source operations
- Conflict resolution will need enhancement to handle source priority conflicts
- MCP configuration merging will require source priority consideration

## Complexity Tracking
*No constitutional violations identified - all requirements align with simplicity principles*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (planner subagent)
- [x] Phase 1: Design complete (planner subagent)
- [x] Phase 2: Task planning complete (planner subagent - describe approach only)
- [ ] Phase 3: Tasks generated (tasks subagent)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

## Hackathon-Specific Considerations

### Time Management (8 hours total)
- **Phase 1: Project Setup** (1.5 hours) - Go modules, Cobra setup, basic structure
- **Phase 2: Core Libraries** (3 hours) - config, source, storage libraries with tests
- **Phase 3: Tool Adapters** (2 hours) - Claude and Cursor configuration generation
- **Phase 4: CLI Integration** (1 hour) - Command wiring and error handling
- **Phase 5: Testing & Documentation** (0.5 hours) - E2E validation and README

### Risk Mitigation
- **Scope Reduction**: Focus on Claude Code + Cursor only (defer other tools)
- **Security Simplification**: Base64 encoding only (defer proper encryption)
- **Source Management**: Local filesystem only (defer Git integration)
- **Error Handling**: Basic error messages (defer detailed diagnostics)
- **Testing Strategy**: Focus on happy path + critical error cases

### MVP Priority Order
1. **Core Data Models** - Project config, MCP config structures
2. **Basic CLI Framework** - init, list commands (no file operations)
3. **Configuration Loading** - Read wexler.yaml, validate source directory
4. **MCP Storage** - BoltDB operations for import/apply
5. **Tool Generation** - Claude CLAUDE.md and Cursor .mdc file creation
6. **Conflict Resolution** - Basic diff display and user prompts
7. **Integration Testing** - End-to-end workflow validation

### Success Metrics
- **5-minute setup**: `wexler init` → `wexler apply` in under 5 minutes
- **Zero-config team onboarding**: Makefile integration works without manual steps  
- **Cross-tool consistency**: Same source generates equivalent configs for different tools
- **Safe operations**: No data loss during conflicts, clear error messages