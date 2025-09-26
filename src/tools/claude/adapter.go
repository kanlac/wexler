package claude

import (
	"fmt"
	"path/filepath"

	"mindful/src/tools/types"
)

// Adapter implements the ToolAdapter interface for Claude Code
type Adapter struct{}

// NewAdapter creates a new Claude adapter instance
func NewAdapter() *Adapter {
	return &Adapter{}
}

// GetToolName returns the name of the tool this adapter serves
func (a *Adapter) GetToolName() string {
	return "claude"
}

// Generate generates configuration files for Claude Code
func (a *Adapter) Generate(config *types.ToolConfig) ([]types.ConfigFile, error) {
	if config == nil {
		return nil, fmt.Errorf("tool config cannot be nil")
	}

	var files []types.ConfigFile

	// Generate CLAUDE.md (memory configuration)
	if config.Memory != nil {
		memoryContent := GenerateClaudeMemoryContent(config.Memory)
		if memoryContent != "" {
			files = append(files, types.ConfigFile{
				Path:    "CLAUDE.md",
				Content: memoryContent,
				Type:    "memory",
			})
		}
	}

	// Generate subagent files in .claude/agents/
	for _, subagent := range config.Subagents {
		if subagent != nil && subagent.Name != "" {
			agentPath := filepath.Join(".claude", "agents", subagent.Name+".mindful.md")
			files = append(files, types.ConfigFile{
				Path:    agentPath,
				Content: subagent.Content,
				Type:    "subagent",
			})
		}
	}

	// Generate .mcp.json
	if config.MCP != nil && len(config.MCP.Servers) > 0 {
		mcpContent, err := GenerateMCPFile(config.MCP)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MCP file: %w", err)
		}
		files = append(files, types.ConfigFile{
			Path:    ".mcp.json",
			Content: mcpContent,
			Type:    "mcp",
		})
	}

	return files, nil
}

// Validate validates the generated configuration files for Claude
func (a *Adapter) Validate(files []types.ConfigFile) error {
	for _, file := range files {
		switch file.Type {
		case "memory":
			if err := validateClaudeMemoryFile(file); err != nil {
				return fmt.Errorf("memory file validation failed for %s: %w", file.Path, err)
			}
		case "mcp":
			if err := validateMCPFile(file); err != nil {
				return fmt.Errorf("MCP file validation failed for %s: %w", file.Path, err)
			}
		case "subagent":
			if err := validateSubagentFile(file); err != nil {
				return fmt.Errorf("subagent file validation failed for %s: %w", file.Path, err)
			}
		}
	}
	return nil
}