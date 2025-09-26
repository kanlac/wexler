package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mindful/src/tools/claude"
	"mindful/src/tools/cursor"
	"mindful/src/source"
)

func TestApplyDualScopeMemory(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create team source directory and memory file
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}

	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamContent := `# Team Development Guidelines

## Code Standards
Use Go conventions across all projects.

## Workflow
Prefer TDD approach for all development.`
	if err := os.WriteFile(teamMemoryPath, []byte(teamContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Create project directory and memory file
	projectMindfulDir := filepath.Join(projectDir, "mindful")
	if err := os.MkdirAll(projectMindfulDir, 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	projectMemoryPath := filepath.Join(projectMindfulDir, "memory.mdc")
	projectContent := `# Project Configuration

## Local Development
Run with make dev for this specific project.

## Database
Uses PostgreSQL with specific schema requirements.`
	if err := os.WriteFile(projectMemoryPath, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Load dual scope source configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(config.Memory)

	// Verify Claude content has both scopes
	if !strings.Contains(claudeContent, "Team Development Guidelines -- Mindful (scope:team)") {
		t.Error("Claude content missing team scope section")
	}
	if !strings.Contains(claudeContent, "Project Configuration -- Mindful (scope:project)") {
		t.Error("Claude content missing project scope section")
	}
	if !strings.Contains(claudeContent, "Use Go conventions across all projects") {
		t.Error("Claude content missing team content")
	}
	if !strings.Contains(claudeContent, "Run with make dev for this specific project") {
		t.Error("Claude content missing project content")
	}
	if !strings.Contains(claudeContent, "<!-- Source: "+teamMemoryPath+" -->") {
		t.Error("Claude content missing team source comment")
	}
	if !strings.Contains(claudeContent, "<!-- Source: "+projectMemoryPath+" -->") {
		t.Error("Claude content missing project source comment")
	}

	// Test Cursor memory generation
	cursorContent := cursor.GenerateCursorMemoryContent(config.Memory)

	// Verify Cursor content has both scopes
	if !strings.Contains(cursorContent, "Team Development Guidelines -- Mindful (scope:team)") {
		t.Error("Cursor content missing team scope section")
	}
	if !strings.Contains(cursorContent, "Project Configuration -- Mindful (scope:project)") {
		t.Error("Cursor content missing project scope section")
	}
	if !strings.Contains(cursorContent, "Use Go conventions across all projects") {
		t.Error("Cursor content missing team content")
	}
	if !strings.Contains(cursorContent, "Run with make dev for this specific project") {
		t.Error("Cursor content missing project content")
	}
	if !strings.HasPrefix(cursorContent, "---\n") {
		t.Error("Cursor content missing frontmatter")
	}

	// Verify structure: team scope should come before project scope
	teamIndex := strings.Index(claudeContent, "scope:team")
	projectIndex := strings.Index(claudeContent, "scope:project")
	if teamIndex == -1 || projectIndex == -1 {
		t.Error("Missing scope indicators in Claude content")
	}
	if teamIndex >= projectIndex {
		t.Error("Team scope should come before project scope in Claude content")
	}

	t.Logf("Generated Claude content:\n%s", claudeContent)
	t.Logf("Generated Cursor content:\n%s", cursorContent)
}

func TestApplyDualScopeMemory_OnlyTeamScope(t *testing.T) {
	// Setup test directories - only team source, no project memory
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create team source directory and memory file
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}

	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamContent := `# Team Guidelines
Use consistent patterns across projects.`
	if err := os.WriteFile(teamMemoryPath, []byte(teamContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Create empty project directory (no mindful/memory.mdc)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Load dual scope source configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(config.Memory)

	// Verify Claude content has only team scope
	if !strings.Contains(claudeContent, "Team Guidelines -- Mindful (scope:team)") {
		t.Error("Claude content missing team scope section")
	}
	if strings.Contains(claudeContent, "scope:project") {
		t.Error("Claude content should not contain project scope when no project memory exists")
	}
}

func TestApplyDualScopeMemory_OnlyProjectScope(t *testing.T) {
	// Setup test directories - only project memory, no team source
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")

	// Create project directory and memory file
	projectMindfulDir := filepath.Join(projectDir, "mindful")
	if err := os.MkdirAll(projectMindfulDir, 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	projectMemoryPath := filepath.Join(projectMindfulDir, "memory.mdc")
	projectContent := `# Project Only
This project has specific requirements.`
	if err := os.WriteFile(projectMemoryPath, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Load dual scope source configuration (empty team source path)
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource("", projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(config.Memory)

	// Verify Claude content has only project scope
	if !strings.Contains(claudeContent, "Project Only -- Mindful (scope:project)") {
		t.Error("Claude content missing project scope section")
	}
	if strings.Contains(claudeContent, "scope:team") {
		t.Error("Claude content should not contain team scope when no team memory exists")
	}
}