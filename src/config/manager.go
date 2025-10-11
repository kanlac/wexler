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

// LoadProject loads mindful/mindful.yaml for the specified project path.
func (m *Manager) LoadProject(projectPath string) (*models.ProjectConfig, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

	configPath := filepath.Join(projectPath, models.DefaultMindfulDirName, "mindful.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("mindful.yaml not found in %s", filepath.Dir(configPath))
		}
		return nil, fmt.Errorf("failed to read mindful.yaml: %w", err)
	}

	var config models.ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mindful.yaml: %w", err)
	}

	if err := m.ValidateProject(projectPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveProject persists mindful/mindful.yaml for the project.
func (m *Manager) SaveProject(projectPath string, config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if projectPath == "" {
		projectPath = "."
	}

	if err := m.ValidateProject(projectPath, config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	configPath := filepath.Join(projectPath, models.DefaultMindfulDirName, "mindful.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create mindful directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write mindful.yaml: %w", err)
	}

	return nil
}

// ValidateProject validates the configuration for completeness and correctness.
func (m *Manager) ValidateProject(projectPath string, config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := ValidateProjectName(config.Name); err != nil {
		return err
	}

	if err := ValidateVersion(config.Version); err != nil {
		return err
	}

	if err := config.Validate(); err != nil {
		return err
	}

	if err := ValidateProjectStructure(projectPath, config); err != nil {
		return err
	}

	return nil
}
