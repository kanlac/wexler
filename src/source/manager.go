package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mindful/src/models"
)

// Manager implements SourceManager interface for source configuration management
type Manager struct{}

// NewManager creates a new SourceManager instance
func NewManager() *Manager {
	return &Manager{}
}

// LoadSource loads source configuration from the specified directory
func (m *Manager) LoadSource(sourcePath string) (*models.SourceConfig, error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("source path cannot be empty")
	}

	// Check if source directory exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory not found: %s", sourcePath)
	}

	config := models.NewSourceConfig()

	// Load memory.mdc if it exists
	memoryPath := filepath.Join(sourcePath, "memory.mdc")
	if _, err := os.Stat(memoryPath); err == nil {
		memory, err := m.ParseMemory(memoryPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory.mdc: %w", err)
		}
		config.Memory = memory
	}

	// Load subagent files
	subagentDir := filepath.Join(sourcePath, "subagents")
	if err := m.loadSubagentsFromDir(subagentDir, config); err != nil {
		return nil, fmt.Errorf("failed to load subagent files: %w", err)
	}

	return config, nil
}

// ListSourceFiles returns a list of all source files in the directory
func (m *Manager) ListSourceFiles(sourcePath string) ([]string, error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("source path cannot be empty")
	}

	// Check if source directory exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory not found: %s", sourcePath)
	}

	var files []string

	// Check for memory.mdc
	memoryPath := filepath.Join(sourcePath, "memory.mdc")
	if _, err := os.Stat(memoryPath); err == nil {
		files = append(files, memoryPath)
	}

	// Find subagent files
	subagentDir := filepath.Join(sourcePath, "subagent")
	if _, err := os.Stat(subagentDir); err == nil {
		err := filepath.Walk(subagentDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".mdc") {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list subagent files: %w", err)
		}
	}

	return files, nil
}

// ParseMemory parses a memory.mdc file into MemoryConfig
func (m *Manager) ParseMemory(filePath string) (*models.MemoryConfig, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("memory file not found: %s", filePath)
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory file: %w", err)
	}

	content := string(data)

	memory := models.NewMemoryConfig()

	// For single file parsing (legacy mode), treat as team content
	if strings.TrimSpace(content) != "" {
		memory.TeamContent = content
		memory.HasTeam = true
		memory.TeamSourcePath = filePath
	}

	return memory, nil
}

// ParseSubagent parses a subagent .mdc file into SubagentConfig
func (m *Manager) ParseSubagent(filePath string) (*models.SubagentConfig, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("subagent file not found: %s", filePath)
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subagent file: %w", err)
	}

	content := string(data)

	// Extract name from filename
	name := strings.TrimSuffix(filepath.Base(filePath), ".mdc")

	return models.NewSubagentConfig(name, content), nil
}

// validateUTF8File validates that a file contains valid UTF-8 content
func (m *Manager) validateUTF8File(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if !strings.Contains(string(data), "\ufffd") {
		return nil // Valid UTF-8
	}

	return fmt.Errorf("file contains invalid UTF-8 sequences: %s", filePath)
}

// loadSubagentsFromDir loads subagent files from a directory into the config
func (m *Manager) loadSubagentsFromDir(subagentDir string, config *models.SourceConfig) error {
	if _, err := os.Stat(subagentDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip silently
	}

	err := filepath.Walk(subagentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process .mdc files
		if !info.IsDir() && strings.HasSuffix(path, ".mdc") {
			subagent, err := m.ParseSubagent(path)
			if err != nil {
				return fmt.Errorf("failed to parse subagent file %s: %w", path, err)
			}

			// Use filename without extension as subagent name
			name := strings.TrimSuffix(info.Name(), ".mdc")
			subagent.Name = name
			config.Subagents[name] = subagent
		}

		return nil
	})

	return err
}

// loadTeamMemory loads team scope memory from external source directory
func (m *Manager) loadTeamMemory(teamSourcePath string, memory *models.MemoryConfig) error {
	if teamSourcePath == "" {
		return nil // No team source path, skip
	}

	teamMemoryPath := filepath.Join(teamSourcePath, "memory.mdc")
	if _, err := os.Stat(teamMemoryPath); os.IsNotExist(err) {
		return nil // File doesn't exist, skip silently
	}

	// Validate UTF-8 encoding
	if err := m.validateUTF8File(teamMemoryPath); err != nil {
		return fmt.Errorf("team memory file encoding validation failed: %w", err)
	}

	data, err := os.ReadFile(teamMemoryPath)
	if err != nil {
		return fmt.Errorf("failed to read team memory file: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) != "" {
		memory.TeamContent = content
		memory.HasTeam = true
		memory.TeamSourcePath = teamMemoryPath
	}

	return nil
}

// loadProjectMemory loads project scope memory from local mindful directory
func (m *Manager) loadProjectMemory(projectPath string, memory *models.MemoryConfig) error {
	if projectPath == "" {
		return nil // No project path, skip
	}

	projectMemoryPath := filepath.Join(projectPath, "mindful", "memory.mdc")
	if _, err := os.Stat(projectMemoryPath); os.IsNotExist(err) {
		return nil // File doesn't exist, skip silently
	}

	// Validate UTF-8 encoding
	if err := m.validateUTF8File(projectMemoryPath); err != nil {
		return fmt.Errorf("project memory file encoding validation failed: %w", err)
	}

	data, err := os.ReadFile(projectMemoryPath)
	if err != nil {
		return fmt.Errorf("failed to read project memory file: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) != "" {
		memory.ProjectContent = content
		memory.HasProject = true
		memory.ProjectSourcePath = projectMemoryPath
	}

	return nil
}

// LoadDualScopeSource loads both team scope (external) and project scope (local) configurations
func (m *Manager) LoadDualScopeSource(teamSourcePath, projectPath string) (*models.SourceConfig, error) {
	config := models.NewSourceConfig()
	memory := models.NewMemoryConfig()

	// Load team scope memory
	if err := m.loadTeamMemory(teamSourcePath, memory); err != nil {
		return nil, err
	}

	// Load project scope memory
	if err := m.loadProjectMemory(projectPath, memory); err != nil {
		return nil, err
	}

	// Only assign memory config if we have content from either scope
	if memory.HasTeam || memory.HasProject {
		config.Memory = memory
	}

	// Load team scope subagents
	if teamSourcePath != "" {
		teamSubagentDir := filepath.Join(teamSourcePath, "subagents")
		if err := m.loadSubagentsFromDir(teamSubagentDir, config); err != nil {
			return nil, fmt.Errorf("failed to load team subagents: %w", err)
		}
	}

	// Load project scope subagents
	if projectPath != "" {
		projectSubagentDir := filepath.Join(projectPath, "mindful", "subagents")
		if err := m.loadSubagentsFromDir(projectSubagentDir, config); err != nil {
			return nil, fmt.Errorf("failed to load project subagents: %w", err)
		}
	}

	return config, nil
}
