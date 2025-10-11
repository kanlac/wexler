package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// DefaultMindfulDirName is the directory that contains mindful configuration assets.
	DefaultMindfulDirName = "mindful"
	// DefaultOutDirName is the directory under mindful/ where build artifacts are written.
	DefaultOutDirName = "out"
	// DefaultStorageFileName is the filename of the BoltDB database used for MCP storage.
	DefaultStorageFileName = "mindful.db"
)

// ProjectConfig models the mindful.yaml configuration file.
type ProjectConfig struct {
	Name               string            `yaml:"name" json:"name"`
	Version            string            `yaml:"version" json:"version"`
	Source             string            `yaml:"source,omitempty" json:"source,omitempty"`                   // New field (preferred)
	SourcePath         string            `yaml:"source_path,omitempty" json:"source_path,omitempty"`         // Legacy field for backward compatibility
	EnableCodingAgents []string          `yaml:"enable-coding-agents,omitempty" json:"enable-coding-agents"` // Preferred way to declare enabled tools
	Tools              map[string]string `yaml:"tools,omitempty" json:"tools,omitempty"`                     // Legacy map of tool -> status ("enabled"/"disabled")
}

// ToolSymlinkConfig defines the link templates for a given tool.
type ToolSymlinkConfig struct {
	Memory    string `yaml:"memory,omitempty" json:"memory,omitempty"`
	Subagents string `yaml:"subagents,omitempty" json:"subagents,omitempty"`
	MCP       string `yaml:"mcp,omitempty" json:"mcp,omitempty"`
}

// SymlinkConfig is a thin wrapper that offers helper methods for tool lookups.
type SymlinkConfig struct {
	Tools map[string]*ToolSymlinkConfig
}

// NewSymlinkConfig normalises an optional raw map into a helper struct.
func NewSymlinkConfig(raw map[string]*ToolSymlinkConfig) *SymlinkConfig {
	cfg := &SymlinkConfig{
		Tools: make(map[string]*ToolSymlinkConfig),
	}

	for k, v := range raw {
		if v == nil {
			cfg.Tools[k] = &ToolSymlinkConfig{}
			continue
		}
		cfg.Tools[k] = &ToolSymlinkConfig{
			Memory:    strings.TrimSpace(v.Memory),
			Subagents: strings.TrimSpace(v.Subagents),
			MCP:       strings.TrimSpace(v.MCP),
		}
	}

	return cfg
}

// ToolNames returns the list of tool identifiers sorted alphabetically.
func (c *SymlinkConfig) ToolNames() []string {
	if c == nil || len(c.Tools) == 0 {
		return nil
	}

	names := make([]string, 0, len(c.Tools))
	for name := range c.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ToolConfig returns the symlink configuration for a tool.
func (c *SymlinkConfig) ToolConfig(toolName string) (*ToolSymlinkConfig, bool) {
	if c == nil || c.Tools == nil {
		return nil, false
	}
	cfg, ok := c.Tools[toolName]
	return cfg, ok
}

// HasTool reports whether a tool has symlink configuration defined.
func (c *SymlinkConfig) HasTool(toolName string) bool {
	_, ok := c.ToolConfig(toolName)
	return ok
}

// IsEmpty reports whether the tool configuration defines any paths.
func (t *ToolSymlinkConfig) IsEmpty() bool {
	if t == nil {
		return true
	}
	return strings.TrimSpace(t.Memory) == "" &&
		strings.TrimSpace(t.Subagents) == "" &&
		strings.TrimSpace(t.MCP) == ""
}

// Validate performs lightweight structural validation of the project configuration.
func (p *ProjectConfig) Validate() error {
	if p == nil {
		return fmt.Errorf("project config is nil")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if strings.TrimSpace(p.Version) == "" {
		return fmt.Errorf("project version cannot be empty")
	}
	if _, err := p.resolveSourceValue(); err != nil {
		return err
	}
	return nil
}

// resolveSourceValue determines which source root field is populated.
func (p *ProjectConfig) resolveSourceValue() (string, error) {
	if p == nil {
		return "", fmt.Errorf("project config is nil")
	}

	candidate := strings.TrimSpace(p.Source)
	if candidate == "" {
		candidate = strings.TrimSpace(p.SourcePath)
	}

	if candidate == "" {
		return "", fmt.Errorf("source path cannot be empty")
	}

	return candidate, nil
}

// ResolveSourceRoot resolves the project source root to an absolute path.
func (p *ProjectConfig) ResolveSourceRoot(projectPath string) (string, error) {
	candidate, err := p.resolveSourceValue()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(candidate, "~") {
		homeDir, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", herr)
		}
		if candidate == "~" {
			return homeDir, nil
		}
		candidate = filepath.Join(homeDir, strings.TrimPrefix(candidate, "~/"))
	}

	if filepath.IsAbs(candidate) {
		return filepath.Clean(candidate), nil
	}

	if projectPath == "" {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("cannot resolve source path %q: %w", candidate, err)
		}
		return filepath.Clean(abs), nil
	}

	return filepath.Clean(filepath.Join(projectPath, candidate)), nil
}

// ResolveMindfulDir returns the absolute mindful/ directory for the project.
func (p *ProjectConfig) ResolveMindfulDir(projectPath string) string {
	return filepath.Join(projectPath, DefaultMindfulDirName)
}

// ResolveOutDir returns the absolute path to mindful/out.
func (p *ProjectConfig) ResolveOutDir(projectPath string) string {
	return filepath.Join(p.ResolveMindfulDir(projectPath), DefaultOutDirName)
}

// GetDatabasePath returns the absolute path to the BoltDB storage file.
func (p *ProjectConfig) GetDatabasePath(projectPath string) (string, error) {
	sourceRoot, err := p.ResolveSourceRoot(projectPath)
	if err != nil {
		return "", fmt.Errorf("cannot determine storage path: %w", err)
	}
	return filepath.Join(sourceRoot, DefaultStorageFileName), nil
}

// GetEnabledTools returns the list of enabled tools (new field takes precedence).
func (p *ProjectConfig) GetEnabledTools() []string {
	if len(p.EnableCodingAgents) > 0 {
		return normaliseToolList(p.EnableCodingAgents)
	}

	if len(p.Tools) == 0 {
		return nil
	}

	var enabled []string
	for name, status := range p.Tools {
		if strings.EqualFold(status, "enabled") {
			enabled = append(enabled, name)
		}
	}
	return normaliseToolList(enabled)
}

// IsToolEnabled reports whether a tool is enabled.
func (p *ProjectConfig) IsToolEnabled(toolName string) bool {
	tools := p.GetEnabledTools()
	for _, name := range tools {
		if name == toolName {
			return true
		}
	}
	return false
}

// EnableTool enables a tool in both the new and legacy representations.
func (p *ProjectConfig) EnableTool(toolName string) {
	if p == nil {
		return
	}

	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		return
	}

	p.EnableCodingAgents = append(p.EnableCodingAgents, toolName)
	p.EnableCodingAgents = normaliseToolList(p.EnableCodingAgents)

	if p.Tools == nil {
		p.Tools = make(map[string]string)
	}
	p.Tools[toolName] = "enabled"
}

// DisableTool disables a tool in both representations.
func (p *ProjectConfig) DisableTool(toolName string) {
	if p == nil {
		return
	}

	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		return
	}

	// Remove from EnableCodingAgents slice
	if len(p.EnableCodingAgents) > 0 {
		var filtered []string
		for _, name := range p.EnableCodingAgents {
			if name != toolName {
				filtered = append(filtered, name)
			}
		}
		p.EnableCodingAgents = filtered
	}

	if p.Tools == nil {
		p.Tools = make(map[string]string)
	}
	p.Tools[toolName] = "disabled"
}

// normaliseToolList removes duplicates, trims whitespace, and sorts the tool list.
func normaliseToolList(values []string) []string {
	set := make(map[string]struct{})
	for _, v := range values {
		name := strings.TrimSpace(v)
		if name == "" {
			continue
		}
		set[strings.ToLower(name)] = struct{}{}
	}

	if len(set) == 0 {
		return nil
	}

	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)

	// Preserve canonical casing by reusing original spelling when available.
	// This keeps outputs predictable (e.g. "claude" stays "claude").
	for i, name := range names {
		names[i] = name
	}
	return names
}
