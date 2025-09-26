package claude

import (
	"fmt"
	"strings"

	"mindful/src/models"
	"mindful/src/tools/common"
	"mindful/src/tools/types"
)

// GenerateClaudeMemoryContent generates dual-scope memory content for CLAUDE.md
func GenerateClaudeMemoryContent(memory *models.MemoryConfig) string {
	if memory == nil {
		return ""
	}

	var parts []string

	// Add team scope section if available
	if memory.HasTeam && strings.TrimSpace(memory.TeamContent) != "" {
		teamSection := common.ProcessMemoryContent(memory.TeamContent, "team", memory.TeamSourcePath)
		parts = append(parts, teamSection)
	}

	// Add project scope section if available
	if memory.HasProject && strings.TrimSpace(memory.ProjectContent) != "" {
		projectSection := common.ProcessMemoryContent(memory.ProjectContent, "project", memory.ProjectSourcePath)
		parts = append(parts, projectSection)
	}

	return strings.Join(parts, "\n\n")
}

// validateClaudeMemoryFile validates Claude memory file content
func validateClaudeMemoryFile(file types.ConfigFile) error {
	if file.Content == "" {
		return nil // Empty content is valid
	}

	// Basic validation - ensure it's not malformed
	if strings.TrimSpace(file.Content) == "" {
		return nil
	}

	// Check that content contains proper dual-scope structure
	content := file.Content
	if !strings.Contains(content, "Mindful (scope:") {
		return fmt.Errorf("invalid memory file format: missing scope headers")
	}

	return nil
}