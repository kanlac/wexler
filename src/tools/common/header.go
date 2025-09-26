package common

import (
	"fmt"
	"strings"
)

// ProcessMemoryContent processes user memory content with intelligent header handling
func ProcessMemoryContent(content, scopeName, sourcePath string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Check if content has level-1 headers
	hasH1 := hasLevelOneHeaders(content)

	if !hasH1 {
		// No level-1 headers: add our header
		return fmt.Sprintf("# Mindful (scope:%s)\n<!-- Source: %s -->\n\n%s",
			scopeName, sourcePath, content)
	} else {
		// Has level-1 headers: add suffix to each one
		return addSuffixToH1Headers(content, scopeName, sourcePath)
	}
}

// hasLevelOneHeaders checks if content contains any level-1 headers
func hasLevelOneHeaders(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && len(trimmed) > 2 {
			return true
		}
	}
	return false
}

// addSuffixToH1Headers adds Mindful suffix to all level-1 headers and includes source comment under each header
func addSuffixToH1Headers(content, scopeName, sourcePath string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)*2) // More space for source comments

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && len(trimmed) > 2 {
			// Extract the header title (remove "# ")
			headerTitle := strings.TrimSpace(trimmed[2:])
			// Add suffix to the header
			modifiedHeader := fmt.Sprintf("# %s -- Mindful (scope:%s)", headerTitle, scopeName)
			result = append(result, modifiedHeader)

			// Add source comment under this header
			result = append(result, fmt.Sprintf("<!-- Source: %s -->", sourcePath))

			// Add empty line after source comment, but only if the next line isn't already empty
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
				result = append(result, "")
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}