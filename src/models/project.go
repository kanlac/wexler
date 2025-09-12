package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultWexlerSource    = "~/.wexler"
	DefaultStorageFileName = "wexler.db"
)

// ProjectConfig represents the main project configuration (wexler.yaml)
type ProjectConfig struct {
	Name       string            `yaml:"name" json:"name"`
	Version    string            `yaml:"version" json:"version"`
	SourcePath string            `yaml:"source_path" json:"source_path"` // 指向全局配置源，如 "~/.wexler"
	Tools      map[string]string `yaml:"tools" json:"tools"`
}

// DefaultProjectConfig returns a project configuration with sensible defaults
// func DefaultProjectConfig(name string) *ProjectConfig {
// 	return &ProjectConfig{
// 		Name:       name,
// 		Version:    "1.0.0",
// 		SourcePath: DefaultWexlerSource,
// 		Tools:      map[string]string{},
// 	}
// }

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

	return nil
}

// GetAbsoluteSourcePath returns the absolute path to the global wexler source directory
func (p *ProjectConfig) GetAbsoluteSourcePath() (string, error) {
	// Handle ~ prefix for user home directory
	if strings.HasPrefix(p.SourcePath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine user home directory: %w", err)
		}
		
		if p.SourcePath == "~" {
			return homeDir, nil
		} else if strings.HasPrefix(p.SourcePath, "~/") {
			return filepath.Join(homeDir, strings.TrimPrefix(p.SourcePath, "~/")), nil
		}
	}
	
	if filepath.IsAbs(p.SourcePath) {
		return p.SourcePath, nil
	}
	
	// 拒绝相对路径
	return "", fmt.Errorf("source path must be absolute or use ~ prefix, got: %s", p.SourcePath)
}

// GetDatabasePath returns the path to the wexler database file
func (p *ProjectConfig) GetDatabasePath() (string, error) {
	sourceAbs, err := p.GetAbsoluteSourcePath()
	if err != nil {
		return "", fmt.Errorf("cannot resolve source path: %w", err)
	}
	return filepath.Join(sourceAbs, DefaultStorageFileName), nil
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
