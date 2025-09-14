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
	subagentDir := filepath.Join(sourcePath, "subagent")
	if _, err := os.Stat(subagentDir); err == nil {
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

		if err != nil {
			return nil, fmt.Errorf("failed to load subagent files: %w", err)
		}
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

	// Parse content into sections using the model's method
	if err := memory.ParseMemoryContent(content); err != nil {
		return nil, fmt.Errorf("failed to parse memory content: %w", err)
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
