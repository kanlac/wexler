package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mindful/src/source"
	"mindful/src/tools/claude"
	"mindful/src/tools/cursor"
)

func TestDualScopeMemory_BothFilesExist(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create directories
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "mindful"), 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	// Create team memory file
	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamContent := "# Team Configuration\nThis is team-wide configuration content."
	if err := os.WriteFile(teamMemoryPath, []byte(teamContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Create project memory file
	projectMemoryPath := filepath.Join(projectDir, "mindful", "memory.mdc")
	projectContent := "# Project Configuration\nThis is project-specific configuration content."
	if err := os.WriteFile(projectMemoryPath, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Load dual scope configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify memory configuration
	if config.Memory == nil {
		t.Fatal("Memory config is nil")
	}

	memory := config.Memory
	if !memory.HasTeam {
		t.Error("Expected HasTeam to be true")
	}
	if !memory.HasProject {
		t.Error("Expected HasProject to be true")
	}

	if memory.TeamContent != teamContent {
		t.Errorf("Team content mismatch. Expected: %q, Got: %q", teamContent, memory.TeamContent)
	}
	if memory.ProjectContent != projectContent {
		t.Errorf("Project content mismatch. Expected: %q, Got: %q", projectContent, memory.ProjectContent)
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(memory)

	// Verify Claude dual-scope structure - updated for new header format
	if !strings.Contains(claudeContent, "Team Configuration -- Mindful (scope:team)") {
		t.Error("Claude content missing team scope header with user title")
	}
	if !strings.Contains(claudeContent, "Project Configuration -- Mindful (scope:project)") {
		t.Error("Claude content missing project scope header with user title")
	}
	if !strings.Contains(claudeContent, "<!-- Source: "+teamMemoryPath+" -->") {
		t.Error("Claude content missing team source path comment")
	}
	if !strings.Contains(claudeContent, "<!-- Source: "+projectMemoryPath+" -->") {
		t.Error("Claude content missing project source path comment")
	}

	// Test Cursor memory generation
	cursorContent := cursor.GenerateCursorMemoryContent(memory)

	// Verify Cursor dual-scope structure - updated for new header format
	if !strings.HasPrefix(cursorContent, "---\n") {
		t.Error("Cursor content missing frontmatter")
	}
	if !strings.Contains(cursorContent, "Team Configuration -- Mindful (scope:team)") {
		t.Error("Cursor content missing team scope header with user title")
	}
	if !strings.Contains(cursorContent, "Project Configuration -- Mindful (scope:project)") {
		t.Error("Cursor content missing project scope header with user title")
	}
}

func TestDualScopeMemory_OnlyTeamFile(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create directories
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}

	// Create team memory file only
	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	teamContent := "# Team Only Configuration\nThis is team-only content."
	if err := os.WriteFile(teamMemoryPath, []byte(teamContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Load dual scope configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify memory configuration
	memory := config.Memory
	if !memory.HasTeam {
		t.Error("Expected HasTeam to be true")
	}
	if memory.HasProject {
		t.Error("Expected HasProject to be false")
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(memory)

	// Verify Claude contains only team scope - updated for new header format
	if !strings.Contains(claudeContent, "Team Only Configuration -- Mindful (scope:team)") {
		t.Error("Claude content missing team scope header with user title")
	}
	if strings.Contains(claudeContent, "Mindful (scope:project)") {
		t.Error("Claude content should not contain project scope header")
	}
}

func TestDualScopeMemory_OnlyProjectFile(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")

	// Create directories
	if err := os.MkdirAll(filepath.Join(projectDir, "mindful"), 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	// Create project memory file only
	projectMemoryPath := filepath.Join(projectDir, "mindful", "memory.mdc")
	projectContent := "# Project Only Configuration\nThis is project-only content."
	if err := os.WriteFile(projectMemoryPath, []byte(projectContent), 0644); err != nil {
		t.Fatalf("Failed to write project memory file: %v", err)
	}

	// Load dual scope configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource("", projectDir) // Empty team source path
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify memory configuration
	memory := config.Memory
	if memory.HasTeam {
		t.Error("Expected HasTeam to be false")
	}
	if !memory.HasProject {
		t.Error("Expected HasProject to be true")
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(memory)

	// Verify Claude contains only project scope - updated for new header format
	if strings.Contains(claudeContent, "Mindful (scope:team)") {
		t.Error("Claude content should not contain team scope header")
	}
	if !strings.Contains(claudeContent, "Project Only Configuration -- Mindful (scope:project)") {
		t.Error("Claude content missing project scope header with user title")
	}
}

func TestDualScopeMemory_NoMemoryFiles(t *testing.T) {
	// Setup test directories (empty)
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create empty directories
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "mindful"), 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	// Load dual scope configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify no memory configuration
	if config.Memory != nil {
		t.Error("Expected Memory config to be nil when no files exist")
	}

	// Test Claude memory generation with nil memory
	claudeContent := claude.GenerateClaudeMemoryContent(nil)
	if claudeContent != "" {
		t.Error("Expected empty Claude content when no memory files exist")
	}

	// Test Cursor memory generation with nil memory
	cursorContent := cursor.GenerateCursorMemoryContent(nil)
	if cursorContent != "" {
		t.Error("Expected empty Cursor content when no memory files exist")
	}
}

func TestDualScopeMemory_UTF8Validation(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")

	// Create directory
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}

	// Create file with valid UTF-8 content
	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	validContent := "# 中文内容\n这是有效的UTF-8内容。\n🚀 Emojis work too!"
	if err := os.WriteFile(teamMemoryPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write team memory file: %v", err)
	}

	// Load dual scope configuration - should succeed
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, "")
	if err != nil {
		t.Fatalf("Failed to load dual scope source with valid UTF-8: %v", err)
	}

	// Verify memory configuration was loaded
	if config.Memory == nil {
		t.Fatal("Memory config is nil")
	}

	if !config.Memory.HasTeam {
		t.Error("Expected HasTeam to be true")
	}
}

func TestDualScopeMemory_EmptyFiles(t *testing.T) {
	// Setup test directories
	tempDir := t.TempDir()
	teamSourceDir := filepath.Join(tempDir, "team_source")
	projectDir := filepath.Join(tempDir, "project")

	// Create directories
	if err := os.MkdirAll(teamSourceDir, 0755); err != nil {
		t.Fatalf("Failed to create team source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "mindful"), 0755); err != nil {
		t.Fatalf("Failed to create project mindful directory: %v", err)
	}

	// Create empty memory files
	teamMemoryPath := filepath.Join(teamSourceDir, "memory.mdc")
	if err := os.WriteFile(teamMemoryPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty team memory file: %v", err)
	}

	projectMemoryPath := filepath.Join(projectDir, "mindful", "memory.mdc")
	if err := os.WriteFile(projectMemoryPath, []byte("   \n\t  \n"), 0644); err != nil {
		t.Fatalf("Failed to write whitespace-only project memory file: %v", err)
	}

	// Load dual scope configuration
	manager := source.NewManager()
	config, err := manager.LoadDualScopeSource(teamSourceDir, projectDir)
	if err != nil {
		t.Fatalf("Failed to load dual scope source: %v", err)
	}

	// Verify no memory configuration (empty files should be silently skipped)
	if config.Memory != nil {
		t.Error("Expected Memory config to be nil when files are empty")
	}

	// Test Claude memory generation
	claudeContent := claude.GenerateClaudeMemoryContent(config.Memory)
	if claudeContent != "" {
		t.Error("Expected empty Claude content when memory files are empty")
	}
}