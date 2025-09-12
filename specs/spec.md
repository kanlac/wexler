# Wexler - AI Configuration Management Tool Specification

## User Scenarios & Testing

### Primary User Story
As a developer using multiple AI coding assistants, I want to maintain a single source of configuration that automatically syncs across all my AI tools, so that I can switch between tools without reconfiguring and ensure my team uses consistent coding standards.

### Acceptance Scenarios
1. **Given** a new project with no AI configuration, **When** I run `wexler init --source=/path/to/configs`, **Then** a wexler.yaml file is created linking to my configuration source
2. **Given** existing AI tool configurations in my project, **When** I run `wexler import`, **Then** MCP configurations are extracted and stored in the central database
3. **Given** a Wexler source with team configurations, **When** I run `wexler apply`, **Then** all specified AI tools receive updated configuration files
4. **Given** sensitive API keys in my configuration, **When** stored in wexler.db, **Then** they are base64 encoded and file permissions are set to 0600
5. **Given** multiple team members using different AI tools, **When** they run `wexler apply` from the same source, **Then** each tool receives equivalent configuration in its native format
6. **Given** existing local configurations that differ from source, **When** I run `wexler apply`, **Then** the system shows differences, warns me, and prompts for confirmation before overwriting
7. **Given** multiple conflicts during apply, **When** user selects "Stop" at the third conflict, **Then** the first two applied changes remain while subsequent changes are skipped

### Edge Cases
- **Source directory doesn't exist**: When using default source (~/.wexler), init creates it with proper structure. When custom source parent directory missing, init fails
- **Apply without initialization**: System fails with error message guiding user to run `wexler init` first
- **Write permission denied**: System fails when unable to write to tool configuration directories
- **Corrupted source files**: System fails when encountering invalid configuration files
- **User declines conflict resolution**: System preserves all changes made before "Stop" was selected, exits without applying remaining changes

## Requirements

### Functional Requirements
- **FR-001**: System MUST provide a CLI interface for all configuration management operations
- **FR-002**: System MUST support initialization of Wexler in any project directory via `wexler init`
- **FR-003**: System MUST import existing AI tool configurations and extract MCP settings via `wexler import`
- **FR-004**: System MUST apply configurations from Wexler source to multiple AI tools via `wexler apply`
- **FR-005**: System MUST list all managed configurations via `wexler list`
- **FR-006**: System MUST store sensitive information (API keys) using base64 encoding for MVP
- **FR-007**: System MUST support local filesystem as configuration source
- **FR-008**: System MUST generate tool-specific configuration files in correct formats (CLAUDE.md for Claude, .mdc files for Cursor)
- **FR-009**: System MUST merge MCP configurations using upsert logic when applying to existing files
- **FR-010**: System MUST maintain file permissions of 0600 for sensitive data storage
- **FR-011**: System MUST identify Wexler-managed files with specific naming conventions (.wexler.md, .wexler.mdc)
- **FR-012**: System MUST support team collaboration through shared configuration sources
- **FR-013**: System MUST detect configuration conflicts one by one, display differences, and provide three options: "Continue", "Continue without asking", or "Stop"
- **FR-014**: System MUST validate source directory accessibility before operations
- **FR-015**: Users MUST be able to specify target tools for import and apply operations
- **FR-016**: System MUST preserve all changes made before user selects "Stop" during conflict resolution
- **FR-017**: System MUST create default source directory (~/.wexler) with proper structure if it doesn't exist during init
- **FR-018**: System MUST fail init operation if specified custom source parent directory doesn't exist
- **FR-019**: System MUST fail apply operation if no wexler.yaml exists in current directory
- **FR-020**: System MUST provide clear error guidance when operations fail due to missing prerequisites
- **FR-021**: System MUST fail operations when encountering corrupted or invalid configuration files
- **FR-022**: System MUST fail when lacking write permissions to target directories
- **FR-023**: System MUST skip remaining conflict prompts when user selects "Continue without asking"

### Key Entities
- **Configuration Source**: Central repository containing team-shared configurations, memory files, subagent/role definitions, and sensitive data storage
- **Project Configuration (wexler.yaml)**: Project-specific settings including source path and enabled AI tools
- **Memory Configuration**: General knowledge and coding standards shared across all AI interactions
- **Subagent/Role Configuration**: Specialized configurations for specific coding tasks (code review, testing, architecture)
- **MCP Configuration**: Model Context Protocol settings including server definitions, API endpoints, and authentication credentials
- **Tool Configuration**: Native configuration files for each AI tool (Claude, Cursor, etc.)
- **Conflict Resolution State**: Tracking mechanism for user choices during conflict resolution (continue, skip all, stop)
- **Apply Progress State**: Record of successfully applied changes during current operation

## Review & Acceptance Checklist

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Execution Status
- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---

## Summary of Resolution

All edge cases have been fully specified with the following key update:

**Partial Apply Behavior** (FR-016, FR-023): When user selects "Stop" during conflict resolution:
- All changes made before "Stop" are preserved (not rolled back)
- Remaining unapplied changes are skipped
- User can run `wexler apply` again to continue from where they stopped

This provides a safe, incremental approach to configuration updates where users maintain control over each change while preserving work already completed.

The specification is complete with all behaviors clearly defined and ready for implementation planning.