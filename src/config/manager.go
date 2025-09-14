package config

import (
	"fmt"
	"os"
	"path/filepath"

	"mindful/src/models"

	"gopkg.in/yaml.v3"
)

// Manager implements ConfigManager interface for project configuration management
type Manager struct{}

// NewManager creates a new ConfigManager instance
func NewManager() *Manager {
	return &Manager{}
}

// LoadProject loads a mindful.yaml configuration from the specified project path
func (m *Manager) LoadProject(projectPath string) (*models.ProjectConfig, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

	configPath := filepath.Join(projectPath, "mindful.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mindful.yaml not found in %s", projectPath)
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mindful.yaml: %w", err)
	}

	// Parse YAML
	var config models.ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mindful.yaml: %w", err)
	}

	return &config, nil
}

// SaveProject saves a ProjectConfig to mindful.yaml in the project directory
func (m *Manager) SaveProject(config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate config before saving
	if err := m.ValidateProject(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Determine project path - assume current directory if not specified
	projectPath := "."
	configPath := filepath.Join(projectPath, "mindful.yaml")

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write mindful.yaml: %w", err)
	}

	return nil
}

// ValidateProject validates a ProjectConfig for completeness and correctness
func (m *Manager) ValidateProject(config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate project name
	if err := ValidateProjectName(config.Name); err != nil {
		return err
	}

	// Validate version
	if err := ValidateVersion(config.Version); err != nil {
		return err
	}

	if config.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}

	// Validate tool configuration
	if err := ValidateToolConfiguration(config); err != nil {
		return err
	}

	return nil
}
