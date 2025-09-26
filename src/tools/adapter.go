package tools

import (
	"fmt"

	"mindful/src/tools/claude"
	"mindful/src/tools/cursor"
	"mindful/src/tools/types"
)

// Re-export types for convenience
type ConfigFile = types.ConfigFile
type ToolConfig = types.ToolConfig
type ToolAdapter = types.ToolAdapter

// NewAdapter creates a new tool adapter for the specified tool
func NewAdapter(toolName string) (types.ToolAdapter, error) {
	switch toolName {
	case "claude":
		return claude.NewAdapter(), nil
	case "cursor":
		return cursor.NewAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
}

// GetSupportedTools returns a list of supported tool names
func GetSupportedTools() []string {
	return []string{"claude", "cursor"}
}