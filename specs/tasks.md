# Wexler Implementation Tasks

**Input**: Design documents from `/specs/`
**Prerequisites**: plan.md (complete), research.md (complete), data-model.md (complete), contracts/ (complete)

Generated from implementation plan with 8-hour hackathon timeline. Tasks are numbered sequentially and marked [P] for parallel execution when applicable.

## Phase 3.1: Setup

- [ ] T001 Create Go project structure with cmd/wexler/, src/, tests/, internal/
- [ ] T002 Initialize Go module and add dependencies: cobra, gopkg.in/yaml.v3, go.etcd.io/bbolt
- [ ] T003 [P] Configure linting with golangci-lint and formatting with gofmt

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests
- [x] T004 [P] Contract test ConfigManager.LoadProject in tests/contract/config_test.go
- [x] T005 [P] Contract test ConfigManager.SaveProject in tests/contract/config_test.go
- [x] T006 [P] Contract test ConfigManager.ValidateProject in tests/contract/config_test.go
- [x] T007 [P] Contract test SourceManager.LoadSource in tests/contract/source_test.go
- [x] T008 [P] Contract test SourceManager.ListSourceFiles in tests/contract/source_test.go
- [x] T009 [P] Contract test SourceManager.ParseMemory in tests/contract/source_test.go
- [x] T010 [P] Contract test SourceManager.ParseSubagent in tests/contract/source_test.go
- [x] T011 [P] Contract test StorageManager.StoreMCP in tests/contract/storage_test.go
- [x] T012 [P] Contract test StorageManager.RetrieveMCP in tests/contract/storage_test.go
- [x] T013 [P] Contract test StorageManager.ListMCP in tests/contract/storage_test.go
- [x] T014 [P] Contract test ToolAdapter.Generate in tests/contract/tools_test.go
- [x] T015 [P] Contract test ToolAdapter.Validate in tests/contract/tools_test.go
- [x] T016 [P] Contract test ToolAdapter.Merge in tests/contract/tools_test.go
- [x] T017 [P] Contract test ApplyManager.ApplyConfig in tests/contract/apply_test.go
- [x] T018 [P] Contract test ApplyManager.DetectConflicts in tests/contract/apply_test.go

### Integration Tests
- [ ] T019 [P] Integration test init command workflow in tests/integration/init_test.go
- [ ] T020 [P] Integration test import command workflow in tests/integration/import_test.go
- [ ] T021 [P] Integration test apply command workflow in tests/integration/apply_test.go
- [ ] T022 [P] Integration test list command workflow in tests/integration/list_test.go
- [ ] T023 [P] Integration test team collaboration scenario in tests/integration/team_test.go
- [ ] T024 [P] Integration test conflict resolution workflow in tests/integration/conflict_test.go
- [ ] T025 [P] Integration test BoltDB operations in tests/integration/storage_integration_test.go
- [ ] T026 [P] Integration test file system operations in tests/integration/filesystem_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Data Models
- [x] T027 [P] ProjectConfig struct and validation in src/models/project.go
- [x] T028 [P] MCPConfig struct in src/models/mcp.go
- [x] T029 [P] SourceConfig, MemoryConfig, SubagentConfig structs in src/models/source.go
- [x] T030 [P] ConflictResolution and ApplyProgress structs in src/models/apply.go
- [x] T031 [P] ToolConfig interface and ConfigFile struct in src/models/tool.go

### Config Library
- [x] T032 ConfigManager implementation in src/config/manager.go
- [x] T033 Project configuration validation logic in src/config/validation.go
- [x] T034 YAML loading and saving operations in src/config/yaml.go

### Source Library  
- [x] T035 SourceManager implementation in src/source/manager.go
- [x] T036 Markdown parsing for memory and subagent configs in src/source/parser.go
- [x] T037 Source directory validation and scanning in src/source/scanner.go

### Storage Library
- [x] T038 StorageManager implementation with BoltDB in src/storage/manager.go
- [x] T039 Base64 encoding/decoding for secrets in src/storage/encoding.go
- [x] T040 Database initialization and bucket management in src/storage/database.go

### Tools Library
- [x] T041 [P] ClaudeAdapter implementation in src/tools/claude.go
- [x] T042 [P] CursorAdapter implementation in src/tools/cursor.go
- [x] T043 ToolAdapter factory function in src/tools/factory.go
- [x] T044 Configuration merging utilities in src/tools/merge.go

### Apply Library
- [x] T045 ApplyManager implementation in src/apply/manager.go
- [x] T046 Conflict detection and resolution in src/apply/conflicts.go
- [x] T047 Progress tracking during apply operations in src/apply/progress.go

## Phase 3.4: CLI Integration

- [ ] T048 Root command setup with cobra in cmd/wexler/main.go
- [ ] T049 Init command implementation in src/cli/init.go
- [ ] T050 Import command implementation in src/cli/import.go
- [ ] T051 Apply command implementation in src/cli/apply.go
- [ ] T052 List command implementation in src/cli/list.go
- [ ] T053 Global flags and error handling in src/cli/root.go
- [ ] T054 User interaction utilities for conflict resolution in src/cli/interact.go

## Phase 3.5: Integration & Polish

### Cross-library Integration
- [ ] T055 Wire all libraries together in main CLI application
- [ ] T056 Add structured logging with context in src/internal/logger.go
- [ ] T057 Implement atomic file operations in src/internal/fileutil.go
- [ ] T058 Add cross-platform permission handling in src/internal/permissions.go

### Testing & Validation
- [ ] T059 [P] Unit tests for configuration validation in tests/unit/validation_test.go
- [ ] T060 [P] Unit tests for markdown parsing in tests/unit/parser_test.go
- [ ] T061 [P] Unit tests for base64 encoding in tests/unit/encoding_test.go
- [ ] T062 [P] Unit tests for conflict detection in tests/unit/conflict_test.go
- [ ] T063 Performance tests ensuring <500ms command execution in tests/performance/
- [ ] T064 End-to-end test following quickstart scenarios in tests/e2e/quickstart_test.go

### Documentation & Polish
- [ ] T065 [P] Add comprehensive error messages and help text
- [ ] T066 [P] Create llms.txt documentation for libraries in docs/
- [ ] T067 Remove code duplication and refactor common patterns
- [ ] T068 Final testing with quickstart.md validation

## Dependencies

### Critical Dependencies (Must Complete Before Others)
- **Setup (T001-T003)** before all other tasks
- **Contract Tests (T004-T018)** before corresponding implementations
- **Integration Tests (T019-T026)** before implementation tasks
- **Models (T027-T031)** before all library implementations

### Library Dependencies
- **T027-T031 (Models)** → **T032-T047 (Library Implementations)**
- **T032-T034 (Config)** → **T048-T054 (CLI)**
- **T035-T037 (Source)** → **T045-T047 (Apply)**
- **T038-T040 (Storage)** → **T045-T047 (Apply)**
- **T041-T044 (Tools)** → **T045-T047 (Apply)**
- **T045-T047 (Apply)** → **T048-T054 (CLI)**

### Polish Dependencies
- **T048-T054 (CLI)** → **T055-T058 (Integration)**
- **T055-T058 (Integration)** → **T059-T068 (Testing & Polish)**

## Parallel Execution Examples

### Phase 3.2: Contract Tests (Launch together)
```bash
# All contract tests can run in parallel - different files, no dependencies
Task T004: "Contract test ConfigManager.LoadProject in tests/contract/config_test.go"
Task T005: "Contract test ConfigManager.SaveProject in tests/contract/config_test.go"  
Task T007: "Contract test SourceManager.LoadSource in tests/contract/source_test.go"
Task T011: "Contract test StorageManager.StoreMCP in tests/contract/storage_test.go"
Task T014: "Contract test ToolAdapter.Generate in tests/contract/tools_test.go"
Task T017: "Contract test ApplyManager.ApplyConfig in tests/contract/apply_test.go"
```

### Phase 3.2: Integration Tests (Launch together)
```bash
# Integration tests can run in parallel - different test scenarios
Task T019: "Integration test init command workflow in tests/integration/init_test.go"
Task T020: "Integration test import command workflow in tests/integration/import_test.go"
Task T023: "Integration test team collaboration scenario in tests/integration/team_test.go"
Task T025: "Integration test BoltDB operations in tests/integration/storage_integration_test.go"
```

### Phase 3.3: Models (Launch together)
```bash
# All model definitions can be created in parallel - different files
Task T027: "ProjectConfig struct and validation in src/models/project.go"
Task T028: "MCPConfig struct in src/models/mcp.go" 
Task T029: "SourceConfig, MemoryConfig, SubagentConfig structs in src/models/source.go"
Task T030: "ConflictResolution and ApplyProgress structs in src/models/apply.go"
Task T031: "ToolConfig interface and ConfigFile struct in src/models/tool.go"
```

### Phase 3.3: Tool Adapters (Launch together)
```bash
# Tool adapters can be implemented in parallel - different tools
Task T041: "ClaudeAdapter implementation in src/tools/claude.go"
Task T042: "CursorAdapter implementation in src/tools/cursor.go"
```

## Validation Checklist

*GATE: Checked before marking tasks complete*

### Contract Completeness
- [x] All ConfigManager operations have contract tests (T004-T006)
- [x] All SourceManager operations have contract tests (T007-T010) 
- [x] All StorageManager operations have contract tests (T011-T013)
- [x] All ToolAdapter operations have contract tests (T014-T016)
- [x] All ApplyManager operations have contract tests (T017-T018)

### Entity Completeness  
- [x] All data model entities have creation tasks (T027-T031)
- [x] ProjectConfig entity has model task (T027)
- [x] MCPConfig entity has model task (T028)
- [x] SourceConfig entity has model task (T029)
- [x] ConflictResolution entity has model task (T030)
- [x] ToolConfig entity has model task (T031)

### Integration Completeness
- [x] All CLI commands have integration tests (T019-T022)
- [x] Team collaboration has integration test (T023)
- [x] Conflict resolution has integration test (T024)
- [x] Database operations have integration test (T025)
- [x] File operations have integration test (T026)

### Implementation Completeness
- [x] All contracts have corresponding implementations
- [x] All tests come before implementations (TDD enforced)
- [x] Parallel tasks are truly independent (different files/no dependencies)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task

### Hackathon Constraints
- [x] Tasks prioritized for 8-hour timeline
- [x] Base64 encoding approach (not AES) documented
- [x] BoltDB single-file approach validated
- [x] Cross-platform considerations included

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **Verify tests fail** before implementing (RED phase of TDD)
- **Commit after each task** for progress tracking
- **Focus on MVP**: Claude Code + Cursor support, base64 encoding, local filesystem
- **Performance target**: <500ms command execution, <10MB memory usage
- **Security approach**: File permissions (0600) + base64 encoding for MVP

## Task Generation Summary

**Total Tasks**: 68
**Parallel Tasks**: 32 (marked with [P])
**Sequential Tasks**: 36 (dependencies require order)

**Estimated Timeline**:
- Setup: 1.5 hours (T001-T003)
- Contract Tests: 2.5 hours (T004-T026)  
- Core Implementation: 2.5 hours (T027-T047)
- CLI Integration: 1 hour (T048-T054)
- Polish & Testing: 0.5 hours (T055-T068)

Tasks follow constitutional principles: TDD enforced, libraries before CLI, real dependencies in tests, and comprehensive error handling throughout.