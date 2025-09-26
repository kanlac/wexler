package cursor

import (
	"encoding/json"
	"fmt"
	"strings"

	"mindful/src/models"
	"mindful/src/tools/types"
)

// GenerateMCPFile generates MCP JSON configuration for Cursor
func GenerateMCPFile(mcp *models.MCPConfig) (string, error) {
	if mcp == nil || len(mcp.Servers) == 0 {
		return `{"mcpServers": {}}`, nil
	}

	// Use the model's ToMCPJSON method
	data, err := mcp.ToMCPJSON()
	if err != nil {
		return "", fmt.Errorf("failed to generate MCP JSON: %w", err)
	}

	return string(data), nil
}

// validateMCPFile validates MCP JSON configuration
func validateMCPFile(file types.ConfigFile) error {
	if file.Content == "" {
		return fmt.Errorf("MCP file content cannot be empty")
	}

	var mcpData interface{}
	if err := json.Unmarshal([]byte(file.Content), &mcpData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	return nil
}

// validateCursorSubagentFile validates Cursor subagent configuration file content
func validateCursorSubagentFile(file types.ConfigFile) error {
	if file.Path == "" {
		return fmt.Errorf("subagent file path cannot be empty")
	}

	// Ensure it has proper extension
	if !strings.HasSuffix(file.Path, ".mindful.mdc") {
		return fmt.Errorf("Cursor subagent file must have .mindful.mdc extension")
	}

	// Basic validation - ensure it has proper frontmatter if content exists
	if file.Content != "" && !strings.HasPrefix(file.Content, "---\n") {
		return fmt.Errorf("invalid Cursor subagent file format: missing frontmatter")
	}

	return nil
}