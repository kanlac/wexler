package models

import (
	"fmt"
)

// ToolConfig represents configuration for a specific AI tool
type ToolConfig struct {
	ToolName  string            `yaml:"tool_name" json:"tool_name"`   // "claude", "cursor", etc.
	Memory    *MemoryConfig     `yaml:"memory" json:"memory"`         // Memory configuration to apply
	Subagents []*SubagentConfig `yaml:"subagents" json:"subagents"`   // Subagent configurations to apply
	MCP       *MCPConfig        `yaml:"mcp" json:"mcp"`               // MCP configuration to apply
}

// NewToolConfig creates a new tool configuration
func NewToolConfig(toolName string) *ToolConfig {
	return &ToolConfig{
		ToolName:  toolName,
		Memory:    NewMemoryConfig(),
		Subagents: []*SubagentConfig{},
		MCP:       NewMCPConfig(),
	}
}

// AddSubagent adds a subagent configuration to the tool config
func (tc *ToolConfig) AddSubagent(subagent *SubagentConfig) {
	tc.Subagents = append(tc.Subagents, subagent)
}

// GetSubagent returns a subagent by name
func (tc *ToolConfig) GetSubagent(name string) (*SubagentConfig, bool) {
	for _, subagent := range tc.Subagents {
		if subagent.Name == name {
			return subagent, true
		}
	}
	return nil, false
}

// ListSubagentNames returns all subagent names
func (tc *ToolConfig) ListSubagentNames() []string {
	names := make([]string, len(tc.Subagents))
	for i, subagent := range tc.Subagents {
		names[i] = subagent.Name
	}
	return names
}

// Validate checks if the tool configuration is valid
func (tc *ToolConfig) Validate() error {
	if tc == nil {
		return fmt.Errorf("tool config is nil")
	}
	
	if tc.ToolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	// Validate memory config
	if tc.Memory != nil {
		if err := tc.Memory.Validate(); err != nil {
			return fmt.Errorf("memory config validation failed: %w", err)
		}
	}
	
	// Validate subagent configs
	for i, subagent := range tc.Subagents {
		if err := subagent.Validate(); err != nil {
			return fmt.Errorf("subagent %d validation failed: %w", i, err)
		}
	}
	
	// Validate MCP config
	if tc.MCP != nil {
		if err := tc.MCP.Validate(); err != nil {
			return fmt.Errorf("MCP config validation failed: %w", err)
		}
	}
	
	return nil
}

// ConfigFile represents a file that will be generated for an AI tool
type ConfigFile struct {
	Path    string `yaml:"path" json:"path"`       // Relative path where the file should be written
	Content string `yaml:"content" json:"content"` // Content of the file
	Type    string `yaml:"type" json:"type"`       // "memory", "subagent", "mcp"
}

// NewConfigFile creates a new config file
func NewConfigFile(path, content, fileType string) *ConfigFile {
	return &ConfigFile{
		Path:    path,
		Content: content,
		Type:    fileType,
	}
}

// Validate checks if the config file is valid
func (cf *ConfigFile) Validate() error {
	if cf == nil {
		return fmt.Errorf("config file is nil")
	}
	
	if cf.Path == "" {
		return fmt.Errorf("config file path cannot be empty")
	}
	
	validTypes := map[string]bool{
		"memory":   true,
		"subagent": true,
		"mcp":      true,
	}
	
	if !validTypes[cf.Type] {
		return fmt.Errorf("invalid config file type: %s", cf.Type)
	}
	
	return nil
}

// GetSize returns the size of the file content in bytes
func (cf *ConfigFile) GetSize() int {
	return len([]byte(cf.Content))
}

// IsEmpty returns true if the file content is empty
func (cf *ConfigFile) IsEmpty() bool {
	return cf.Content == ""
}

// Clone creates a deep copy of the config file
func (cf *ConfigFile) Clone() *ConfigFile {
	if cf == nil {
		return nil
	}
	
	return &ConfigFile{
		Path:    cf.Path,
		Content: cf.Content,
		Type:    cf.Type,
	}
}

// ToolAdapter defines the interface that all tool adapters must implement
type ToolAdapter interface {
	// Generate creates configuration files for the specific AI tool
	Generate(config *ToolConfig) ([]*ConfigFile, error)
	
	// Validate checks if the generated configuration files are valid for the tool
	Validate(files []*ConfigFile) error
	
	// Merge combines existing files with new files, detecting conflicts
	Merge(existing []*ConfigFile, new []*ConfigFile) ([]*ConfigFile, *ConflictResult, error)
}

// ToolAdapterFactory defines the interface for creating tool adapters
type ToolAdapterFactory interface {
	// CreateAdapter creates a tool adapter for the specified tool name
	CreateAdapter(toolName string) (ToolAdapter, error)
	
	// ListSupportedTools returns a list of supported tool names
	ListSupportedTools() []string
	
	// IsToolSupported checks if a tool is supported
	IsToolSupported(toolName string) bool
}

// ToolInfo contains metadata about a supported AI tool
type ToolInfo struct {
	Name        string   `yaml:"name" json:"name"`               // Tool name (e.g., "claude")
	DisplayName string   `yaml:"display_name" json:"display_name"` // Human-readable name (e.g., "Claude Code")
	Version     string   `yaml:"version" json:"version"`         // Supported version
	ConfigFiles []string `yaml:"config_files" json:"config_files"` // List of config files it generates
	Description string   `yaml:"description" json:"description"` // Tool description
}

// NewToolInfo creates a new tool info
func NewToolInfo(name, displayName, version, description string, configFiles []string) *ToolInfo {
	return &ToolInfo{
		Name:        name,
		DisplayName: displayName,
		Version:     version,
		ConfigFiles: configFiles,
		Description: description,
	}
}

// Validate checks if the tool info is valid
func (ti *ToolInfo) Validate() error {
	if ti == nil {
		return fmt.Errorf("tool info is nil")
	}
	
	if ti.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	if ti.DisplayName == "" {
		return fmt.Errorf("tool display name cannot be empty")
	}
	
	return nil
}

// Clone creates a deep copy of the tool config
func (tc *ToolConfig) Clone() *ToolConfig {
	if tc == nil {
		return nil
	}
	
	clone := &ToolConfig{
		ToolName: tc.ToolName,
	}
	
	// Clone memory config
	if tc.Memory != nil {
		clone.Memory = &MemoryConfig{
			Content:  tc.Memory.Content,
			Sections: make(map[string]string),
		}
		for name, content := range tc.Memory.Sections {
			clone.Memory.Sections[name] = content
		}
	}
	
	// Clone subagents
	clone.Subagents = make([]*SubagentConfig, len(tc.Subagents))
	for i, subagent := range tc.Subagents {
		clone.Subagents[i] = &SubagentConfig{
			Name:    subagent.Name,
			Content: subagent.Content,
		}
	}
	
	// Clone MCP config
	if tc.MCP != nil {
		clone.MCP = tc.MCP.Clone()
	}
	
	return clone
}