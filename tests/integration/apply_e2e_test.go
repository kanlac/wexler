package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mindful/src/apply"
	"mindful/src/config"
	"mindful/src/models"
	"mindful/src/source"
	"mindful/src/tools/claude"
	"mindful/src/tools/cursor"
)

func TestMindfulApplyE2E_DualScopeMemory(t *testing.T) {
	// Setup complete test environment
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team-configs")
	projectDir := filepath.Join(tempDir, "my-project")
	projectMindfulDir := filepath.Join(projectDir, "mindful")

	// 1. Create team source directory structure (simulating team-shared configs)
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(teamSourceDir, "subagent"), 0755); err != nil {
		t.Fatalf("Failed to create team subagent directory: %v", err)
	}

	// Create team memory file
	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamMemoryContent := `# Team Development Standards

## Code Guidelines
- Always use Go conventions
- Write comprehensive tests
- Follow TDD methodology

## Review Process
- All code must be peer reviewed
- Use semantic commit messages`

	if err := os.WriteFile(teamMemoryPath, []byte(teamMemoryContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// 2. Create project directory structure (simulating user project)
	if err := os.MkdirAll(projectMindfulDir, 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	// Create project memory file
	projectMemoryPath := filepath.Join(projectMindfulDir, "memory.mdc")
	projectMemoryContent := `# E2E Test Project Configuration

## Database Setup
This project uses PostgreSQL with migrations in db/migrations/

## Local Development Environment
- Run with: make dev
- Test with: make test
- Build with: make build

## API Endpoints
REST API with /api/v1 prefix`

	if err := os.WriteFile(projectMemoryPath, []byte(projectMemoryContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Create project configuration file (mindful.yaml)
	projectConfig := &models.ProjectConfig{
		Name:       "e2e-test-project",
		Version:    "1.0.0",
		SourcePath: teamSourceDir,
		Tools: map[string]string{
			"claude": "enabled",
			"cursor": "enabled",
		},
	}

	configManager := config.NewManager()
	// Temporarily change to project directory for saving config
	originalDir, _ := os.Getwd()
	os.Chdir(projectDir)
	if err := configManager.SaveProject(projectConfig); err != nil {
		t.Fatalf("Failed to save project config: %v", err)
	}
	os.Chdir(originalDir)

	// 3. Test LoadDualScopeSource directly
	sourceManager := source.NewManager()
	sourceConfig, err := sourceManager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify dual scope memory was loaded
	if sourceConfig.Memory == nil {
		t.Fatal("Memory config is nil")
	}
	if !sourceConfig.Memory.HasTeam {
		t.Error("Expected HasTeam to be true")
	}
	if !sourceConfig.Memory.HasProject {
		t.Error("Expected HasProject to be true")
	}
	if !strings.Contains(sourceConfig.Memory.TeamContent, "Team Development Standards") {
		t.Error("Team content not loaded correctly")
	}
	if !strings.Contains(sourceConfig.Memory.ProjectContent, "E2E Test Project Configuration") {
		t.Error("Project content not loaded correctly")
	}

	// 4. Test Claude memory generation with dual scope
	claudeMemory := claude.GenerateClaudeMemoryContent(sourceConfig.Memory)

	// Verify Claude memory contains both scopes
	if !strings.Contains(claudeMemory, "Team Development Standards -- Mindful (scope:team)") {
		t.Error("Claude memory missing team scope header")
	}
	if !strings.Contains(claudeMemory, "E2E Test Project Configuration -- Mindful (scope:project)") {
		t.Error("Claude memory missing project scope header")
	}
	if !strings.Contains(claudeMemory, "Always use Go conventions") {
		t.Error("Claude memory missing team content")
	}
	if !strings.Contains(claudeMemory, "Run with: make dev") {
		t.Error("Claude memory missing project content")
	}
	if !strings.Contains(claudeMemory, "<!-- Source: "+teamMemoryPath+" -->") {
		t.Error("Claude memory missing team source comment")
	}
	if !strings.Contains(claudeMemory, "<!-- Source: "+projectMemoryPath+" -->") {
		t.Error("Claude memory missing project source comment")
	}

	// 5. Test Cursor memory generation with dual scope
	cursorMemory := cursor.GenerateCursorMemoryContent(sourceConfig.Memory)

	// Verify Cursor memory contains both scopes
	if !strings.Contains(cursorMemory, "Team Development Standards -- Mindful (scope:team)") {
		t.Error("Cursor memory missing team scope header")
	}
	if !strings.Contains(cursorMemory, "E2E Test Project Configuration -- Mindful (scope:project)") {
		t.Error("Cursor memory missing project scope header")
	}
	if !strings.Contains(cursorMemory, "Write comprehensive tests") {
		t.Error("Cursor memory missing team content")
	}
	if !strings.Contains(cursorMemory, "PostgreSQL with migrations") {
		t.Error("Cursor memory missing project content")
	}
	if !strings.HasPrefix(cursorMemory, "---\n") {
		t.Error("Cursor memory missing frontmatter")
	}

	// 6. Test Apply Manager functionality
	applyManager := apply.NewManager()

	// Create MCP config (empty for this test)
	mcpConfig := models.NewMCPConfig()

	// Test Claude apply configuration
	claudeApplyConfig := &models.ApplyConfig{
		ProjectPath: projectDir,
		ToolName:    "claude",
		Source:      sourceConfig,
		MCP:         mcpConfig,
		DryRun:      true, // Use dry run to avoid actual file writes
		Force:       false,
	}

	// Test conflict detection (should be none for new project)
	conflicts, err := applyManager.DetectConflicts(claudeApplyConfig)
	if err != nil {
		t.Fatalf("Failed to detect conflicts: %v", err)
	}
	if len(conflicts) > 0 {
		t.Logf("Found %d conflicts (expected for dry run): %v", len(conflicts), conflicts)
	}

	// Test apply (dry run)
	result, err := applyManager.ApplyConfig(claudeApplyConfig)
	if err != nil {
		t.Fatalf("Failed to apply Claude config (dry run): %v", err)
	}
	if !result.Success {
		t.Errorf("Apply result not successful: %s", result.Error)
	}

	// Verify proper dual-scope structure in generated content
	t.Logf("=== Generated Claude Memory Content ===\n%s", claudeMemory)
	t.Logf("=== Generated Cursor Memory Content ===\n%s", cursorMemory)

	// Test scope ordering (team scope should come before project scope)
	teamIndex := strings.Index(claudeMemory, "scope:team")
	projectIndex := strings.Index(claudeMemory, "scope:project")
	if teamIndex == -1 || projectIndex == -1 {
		t.Error("Missing scope indicators")
	}
	if teamIndex >= projectIndex {
		t.Error("Team scope should appear before project scope")
	}
}

func TestMindfulApplyE2E_OnlyProjectScope(t *testing.T) {
	// Test case where only project memory exists, no team source
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project-only")
	projectMindfulDir := filepath.Join(projectDir, "mindful")

	// Create only project memory
	if err := os.MkdirAll(projectMindfulDir, 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	projectMemoryPath := filepath.Join(projectMindfulDir, "memory.mdc")
	projectMemoryContent := `# Standalone Project

## Independent Development
This project doesn't rely on team-wide configurations.`

	if err := os.WriteFile(projectMemoryPath, []byte(projectMemoryContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Load with empty team source
	sourceManager := source.NewManager()
	sourceConfig, err := sourceManager.LoadDualScopeSource("", projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Should have only project scope
	if sourceConfig.Memory == nil {
		t.Fatal("Memory config is nil")
	}
	if sourceConfig.Memory.HasTeam {
		t.Error("Expected HasTeam to be false")
	}
	if !sourceConfig.Memory.HasProject {
		t.Error("Expected HasProject to be true")
	}

	// Test generated content
	claudeMemory := claude.GenerateClaudeMemoryContent(sourceConfig.Memory)

	if strings.Contains(claudeMemory, "scope:team") {
		t.Error("Should not contain team scope when no team memory exists")
	}
	if !strings.Contains(claudeMemory, "Standalone Project -- Mindful (scope:project)") {
		t.Error("Missing project scope header")
	}
}

func TestMindfulApplyE2E_OnlyTeamScope(t *testing.T) {
	// Test case where only team source exists, no project memory
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team-only")
	projectDir := filepath.Join(tempDir, "empty-project")

	// Create only team memory
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create empty project directory: %v", err)
	}

	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamMemoryContent := `# Company-wide Standards

## Security Guidelines
All projects must follow these security practices.`

	if err := os.WriteFile(teamMemoryPath, []byte(teamMemoryContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Load dual scope
	sourceManager := source.NewManager()
	sourceConfig, err := sourceManager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Should have only team scope
	if sourceConfig.Memory == nil {
		t.Fatal("Memory config is nil")
	}
	if !sourceConfig.Memory.HasTeam {
		t.Error("Expected HasTeam to be true")
	}
	if sourceConfig.Memory.HasProject {
		t.Error("Expected HasProject to be false")
	}

	// Test generated content
	claudeMemory := claude.GenerateClaudeMemoryContent(sourceConfig.Memory)

	if strings.Contains(claudeMemory, "scope:project") {
		t.Error("Should not contain project scope when no project memory exists")
	}
	if !strings.Contains(claudeMemory, "Company-wide Standards -- Mindful (scope:team)") {
		t.Error("Missing team scope header")
	}
}