package cursor

import (
	"fmt"
	"strings"

	"mindful/src/models"
	"mindful/src/tools/common"
	"mindful/src/tools/types"
)

// GenerateCursorMemoryContent generates dual-scope memory content for Cursor
func GenerateCursorMemoryContent(memory *models.MemoryConfig) string {
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

	content := strings.Join(parts, "\n\n")
	if content == "" {
		return ""
	}

	// Add frontmatter
	frontmatter := `---
description: General Memories
globs:
alwaysApply: true
---

`

	return frontmatter + content
}

// generateCursorSubagentContent generates Cursor subagent file content with frontmatter
func generateCursorSubagentContent(content, description string) string {
	frontmatter := fmt.Sprintf(`---
description: %s
globs:
alwaysApply: true
---

`, description)

	return frontmatter + strings.TrimSpace(content)
}

// extractDescriptionFromContent extracts description from markdown content
// First tries to find first level-1 header (# Title), then falls back to filename
func extractDescriptionFromContent(content, fallbackName string) string {
	if strings.TrimSpace(content) == "" {
		return fallbackName
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for level-1 header
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}

	// No title found, use filename
	return fallbackName
}

// validateCursorMemoryFile validates Cursor memory file content
func validateCursorMemoryFile(file types.ConfigFile) error {
	if file.Content == "" {
		return nil // Empty content is valid
	}

	// Basic validation - ensure it has proper frontmatter
	content := file.Content
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("invalid Cursor memory file format: missing frontmatter")
	}

	// Check for proper dual-scope structure if content exists
	if strings.Contains(content, "Mindful") && !strings.Contains(content, "Mindful (scope:") {
		return fmt.Errorf("invalid memory file format: missing scope headers")
	}

	return nil
}