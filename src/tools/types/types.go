package types

import "mindful/src/models"

// ConfigFile represents a generated configuration file
type ConfigFile struct {
	Path    string `yaml:"path" json:"path"`
	Content string `yaml:"content" json:"content"`
	Type    string `yaml:"type" json:"type"` // "memory", "subagent", "mcp"
}

// ToolConfig represents input configuration for tool adaptation
type ToolConfig struct {
	ToolName  string                   `yaml:"tool_name" json:"tool_name"`
	Memory    *models.MemoryConfig     `yaml:"memory" json:"memory"`
	Subagents []*models.SubagentConfig `yaml:"subagents" json:"subagents"`
	MCP       *models.MCPConfig        `yaml:"mcp" json:"mcp"`
}

// ToolAdapter defines the interface for tool-specific configuration generators
type ToolAdapter interface {
	// Generate generates configuration files for the specific tool
	Generate(config *ToolConfig) ([]ConfigFile, error)

	// GetToolName returns the name of the tool this adapter serves
	GetToolName() string

	// Validate validates the generated configuration files
	Validate(files []ConfigFile) error
}