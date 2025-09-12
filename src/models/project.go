package models

import (
	"fmt"
	"path/filepath"
)

// ProjectConfig represents the main project configuration (wexler.yaml)
type ProjectConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version" json:"version"`
	SourcePath  string            `yaml:"source_path" json:"source_path"`
	StoragePath string            `yaml:"storage_path" json:"storage_path"`
	Tools       map[string]string `yaml:"tools" json:"tools"`
}

// DefaultProjectConfig returns a project configuration with sensible defaults
func DefaultProjectConfig(name string) *ProjectConfig {
	return &ProjectConfig{
		Name:        name,
		Version:     "1.0.0",
		SourcePath:  "source",
		StoragePath: ".wexler",
		Tools:       map[string]string{},
	}
}

// Validate checks if the project configuration is valid
func (p *ProjectConfig) Validate() error {
	if p == nil {
		return fmt.Errorf("project config is nil")
	}
	
	if p.Name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	
	if p.Version == "" {
		return fmt.Errorf("project version cannot be empty")
	}
	
	if p.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	
	if p.StoragePath == "" {
		return fmt.Errorf("storage path cannot be empty")
	}
	
	return nil
}

// GetAbsoluteSourcePath returns the absolute path to the source directory
func (p *ProjectConfig) GetAbsoluteSourcePath(projectRoot string) string {
	if filepath.IsAbs(p.SourcePath) {
		return p.SourcePath
	}
	return filepath.Join(projectRoot, p.SourcePath)
}

// GetAbsoluteStoragePath returns the absolute path to the storage directory
func (p *ProjectConfig) GetAbsoluteStoragePath(projectRoot string) string {
	if filepath.IsAbs(p.StoragePath) {
		return p.StoragePath
	}
	return filepath.Join(projectRoot, p.StoragePath)
}

// IsToolEnabled checks if a specific tool is enabled in the project
func (p *ProjectConfig) IsToolEnabled(toolName string) bool {
	if p.Tools == nil {
		return false
	}
	
	status, exists := p.Tools[toolName]
	if !exists {
		return false
	}
	
	return status == "enabled"
}

// EnableTool enables a specific tool in the project configuration
func (p *ProjectConfig) EnableTool(toolName string) {
	if p.Tools == nil {
		p.Tools = make(map[string]string)
	}
	p.Tools[toolName] = "enabled"
}

// DisableTool disables a specific tool in the project configuration
func (p *ProjectConfig) DisableTool(toolName string) {
	if p.Tools == nil {
		p.Tools = make(map[string]string)
	}
	p.Tools[toolName] = "disabled"
}

// GetEnabledTools returns a list of all enabled tools
func (p *ProjectConfig) GetEnabledTools() []string {
	var enabled []string
	for toolName, status := range p.Tools {
		if status == "enabled" {
			enabled = append(enabled, toolName)
		}
	}
	return enabled
}