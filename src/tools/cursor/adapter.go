package cursor

import (
	"fmt"
	"path/filepath"

	"mindful/src/tools/types"
)

// Adapter implements the ToolAdapter interface for Cursor
type Adapter struct{}

// NewAdapter creates a new Cursor adapter instance
func NewAdapter() *Adapter {
	return &Adapter{}
}

// GetToolName returns the name of the tool this adapter serves
func (a *Adapter) GetToolName() string {
	return "cursor"
}

// Generate generates configuration files for Cursor
func (a *Adapter) Generate(config *types.ToolConfig) ([]types.ConfigFile, error) {
	if config == nil {
		return nil, fmt.Errorf("tool config cannot be nil")
	}

	var files []types.ConfigFile

	// Generate .cursor/rules/general.mindful.mdc (memory configuration)
	if config.Memory != nil {
		cursorContent := GenerateCursorMemoryContent(config.Memory)
		if cursorContent != "" {
			files = append(files, types.ConfigFile{
				Path:    ".cursor/rules/general.mindful.mdc",
				Content: cursorContent,
				Type:    "memory",
			})
		}
	}

	// Generate subagent files in .cursor/rules/
	for _, subagent := range config.Subagents {
		if subagent != nil && subagent.Name != "" {
			description := extractDescriptionFromContent(subagent.Content, subagent.Name)
			cursorContent := generateCursorSubagentContent(subagent.Content, description)
			rulePath := filepath.Join(".cursor", "rules", subagent.Name+".mindful.mdc")
			files = append(files, types.ConfigFile{
				Path:    rulePath,
				Content: cursorContent,
				Type:    "subagent",
			})
		}
	}

	// Generate .cursor/mcp.json
	if config.MCP != nil && len(config.MCP.Servers) > 0 {
		mcpContent, err := GenerateMCPFile(config.MCP)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MCP file: %w", err)
		}
		files = append(files, types.ConfigFile{
			Path:    ".cursor/mcp.json",
			Content: mcpContent,
			Type:    "mcp",
		})
	}

	return files, nil
}

// Validate validates the generated configuration files for Cursor
func (a *Adapter) Validate(files []types.ConfigFile) error {
	for _, file := range files {
		switch file.Type {
		case "memory":
			if err := validateCursorMemoryFile(file); err != nil {
				return fmt.Errorf("memory file validation failed for %s: %w", file.Path, err)
			}
		case "mcp":
			if err := validateMCPFile(file); err != nil {
				return fmt.Errorf("MCP file validation failed for %s: %w", file.Path, err)
			}
		case "subagent":
			if err := validateCursorSubagentFile(file); err != nil {
				return fmt.Errorf("subagent file validation failed for %s: %w", file.Path, err)
			}
		}
	}
	return nil
}