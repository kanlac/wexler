package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadYAML loads a YAML file and unmarshals it into the provided interface
func LoadYAML(filePath string, target interface{}) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML from '%s': %w", filePath, err)
	}

	return nil
}

// SaveYAML marshals the provided data to YAML and saves it to the specified file
func SaveYAML(filePath string, data interface{}) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file '%s': %w", filePath, err)
	}

	return nil
}

// CreateDefaultConfig creates a default mindful.yaml configuration file
// func CreateDefaultConfig(projectPath, projectName string) error {
// 	if projectPath == "" {
// 		return fmt.Errorf("project path cannot be empty")
// 	}
// 	if projectName == "" {
// 		return fmt.Errorf("project name cannot be empty")
// 	}

// 	config := models.DefaultProjectConfig(projectName)

// 	configPath := filepath.Join(projectPath, "mindful.yaml")
// 	return SaveYAML(configPath, config)
// }

// BackupConfig creates a backup of the existing configuration file
func BackupConfig(projectPath string) error {
	if projectPath == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	configPath := filepath.Join(projectPath, "mindful.yaml")

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("no mindful.yaml found to backup")
	}

	// Create backup with timestamp
	backupPath := configPath + ".backup"

	// Read original
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config for backup: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// RestoreConfig restores configuration from backup
func RestoreConfig(projectPath string) error {
	if projectPath == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	configPath := filepath.Join(projectPath, "mindful.yaml")
	backupPath := configPath + ".backup"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	// Copy backup to main config
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}
